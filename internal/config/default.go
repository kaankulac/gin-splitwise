package config

import "time"

var defaultConfig = map[string]interface{}{
	"server.port": 8080,
	"server.readTimeout": "5s",
	"server.writeTimeout": "10s",
	"server.gracefulTimeout": "30s",

	"logging.level": -1,
	"logging.encoding": "console",
	"logging.development": true,

	"jwt.secret": "secret",
	"jwt.sessionTime": "86400s",

	"db.dataSourceName": "postgres://postgres:123123@localhost:5432/splitwise?sslmode=disable",
	"db.logLevel": 1,
	"db.migrate.enable": false,
	"db.migrate.dir": "",
	"db.pool.maxOpen": 10,
	"db.pool.maxIdle": 5,
	"db.pool.maxLifeTime": "5m",

	"cache.enabled": false,
	"cache.prefix": "splitwise-",
	"cache.type": "redis",
	"cache.ttl": 60 * time.Second,
	"cache.redis.cluster": false,
	"cache.redis.endpoints": []string{"localhost:6379"},
	"cache.redis.readTimeout": "3s",
	"cache.redis.writeTimeout": "3s",
	"cache.redis.dialTimeout": "5s",
	"cache.redis.poolSize": 10,
	"cache.redis.maxConnAge": "0",
	"cache.redis.idleTimeout": "5m",

	"metrics.namespace": "splitwise_server",
	"metrics.subsystem": "",
}