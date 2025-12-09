import { useState, useEffect } from 'react';
import { getWebhookToken, getTemplates } from '../services/api';
import { WebhookConfig } from '../components/WebhookConfig';
import { MessageTemplate } from '../types';

export function Webhook() {
  const [webhookToken, setWebhookToken] = useState('');
  const [templates, setTemplates] = useState<MessageTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const [tokenData, templatesData] = await Promise.all([
        getWebhookToken(),
        getTemplates(),
      ]);
      setWebhookToken(tokenData.token || '');
      setTemplates(templatesData);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载数据失败');
    } finally {
      setLoading(false);
    }
  };

  const showSuccess = (msg: string) => {
    setSuccess(msg);
    setTimeout(() => setSuccess(null), 3000);
  };

  const showError = (msg: string) => {
    setError(msg);
  };

  if (loading) {
    return <div className="loading">加载中...</div>;
  }

  return (
    <div>
      <h1 className="page-title">Webhook 配置</h1>
      {error && <div className="error-message">{error}</div>}
      {success && <div className="success-message">{success}</div>}
      
      <WebhookConfig 
        token={webhookToken}
        templates={templates}
        onTokenChange={setWebhookToken}
        onSuccess={showSuccess}
        onError={showError}
      />
    </div>
  );
}
