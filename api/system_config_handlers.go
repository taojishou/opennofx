package api

import (
	"database/sql"
	"log"
	"net/http"

	"nofx/database"

	"github.com/gin-gonic/gin"
)

// handleGetSystemConfigs 获取系统配置列表
func (s *Server) handleGetSystemConfigs(c *gin.Context) {
	// 获取系统数据库连接
	systemConn, err := database.NewSystemConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "连接系统数据库失败"})
		return
	}
	defer systemConn.Close()

	// 获取配置类型参数（可选）
	configType := c.Query("type") // api, market, database, risk, indicator, pool, trading, backup

	var query string
	var rows *sql.Rows

	if configType != "" {
		query = `SELECT key, value, description, config_type, updated_at FROM system_configs WHERE config_type = ? ORDER BY key`
		rows, err = systemConn.DB().Query(query, configType)
	} else {
		query = `SELECT key, value, description, config_type, updated_at FROM system_configs ORDER BY config_type, key`
		rows, err = systemConn.DB().Query(query)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询配置失败"})
		return
	}
	defer rows.Close()

	var configs []gin.H
	for rows.Next() {
		var key, value, description, cfgType, updatedAt string
		if err := rows.Scan(&key, &value, &description, &cfgType, &updatedAt); err != nil {
			continue
		}
		configs = append(configs, gin.H{
			"key":         key,
			"value":       value,
			"description": description,
			"type":        cfgType,
			"updated_at":  updatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"configs": configs,
	})
}

// handleUpdateSystemConfig 更新系统配置
func (s *Server) handleUpdateSystemConfig(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 获取系统数据库连接
	systemConn, err := database.NewSystemConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "连接系统数据库失败"})
		return
	}
	defer systemConn.Close()

	// 更新配置
	_, err = systemConn.DB().Exec(`
		UPDATE system_configs SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE key = ?
	`, req.Value, req.Key)

	if err != nil {
		log.Printf("更新配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新配置失败"})
		return
	}

	// 重新加载全局配置（热重载）
	database.ReloadGlobalConfig()

	log.Printf("✓ 配置已更新: %s = %s", req.Key, req.Value)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置更新成功",
	})
}

// handleBatchUpdateConfigs 批量更新配置
func (s *Server) handleBatchUpdateConfigs(c *gin.Context) {
	var req struct {
		Configs []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"configs"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 获取系统数据库连接
	systemConn, err := database.NewSystemConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "连接系统数据库失败"})
		return
	}
	defer systemConn.Close()

	// 开始事务
	tx, err := systemConn.DB().Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "开始事务失败"})
		return
	}
	defer tx.Rollback()

	// 批量更新
	for _, cfg := range req.Configs {
		_, err = tx.Exec(`
			UPDATE system_configs SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE key = ?
		`, cfg.Value, cfg.Key)

		if err != nil {
			log.Printf("更新配置失败 [%s]: %v", cfg.Key, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新配置失败"})
			return
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败"})
		return
	}

	// 重新加载全局配置
	database.ReloadGlobalConfig()

	log.Printf("✓ 批量更新 %d 个配置", len(req.Configs))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置批量更新成功",
		"count":   len(req.Configs),
	})
}

// handleGetConfigByType 按类型获取配置
func (s *Server) handleGetConfigByType(c *gin.Context) {
	configType := c.Param("type")

	// 获取系统数据库连接
	systemConn, err := database.NewSystemConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "连接系统数据库失败"})
		return
	}
	defer systemConn.Close()

	helper := database.NewConfigHelper(systemConn.DB())
	configs, err := helper.GetAllByType(configType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"type":    configType,
		"configs": configs,
	})
}

// handleResetConfig 重置配置为默认值
func (s *Server) handleResetConfig(c *gin.Context) {
	_ = c.Param("key") // 配置键名

	// 这里应该有一个默认值映射，为了简化先返回错误
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "重置功能尚未实现，请手动修改配置值",
	})
}
