package main

import (
	"fmt"
	"log"
	"nofx/database"
	"os"
)

// migrate_config 配置迁移工具
// 用法: go run cmd/migrate_config.go [config.json路径]
func main() {
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║   配置迁移工具: config.json → database       ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
	fmt.Println()

	// 获取配置文件路径
	configFile := "config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	log.Printf("📋 配置文件: %s", configFile)

	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatalf("❌ 配置文件不存在: %s", configFile)
	}

	// 创建数据库管理器
	manager, err := database.NewManager()
	if err != nil {
		log.Fatalf("❌ 创建数据库管理器失败: %v", err)
	}
	defer manager.Close()

	// 执行迁移
	if err := database.MigrateFromConfigFile(configFile, manager); err != nil {
		log.Fatalf("❌ 配置迁移失败: %v", err)
	}

	fmt.Println()
	fmt.Println("✅ 配置迁移成功！")
	fmt.Println()
	fmt.Println("现在可以：")
	fmt.Println("  1. 备份 config.json 文件")
	fmt.Println("  2. 运行主程序将从数据库读取配置")
	fmt.Println("  3. 通过API动态管理配置")
}
