package orm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/charlesbases/hfw/database"
	"github.com/charlesbases/hfw/database/orm/driver"
	"github.com/charlesbases/logger"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

var (
	once      sync.Once
	defaultDB *gorm.DB
)

// New new db
func New(fn driver.Dialector, opts ...database.Option) *gorm.DB {
	var options = new(database.Options)
	for _, opt := range opts {
		opt(options)
	}

	gormDB, err := gorm.Open(fn.Dialector(options), &gorm.Config{Logger: custom(fn.Type(), options.ShowSQL)})
	if err != nil {
		logger.Fatalf("database connect failed. %v", err)
	}

	db, err := gormDB.DB()
	if err != nil {
		logger.Fatalf("database connect failed. %v", err)
	}

	db.SetMaxIdleConns(options.MaxIdleConns)
	db.SetMaxOpenConns(options.MaxOpenConns)

	return gormDB
}

// Init init defaultDB
func Init(fn driver.Dialector, opts ...database.Option) {
	once.Do(func() {
		var options = database.DefaultOptions()
		for _, opt := range opts {
			opt(options)
		}

		gormDB, err := gorm.Open(fn.Dialector(options), &gorm.Config{Logger: custom(fn.Type(), options.ShowSQL)})
		if err != nil {
			logger.Fatalf("database connect failed. %v", err)
		}

		defaultDB = gormDB

		db, err := gormDB.DB()
		if err != nil {
			logger.Fatalf("database connect failed. %v", err)
		}

		db.SetMaxIdleConns(options.MaxIdleConns)
		db.SetMaxOpenConns(options.MaxOpenConns)
	})
}

// DB return defaultDB
func DB() *gorm.DB {
	if defaultDB != nil {
		return defaultDB
	} else {
		logger.Fatal(database.ErrorDatabaseNil)
		return nil
	}
}

// Transaction .
func Transaction(gormDB *gorm.DB, fs ...func(tx *gorm.DB) error) error {
	tx := gormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, f := range fs {
		if err := f(tx); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// l .
type l struct {
	debug  bool
	driver string
}

// custom .
func custom(dt driver.Type, debug bool) glogger.Interface {
	return &l{driver: fmt.Sprintf("[%s] >>> ", dt), debug: debug}
}

func (l *l) LogMode(level glogger.LogLevel) glogger.Interface {
	return l
}

func (l *l) Info(ctx context.Context, format string, v ...interface{}) {
	logger.Infof(l.driver+format, v...)
}

func (l *l) Warn(ctx context.Context, format string, v ...interface{}) {
	logger.Warnf(l.driver+format, v...)
}

func (l *l) Error(ctx context.Context, format string, v ...interface{}) {
	logger.Errorf(l.driver+format, v...)
}

func (l *l) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.debug {
		elapsed := time.Since(begin)
		sql, rows := fc()

		switch {
		case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
			logger.Errorf(l.driver+"%s | %v", sql, err)
		default:
			logger.Debugf(l.driver+"%s | %d rows | %v", sql, rows, elapsed)
		}
	}
}
