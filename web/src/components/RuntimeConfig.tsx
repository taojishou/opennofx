import { useState, useEffect } from 'react';
import { Card, Button, Input, Select, Badge } from './ui';
import { useToast } from './ui/Toast';
import { RuntimeConfigItem, ConfigGroup } from '../types/config';
import { theme } from '../styles/theme';

const CONFIG_TYPES = [
  { value: 'all', label: '全部配置' },
  { value: 'risk', label: '风险管理' },
  { value: 'indicator', label: '技术指标' },
  { value: 'database', label: '查询限制' },
  { value: 'pool', label: '币种池' },
  { value: 'trading', label: '交易配置' },
  { value: 'market', label: '市场数据' },
  { value: 'api', label: 'API配置' },
  { value: 'backup', label: '备份配置' },
];

const TYPE_COLORS: { [key: string]: 'success' | 'error' | 'warning' | 'info' | 'purple' | 'default' } = {
  risk: 'error',
  indicator: 'info',
  database: 'success',
  pool: 'purple',
  trading: 'warning',
  market: 'info',
  api: 'purple',
  backup: 'default',
};

export default function RuntimeConfig() {
  const [configs, setConfigs] = useState<RuntimeConfigItem[]>([]);
  const [filteredConfigs, setFilteredConfigs] = useState<ConfigGroup>({});
  const [loading, setLoading] = useState(false);
  const [selectedType, setSelectedType] = useState('all');
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [editValue, setEditValue] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const toast = useToast();

  const API_BASE = 'http://localhost:8080';

  const loadConfigs = async () => {
    setLoading(true);
    try {
      const url = selectedType === 'all' 
        ? `${API_BASE}/api/system/configs`
        : `${API_BASE}/api/system/configs/${selectedType}`;
      
      const response = await fetch(url);
      const data = await response.json();
      const configList = data.configs || [];
      setConfigs(configList);
      
      const grouped = configList.reduce((acc: ConfigGroup, config: RuntimeConfigItem) => {
        if (!acc[config.type]) {
          acc[config.type] = [];
        }
        acc[config.type].push(config);
        return acc;
      }, {});
      
      setFilteredConfigs(grouped);
    } catch (error) {
      console.error('加载配置失败:', error);
      toast.error('加载配置失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConfigs();
  }, [selectedType]);

  useEffect(() => {
    if (searchTerm) {
      const filtered = configs.filter(c => 
        c.key.toLowerCase().includes(searchTerm.toLowerCase()) ||
        c.description.toLowerCase().includes(searchTerm.toLowerCase())
      );
      const grouped = filtered.reduce((acc: ConfigGroup, config: RuntimeConfigItem) => {
        if (!acc[config.type]) {
          acc[config.type] = [];
        }
        acc[config.type].push(config);
        return acc;
      }, {});
      setFilteredConfigs(grouped);
    } else {
      const grouped = configs.reduce((acc: ConfigGroup, config: RuntimeConfigItem) => {
        if (!acc[config.type]) {
          acc[config.type] = [];
        }
        acc[config.type].push(config);
        return acc;
      }, {});
      setFilteredConfigs(grouped);
    }
  }, [searchTerm, configs]);

  const handleEdit = (key: string, value: string) => {
    setEditingKey(key);
    setEditValue(value);
  };

  const handleSave = async (key: string) => {
    try {
      await fetch(`${API_BASE}/api/system/configs`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key, value: editValue }),
      });
      
      toast.success('✓ 配置已更新（已自动热重载）');
      setEditingKey(null);
      loadConfigs();
    } catch (error) {
      console.error('更新配置失败:', error);
      toast.error('更新配置失败');
    }
  };

  const handleCancel = () => {
    setEditingKey(null);
    setEditValue('');
  };

  const getTypeLabel = (type: string) => {
    const found = CONFIG_TYPES.find(t => t.value === type);
    return found ? found.label : type;
  };

  return (
    <div>
      <div className="mb-6">
        <h2 className="text-2xl font-bold mb-2" style={{ color: theme.colors.text.primary }}>
          运行时参数配置
        </h2>
        <p className="text-sm" style={{ color: theme.colors.text.secondary }}>
          管理系统运行时参数（风险阈值、技术指标、查询限制等），修改后自动热重载生效
        </p>
      </div>

      {/* 工具栏 */}
      <div className="mb-6 flex gap-4 items-center">
        <Input
          type="text"
          placeholder="搜索配置项..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          fullWidth
        />
        
        <Select
          value={selectedType}
          onChange={(e) => setSelectedType(e.target.value)}
          options={CONFIG_TYPES}
        />

        <Button
          variant="primary"
          onClick={loadConfigs}
          isLoading={loading}
        >
          {loading ? '加载中...' : '🔄 刷新'}
        </Button>
      </div>

      {/* 配置列表 */}
      <div className="space-y-6">
        {Object.keys(filteredConfigs).sort().map(type => (
          <Card key={type}>
            <div className="flex items-center gap-3 mb-4">
              <Badge variant={TYPE_COLORS[type] || 'default'}>
                {getTypeLabel(type)}
              </Badge>
              <span className="text-sm" style={{ color: theme.colors.text.secondary }}>
                {filteredConfigs[type].length} 项配置
              </span>
            </div>
            
            <div className="space-y-3">
              {filteredConfigs[type].map((config) => (
                <div
                  key={config.key}
                  className="p-4 rounded-lg transition-colors hover:bg-opacity-80"
                  style={{ background: theme.colors.background.tertiary }}
                >
                  <div className="flex items-start gap-4">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <code
                          className="text-sm font-mono px-2 py-1 rounded"
                          style={{ color: theme.colors.brand.primary, background: theme.colors.brand.light }}
                        >
                          {config.key}
                        </code>
                      </div>
                      <p className="text-sm mb-2" style={{ color: theme.colors.text.secondary }}>
                        {config.description}
                      </p>
                      
                      {editingKey === config.key ? (
                        <div className="flex gap-2 items-center">
                          <Input
                            type="text"
                            value={editValue}
                            onChange={(e) => setEditValue(e.target.value)}
                            fullWidth
                            autoFocus
                          />
                          <Button variant="success" onClick={() => handleSave(config.key)}>
                            ✓ 保存
                          </Button>
                          <Button variant="secondary" onClick={handleCancel}>
                            ✗ 取消
                          </Button>
                        </div>
                      ) : (
                        <div className="flex items-center gap-2">
                          <span className="font-mono font-semibold" style={{ color: '#60a5fa' }}>
                            {config.value}
                          </span>
                          <button
                            onClick={() => handleEdit(config.key, config.value)}
                            className="text-sm font-semibold transition-colors"
                            style={{ color: theme.colors.brand.primary }}
                          >
                            ✏️ 编辑
                          </button>
                        </div>
                      )}
                      
                      <p className="text-xs mt-1" style={{ color: theme.colors.text.tertiary }}>
                        最后更新: {new Date(config.updated_at).toLocaleString('zh-CN')}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        ))}
      </div>

      {Object.keys(filteredConfigs).length === 0 && !loading && (
        <div className="text-center py-12" style={{ color: theme.colors.text.secondary }}>
          <div className="text-6xl mb-4 opacity-30">⚙️</div>
          <div className="text-lg font-semibold">
            {searchTerm ? '未找到匹配的配置项' : '暂无配置数据'}
          </div>
        </div>
      )}
    </div>
  );
}
