package driver

import (
	"github.com/charlesbases/hfw/database"
	"github.com/charlesbases/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQL .
var MySQL *m

type m struct{}

// Dialector .
func (m *m) Dialector(opts *database.Options) gorm.Dialector {
	if len(opts.Address) == 0 {
		logger.Fatal(database.ErrorInvaildDsn)
	}
	return mysql.Open(opts.Address)
}

// Type .
func (m *m) Type() Type {
	return TypeMysql
}
