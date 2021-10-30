package server

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func newPG(conf *PGConfig, logger *zap.Logger) (*sqlx.DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", conf.User, conf.Password, conf.DB)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		logger.Error("can't connect to pg", zap.Error(err))
		return nil, err
	}

	return db, nil
}
