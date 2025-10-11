package utils

import (
	"bytes"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

var (
	positiveKeywords = []string{"content", "article", "post", "main", "entry", "text", "body"}
	negativeKeywords = []string{"nav", "footer", "header", "sidebar", "ad", "comment", "menu", "widget"}
)

func StripHTMLTags(content []byte) string {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return stripHTMLTagsSimple(content)
	}

	var result strings.Builder
	extractHTMLText(doc, &result)

	text := result.String()
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}

func extractHTMLText(n *html.Node, result *strings.Builder) {
	if n.Type == html.TextNode {
		result.WriteString(n.Data)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractHTMLText(c, result)
	}
}

func stripHTMLTagsSimple(content []byte) string {
	var result strings.Builder
	inTag := false

	for _, char := range content {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			result.WriteByte(char)
		}
	}

	return strings.TrimSpace(result.String())
}

// ExtractTitle 从HTML内容中提取title标签的文本
func ExtractTitle(content []byte) string {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return ""
	}

	title := findTitle(doc)
	if title != "" {
		return strings.TrimSpace(title)
	}
	return ""
}

// findTitle 递归查找title标签
func findTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		// 提取title标签内的文本内容
		var result strings.Builder
		extractHTMLText(n, &result)
		return result.String()
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := findTitle(c); title != "" {
			return title
		}
	}
	return ""
}

// ExtractMainContent 从HTML内容中提取主要正文内容，过滤掉页脚、导航、广告等无关内容
func ExtractMainContent(content []byte) string {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return StripHTMLTags(content)
	}

	// 首先尝试查找常见的主要内容标签
	mainContent := findMainContentByTags(doc)
	if mainContent != "" {
		return strings.TrimSpace(mainContent)
	}

	// 如果没有找到明确的主要内容标签，使用启发式方法
	mainContent = findMainContentHeuristic(doc)
	return strings.TrimSpace(mainContent)
}

// findMainContentByTags 通过常见的语义化标签查找主要内容
func findMainContentByTags(n *html.Node) string {
	// 优先级顺序：article > main > .content > #content > .post > .entry
	tagPriorities := []struct {
		tagName   string
		attrKey   string
		attrValue string
		priority  int
	}{
		{"article", "", "", 1},
		{"main", "", "", 2},
		{"div", "class", "main-content", 3},
		{"div", "class", "article-content", 3},
		{"div", "class", "content", 3},
		{"div", "id", "content", 4},
		{"div", "id", "main", 4},
		{"div", "class", "post", 5},
		{"div", "class", "entry", 5},
		{"div", "class", "post-content", 5},
		{"section", "class", "content", 6},
	}

	var bestMatch *html.Node
	bestPriority := 999

	findBestMatch(n, tagPriorities, &bestMatch, &bestPriority)

	if bestMatch != nil {
		var result strings.Builder
		extractMainText(bestMatch, &result)
		return result.String()
	}

	return ""
}

// findBestMatch 递归查找最佳匹配的内容标签
func findBestMatch(n *html.Node, priorities []struct {
	tagName   string
	attrKey   string
	attrValue string
	priority  int
}, bestMatch **html.Node, bestPriority *int) {
	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			for _, neg := range negativeKeywords {
				if strings.Contains(strings.ToLower(attr.Val), neg) {
					goto nextNode
				}
			}
		}
		for _, p := range priorities {
			if n.Data == p.tagName && p.priority <= *bestPriority {
				if p.attrKey == "" {
					// 直接匹配标签名
					*bestMatch = n
					*bestPriority = p.priority
				} else {
					// 匹配属性
					for _, attr := range n.Attr {
						if attr.Key == p.attrKey && strings.Contains(strings.ToLower(attr.Val), p.attrValue) {
							*bestMatch = n
							*bestPriority = p.priority
							break
						}
					}
				}
			}
		}
	}

nextNode:
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findBestMatch(c, priorities, bestMatch, bestPriority)
	}
}

// findMainContentHeuristic 使用启发式方法查找主要内容
func findMainContentHeuristic(n *html.Node) string {
	// 收集所有可能的内容块
	contentBlocks := make([]contentBlock, 0)
	collectContentBlocks(n, &contentBlocks, 0)

	// 根据启发式规则评分
	var bestBlock contentBlock
	bestScore := 0

	for _, block := range contentBlocks {
		score := calculateContentScore(block)
		if score > bestScore {
			bestScore = score
			bestBlock = block
		}
	}

	if bestBlock.text != "" {
		return bestBlock.text
	}

	// 如果没有找到合适的块，返回过滤后的全部内容
	var result strings.Builder
	extractFilteredText(n, &result)
	return result.String()
}

// contentBlock 内容块结构
type contentBlock struct {
	text        string
	textLength  int
	tagName     string
	className   string
	id          string
	depth       int
	pCount      int     // 段落数量
	linkDensity float64 // 链接密度
}

// collectContentBlocks 收集内容块
func collectContentBlocks(n *html.Node, blocks *[]contentBlock, depth int) {
	if n.Type == html.ElementNode {
		// 跳过明显的非内容标签
		if isNonContentTag(n.Data) {
			return
		}

		// 如果是可能包含内容的标签，提取文本
		if isContentTag(n.Data) {
			var textBuilder strings.Builder
			extractMainText(n, &textBuilder)
			text := strings.TrimSpace(textBuilder.String())

			if len(text) > 50 { // 只考虑有足够文本的块
				block := contentBlock{
					text:        text,
					textLength:  len(text),
					tagName:     n.Data,
					depth:       depth,
					pCount:      countParagraphs(n),
					linkDensity: calculateLinkDensity(n),
				}

				// 获取class和id属性
				for _, attr := range n.Attr {
					switch attr.Key {
					case "class":
						block.className = attr.Val
					case "id":
						block.id = attr.Val
					}
				}

				*blocks = append(*blocks, block)
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectContentBlocks(c, blocks, depth+1)
	}
}

// calculateContentScore 计算内容块得分
func calculateContentScore(block contentBlock) int {
	score := 0

	// 基础分数：文本长度
	score += block.textLength / 10

	// 段落数量加分
	score += block.pCount * 20

	// 标签类型加分
	switch block.tagName {
	case "article":
		score += 100
	case "main":
		score += 80
	case "div":
		score += 10
	case "section":
		score += 30
	}

	// class和id加分
	classLower := strings.ToLower(block.className)
	idLower := strings.ToLower(block.id)

	for _, keyword := range positiveKeywords {
		if strings.Contains(classLower, keyword) || strings.Contains(idLower, keyword) {
			score += 50
		}
	}

	for _, keyword := range negativeKeywords {
		if strings.Contains(classLower, keyword) || strings.Contains(idLower, keyword) {
			score -= 100
		}
	}

	// 链接密度惩罚（链接过多的内容可能是导航或广告）
	if block.linkDensity > 0.3 {
		score -= int(block.linkDensity * 100)
	}

	// 深度惩罚（过深的嵌套可能不是主要内容）
	if block.depth > 5 {
		score -= (block.depth - 5) * 10
	}

	return score
}

// isNonContentTag 判断是否为非内容标签
func isNonContentTag(tagName string) bool {
	nonContentTags := []string{
		"script", "style", "nav", "header", "footer", "aside",
		"menu", "form", "input", "button", "select", "textarea",
		"iframe", "embed", "object", "applet",
	}

	return slices.Contains(nonContentTags, tagName)
}

// isContentTag 判断是否为可能包含内容的标签
func isContentTag(tagName string) bool {
	contentTags := []string{
		"article", "main", "section", "div", "p", "span",
		"h1", "h2", "h3", "h4", "h5", "h6",
	}

	return slices.Contains(contentTags, tagName)
}

// extractMainText 提取主要文本内容，跳过非内容元素
func extractMainText(n *html.Node, result *strings.Builder) {
	switch n.Type {
	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text != "" {
			result.WriteString(text)
			result.WriteString(" ")
		}
	case html.ElementNode:
		// 跳过非内容标签
		if isNonContentTag(n.Data) {
			return
		}

		// 跳过明显的非内容class
		for _, attr := range n.Attr {
			if attr.Key == "class" || attr.Key == "id" {
				attrLower := strings.ToLower(attr.Val)
				skipKeywords := []string{"nav", "footer", "header", "sidebar", "ad", "advertisement", "comment", "menu", "widget", "social", "share"}
				for _, keyword := range skipKeywords {
					if strings.Contains(attrLower, keyword) {
						return
					}
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractMainText(c, result)
	}
}

// extractFilteredText 提取过滤后的文本
func extractFilteredText(n *html.Node, result *strings.Builder) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			result.WriteString(text)
			result.WriteString(" ")
		}
	} else if n.Type == html.ElementNode && !isNonContentTag(n.Data) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractFilteredText(c, result)
		}
	}
}

// countParagraphs 计算段落数量
func countParagraphs(n *html.Node) int {
	count := 0
	if n.Type == html.ElementNode && n.Data == "p" {
		count++
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		count += countParagraphs(c)
	}

	return count
}

// calculateLinkDensity 计算链接密度
func calculateLinkDensity(n *html.Node) float64 {
	totalText := 0
	linkText := 0

	calculateLinkDensityRecursive(n, &totalText, &linkText, false)

	if totalText == 0 {
		return 0
	}

	return float64(linkText) / float64(totalText)
}

// calculateLinkDensityRecursive 递归计算链接密度
func calculateLinkDensityRecursive(n *html.Node, totalText, linkText *int, inLink bool) {
	switch n.Type {
	case html.TextNode:
		textLen := len(strings.TrimSpace(n.Data))
		*totalText += textLen
		if inLink {
			*linkText += textLen
		}
	case html.ElementNode:
		isLink := n.Data == "a"
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			calculateLinkDensityRecursive(c, totalText, linkText, inLink || isLink)
		}
	}
}
