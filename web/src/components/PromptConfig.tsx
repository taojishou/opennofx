import { useState, useEffect } from 'react';

interface PromptSection {
  id: number;
  section_name: string;
  title: string;
  content: string;
  enabled: boolean;
  display_order: number;
  updated_at: string;
}

interface PromptConfigProps {
  traderId: string;
}

export default function PromptConfig({ traderId }: PromptConfigProps) {
  const [sections, setSections] = useState<PromptSection[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editContent, setEditContent] = useState('');
  const [saving, setSaving] = useState(false);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [preview, setPreview] = useState('');
  const [showAddForm, setShowAddForm] = useState(false);
  const [newSection, setNewSection] = useState({ section_name: '', title: '', content: '', enabled: true });

  useEffect(() => {
    loadPrompts();
  }, [traderId]);

  const loadPrompts = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/prompts?trader_id=${traderId}`);
      const data = await response.json();
      if (data.success) {
        setSections(data.data || []);
      }
    } catch (error) {
      console.error('加载Prompt配置失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleToggle = async (sectionName: string, enabled: boolean) => {
    try {
      const response = await fetch(`/api/prompts/toggle?trader_id=${traderId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ section_name: sectionName, enabled: !enabled }),
      });
      const data = await response.json();
      if (data.success) {
        setSections(prev =>
          prev.map(s => (s.section_name === sectionName ? { ...s, enabled: !enabled } : s))
        );
      }
    } catch (error) {
      console.error('切换状态失败:', error);
      alert('切换状态失败，请重试');
    }
  };

  const handleEdit = (section: PromptSection) => {
    setEditingId(section.id);
    setEditContent(section.content);
  };

  const handleSave = async (section: PromptSection) => {
    try {
      setSaving(true);
      const response = await fetch(`/api/prompts/update?trader_id=${traderId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          section_name: section.section_name,
          title: section.title,
          content: editContent,
          enabled: section.enabled,
          display_order: section.display_order,
        }),
      });
      const data = await response.json();
      if (data.success) {
        setSections(prev =>
          prev.map(s => (s.id === section.id ? { ...s, content: editContent } : s))
        );
        setEditingId(null);
        alert('保存成功！');
      }
    } catch (error) {
      console.error('保存失败:', error);
      alert('保存失败，请重试');
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    setEditingId(null);
    setEditContent('');
  };

  const handleAdd = async () => {
    if (!newSection.section_name || !newSection.title || !newSection.content) {
      alert('请填写完整信息');
      return;
    }

    try {
      setSaving(true);
      const response = await fetch(`/api/prompts/add?trader_id=${traderId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newSection),
      });
      const data = await response.json();
      if (data.success) {
        await loadPrompts();
        setShowAddForm(false);
        setNewSection({ section_name: '', title: '', content: '', enabled: true });
        alert('添加成功！');
      } else {
        alert('添加失败: ' + (data.error || '未知错误'));
      }
    } catch (error) {
      console.error('添加失败:', error);
      alert('添加失败，请重试');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (sectionName: string) => {
    if (!confirm(`确定要删除 "${sectionName}" 吗？此操作不可撤销！`)) {
      return;
    }

    try {
      const response = await fetch(`/api/prompts/delete?trader_id=${traderId}&section_name=${sectionName}`, {
        method: 'DELETE',
      });
      const data = await response.json();
      if (data.success) {
        await loadPrompts();
        alert('删除成功！');
      } else {
        alert('删除失败: ' + (data.error || '未知错误'));
      }
    } catch (error) {
      console.error('删除失败:', error);
      alert('删除失败，请重试');
    }
  };

  const handlePreview = async () => {
    try {
      const response = await fetch(`/api/prompts/preview?trader_id=${traderId}`);
      const data = await response.json();
      if (data.success) {
        setPreview(data.data.prompt);
        setPreviewOpen(true);
      }
    } catch (error) {
      console.error('预览失败:', error);
      alert('预览失败，请重试');
    }
  };

  if (loading) {
    return (
      <div className="rounded-2xl p-8" style={{ background: '#1E2329', border: '1px solid #2B3139' }}>
        <div style={{ color: '#848E9C' }}>⏳ 加载中...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 标题栏 */}
      <div className="relative rounded-2xl p-6 overflow-hidden" style={{
        background: 'linear-gradient(135deg, rgba(139, 92, 246, 0.15) 0%, rgba(99, 102, 241, 0.1) 50%, rgba(30, 35, 41, 0.8) 100%)',
        border: '1px solid rgba(139, 92, 246, 0.3)',
        boxShadow: '0 8px 32px rgba(139, 92, 246, 0.2)'
      }}>
        <div className="absolute top-0 right-0 w-96 h-96 rounded-full opacity-10" style={{
          background: 'radial-gradient(circle, #8B5CF6 0%, transparent 70%)',
          filter: 'blur(60px)'
        }} />
        <div className="relative flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 rounded-2xl flex items-center justify-center text-3xl" style={{
              background: 'linear-gradient(135deg, #8B5CF6 0%, #6366F1 100%)',
              boxShadow: '0 8px 24px rgba(139, 92, 246, 0.5)',
              border: '2px solid rgba(255, 255, 255, 0.1)'
            }}>
              ⚙️
            </div>
            <div>
              <h2 className="text-3xl font-bold mb-1" style={{
                color: '#EAECEF',
                textShadow: '0 2px 8px rgba(139, 92, 246, 0.3)'
              }}>
                AI Prompt 配置
              </h2>
              <p className="text-base" style={{ color: '#A78BFA' }}>
                动态调整AI交易策略，实时生效
              </p>
            </div>
          </div>
          <div className="flex gap-3">
            <button
              onClick={() => setShowAddForm(!showAddForm)}
              className="px-6 py-3 rounded-xl font-bold transition-all hover:scale-105"
              style={{
                background: 'linear-gradient(135deg, #10B981 0%, #0ECB81 100%)',
                color: '#FFFFFF',
                boxShadow: '0 4px 16px rgba(16, 185, 129, 0.3)'
              }}
            >
              ➕ 新增Prompt
            </button>
            <button
              onClick={handlePreview}
              className="px-6 py-3 rounded-xl font-bold transition-all hover:scale-105"
              style={{
                background: 'linear-gradient(135deg, #F0B90B 0%, #FCD535 100%)',
                color: '#1E2329',
                boxShadow: '0 4px 16px rgba(240, 185, 11, 0.3)'
              }}
            >
              👁️ 预览完整Prompt
            </button>
          </div>
        </div>
      </div>

      {/* 新增Prompt表单 */}
      {showAddForm && (
        <div className="rounded-2xl p-6" style={{
          background: 'rgba(30, 35, 41, 0.8)',
          border: '1px solid rgba(139, 92, 246, 0.3)',
          boxShadow: '0 4px 16px rgba(139, 92, 246, 0.1)'
        }}>
          <h3 className="text-xl font-bold mb-4" style={{ color: '#EAECEF' }}>
            ➕ 新增Prompt Section
          </h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>
                Section Name (英文标识，唯一)
              </label>
              <input
                type="text"
                value={newSection.section_name}
                onChange={(e) => setNewSection({ ...newSection, section_name: e.target.value })}
                placeholder="例如: my_custom_rule"
                className="w-full rounded-xl p-3"
                style={{
                  background: 'rgba(0, 0, 0, 0.3)',
                  border: '1px solid rgba(139, 92, 246, 0.3)',
                  color: '#EAECEF',
                  outline: 'none'
                }}
              />
            </div>
            <div>
              <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>
                标题 (可包含emoji)
              </label>
              <input
                type="text"
                value={newSection.title}
                onChange={(e) => setNewSection({ ...newSection, title: e.target.value })}
                placeholder="例如: 🎯 我的自定义规则"
                className="w-full rounded-xl p-3"
                style={{
                  background: 'rgba(0, 0, 0, 0.3)',
                  border: '1px solid rgba(139, 92, 246, 0.3)',
                  color: '#EAECEF',
                  outline: 'none'
                }}
              />
            </div>
            <div>
              <label className="block text-sm mb-2" style={{ color: '#848E9C' }}>
                内容 (Markdown格式)
              </label>
              <textarea
                value={newSection.content}
                onChange={(e) => setNewSection({ ...newSection, content: e.target.value })}
                rows={10}
                placeholder="输入Prompt内容..."
                className="w-full rounded-xl p-4 font-mono text-sm leading-relaxed resize-y"
                style={{
                  background: 'rgba(0, 0, 0, 0.3)',
                  border: '1px solid rgba(139, 92, 246, 0.3)',
                  color: '#E0E7FF',
                  outline: 'none'
                }}
              />
            </div>
            <div className="flex gap-3">
              <button
                onClick={handleAdd}
                disabled={saving}
                className="px-6 py-3 rounded-xl font-bold transition-all hover:scale-105 disabled:opacity-50"
                style={{
                  background: 'linear-gradient(135deg, #10B981 0%, #0ECB81 100%)',
                  color: '#FFFFFF',
                  boxShadow: '0 4px 16px rgba(16, 185, 129, 0.3)'
                }}
              >
                {saving ? '⏳ 添加中...' : '✅ 确认添加'}
              </button>
              <button
                onClick={() => {
                  setShowAddForm(false);
                  setNewSection({ section_name: '', title: '', content: '', enabled: true });
                }}
                disabled={saving}
                className="px-6 py-3 rounded-xl font-bold transition-all hover:scale-105 disabled:opacity-50"
                style={{
                  background: 'rgba(248, 113, 113, 0.2)',
                  color: '#FCA5A5',
                  border: '1px solid rgba(248, 113, 113, 0.3)'
                }}
              >
                ❌ 取消
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Prompt部分列表 */}
      <div className="space-y-4">
        {sections.map((section) => {
          const isEditing = editingId === section.id;
          
          return (
            <div
              key={section.id}
              className="rounded-2xl overflow-hidden transition-all hover:scale-[1.01]"
              style={{
                background: 'rgba(30, 35, 41, 0.6)',
                border: section.enabled ? '1px solid rgba(139, 92, 246, 0.3)' : '1px solid rgba(71, 85, 105, 0.3)',
                boxShadow: section.enabled ? '0 4px 16px rgba(139, 92, 246, 0.1)' : 'none'
              }}
            >
              {/* Header */}
              <div className="p-6 border-b" style={{ borderColor: 'rgba(71, 85, 105, 0.3)' }}>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-4">
                    <div className="text-2xl">{section.title.split(' ')[0]}</div>
                    <div>
                      <h3 className="text-xl font-bold" style={{ color: '#EAECEF' }}>
                        {section.title}
                      </h3>
                      <div className="text-sm" style={{ color: '#848E9C' }}>
                        {section.section_name}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <label className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={section.enabled}
                        onChange={() => handleToggle(section.section_name, section.enabled)}
                        className="w-5 h-5 rounded cursor-pointer"
                        style={{
                          accentColor: '#8B5CF6'
                        }}
                      />
                      <span className="text-sm font-semibold" style={{ 
                        color: section.enabled ? '#A78BFA' : '#848E9C' 
                      }}>
                        {section.enabled ? '已启用' : '已禁用'}
                      </span>
                    </label>
                    {!isEditing ? (
                      <div className="flex gap-2">
                        <button
                          onClick={() => handleEdit(section)}
                          className="px-4 py-2 rounded-lg font-semibold transition-all hover:scale-105"
                          style={{
                            background: 'rgba(139, 92, 246, 0.2)',
                            color: '#A78BFA',
                            border: '1px solid rgba(139, 92, 246, 0.3)'
                          }}
                        >
                          ✏️ 编辑
                        </button>
                        <button
                          onClick={() => handleDelete(section.section_name)}
                          className="px-4 py-2 rounded-lg font-semibold transition-all hover:scale-105"
                          style={{
                            background: 'rgba(248, 113, 113, 0.2)',
                            color: '#FCA5A5',
                            border: '1px solid rgba(248, 113, 113, 0.3)'
                          }}
                        >
                          🗑️ 删除
                        </button>
                      </div>
                    ) : null}
                  </div>
                </div>
              </div>

              {/* Content */}
              <div className="p-6">
                {isEditing ? (
                  <div className="space-y-4">
                    <textarea
                      value={editContent}
                      onChange={(e) => setEditContent(e.target.value)}
                      rows={15}
                      className="w-full rounded-xl p-4 font-mono text-sm leading-relaxed resize-y"
                      style={{
                        background: 'rgba(0, 0, 0, 0.3)',
                        border: '1px solid rgba(139, 92, 246, 0.3)',
                        color: '#E0E7FF',
                        outline: 'none'
                      }}
                    />
                    <div className="flex gap-3">
                      <button
                        onClick={() => handleSave(section)}
                        disabled={saving}
                        className="px-6 py-3 rounded-xl font-bold transition-all hover:scale-105 disabled:opacity-50"
                        style={{
                          background: 'linear-gradient(135deg, #10B981 0%, #0ECB81 100%)',
                          color: '#FFFFFF',
                          boxShadow: '0 4px 16px rgba(16, 185, 129, 0.3)'
                        }}
                      >
                        {saving ? '⏳ 保存中...' : '✅ 保存'}
                      </button>
                      <button
                        onClick={handleCancel}
                        disabled={saving}
                        className="px-6 py-3 rounded-xl font-bold transition-all hover:scale-105 disabled:opacity-50"
                        style={{
                          background: 'rgba(248, 113, 113, 0.2)',
                          color: '#FCA5A5',
                          border: '1px solid rgba(248, 113, 113, 0.3)'
                        }}
                      >
                        ❌ 取消
                      </button>
                    </div>
                  </div>
                ) : (
                  <pre
                    className="whitespace-pre-wrap font-mono text-sm leading-relaxed"
                    style={{
                      color: section.enabled ? '#CBD5E1' : '#64748B'
                    }}
                  >
                    {section.content}
                  </pre>
                )}
              </div>

              {/* Footer */}
              {!isEditing && (
                <div className="px-6 py-3 border-t" style={{ 
                  borderColor: 'rgba(71, 85, 105, 0.3)',
                  background: 'rgba(0, 0, 0, 0.2)'
                }}>
                  <div className="text-xs" style={{ color: '#64748B' }}>
                    最后更新: {new Date(section.updated_at).toLocaleString('zh-CN')}
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* 预览对话框 */}
      {previewOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-4"
          style={{ background: 'rgba(0, 0, 0, 0.8)' }}
          onClick={() => setPreviewOpen(false)}
        >
          <div
            className="rounded-2xl p-6 max-w-4xl w-full max-h-[80vh] overflow-y-auto"
            style={{
              background: '#1E2329',
              border: '1px solid rgba(139, 92, 246, 0.3)',
              boxShadow: '0 20px 60px rgba(0, 0, 0, 0.5)'
            }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-2xl font-bold" style={{ color: '#EAECEF' }}>
                完整System Prompt预览
              </h3>
              <button
                onClick={() => setPreviewOpen(false)}
                className="text-2xl hover:scale-110 transition-transform"
                style={{ color: '#848E9C' }}
              >
                ✕
              </button>
            </div>
            <pre
              className="whitespace-pre-wrap font-mono text-sm leading-relaxed p-4 rounded-xl"
              style={{
                background: 'rgba(0, 0, 0, 0.3)',
                color: '#CBD5E1',
                border: '1px solid rgba(71, 85, 105, 0.3)'
              }}
            >
              {preview}
            </pre>
            <div className="mt-4 flex justify-end">
              <button
                onClick={() => {
                  navigator.clipboard.writeText(preview);
                  alert('已复制到剪贴板！');
                }}
                className="px-6 py-3 rounded-xl font-bold transition-all hover:scale-105"
                style={{
                  background: 'rgba(139, 92, 246, 0.2)',
                  color: '#A78BFA',
                  border: '1px solid rgba(139, 92, 246, 0.3)'
                }}
              >
                📋 复制到剪贴板
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 说明卡片 */}
      <div className="rounded-2xl p-6" style={{
        background: 'linear-gradient(135deg, rgba(240, 185, 11, 0.1) 0%, rgba(252, 213, 53, 0.05) 100%)',
        border: '1px solid rgba(240, 185, 11, 0.2)'
      }}>
        <div className="flex items-start gap-4">
          <div className="text-2xl">💡</div>
          <div>
            <h4 className="font-bold mb-2" style={{ color: '#FCD34D' }}>使用提示</h4>
            <ul className="space-y-2 text-sm" style={{ color: '#CBD5E1' }}>
              <li>• 修改后立即生效，AI下次决策时会使用新配置</li>
              <li>• 可以禁用某些部分进行A/B测试</li>
              <li>• 支持变量: {'{{accountEquity}}'}, {'{{btcEthLeverage}}'}, {'{{altcoinLeverage}}'}</li>
              <li>• 建议小幅调整并观察效果，避免大幅改动</li>
              <li>• 点击"预览完整Prompt"可查看AI实际接收的内容</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}
