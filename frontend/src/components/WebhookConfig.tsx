import { useState } from 'react';
import { generateWebhookToken } from '../services/api';

interface Props {
  token: string;
  onTokenChange: (token: string) => void;
  onSuccess: (msg: string) => void;
  onError: (msg: string) => void;
}

export function WebhookConfig({ token, onTokenChange, onSuccess, onError }: Props) {
  const [generating, setGenerating] = useState(false);

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
            <input type="text" className="form-input" value="http://localhost:8080/webhook/send" readOnly />
            <button type="button" className="btn btn-secondary" onClick={() => copyToClipboard('http://localhost:8080/webhook/send')}>复制</button>
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
          <p><strong>使用方法：</strong></p>
          <pre>{`POST http://localhost:8080/webhook/send
Headers:
  Authorization: Bearer <token>
  Content-Type: application/json
Body:
{
  "title": "消息标题",
  "content": "消息内容",
  "recipientIds": [1, 2]  // 可选，不传则发送给所有接收者
}

curl 示例：
curl -X POST http://localhost:8080/webhook/send \\
  -H "Authorization: Bearer ${token || '<your-token>'}" \\
  -H "Content-Type: application/json" \\
  -d '{"title":"测试标题","content":"测试内容"}'`}</pre>
        </div>
      </div>
    </div>
  );
}
