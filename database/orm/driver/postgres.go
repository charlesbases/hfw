package driver

import (
	"github.com/charlesbases/hfw/database"
	"github.com/charlesbases/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Postgres Postgres
var Postgres *p

type p struct{}

// Dialector .
func (p *p) Dialector(opts *database.Options) gorm.Dialector {
	if len(opts.Address) == 0 {
		logger.Fatal(database.ErrorInvaildDsn)
	}
	return postgres.Open(opts.Address)
}

// Type .
func (p *p) Type() Type {
	return TypePostgres
}
