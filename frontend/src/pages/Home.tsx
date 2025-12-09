import { useState, useEffect } from 'react';
import { Recipient, SendMessageResponse } from '../types';
import { getRecipients, sendMessage } from '../services/api';

export function Home() {
  const [recipients, setRecipients] = useState<Recipient[]>([]);
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [result, setResult] = useState<SendMessageResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadRecipients();
  }, []);

  const loadRecipients = async () => {
    try {
      setLoading(true);
      const data = await getRecipients();
      setRecipients(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载接收者失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSelectAll = () => {
    setSelectedIds(new Set(recipients.map((r) => r.id)));
  };

  const handleDeselectAll = () => {
    setSelectedIds(new Set());
  };

  const handleToggleRecipient = (id: number) => {
    const newSelected = new Set(selectedIds);
    if (newSelected.has(id)) {
      newSelected.delete(id);
    } else {
      newSelected.add(id);
    }
    setSelectedIds(newSelected);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setResult(null);

    // Validation
    if (!title.trim()) {
      setError('请输入消息标题');
      return;
    }
    if (!content.trim()) {
      setError('请输入消息内容');
      return;
    }
    if (selectedIds.size === 0) {
      setError('请选择至少一个接收者');
      return;
    }

    try {
      setSending(true);
      const response = await sendMessage({
        title: title.trim(),
        content: content.trim(),
        recipientIds: Array.from(selectedIds),
      });
      setResult(response);
      
      // Clear form on success
      if (response.totalFailed === 0) {
        setTitle('');
        setContent('');
        setSelectedIds(new Set());
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '发送消息失败');
    } finally {
      setSending(false);
    }
  };

  if (loading) {
    return <div className="loading">加载中...</div>;
  }

  return (
    <div>
      <h1 className="page-title">发送消息</h1>
      
      <form onSubmit={handleSubmit} className="card">
        <div className="form-group">
          <label className="form-label">消息标题</label>
          <input
            type="text"
            className="form-input"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="请输入消息标题"
          />
        </div>

        <div className="form-group">
          <label className="form-label">消息内容</label>
          <textarea
            className="form-textarea"
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="请输入消息内容"
          />
        </div>

        <div className="recipients-section">
          <div className="recipients-header">
            <label className="form-label">选择接收者</label>
            <div className="recipients-actions">
              <button
                type="button"
                className="btn btn-secondary btn-small"
                onClick={handleSelectAll}
              >
                全选
              </button>
              <button
                type="button"
                className="btn btn-secondary btn-small"
                onClick={handleDeselectAll}
              >
                取消全选
              </button>
            </div>
          </div>

          {recipients.length === 0 ? (
            <div className="empty-state">
              暂无接收者，请先在设置页面添加
            </div>
          ) : (
            <div className="recipients-list">
              {recipients.map((recipient) => (
                <label key={recipient.id} className="recipient-checkbox">
                  <input
                    type="checkbox"
                    checked={selectedIds.has(recipient.id)}
                    onChange={() => handleToggleRecipient(recipient.id)}
                  />
                  <span>
                    <span className="recipient-name">{recipient.name}</span>
                    <span className="recipient-openid"> ({recipient.openId})</span>
                  </span>
                </label>
              ))}
            </div>
          )}
        </div>

        {error && <div className="error-message">{error}</div>}

        <button
          type="submit"
          className="btn btn-primary"
          disabled={sending}
        >
          {sending ? '发送中...' : '发送消息'}
        </button>
      </form>

      {result && (
        <div className={`send-result ${result.totalFailed === 0 ? 'success' : 'error'}`}>
          <div className="result-summary">
            发送完成: 成功 {result.totalSent} 条，失败 {result.totalFailed} 条
          </div>
          <div className="result-details">
            {result.results.map((r, index) => (
              <div key={index} className={`result-item ${r.success ? '' : 'failed'}`}>
                {r.recipientName}: {r.success ? '成功' : `失败 - ${r.error}`}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
