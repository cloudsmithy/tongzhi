import { useState } from 'react';
import { generateWebhookToken } from '../services/api';
import { MessageTemplate } from '../types';

interface Props {
  token: string;
  templates: MessageTemplate[];
  onTokenChange: (token: string) => void;
  onSuccess: (msg: string) => void;
  onError: (msg: string) => void;
}

export function WebhookConfig({ token, templates, onTokenChange, onSuccess, onError }: Props) {
  const [generating, setGenerating] = useState(false);
  
  // 获取示例用的 templateKey
  const exampleTemplateKey = templates.length > 0 ? templates[0].key : '订单通知';
  const webhookUrl = `${window.location.origin}/api/webhook/send`;

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    onSuccess('已复制到剪贴板');
  };

  const handleGenerate = async () => {
    if (token && !confirm('确定要重新生成 Token 吗？旧 Token 将失效。')) return;
    try {
      setGenerating(true);
      const newToken = await generateWebhookToken();
      onTokenChange(newToken);
      onSuccess('Webhook Token 生成成功');
    } catch (err) {
      onError(err instanceof Error ? err.message : '生成 Token 失败');
    } finally {
      setGenerating(false);
    }
  };

  return (
    <div className="card settings-section">
      <h2 className="section-title">Webhook 配置</h2>
      <div className="webhook-info">
        <div className="form-group">
          <label className="form-label">Webhook URL</label>
          <div className="input-with-button">
            <input type="text" className="form-input" value={webhookUrl} readOnly />
            <button type="button" className="btn btn-secondary" onClick={() => copyToClipboard(webhookUrl)}>复制</button>
          </div>
        </div>
        <div className="form-group">
          <label className="form-label">Token</label>
          <div className="input-with-button">
            <input type="text" className="form-input" value={token || '未生成'} readOnly />
            {token && <button type="button" className="btn btn-secondary" onClick={() => copyToClipboard(token)}>复制</button>}
            <button type="button" className="btn btn-primary" onClick={handleGenerate} disabled={generating}>
              {generating ? '生成中...' : token ? '重新生成' : '生成 Token'}
            </button>
          </div>
        </div>
        <div className="webhook-usage">
          <p><strong>请求格式：</strong></p>
          <textarea
            className="form-input webhook-example"
            readOnly
            value={`{
  "templateKey": "${exampleTemplateKey}",
  "keywords": {
    "first": "您有一条新消息",
    "keyword1": "订单编号: 2024120901",
    "keyword2": "金额: ¥99.00",
    "keyword3": "时间: 2024-12-09 15:30",
    "remark": "点击查看详情"
  },
  "recipientIds": [1]
}`}
          />
          <p style={{ marginTop: '15px' }}><strong>curl 示例：</strong></p>
          <pre>{`curl -X POST ${webhookUrl} \\
  -H "Authorization: Bearer ${token || '<your-token>'}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "templateKey": "${exampleTemplateKey}",
    "keywords": {
      "first": "您有一条新消息",
      "keyword1": "订单编号: 2024120901"
    }
  }'`}</pre>
        </div>
      </div>
    </div>
  );
}
