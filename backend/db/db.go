package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Dao *gorm.DB

// dbType 记录当前使用的数据库类型: "sqlite" 或 "postgres"
var dbType = "sqlite"

// GetDBType 返回当前数据库类型
func GetDBType() string {
	return dbType
}

// getEnv 获取环境变量，若为空则返回默认值
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Init(sqlitePath string) {
	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second * 3,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			LogLevel:                  logger.Silent,
		},
	)

	var openDb *gorm.DB
	var err error

	// 通过环境变量 GO_STOCK_DB 切换数据库类型: "postgres" 或 "sqlite"（默认）
	driver := getEnv("GO_STOCK_DB", "sqlite")

	switch driver {
	case "postgres":
		// PostgreSQL 连接参数均从环境变量读取，方便配置
		pgHost := getEnv("GO_STOCK_PG_HOST", "localhost")
		pgPort := getEnv("GO_STOCK_PG_PORT", "5432")
		pgUser := getEnv("GO_STOCK_PG_USER", "postgres")
		pgPass := getEnv("GO_STOCK_PG_PASSWORD", "postgres")
		pgDBName := getEnv("GO_STOCK_PG_DBNAME", "gostock")
		pgSSLMode := getEnv("GO_STOCK_PG_SSLMODE", "disable")
		pgTimeZone := getEnv("GO_STOCK_PG_TIMEZONE", "Asia/Shanghai")

		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			pgHost, pgPort, pgUser, pgPass, pgDBName, pgSSLMode, pgTimeZone,
		)
		log.Printf("[db] connecting to PostgreSQL: %s:%s/%s", pgHost, pgPort, pgDBName)

		openDb, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger:                                   dbLogger,
			DisableForeignKeyConstraintWhenMigrating: true,
			SkipDefaultTransaction:                   true,
			PrepareStmt:                              true,
		})
		if err != nil {
			log.Fatalf("PostgreSQL connection error: %s", err.Error())
		}

		dbCon, err := openDb.DB()
		if err != nil {
			log.Fatalf("openDb.DB error: %s", err.Error())
		}
		dbCon.SetMaxIdleConns(5)
		dbCon.SetMaxOpenConns(30)
		dbCon.SetConnMaxLifetime(time.Hour)

		dbType = "postgres"

	default:
		// SQLite（默认行为，完全向后兼容）
		if sqlitePath == "" {
			sqlitePath = "data/stock.db?_busy_timeout=10000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=-524288"
		}
		log.Printf("[db] connecting to SQLite: %s", sqlitePath)

		openDb, err = gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{
			Logger:                                   dbLogger,
			DisableForeignKeyConstraintWhenMigrating: true,
			SkipDefaultTransaction:                   true,
			PrepareStmt:                              true,
		})
		if err != nil {
			log.Fatalf("SQLite connection error: %s", err.Error())
		}

		// 兜底：确保 busy_timeout / WAL / synchronous 生效
		_ = openDb.Exec("PRAGMA busy_timeout=10000").Error
		_ = openDb.Exec("PRAGMA journal_mode=WAL").Error
		_ = openDb.Exec("PRAGMA synchronous=NORMAL").Error

		dbCon, err := openDb.DB()
		if err != nil {
			log.Fatalf("openDb.DB error: %s", err.Error())
		}
		// SQLite 写入是串行锁模型：连接开太多会放大锁竞争导致 SQLITE_BUSY
		dbCon.SetMaxIdleConns(1)
		dbCon.SetMaxOpenConns(5)
		dbCon.SetConnMaxLifetime(time.Hour)

		dbType = "sqlite"
	}

	Dao = openDb
	AutoMigrate()
}
