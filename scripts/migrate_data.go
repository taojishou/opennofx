package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// 数据迁移脚本：将旧的decision_logs目录迁移到新的data目录结构

func main() {
	fmt.Println("🔄 开始数据迁移...")
	
	// 检查旧目录是否存在
	oldDir := "decision_logs"
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		fmt.Println("✅ 未发现旧的decision_logs目录，无需迁移")
		return
	}
	
	// 创建新的data目录结构
	newBaseDir := "data"
	newTradersDir := filepath.Join(newBaseDir, "traders")
	backupDir := filepath.Join(newBaseDir, "backups")
	logsDir := filepath.Join(newBaseDir, "logs")
	
	dirs := []string{newTradersDir, backupDir, logsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("❌ 创建目录失败 %s: %v", dir, err)
		}
	}
	
	// 扫描旧目录中的交易员数据
	entries, err := os.ReadDir(oldDir)
	if err != nil {
		log.Fatalf("❌ 读取旧目录失败: %v", err)
	}
	
	migratedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		traderID := entry.Name()
		oldTraderDir := filepath.Join(oldDir, traderID)
		newTraderDir := filepath.Join(newTradersDir, traderID)
		
		fmt.Printf("📁 迁移交易员数据: %s\n", traderID)
		
		// 创建新的交易员目录
		if err := os.MkdirAll(newTraderDir, 0755); err != nil {
			log.Printf("⚠️ 创建交易员目录失败 %s: %v", traderID, err)
			continue
		}
		
		// 迁移数据库文件
		oldDBPath := filepath.Join(oldTraderDir, "decisions.db")
		newDBPath := filepath.Join(newTraderDir, "decisions.db")
		
		if _, err := os.Stat(oldDBPath); err == nil {
			if err := copyFile(oldDBPath, newDBPath); err != nil {
				log.Printf("⚠️ 迁移数据库文件失败 %s: %v", traderID, err)
				continue
			}
			fmt.Printf("  ✅ 数据库文件已迁移: %s\n", newDBPath)
		}
		
		// 迁移其他文件
		if err := migrateDirectory(oldTraderDir, newTraderDir); err != nil {
			log.Printf("⚠️ 迁移目录失败 %s: %v", traderID, err)
			continue
		}
		
		migratedCount++
	}
	
	// 创建备份
	if migratedCount > 0 {
		timestamp := time.Now().Format("20060102_150405")
		backupPath := filepath.Join(backupDir, "migration_backup_"+timestamp)
		
		fmt.Printf("📦 创建迁移备份: %s\n", backupPath)
		if err := copyDirectory(oldDir, backupPath); err != nil {
			log.Printf("⚠️ 创建备份失败: %v", err)
		} else {
			fmt.Printf("  ✅ 备份已创建: %s\n", backupPath)
		}
	}
	
	fmt.Printf("\n🎉 数据迁移完成！\n")
	fmt.Printf("  - 迁移的交易员数量: %d\n", migratedCount)
	fmt.Printf("  - 新数据目录: %s\n", newBaseDir)
	fmt.Printf("  - 旧数据目录: %s (建议确认迁移成功后删除)\n", oldDir)
	
	if migratedCount > 0 {
		fmt.Println("\n⚠️  重要提醒:")
		fmt.Println("  1. 请验证新目录中的数据完整性")
		fmt.Println("  2. 确认系统正常运行后，可以删除旧的decision_logs目录")
		fmt.Println("  3. 备份文件已保存在data/backups目录中")
	}
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	return err
}

// copyDirectory 复制目录
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		
		dstPath := filepath.Join(dst, relPath)
		
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		
		return copyFile(path, dstPath)
	})
}

// migrateDirectory 迁移目录中的文件
func migrateDirectory(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		
		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := migrateDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	
	return nil
}