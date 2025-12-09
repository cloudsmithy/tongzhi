import { useState, useEffect } from 'react';
import { Recipient, SendMessageResponse, MessageTemplate } from '../types';
import { getRecipients, getTemplates, sendMessage } from '../services/api';

export function Home() {
  const [recipients, setRecipients] = useState<Recipient[]>([]);
  const [templates, setTemplates] = useState<MessageTemplate[]>([]);
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [selectedTemplate, setSelectedTemplate] = useState('');
  const [keywords, setKeywords] = useState<{ key: string; value: string }[]>([
    { key: 'first', value: '' },
  ]);
  const [customMode, setCustomMode] = useState(false);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [result, setResult] = useState<SendMessageResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [recipientsData, templatesData] = await Promise.all([
        getRecipients(),
        getTemplates(),
      ]);
      setRecipients(recipientsData);
      setTemplates(templatesData);
      if (templatesData.length > 0) {
        setSelectedTemplate(templatesData[0].key);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载数据失败');
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

  const handleKeywordChange = (index: number, field: 'key' | 'value', val: string) => {
    const newKeywords = [...keywords];
    newKeywords[index][field] = val;
    setKeywords(newKeywords);
  };

  const addKeyword = () => {
    const keywordCount = keywords.filter(kw => kw.key.startsWith('keyword')).length;
    setKeywords([...keywords, { key: `keyword${keywordCount + 1}`, value: '' }]);
  };

  const removeKeyword = (index: number) => {
    if (keywords.length > 1) {
      setKeywords(keywords.filter((_, i) => i !== index));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setResult(null);

    if (!selectedTemplate) {
      setError('请选择消息模板');
      return;
    }
    if (selectedIds.size === 0) {
      setError('请选择至少一个接收者');
      return;
    }

    // Build keywords map
    const keywordsMap: Record<string, string> = {};
    for (const kw of keywords) {
      if (kw.key.trim() && kw.value.trim()) {
        keywordsMap[kw.key.trim()] = kw.value.trim();
      }
    }

    if (Object.keys(keywordsMap).length === 0) {
      setError('请至少填写一个字段');
      return;
    }

    try {
      setSending(true);
      const response = await sendMessage({
        templateKey: selectedTemplate,
        keywords: keywordsMap,
        recipientIds: Array.from(selectedIds),
      });
      setResult(response);
      
      if (response.totalFailed === 0) {
        setKeywords(keywords.map(kw => ({ ...kw, value: '' })));
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
          <label className="form-label">选择模板</label>
          {templates.length === 0 ? (
            <div className="empty-state">暂无模板，请先在设置页面添加</div>
          ) : (
            <select
              className="form-input"
              value={selectedTemplate}
              onChange={(e) => setSelectedTemplate(e.target.value)}
            >
              {templates.map((t) => (
                <option key={t.id} value={t.key}>{t.name}</option>
              ))}
            </select>
          )}
        </div>

        <div className="form-group">
          <div className="form-label-row">
            <label className="form-label">消息内容</label>
            <label className="custom-toggle">
              <input
                type="checkbox"
                checked={customMode}
                onChange={(e) => setCustomMode(e.target.checked)}
              />
              <span>自定义字段</span>
            </label>
          </div>
          
          {keywords.map((kw, index) => (
            <div key={index} className="keyword-row">
              {customMode && (
                <input
                  type="text"
                  className="form-input keyword-key"
                  value={kw.key}
                  onChange={(e) => handleKeywordChange(index, 'key', e.target.value)}
                  placeholder="字段名"
                />
              )}
              <textarea
                className="form-input keyword-value"
                value={kw.value}
                onChange={(e) => handleKeywordChange(index, 'value', e.target.value)}
                placeholder={customMode ? '字段值' : `输入内容 (${kw.key})`}
              />
              {keywords.length > 1 && (
                <button
                  type="button"
                  className="btn btn-danger btn-small"
                  onClick={() => removeKeyword(index)}
                >
                  −
                </button>
              )}
            </div>
          ))}
          <button type="button" className="btn btn-secondary btn-small btn-add" onClick={addKeyword}>
            + 添加字段
          </button>
        </div>

        <div className="recipients-section">
          <div className="recipients-header">
            <label className="form-label">选择接收者</label>
            <div className="recipients-actions">
              <button type="button" className="btn btn-secondary btn-small" onClick={handleSelectAll}>
                全选
              </button>
              <button type="button" className="btn btn-secondary btn-small" onClick={handleDeselectAll}>
                取消全选
              </button>
            </div>
          </div>

          {recipients.length === 0 ? (
            <div className="empty-state">暂无接收者，请先在设置页面添加</div>
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

        <button type="submit" className="btn btn-primary" disabled={sending || templates.length === 0}>
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
