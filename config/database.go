package config

import (
	"log"
	"os"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"homework-blog/models"
)

var DB *gorm.DB

// InitDatabase 初始化SQLite数据库连接
func InitDatabase() {
	var err error

	// 从环境变量获取SQLite数据库文件路径
	dbPath := getEnv("DB_PATH", "blog.db")

	// 连接SQLite数据库（纯Go驱动，无需CGO）
	DB, err = gorm.Open(sqlite.Open(buildDSN(dbPath)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to SQLite database:", err)
	}

	// 自动迁移数据库表结构
	err = DB.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Comment{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("SQLite database connected and migrated successfully")
}

// buildDSN 拼接SQLite连接串，附带并发与外键相关的pragma
func buildDSN(path string) string {
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}

	options := []string{
		"_pragma=busy_timeout(5000)", // 写锁冲突时等待5秒，避免 database is locked
		"_pragma=journal_mode(WAL)",  // WAL模式提升读写并发
		"_pragma=foreign_keys(1)",    // SQLite默认不启用外键约束
	}

	return path + separator + strings.Join(options, "&")
}

// GetDB 获取数据库连接实例
func GetDB() *gorm.DB {
	return DB
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
