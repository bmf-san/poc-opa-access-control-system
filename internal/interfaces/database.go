package interfaces

import (
	"context"
)

type DBConn interface {
	Query(ctx context.Context, sql string, args ...interface{}) (DBRows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) DBRow
}

type DBRows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...interface{}) error
}

type DBRow interface {
	Scan(dest ...interface{}) error
}
