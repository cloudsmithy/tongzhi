import { useState } from 'react';
import { Recipient } from '../types';
import { createRecipient, updateRecipient, deleteRecipient } from '../services/api';

interface Props {
  recipients: Recipient[];
  onReload: () => void;
  onError: (msg: string) => void;
}

export function RecipientManager({ recipients, onReload, onError }: Props) {
  const [newOpenId, setNewOpenId] = useState('');
  const [newName, setNewName] = useState('');
  const [adding, setAdding] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState('');

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newOpenId.trim() || !newName.trim()) {
      onError('请填写完整的 OpenID 和名称');
      return;
    }
    try {
      setAdding(true);
      await createRecipient({ openId: newOpenId.trim(), name: newName.trim() });
      setNewOpenId('');
      setNewName('');
      onReload();
    } catch (err) {
      onError(err instanceof Error ? err.message : '添加接收者失败');
    } finally {
      setAdding(false);
    }
  };

  const handleSaveEdit = async (id: number) => {
    if (!editName.trim()) {
      onError('名称不能为空');
      return;
    }
    try {
      await updateRecipient(id, { name: editName.trim() });
      setEditingId(null);
      setEditName('');
      onReload();
    } catch (err) {
      onError(err instanceof Error ? err.message : '更新接收者失败');
    }
  };

  const handleDelete = async (id: number, name: string) => {
    if (!confirm(`确定要删除接收者 "${name}" 吗？`)) return;
    try {
      await deleteRecipient(id);
      onReload();
    } catch (err) {
      onError(err instanceof Error ? err.message : '删除接收者失败');
    }
  };

  return (
    <>
      <div className="card settings-section">
        <h2 className="section-title">添加接收者</h2>
        <form onSubmit={handleAdd} className="add-form">
          <div className="form-group">
            <label className="form-label">微信 OpenID</label>
            <input type="text" className="form-input" value={newOpenId}
              onChange={(e) => setNewOpenId(e.target.value)} placeholder="请输入微信 OpenID" />
          </div>
          <div className="form-group">
            <label className="form-label">名称</label>
            <input type="text" className="form-input" value={newName}
              onChange={(e) => setNewName(e.target.value)} placeholder="请输入名称" />
          </div>
          <button type="submit" className="btn btn-primary" disabled={adding}>
            {adding ? '添加中...' : '添加'}
          </button>
        </form>
      </div>

      <div className="card settings-section">
        <h2 className="section-title">接收者列表</h2>
        {recipients.length === 0 ? (
          <div className="empty-state">暂无接收者</div>
        ) : (
          <table className="recipient-table">
            <thead>
              <tr><th>ID</th><th>名称</th><th>OpenID</th><th>操作</th></tr>
            </thead>
            <tbody>
              {recipients.map((r) => (
                <tr key={r.id}>
                  <td>{r.id}</td>
                  <td>
                    {editingId === r.id ? (
                      <input type="text" className="form-input" value={editName}
                        onChange={(e) => setEditName(e.target.value)} autoFocus />
                    ) : r.name}
                  </td>
                  <td>{r.openId}</td>
                  <td>
                    <div className="action-buttons">
                      {editingId === r.id ? (
                        <>
                          <button className="btn btn-primary btn-small" onClick={() => handleSaveEdit(r.id)}>保存</button>
                          <button className="btn btn-secondary btn-small" onClick={() => { setEditingId(null); setEditName(''); }}>取消</button>
                        </>
                      ) : (
                        <>
                          <button className="btn btn-secondary btn-small" onClick={() => { setEditingId(r.id); setEditName(r.name); }}>编辑</button>
                          <button className="btn btn-danger btn-small" onClick={() => handleDelete(r.id, r.name)}>删除</button>
                        </>
                      )}
                    </div>
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
