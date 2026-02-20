package pgxpool

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Config struct {
	ConnString      string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime any
	MaxConnIdleTime any
}

type Pool struct{}

type Row struct {
	values []interface{}
}

type Tx struct{}

func ParseConfig(connString string) (*Config, error) { return &Config{ConnString: connString}, nil }
func NewWithConfig(ctx context.Context, cfg *Config) (*Pool, error) { _ = ctx; _ = cfg; return &Pool{}, nil }
func (p *Pool) Ping(ctx context.Context) error { _ = ctx; return nil }
func (p *Pool) Close() {}
func (p *Pool) Exec(ctx context.Context, sql string, args ...interface{}) (string, error) { _ = ctx; _ = sql; _ = args; return "", nil }
func (p *Pool) QueryRow(ctx context.Context, sql string, args ...interface{}) *Row { _ = ctx; _ = sql; _ = args; return &Row{} }
func (r *Row) Scan(dest ...interface{}) error {
	for _, d := range dest {
		switch v := d.(type) {
		case *bool:
			*v = false
		}
	}
	return nil
}
func (p *Pool) BeginTx(ctx context.Context, opts pgx.TxOptions) (*Tx, error) { _ = ctx; _ = opts; return &Tx{}, nil }
func (t *Tx) Exec(ctx context.Context, sql string, args ...interface{}) (string, error) { _ = ctx; _ = sql; _ = args; return "", nil }
func (t *Tx) Commit(ctx context.Context) error { _ = ctx; return nil }
func (t *Tx) Rollback(ctx context.Context) error { _ = ctx; return nil }
