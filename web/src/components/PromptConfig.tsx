import { useState, useEffect } from 'react';
import { Card, Button, Input, TextArea, Switch, Modal } from './ui';
import { useToast } from './ui/Toast';
import { PromptSection } from '../types/config';
import { theme } from '../styles/theme';

interface PromptConfigProps {
  traderId: string;
}

export default function PromptConfig({ traderId }: PromptConfigProps) {
  const [sections, setSections] = useState<PromptSection[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [editMode, setEditMode] = useState(false);
  const [editContent, setEditContent] = useState('');
  const [editTitle, setEditTitle] = useState('');
  const [saving, setSaving] = useState(false);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [preview, setPreview] = useState('');
  const [showAddForm, setShowAddForm] = useState(false);
  const [activeTab, setActiveTab] = useState<'system' | 'user'>('system');
  const [searchTerm, setSearchTerm] = useState('');
  const [newSection, setNewSection] = useState({ 
    section_name: '', 
    title: '', 
    content: '', 
    prompt_type: 'system' as 'system' | 'user',
    enabled: true 
  });
  const toast = useToast();

  // 筛选和搜索
  const filteredSections = sections
    .filter(s => s.prompt_type === activeTab)
    .filter(s => searchTerm === '' || 
      s.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      s.section_name.toLowerCase().includes(searchTerm.toLowerCase())
    )
    .sort((a, b) => a.display_order - b.display_order);

  const selectedSection = sections.find(s => s.id === selectedId);

  useEffect(() => {
    loadPrompts();
  }, [traderId]);

  useEffect(() => {
    // 切换tab时选择第一个
    if (filteredSections.length > 0 && !selectedId) {
      setSelectedId(filteredSections[0].id);
    }
  }, [activeTab, filteredSections.length]);

  const loadPrompts = async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/prompts?trader_id=${traderId}`);
      const data = await response.json();
      if (data.success) {
        setSections(data.data || []);
        if (data.data?.length > 0) {
          setSelectedId(data.data[0].id);
        }
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
        toast.success(enabled ? '已禁用' : '已启用');
      }
    } catch (error) {
      console.error('切换状态失败:', error);
      toast.error('切换状态失败');
    }
  };

  const handleEdit = () => {
    if (selectedSection) {
      setEditMode(true);
      setEditContent(selectedSection.content);
      setEditTitle(selectedSection.title);
    }
  };

  const handleSave = async () => {
    if (!selectedSection) return;

    try {
      setSaving(true);
      const response = await fetch(`/api/prompts/update?trader_id=${traderId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          section_name: selectedSection.section_name,
          title: editTitle,
          content: editContent,
          prompt_type: selectedSection.prompt_type,
          enabled: selectedSection.enabled,
          display_order: selectedSection.display_order,
        }),
      });
      const data = await response.json();
      if (data.success) {
        setSections(prev =>
          prev.map(s => (s.id === selectedSection.id ? { ...s, content: editContent, title: editTitle } : s))
        );
        setEditMode(false);
        toast.success('保存成功！');
      }
    } catch (error) {
      console.error('保存失败:', error);
      toast.error('保存失败');
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    setEditMode(false);
    setEditContent('');
    setEditTitle('');
  };

  const handleAdd = async () => {
    if (!newSection.section_name || !newSection.title || !newSection.content) {
      toast.error('请填写完整信息');
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
        setNewSection({ 
          section_name: '', 
          title: '', 
          content: '', 
          prompt_type: activeTab,
          enabled: true 
        });
        toast.success('添加成功！');
      } else {
        toast.error('添加失败: ' + (data.error || '未知错误'));
      }
    } catch (error) {
      console.error('添加失败:', error);
      toast.error('添加失败');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!selectedSection) return;
    if (!confirm(`确定要删除 "${selectedSection.title}" 吗？`)) return;

    try {
      const response = await fetch(`/api/prompts/delete?trader_id=${traderId}&section_name=${selectedSection.section_name}`, {
        method: 'DELETE',
      });
      const data = await response.json();
      if (data.success) {
        await loadPrompts();
        setSelectedId(null);
        toast.success('删除成功！');
      }
    } catch (error) {
      console.error('删除失败:', error);
      toast.error('删除失败');
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
      toast.error('预览失败');
    }
  };

  if (loading) {
    return (
      <Card>
        <div style={{ color: theme.colors.text.secondary }}>⏳ 加载中...</div>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {/* 顶部工具栏 */}
      <Card variant="purple">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="text-2xl">⚙️</div>
            <div>
              <h2 className="text-xl font-bold" style={{ color: theme.colors.brand.primary }}>
                Prompt配置管理
              </h2>
              <p className="text-sm" style={{ color: theme.colors.text.secondary }}>
                管理AI决策的系统提示和用户数据模板
              </p>
            </div>
          </div>
          <div className="flex gap-2">
            <Button variant="success" size="sm" onClick={() => setShowAddForm(true)}>
              ➕ 新增
            </Button>
            <Button variant="primary" size="sm" onClick={handlePreview}>
              👁️ 预览
            </Button>
          </div>
        </div>
      </Card>

      {/* 主内容区 - 左右分栏 */}
      <div style={{ display: 'grid', gridTemplateColumns: '320px 1fr', gap: '1rem', minHeight: '600px' }}>
        {/* 左侧列表 */}
        <div className="space-y-3">
          {/* 标签切换 */}
          <Card>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button
                onClick={() => setActiveTab('system')}
                style={{
                  flex: 1,
                  padding: '0.5rem 1rem',
                  borderRadius: theme.radius.md,
                  border: 'none',
                  background: activeTab === 'system' ? theme.colors.purple.gradient : theme.colors.background.tertiary,
                  color: theme.colors.text.primary,
                  fontWeight: activeTab === 'system' ? 'bold' : 'normal',
                  cursor: 'pointer',
                  transition: 'all 0.2s',
                }}
              >
                🧠 System ({sections.filter(s => s.prompt_type === 'system').length})
              </button>
              <button
                onClick={() => setActiveTab('user')}
                style={{
                  flex: 1,
                  padding: '0.5rem 1rem',
                  borderRadius: theme.radius.md,
                  border: 'none',
                  background: activeTab === 'user' ? theme.colors.brand.gradient : theme.colors.background.tertiary,
                  color: theme.colors.text.primary,
                  fontWeight: activeTab === 'user' ? 'bold' : 'normal',
                  cursor: 'pointer',
                  transition: 'all 0.2s',
                }}
              >
                📊 User ({sections.filter(s => s.prompt_type === 'user').length})
              </button>
            </div>
          </Card>

          {/* 搜索框 */}
          <Card>
            <Input
              placeholder="🔍 搜索..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              fullWidth
            />
          </Card>

          {/* 列表 */}
          <Card style={{ maxHeight: '500px', overflow: 'auto' }}>
            <div className="space-y-2">
              {filteredSections.length === 0 ? (
                <div style={{ color: theme.colors.text.tertiary, textAlign: 'center', padding: '2rem' }}>
                  暂无配置
                </div>
              ) : (
                filteredSections.map((section) => (
                  <div
                    key={section.id}
                    onClick={() => {
                      setSelectedId(section.id);
                      setEditMode(false);
                    }}
                    style={{
                      padding: '0.75rem',
                      borderRadius: theme.radius.md,
                      background: selectedId === section.id 
                        ? theme.colors.purple.light 
                        : 'transparent',
                      border: selectedId === section.id
                        ? `1px solid ${theme.colors.purple.border}`
                        : '1px solid transparent',
                      cursor: 'pointer',
                      transition: 'all 0.2s',
                    }}
                    onMouseEnter={(e) => {
                      if (selectedId !== section.id) {
                        e.currentTarget.style.background = theme.colors.background.tertiary;
                      }
                    }}
                    onMouseLeave={(e) => {
                      if (selectedId !== section.id) {
                        e.currentTarget.style.background = 'transparent';
                      }
                    }}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2" style={{ flex: 1, minWidth: 0 }}>
                        <span>{section.title.split(' ')[0]}</span>
                        <span
                          style={{
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                            color: theme.colors.text.primary,
                            fontSize: '0.9rem',
                            fontWeight: selectedId === section.id ? 'bold' : 'normal',
                          }}
                        >
                          {section.title.split(' ').slice(1).join(' ')}
                        </span>
                      </div>
                      <div
                        style={{
                          width: '8px',
                          height: '8px',
                          borderRadius: '50%',
                          background: section.enabled ? theme.colors.success.main : theme.colors.text.tertiary,
                        }}
                      />
                    </div>
                    <div style={{ fontSize: '0.75rem', color: theme.colors.text.tertiary, marginTop: '0.25rem' }}>
                      {section.section_name}
                    </div>
                  </div>
                ))
              )}
            </div>
          </Card>
        </div>

        {/* 右侧详情 */}
        <Card>
          {!selectedSection ? (
            <div style={{ 
              display: 'flex', 
              alignItems: 'center', 
              justifyContent: 'center', 
              height: '100%',
              color: theme.colors.text.tertiary 
            }}>
              请从左侧选择一个Prompt
            </div>
          ) : (
            <div className="space-y-4">
              {/* 头部 */}
              <div className="flex items-start justify-between">
                <div style={{ flex: 1 }}>
                  {editMode ? (
                    <Input
                      value={editTitle}
                      onChange={(e) => setEditTitle(e.target.value)}
                      fullWidth
                      style={{ fontSize: '1.25rem', fontWeight: 'bold' }}
                    />
                  ) : (
                    <>
                      <h3 className="text-2xl font-bold mb-2" style={{ color: theme.colors.text.primary }}>
                        {selectedSection.title}
                      </h3>
                      <div className="flex items-center gap-3 text-sm" style={{ color: theme.colors.text.secondary }}>
                        <span>{selectedSection.section_name}</span>
                        <span 
                          className="px-2 py-0.5 rounded text-xs" 
                          style={{
                            background: selectedSection.prompt_type === 'system' 
                              ? theme.colors.purple.light 
                              : theme.colors.brand.light,
                            color: theme.colors.text.primary,
                          }}
                        >
                          {selectedSection.prompt_type === 'system' ? 'System' : 'User'}
                        </span>
                        <span style={{ fontSize: '0.75rem', color: theme.colors.text.tertiary }}>
                          更新: {new Date(selectedSection.updated_at).toLocaleString('zh-CN', {
                            month: 'numeric',
                            day: 'numeric',
                            hour: '2-digit',
                            minute: '2-digit',
                          })}
                        </span>
                      </div>
                    </>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  <Switch
                    checked={selectedSection.enabled}
                    onChange={() => handleToggle(selectedSection.section_name, selectedSection.enabled)}
                    label={selectedSection.enabled ? '启用' : '禁用'}
                  />
                  {!editMode ? (
                    <>
                      <Button variant="purple" size="sm" onClick={handleEdit}>
                        ✏️ 编辑
                      </Button>
                      <Button variant="danger" size="sm" onClick={handleDelete}>
                        🗑️
                      </Button>
                    </>
                  ) : (
                    <>
                      <Button variant="success" size="sm" onClick={handleSave} isLoading={saving}>
                        ✅ 保存
                      </Button>
                      <Button variant="danger" size="sm" onClick={handleCancel} disabled={saving}>
                        ❌ 取消
                      </Button>
                    </>
                  )}
                </div>
              </div>

              {/* 内容区 */}
              <div style={{ borderTop: `1px solid ${theme.colors.border.primary}`, paddingTop: '1rem' }}>
                {editMode ? (
                  <TextArea
                    value={editContent}
                    onChange={(e) => setEditContent(e.target.value)}
                    rows={20}
                    fullWidth
                    style={{ fontFamily: 'monospace', fontSize: '0.9rem' }}
                  />
                ) : (
                  <pre
                    className="whitespace-pre-wrap font-mono"
                    style={{
                      color: theme.colors.text.secondary,
                      fontSize: '0.9rem',
                      lineHeight: '1.6',
                      maxHeight: '500px',
                      overflow: 'auto',
                    }}
                  >
                    {selectedSection.content}
                  </pre>
                )}
              </div>
            </div>
          )}
        </Card>
      </div>

      {/* 新增对话框 */}
      <Modal
        isOpen={showAddForm}
        onClose={() => setShowAddForm(false)}
        title="➕ 新增Prompt"
        maxWidth="2xl"
      >
        <div className="space-y-4">
          <Input
            label="Section Name"
            placeholder="例如: my_custom_rule"
            value={newSection.section_name}
            onChange={(e) => setNewSection({ ...newSection, section_name: e.target.value })}
            fullWidth
          />
          <Input
            label="标题"
            placeholder="例如: 🎯 我的自定义规则"
            value={newSection.title}
            onChange={(e) => setNewSection({ ...newSection, title: e.target.value })}
            fullWidth
          />
          <div>
            <label className="block text-sm font-medium mb-2" style={{ color: theme.colors.text.primary }}>
              类型
            </label>
            <select
              value={newSection.prompt_type}
              onChange={(e) => setNewSection({ ...newSection, prompt_type: e.target.value as 'system' | 'user' })}
              className="w-full px-4 py-2 rounded-lg border"
              style={{
                background: theme.colors.background.secondary,
                borderColor: theme.colors.border.primary,
                color: theme.colors.text.primary,
              }}
            >
              <option value="system">System (静态规则)</option>
              <option value="user">User (动态数据)</option>
            </select>
          </div>
          <TextArea
            label="内容"
            placeholder="输入Prompt内容..."
            rows={12}
            value={newSection.content}
            onChange={(e) => setNewSection({ ...newSection, content: e.target.value })}
            fullWidth
          />
          <div className="flex gap-3 justify-end">
            <Button variant="danger" onClick={() => setShowAddForm(false)}>
              取消
            </Button>
            <Button variant="success" onClick={handleAdd} isLoading={saving}>
              {saving ? '添加中...' : '确认添加'}
            </Button>
          </div>
        </div>
      </Modal>

      {/* 预览对话框 */}
      <Modal
        isOpen={previewOpen}
        onClose={() => setPreviewOpen(false)}
        title="完整Prompt预览"
        maxWidth="4xl"
        footer={
          <Button
            variant="purple"
            onClick={() => {
              navigator.clipboard.writeText(preview);
              toast.success('已复制！');
            }}
          >
            📋 复制
          </Button>
        }
      >
        <pre
          className="whitespace-pre-wrap font-mono text-sm leading-relaxed p-4 rounded-xl"
          style={{
            background: theme.colors.background.primary,
            color: theme.colors.text.secondary,
            border: `1px solid ${theme.colors.border.secondary}`,
            maxHeight: '600px',
            overflow: 'auto',
          }}
        >
          {preview}
        </pre>
      </Modal>
    </div>
  );
}
