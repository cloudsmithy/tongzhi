import { useState, useEffect } from 'react';
import { Recipient, WeChatConfig, MessageTemplate } from '../types';
import { getRecipients, getWeChatConfig, getTemplates } from '../services/api';
import { WeChatConfig as WeChatConfigComponent } from '../components/WeChatConfig';
import { RecipientManager } from '../components/RecipientManager';
import { TemplateManager } from '../components/TemplateManager';

export function Settings() {
  const [recipients, setRecipients] = useState<Recipient[]>([]);
  const [templates, setTemplates] = useState<MessageTemplate[]>([]);
  const [wechatConfig, setWechatConfig] = useState<WeChatConfig>({ appId: '', appSecret: '', templateId: '' });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const [recipientsData, configData, templatesData] = await Promise.all([
        getRecipients(),
        getWeChatConfig(),
        getTemplates(),
      ]);
      setRecipients(recipientsData);
      setWechatConfig(configData);
      setTemplates(templatesData);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载数据失败');
    } finally {
      setLoading(false);
    }
  };

  const loadRecipients = async () => {
    try {
      const data = await getRecipients();
      setRecipients(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载接收者失败');
    }
  };

  const loadTemplates = async () => {
    try {
      const data = await getTemplates();
      setTemplates(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载模板失败');
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
      <h1 className="page-title">系统设置</h1>
      {error && <div className="error-message">{error}</div>}
      {success && <div className="success-message">{success}</div>}
      
      <WeChatConfigComponent
        config={wechatConfig}
        onConfigChange={setWechatConfig}
        onSuccess={showSuccess}
        onError={showError}
      />
      
      <RecipientManager
        recipients={recipients}
        onReload={loadRecipients}
        onError={showError}
      />
      
      <TemplateManager
        templates={templates}
        onReload={loadTemplates}
        onError={showError}
      />
    </div>
  );
}
