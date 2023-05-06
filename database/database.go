package database

import (
	"errors"
)

var (
	// ErrorInvaildDsn invalid addrs
	ErrorInvaildDsn = errors.New("database: invalid dsn")
	// ErrorDatabaseNil db is not initialized or closed
	ErrorDatabaseNil = errors.New("database: db is not initialized or closed")
)

type Driver string

// Options .
type Options struct {
	// Address address
	Address string
	// MaxIdleConns 连接池空闲连接数
	MaxIdleConns int
	// MaxOpenConns 连接池最大连接数
	MaxOpenConns int
	// ShowSQL 是否显示 sql 日志
	ShowSQL bool
}

// DefaultOptions .
func DefaultOptions() *Options {
	return &Options{
		ShowSQL: false,
	}
}

type Option func(opts *Options)

// Address .
func Address(address string) Option {
	return func(opts *Options) {
		opts.Address = address
	}
}

// ShowSQL .
func ShowSQL(b bool) Option {
	return func(opts *Options) {
		opts.ShowSQL = b
	}
}
