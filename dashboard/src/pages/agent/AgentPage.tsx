import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Card, Layout, List, Input, Button, Typography, message as antdMessage, Avatar, Popconfirm, Collapse } from 'antd';
import { EditOutlined, SendOutlined, UserOutlined, RobotOutlined, DeleteOutlined } from '@ant-design/icons';
import ReactMarkdown, { type Components } from 'react-markdown';
import './AgentPage.css';
import { conversationApi, type ConversationSpec, type ListConversationsResponse, type ChatMessageSpec, type ListChatMessagesResponse } from '../../services/api';
import type { ListParams } from '../../types';

const { Sider, Content } = Layout;
const { Text } = Typography;

// 类型从 services/api 导入

// SSE 消息结构，与后端 AgentThought 对齐
type ThoughtPhase = 'thought' | 'tool' | 'observation';
type ToolCall = { id: string; tool_name: string; args: string };
type AgentThought = {
  id: string;
  phase?: ThoughtPhase;
  thought?: string;
  observation?: string;
  toolCall?: ToolCall | null;
};

// 前端内联渲染片段：文本或工具卡片
// 移除自定义 ChatFragment/ChatMessageEx，直接使用 ChatMessageSpec

const DEFAULT_PAGE_SIZE = 50;

const AgentPage: React.FC = () => {
  const [conversations, setConversations] = useState<ConversationSpec[]>([]);
  const [selected, setSelected] = useState<ConversationSpec | null>(null);
  const [messages, setMessages] = useState<ChatMessageSpec[]>([]);
  const messagesRef = useRef<ChatMessageSpec[]>([]);
  const [loadingConvs, setLoadingConvs] = useState(false);
  const [loadingMsgs, setLoadingMsgs] = useState(false);
  const [input, setInput] = useState('');
  const streamAbortRef = useRef<AbortController | null>(null);
  const contentRef = useRef<HTMLDivElement>(null);

  // 始终保持 ref 与状态同步，便于直接赋值更新而不使用函数式 set
  useEffect(() => {
    messagesRef.current = messages;
  }, [messages]);

  const convParams: ListParams = useMemo(() => ({ start: 0, limit: DEFAULT_PAGE_SIZE }), []);

  const fetchConversations = async () => {
    setLoadingConvs(true);
    try {
      const data: ListConversationsResponse = await conversationApi.list(convParams);
      setConversations(data.items || []);
      if (!selected && data.items?.length) {
        setSelected(data.items[0]);
      }
    } catch (err) {
      antdMessage.error('获取对话列表失败');
    } finally {
      setLoadingConvs(false);
    }
  };

  const fetchMessages = async (conv: ConversationSpec | null) => {
    if (!conv) return;
    setLoadingMsgs(true);
    try {
      const data: ListChatMessagesResponse = await conversationApi.listMessages(conv.uuid, { start: 0, limit: DEFAULT_PAGE_SIZE });
      const items = (data.items || []);
      setMessages(items);
      // 滚动到底部
      setTimeout(() => {
        contentRef.current?.scrollTo({ top: contentRef.current.scrollHeight, behavior: 'smooth' });
      }, 50);
    } catch (err) {
      antdMessage.error('获取历史消息失败');
    } finally {
      setLoadingMsgs(false);
    }
  };

  useEffect(() => {
    fetchConversations();
  }, []);

  useEffect(() => {
    fetchMessages(selected);
  }, [selected?.id]);

  // 消息数量变化时自动滚动到底部，确保新消息可见
  useEffect(() => {
    contentRef.current?.scrollTo({ top: contentRef.current.scrollHeight, behavior: 'smooth' });
  }, [messages.length]);

  const handleSend = async () => {
    if (!selected || !input.trim()) return;
    // 先追加本地消息（用户提问）
    const localMsg: ChatMessageSpec = {
      id: Date.now(),
      conversationId: selected.id,
      query: input,
      answer: '',
      createTime: new Date().toISOString(),
    };
    {
      const next = [...messagesRef.current, localMsg];
      messagesRef.current = next;
      setMessages(next);
    }
    setInput('');
    // 发送后立即滚动到最新消息位置
    setTimeout(() => {
      contentRef.current?.scrollTo({ top: contentRef.current.scrollHeight, behavior: 'smooth' });
    }, 50);

    // 取消已有流
    streamAbortRef.current?.abort();
    const ac = new AbortController();
    streamAbortRef.current = ac;

    try {
      const resp = await conversationApi.chatStream(selected.uuid, { query: localMsg.query });

      if (!resp.ok) {
        throw new Error('网络错误: ' + resp.status);
      }

      const reader = resp.body?.getReader();
      const decoder = new TextDecoder('utf-8');
      let assistantMsgIndex = -1;

      // 解析 SSE：逐行处理 "data: {...}\n"；事件不会有多行
      const appendAssistant = (id: string, delta: string) => {
        const next = messagesRef.current.slice();
        if (assistantMsgIndex === -1) {
          assistantMsgIndex = next.length;
          next.push({
            id: Date.now() + 1,
            conversationId: selected.id,
            query: '',
            answer: delta,
            agentThoughts: delta ? [{ id: id, phase: 'thought', thought: delta }] : [],
            createTime: new Date().toISOString(),
          });
        } else {
          const m = next[assistantMsgIndex];
          const currentAnswer = m.answer || '';
          // 若收到的是完整内容（如工具阶段回填完整文本），避免重复追加
          if (delta.startsWith(currentAnswer)) {
            m.answer = delta;
          } else {
            m.answer = currentAnswer + delta;
          }
          m.agentThoughts = m.agentThoughts || [];
          const last = m.agentThoughts[m.agentThoughts.length - 1];
          if (last && last.phase === 'thought' && last.id === (id || last.id)) {
            const currThought = last.thought || '';
            last.thought = delta.startsWith(currThought) ? delta : currThought + delta;
          } else {
            m.agentThoughts.push({ id: id || 'unknown', phase: 'thought', thought: delta });
          }
        }
        messagesRef.current = next;
        setMessages(next);
      };

      const appendToolInfo = (id: string, tool: ToolCall | null, args: string, observation?: string) => {
        const next = messagesRef.current.slice();
        if (assistantMsgIndex === -1) {
          assistantMsgIndex = next.length;
          next.push({
            id: Date.now() + 1,
            conversationId: selected.id,
            query: '',
            answer: '',
            agentThoughts: [],
            createTime: new Date().toISOString(),
          });
        }
        const m = next[assistantMsgIndex];
        m.agentThoughts = m.agentThoughts || [];
        if (tool) {
          const toolName = tool.tool_name || '未知工具';
          const toolArgs = tool.args || args || '';
          const fragId = tool.id || id || 'unknown';
          m.agentThoughts.push({ id: fragId, phase: 'tool', toolCall: { name: toolName, args: toolArgs } });
        }
        // 观察阶段：仅追加为 observation 片段，不并入最终 answer
        if (observation) {
          // 将观察结果绑定到当前工具调用 id（若无则使用传入 id），确保后续渲染按 id 分组
          const obsId = id || (m.agentThoughts.length ? m.agentThoughts[m.agentThoughts.length - 1].id : 'unknown');
          m.agentThoughts.push({ id: obsId, phase: 'observation', observation });
        }
        messagesRef.current = next;
        setMessages(next);
      };

      let buffer = '';
      let stopStreaming = false;
      let newlineIndex: number;
      while (reader && !stopStreaming) {
        const { done, value } = await reader.read();
        if (done) {
          // 流自然结束（未显式发送 [DONE]），也视为完成
          stopStreaming = true;
          break;
        }
        buffer += decoder.decode(value, { stream: true });
        // 处理所有完整行；保留未完成的最后一行
        while ((newlineIndex = buffer.indexOf('\n')) !== -1) {
          const line = buffer.slice(0, newlineIndex).trim();
          buffer = buffer.slice(newlineIndex + 1);
          if (!line.startsWith('data:')) continue;
          const payload = line.slice(5).trim();
          if (!payload) continue;
          if (payload === '[DONE]') {
            stopStreaming = true;
            buffer = '';
            break;
          }
          try {
            const msg: AgentThought = JSON.parse(payload);
            switch (msg.phase) {
              case 'thought':
                appendAssistant(msg.id, msg.thought || '');
                break;
              case 'tool':
                appendToolInfo(msg.toolCall?.id || msg.id, msg.toolCall || null, msg.toolCall?.args || '', undefined);
                break;
              case 'observation':
                if (msg.observation) {
                  appendToolInfo(msg.toolCall?.id || msg.id, null, '', msg.observation || '');
                }
                break;
            }
          } catch (e) {
            console.error('解析 JSON 失败:', e);
          }
        }
        // 流式接收过程中，持续滚动到底部
        contentRef.current?.scrollTo({ top: contentRef.current.scrollHeight, behavior: 'smooth' });
      }

      // 流结束后（收到 [DONE]），若当前对话标题为空则生成并更新标题
      if (stopStreaming && selected && (!selected.title || !selected.title.trim())) {
        try {
          const { title } = await conversationApi.genTitle(selected.uuid);
          if (title && title.trim()) {
            setConversations((prev) => prev.map((c) => (
              c.uuid === selected.uuid ? { ...c, title } : c
            )));
            setSelected((prev) => (prev && prev.uuid === selected.uuid ? { ...prev, title } : prev));
          }
        } catch (e) {
          console.error('生成对话标题失败:', e);
        }
      }
    } catch (err) {
      antdMessage.error('发送或接收失败');
    }
  };

  const handleCreateConversation = async () => {
    try {
      const resp = await conversationApi.create({});
      const conv = await conversationApi.get(resp.uuid);
      setConversations((prev) => [conv, ...prev]);
      setSelected(conv);
      setMessages([]);
    } catch (err) {
      antdMessage.error('创建对话失败');
    }
  };

  const handleDeleteConversation = async (conv: ConversationSpec) => {
    if (!conv?.uuid) return;
    try {
      await conversationApi.delete(conv.uuid);
      const filtered = conversations.filter((c) => c.uuid !== conv.uuid);
      setConversations(filtered);
      if (selected?.uuid === conv.uuid) {
        // 如删除的是当前对话，重置选择与消息，并中止任何进行中的流
        streamAbortRef.current?.abort();
        const nextSelected = filtered.length ? filtered[0] : null;
        setSelected(nextSelected);
        setMessages([]);
      }
      antdMessage.success('删除成功');
    } catch (err) {
      antdMessage.error('删除对话失败');
    }
  };

  return (
    <Layout style={{ height: '100%', background: 'transparent' }}>
      <Sider width={280} theme="light" style={{ borderRight: '1px solid #f0f0f0' }}>
        <Card bordered={false} title="对话列表" loading={loadingConvs} style={{ height: '100%' }}>
          <div style={{ marginBottom: 12 }}>
            <Button block icon={<EditOutlined />} onClick={handleCreateConversation}>
              开启新对话
            </Button>
          </div>
          <List
            split={false}
            itemLayout="horizontal"
            dataSource={conversations}
            renderItem={(item) => (
              <List.Item
                className="agent-conv-item"
                style={{
                  cursor: 'pointer',
                  background: selected?.id === item.id ? '#f5f5f5' : undefined,
                  borderRadius: selected?.id === item.id ? 12 : undefined,
                  padding: '6px 8px',
                }}
                onClick={() => setSelected(item)}
                actions={[
                  (
                    <Popconfirm
                      title="确认删除该对话？"
                      description="删除后无法恢复，是否继续？"
                      onConfirm={() => handleDeleteConversation(item)}
                    >
                      <DeleteOutlined
                        style={{ color: '#ff4d4f', fontSize: 16}}
                        onClick={(e) => e.stopPropagation()}
                      />
                    </Popconfirm>
                  ),
                ]}
              >
                <List.Item.Meta
                  title={
                    <Text
                      strong
                      ellipsis={{ tooltip: true }}
                      style={{
                        fontSize: 16,
                        display: 'block',
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                      }}
                    >
                      {item.title || item.uuid}
                    </Text>
                  }
                />
              </List.Item>
            )}
          />
        </Card>
      </Sider>
      <Content>
        <Layout style={{ height: '100%' }}>
          <Content style={{ padding: 16, height: 'calc(100vh - 140px)' }}>
            <Card
              bordered={false}
              title="消息"
              loading={loadingMsgs}
              style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
              headStyle={{ padding: '8px 16px' }}
              bodyStyle={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden', minHeight: 0 }}
            >
              <div ref={contentRef} style={{ flex: 1, overflowY: 'auto', paddingRight: 24, minHeight: 0 }}>
                {messages.map((m) => (
                  <div key={m.id} style={{ marginBottom: 12 }}>
                    {m.query && (
                      <div
                        style={{
                          display: 'flex',
                          justifyContent: 'flex-end',
                          alignItems: 'flex-end',
                          gap: 8,
                          marginBottom: 8,
                        }}
                      >
                        <div
                          style={{
                            maxWidth: '72%',
                            background: '#e6f4ff',
                            border: '1px solid #91caff',
                            borderRadius: 12,
                            padding: '5px 12px',
                            lineHeight: 1.4,
                            wordBreak: 'break-word',
                            whiteSpace: 'pre-wrap',
                            boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
                          }}
                        >
                          <ReactMarkdown components={markdownComponents}>{m.query}</ReactMarkdown>
                        </div>
                        <Avatar size={32} icon={<UserOutlined />} />
                      </div>
                    )}
                    {(m.answer || (m.agentThoughts && m.agentThoughts.length > 0)) && (
                      <div
                        style={{
                          display: 'flex',
                          justifyContent: 'flex-start',
                          alignItems: 'flex-end',
                          gap: 8,
                        }}
                      >
                        <Avatar size={32} icon={<RobotOutlined />} />
                        <div
                          style={{
                            maxWidth: '72%',
                            background: '#fafafa',
                            border: '1px solid #d9d9d9',
                            borderRadius: 12,
                            padding: '5px 12px',
                            lineHeight: 1.4,
                            wordBreak: 'break-word',
                            whiteSpace: 'pre-wrap',
                            boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
                          }}
                        >
                          {m.agentThoughts && m.agentThoughts.length > 0 ? (
                            (() => {
                              const elements: React.ReactNode[] = [];
                              for (let i = 0; i < m.agentThoughts.length; i++) {
                                const t = m.agentThoughts[i];
                                if (t.toolCall) {
                                  const results: string[] = [];
                                  // 兼容历史数据：当前条目可能同时包含 toolCall 和 observation
                                  if (t.observation) {
                                    results.push(t.observation);
                                  }
                                  // 继续收集后续连续的 observation 作为结果
                                  let j = i + 1;
                                  while (
                                    j < m.agentThoughts.length &&
                                    m.agentThoughts[j].phase === 'observation' && m.agentThoughts[j].id === t.id
                                  ) {
                                    const tf = m.agentThoughts[j];
                                    if (tf.observation) {
                                      results.push(tf.observation);
                                    }
                                    j++;
                                  }
                                  const resultText = results.join('');
                                  // 若此前没有显式 thought 片段，但当前 tool 事件携带 thought，则先渲染思考内容
                                  const hasPrevThought = m.agentThoughts.slice(0, i).some((x) => x.phase === 'thought' && !!x.thought);
                                  if (t.thought && !hasPrevThought) {
                                    elements.push(
                                      <ReactMarkdown key={`frag-prethought-${m.id}-${i}`} components={markdownComponents}>
                                        {t.thought}
                                      </ReactMarkdown>
                                    );
                                  }
                                  elements.push(
                                    <Card
                                      bordered={false}
                                      key={`frag-tool-${m.id}-${i}`}
                                      size="small"
                                      style={{
                                        margin: '8px 0',
                                        borderRadius: 12,
                                        boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
                                      }}
                                      headStyle={{ borderRadius: 12 }}
                                      bodyStyle={{ borderRadius: 12 }}
                                      title={`🔧 工具: ${t.toolCall?.name ?? (t as any).toolCall?.tool_name ?? '未知工具'}`}
                                    >
                                      <Collapse bordered={false} ghost>
                                        <Collapse.Panel header="参数（点击展开查看 JSON）" key={`params-${m.id}-${i}`}>
                                          <pre
                                            style={{
                                              margin: 0,
                                              whiteSpace: 'pre-wrap',
                                              wordBreak: 'break-word',
                                              lineHeight: 1.5,
                                            }}
                                          >
                                            {(() => {
                                              const s = t.toolCall?.args || '';
                                              try {
                                                const obj = JSON.parse(s);
                                                return JSON.stringify(obj, null, 2);
                                              } catch {
                                                return s;
                                              }
                                            })()}
                                          </pre>
                                        </Collapse.Panel>
                                        {resultText ? (
                                          <Collapse.Panel header="结果（点击展开查看）" key={`result-${m.id}-${i}`}>
                                            <pre
                                              style={{
                                                margin: 0,
                                                whiteSpace: 'pre-wrap',
                                                wordBreak: 'break-word',
                                                lineHeight: 1.5,
                                              }}
                                            >
                                              {(() => {
                                                const s = resultText;
                                                try {
                                                  const obj = JSON.parse(s);
                                                  return JSON.stringify(obj, null, 2);
                                                } catch {
                                                  return s;
                                                }
                                              })()}
                                            </pre>
                                          </Collapse.Panel>
                                        ) : null}
                                      </Collapse>
                                    </Card>
                                  );
                                  // 跳过已作为结果展示的文本片段
                                  i = j - 1;
                                } else if (t.phase === 'thought' && t.thought) {
                                  elements.push(
                                    <ReactMarkdown key={`frag-thought-${m.id}-${i}`} components={markdownComponents}>
                                      {t.thought}
                                    </ReactMarkdown>
                                  );
                                } else if (t.phase === 'observation' && t.observation) {
                                  // 孤立的 observation（没有前置 tool），按纯文本显示
                                  elements.push(
                                    <ReactMarkdown key={`frag-obs-${m.id}-${i}`} components={markdownComponents}>
                                      {t.observation}
                                    </ReactMarkdown>
                                  );
                                }
                              }
                              return elements;
                            })()
                          ) : (
                            <ReactMarkdown components={markdownComponents}>{m.answer}</ReactMarkdown>
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>
              <div style={{ padding: '8px 24px', borderTop: '1px solid #f0f0f0' }}>
                <div style={{ display: 'flex', alignItems: 'flex-end', gap: 8, border: '1px solid #d9d9d9', borderRadius: 12, padding: 8, background: '#fff' }}>
                  <Input.TextArea
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    autoSize={{ minRows: 2, maxRows: 6 }}
                    placeholder="请输入问题，Shift+Enter 换行，Enter 发送"
                    style={{ flex: 1, border: 'none', boxShadow: 'none' }}
                    onPressEnter={(e) => {
                      if (!e.shiftKey) {
                        e.preventDefault();
                        handleSend();
                      }
                    }}
                  />
                  <Button
                    type="primary"
                    icon={<SendOutlined />}
                    onClick={handleSend}
                    disabled={!selected}
                  >
                    发送
                  </Button>
                </div>
              </div>
            </Card>
          </Content>
        </Layout>
      </Content>
    </Layout>
  );
};

export default AgentPage;
// 紧凑的 Markdown 渲染，压缩段落与列表间距
const markdownComponents: Components = {
  p: ({ children, ...props }) => (
    <p {...props} style={{ margin: '4px 0' }}>{children}</p>
  ),
  ul: ({ children, ...props }) => (
    <ul {...props} style={{ margin: '4px 0', paddingLeft: 18 }}>{children}</ul>
  ),
  ol: ({ children, ...props }) => (
    <ol {...props} style={{ margin: '4px 0', paddingLeft: 18 }}>{children}</ol>
  ),
  li: ({ children, ...props }) => (
    <li {...props} style={{ margin: '2px 0' }}>{children}</li>
  ),
  h1: ({ children, ...props }) => (
    <h1 {...props} style={{ margin: '8px 0 4px' }}>{children}</h1>
  ),
  h2: ({ children, ...props }) => (
    <h2 {...props} style={{ margin: '8px 0 4px' }}>{children}</h2>
  ),
  h3: ({ children, ...props }) => (
    <h3 {...props} style={{ margin: '8px 0 4px' }}>{children}</h3>
  ),
  h4: ({ children, ...props }) => (
    <h4 {...props} style={{ margin: '8px 0 4px' }}>{children}</h4>
  ),
  h5: ({ children, ...props }) => (
    <h5 {...props} style={{ margin: '8px 0 4px' }}>{children}</h5>
  ),
  h6: ({ children, ...props }) => (
    <h6 {...props} style={{ margin: '8px 0 4px' }}>{children}</h6>
  ),
};