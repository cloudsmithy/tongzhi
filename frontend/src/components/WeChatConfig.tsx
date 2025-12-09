import { useState } from 'react';
import { WeChatConfig as WeChatConfigType } from '../types';
import { saveWeChatConfig } from '../services/api';

interface Props {
  config: WeChatConfigType;
  onConfigChange: (config: WeChatConfigType) => void;
  onSuccess: (msg: string) => void;
  onError: (msg: string) => void;
}

export function WeChatConfig({ config, onConfigChange, onSuccess, onError }: Props) {
  const [saving, setSaving] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!config.appId.trim() || !config.templateId.trim()) {
      onError('请填写 AppID 和模板 ID');
      return;
    }
    try {
      setSaving(true);
      await saveWeChatConfig(config);
      onSuccess('微信配置保存成功');
    } catch (err) {
      onError(err instanceof Error ? err.message : '保存配置失败');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="card settings-section">
      <h2 className="section-title">微信测试号配置</h2>
      <form onSubmit={handleSubmit} className="add-form">
        <div className="form-group">
          <label className="form-label">AppID</label>
          <input type="text" className="form-input" value={config.appId}
            onChange={(e) => onConfigChange({...config, appId: e.target.value})}
            placeholder="请输入微信测试号 AppID" />
        </div>
        <div className="form-group">
          <label className="form-label">AppSecret</label>
          <input type="password" className="form-input" value={config.appSecret}
            onChange={(e) => onConfigChange({...config, appSecret: e.target.value})}
            placeholder="请输入 AppSecret（留空保持原值）" />
        </div>
        <div className="form-group">
          <label className="form-label">模板 ID</label>
          <input type="text" className="form-input" value={config.templateId}
            onChange={(e) => onConfigChange({...config, templateId: e.target.value})}
            placeholder="请输入消息模板 ID" />
        </div>
        <button type="submit" className="btn btn-primary" disabled={saving}>
          {saving ? '保存中...' : '保存配置'}
        </button>
      </form>
    </div>
  );
}
