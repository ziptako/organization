package svc

import (
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/ziptako/organization/internal/config"
)

type ServiceContext struct {
	Config    config.Config
	SqlConn   sqlx.SqlConn
	CacheConf cache.CacheConf
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewSqlConn("postgres", c.DataSource)
	return &ServiceContext{
		Config:    c,
		SqlConn:   conn,
		CacheConf: c.Cache, // 确保 CacheConf 被正确传递
	}
}
