import { useState } from 'react';
import { MessageTemplate } from '../types';
import { createTemplate, deleteTemplate } from '../services/api';

interface Props {
  templates: MessageTemplate[];
  onReload: () => void;
  onError: (msg: string) => void;
}

export function TemplateManager({ templates, onReload, onError }: Props) {
  const [newName, setNewName] = useState('');
  const [newTemplateId, setNewTemplateId] = useState('');
  const [adding, setAdding] = useState(false);

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName.trim() || !newTemplateId.trim()) {
      onError('请填写模板名称和模板ID');
      return;
    }
    try {
      setAdding(true);
      // key 和 name 使用相同的值
      await createTemplate({
        key: newName.trim(),
        templateId: newTemplateId.trim(),
        name: newName.trim(),
      });
      setNewName('');
      setNewTemplateId('');
      onReload();
    } catch (err) {
      onError(err instanceof Error ? err.message : '添加模板失败');
    } finally {
      setAdding(false);
    }
  };

  const handleDelete = async (id: number, name: string) => {
    if (!confirm(`确定要删除模板 "${name}" 吗？`)) return;
    try {
      await deleteTemplate(id);
      onReload();
    } catch (err) {
      onError(err instanceof Error ? err.message : '删除模板失败');
    }
  };

  return (
    <>
      <div className="card settings-section">
        <h2 className="section-title">添加消息模板</h2>
        <form onSubmit={handleAdd} className="add-form template-form">
          <div className="form-group">
            <label className="form-label">模板名称</label>
            <input
              type="text"
              className="form-input"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="如：订单通知"
            />
          </div>
          <div className="form-group">
            <label className="form-label">微信模板 ID</label>
            <input
              type="text"
              className="form-input"
              value={newTemplateId}
              onChange={(e) => setNewTemplateId(e.target.value)}
              placeholder="从微信测试号后台获取"
            />
          </div>
          <button type="submit" className="btn btn-primary" disabled={adding}>
            {adding ? '添加中...' : '添加模板'}
          </button>
        </form>
      </div>

      <div className="card settings-section">
        <h2 className="section-title">模板列表</h2>
        {templates.length === 0 ? (
          <div className="empty-state">暂无模板</div>
        ) : (
          <table className="recipient-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>名称</th>
                <th>模板ID</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {templates.map((t) => (
                <tr key={t.id}>
                  <td>{t.id}</td>
                  <td>{t.name}</td>
                  <td className="template-id-cell">{t.templateId}</td>
                  <td>
                    <button
                      className="btn btn-danger btn-small"
                      onClick={() => handleDelete(t.id, t.name)}
                    >
                      删除
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </>
  );
}
