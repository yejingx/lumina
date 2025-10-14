import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Card, Layout, List, Input, Button, Typography, message as antdMessage } from 'antd';
import { EditOutlined, SendOutlined } from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import { conversationApi, type ConversationSpec, type ListConversationsResponse, type ChatMessageSpec, type ListChatMessagesResponse } from '../../services/api';
import type { ListParams } from '../../types';

const { Sider, Content } = Layout;
const { Text } = Typography;

// 类型从 services/api 导入

// SSE 消息结构，与后端 AgentThought 对齐
type ThoughtPhase = 'thought' | 'tool' | 'observation';
type ToolCall = { id: string; tool_name: string; args: string };
type AgentThought = {
  phase?: ThoughtPhase;
  id?: string;
  thought?: string;
  observation?: string;
  toolCall?: ToolCall | null;
};

const DEFAULT_PAGE_SIZE = 50;

const AgentPage: React.FC = () => {
  const [conversations, setConversations] = useState<ConversationSpec[]>([]);
  const [selected, setSelected] = useState<ConversationSpec | null>(null);
  const [messages, setMessages] = useState<ChatMessageSpec[]>([]);
  const [loadingConvs, setLoadingConvs] = useState(false);
  const [loadingMsgs, setLoadingMsgs] = useState(false);
  const [input, setInput] = useState('');
  const streamAbortRef = useRef<AbortController | null>(null);
  const contentRef = useRef<HTMLDivElement>(null);

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
      setMessages(data.items || []);
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

  const handleSend = async () => {
    if (!selected || !input.trim()) return;
    // 先追加本地消息（用户提问）
    const localMsg: ChatMessageSpec = {
      id: Date.now(),
      conversationId: selected.id,
      query: input,
      answer: '',
      agentThoughts: [],
      createTime: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, localMsg]);
    setInput('');

    // 取消已有流
    streamAbortRef.current?.abort();
    const ac = new AbortController();
    streamAbortRef.current = ac;

    try {
      const resp = await conversationApi.chatStream(selected.uuid, { query: localMsg.query }, ac.signal);

      if (!resp.ok) {
        throw new Error('网络错误: ' + resp.status);
      }

      const reader = resp.body?.getReader();
      const decoder = new TextDecoder('utf-8');
      let assistantMsgIndex = -1;

      // 解析 SSE：形如 "data: {...}\n"，也可能包含多行
      let buffer = '';
      const appendAssistant = (delta: string) => {
        setMessages((prev) => {
          const next = [...prev];
          if (assistantMsgIndex === -1) {
            assistantMsgIndex = next.length;
            next.push({
              id: Date.now() + 1,
              conversationId: selected.id,
              query: '',
              answer: delta,
              agentThoughts: [],
              createTime: new Date().toISOString(),
            });
          } else {
            const m = next[assistantMsgIndex];
            m.answer = (m.answer || '') + delta;
          }
          return next;
        });
      };

      const appendToolInfo = (tool: ToolCall | null, args: string, observation?: string) => {
        const toolText = tool ? `\n\n🔧 调用工具: ${tool.tool_name}, 参数: ${tool.args}\n\n` : '';
        const obsText = observation ? observation : '';
        appendAssistant(toolText + obsText);
      };

      while (reader) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        // 保留最后一行（可能未完整）
        buffer = lines.pop() || '';
        for (const line of lines) {
          const trimmed = line.trim();
          if (!trimmed.startsWith('data:')) continue;
          const payload = trimmed.substring(5).trim();
          if (payload === '[DONE]') {
            buffer = '';
            break;
          }
          try {
            const msg: AgentThought = JSON.parse(payload);
            switch (msg.phase) {
              case 'thought':
                if (msg.thought) appendAssistant(msg.thought);
                break;
              case 'tool':
                appendToolInfo(msg.toolCall || null, msg.toolCall?.args || '');
                break;
              case 'observation':
                appendToolInfo(msg.toolCall || null, msg.toolCall?.args || '', msg.observation || '');
                break;
              default:
                // 非结构化内容，直接追加
                appendAssistant(payload);
            }
          } catch (e) {
            // 不是 JSON，直接展示原始内容
            appendAssistant(payload);
          }
        }
        // 滚动到底部
        contentRef.current?.scrollTo({ top: contentRef.current.scrollHeight });
      }
    } catch (err) {
      antdMessage.error('发送或接收失败');
    }
  };

  const handleCreateConversation = async () => {
    try {
      const resp = await conversationApi.create({ title: '新对话' });
      const conv = await conversationApi.get(resp.uuid);
      setConversations((prev) => [conv, ...prev]);
      setSelected(conv);
      setMessages([]);
    } catch (err) {
      antdMessage.error('创建对话失败');
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
            itemLayout="horizontal"
            dataSource={conversations}
            renderItem={(item) => (
              <List.Item
                style={{ cursor: 'pointer', background: selected?.id === item.id ? '#f5f5f5' : undefined }}
                onClick={() => setSelected(item)}
              >
                <List.Item.Meta
                  title={<Text strong>{item.title || item.uuid}</Text>}
                  description={<Text type="secondary">{new Date(item.createTime).toLocaleString()}</Text>}
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
            >
              <div ref={contentRef} style={{ flex: 1, overflowY: 'auto' }}>
                {messages.map((m) => (
                  <div key={m.id} style={{ marginBottom: 16 }}>
                    {m.query && (
                      <Card size="small" style={{ marginBottom: 8, background: '#f0f9ff' }}>
                        <ReactMarkdown>{m.query}</ReactMarkdown>
                      </Card>
                    )}
                    {m.answer && (
                      <Card size="small" style={{ background: '#fafafa' }}>
                        <ReactMarkdown>{m.answer}</ReactMarkdown>
                      </Card>
                    )}
                  </div>
                ))}
              </div>
              <div style={{ paddingTop: 8, borderTop: '1px solid #f0f0f0', position: 'relative' }}>
                <Input.TextArea
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  autoSize={{ minRows: 2, maxRows: 6 }}
                  placeholder="请输入问题，Shift+Enter 换行，Enter 发送"
                  style={{ paddingRight: 90 }}
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
                  style={{ position: 'absolute', right: 8, bottom: 8 }}
                >
                  发送
                </Button>
              </div>
            </Card>
          </Content>
        </Layout>
      </Content>
    </Layout>
  );
};

export default AgentPage;