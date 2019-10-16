package pg

import (
	"github.com/jackc/pgx"
)

func New(connStr string, maxConnections int) (*pgx.ConnPool, error) {
	connCnf, err := pgx.ParseConnectionString(connStr)
	if err != nil {
		return nil, err
	}

	connCnf.PreferSimpleProtocol = true

	cp, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     connCnf,
		MaxConnections: maxConnections,
	})
	if err != nil {
		return nil, err
	}

	return cp, nil
}
