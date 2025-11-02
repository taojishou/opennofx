#!/bin/bash

# 数据迁移脚本
# 将旧的decision_logs目录迁移到新的data目录结构

set -e

echo "🔄 NoFX 数据迁移工具"
echo "===================="

# 检查是否在正确的目录
if [ ! -f "go.mod" ]; then
    echo "❌ 请在项目根目录运行此脚本"
    exit 1
fi

# 创建scripts目录（如果不存在）
mkdir -p scripts

# 编译迁移工具
echo "📦 编译迁移工具..."
cd scripts
go mod init migrate 2>/dev/null || true
go build -o migrate_tool migrate_data.go

# 运行迁移
echo "🚀 开始数据迁移..."
./migrate_tool

# 清理临时文件
rm -f migrate_tool go.mod go.sum

echo ""
echo "✅ 迁移脚本执行完成！"
echo ""
echo "📋 后续步骤："
echo "1. 重启应用程序以使用新的数据目录"
echo "2. 验证数据完整性"
echo "3. 确认无误后删除旧的decision_logs目录"