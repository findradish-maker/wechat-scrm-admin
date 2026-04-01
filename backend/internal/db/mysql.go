package db

import (
	"database/sql"
	"fmt"
	"time"
	"wechat-enterprise-backend/internal/config"
	"wechat-enterprise-backend/internal/domain"

	_ "github.com/go-sql-driver/mysql"
	mysqlDriver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connect(cfg config.MySQLConfig) (*gorm.DB, error) {
	if err := ensureDatabase(cfg); err != nil {
		return nil, err
	}

	driverConfig := newMySQLConfig(cfg, true)
	gdb, err := gorm.Open(gormmysql.Open(driverConfig.FormatDSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := gdb.AutoMigrate(
		&domain.AdminUser{},
		&domain.WechatAccount{},
		&domain.LoginSession{},
		&domain.WechatContact{},
		&domain.WechatMessage{},
		&domain.AIConversationSetting{},
	); err != nil {
		return nil, err
	}

	return gdb, nil
}

func ensureDatabase(cfg config.MySQLConfig) error {
	driverConfig := newMySQLConfig(cfg, false)
	dsn := driverConfig.FormatDSN()

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return err
	}

	_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.Database))
	return err
}

func newMySQLConfig(cfg config.MySQLConfig, withDB bool) *mysqlDriver.Config {
	driverConfig := mysqlDriver.NewConfig()
	driverConfig.User = cfg.User
	driverConfig.Passwd = cfg.Password
	driverConfig.Net = "tcp"
	driverConfig.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	driverConfig.ParseTime = true
	driverConfig.Loc = time.Local
	driverConfig.Params = map[string]string{
		"charset": "utf8mb4",
	}
	if withDB {
		driverConfig.DBName = cfg.Database
	}
	return driverConfig
}
