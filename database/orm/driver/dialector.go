package driver

import (
	"github.com/charlesbases/hfw/database"
	"gorm.io/gorm"
)

const (
	// TypeMysql mysql
	TypeMysql Type = "Mysql"
	// TypePostgres postgres
	TypePostgres Type = "Postgres"
)

type Type string

type Dialector interface {
	Dialector(optons *database.Options) gorm.Dialector
	Type() Type
}
