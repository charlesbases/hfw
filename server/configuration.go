package server

import (
	"time"

	"github.com/charlesbases/hfw/database"
)

// Configuration .
type Configuration struct {
	// Name server name
	Name string
	// Port http port
	Port string
	// Metrics metrics
	Metrics *Metrics
	// JWT jwt
	JWT *Jwt
	// Database database
	Database *Database
}

// Metrics .
type Metrics struct {
	Enable bool
}

// Jwt .
type Jwt struct {
	// Enable enable jwt
	Enable bool
	// Secret jwt secret
	Secret string
	// Intercept jwt 拦截器
	Intercept *JwtIntercept
}

// JwtIntercept .
type JwtIntercept struct {
	// Enable enable intercept
	Enable bool
	// Includes 进行 token 校验。优先级：高
	Includes []string
	// Encludes 忽略 token 校验。优先级：低
	Encludes []string
}

type DatabaseType string

// Database .
type Database struct {
	// Enable enable database
	Enable bool
	// Driver database.Driver
	Driver database.Driver
	// Addrs database dsn
	Addrs []string
	// Debug show sql
	Debug bool
	// Timeout timeout
	Timeout time.Duration
}

// Broker .
type Broker struct {
	// Enable enable broker
	Enable bool
}
