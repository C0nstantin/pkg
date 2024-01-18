package pgx_client

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.wm.local/wm/pkg/log"
)

type Client interface {
	DB() PgxPoolIface
	Close() error
}

func NewClient(ctx context.Context, dsn string) (Client, error) {
	return newClient(ctx, dsn)
}

func newClient(ctx context.Context, dsn string) (Client, error) {
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &client{db: &pgDB{conn: conn}}, nil

}

type PgxPoolIface interface {
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	AcquireAllIdle(ctx context.Context) []*pgxpool.Conn
	AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Close()
	Config() *pgxpool.Config
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Reset()
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Stat() *pgxpool.Stat
}
type pgDB struct {
	conn PgxPoolIface
}

func (p *pgDB) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return p.conn.Acquire(ctx)
}

func (p *pgDB) AcquireAllIdle(ctx context.Context) []*pgxpool.Conn {
	return p.conn.AcquireAllIdle(ctx)
}

func (p *pgDB) AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error {
	return p.conn.AcquireFunc(ctx, f)
}

func (p *pgDB) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.conn.Begin(ctx)
}

func (p *pgDB) Config() *pgxpool.Config {
	return p.conn.Config()
}

func (p *pgDB) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return p.conn.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (p *pgDB) Reset() {
	p.conn.Reset()
}

func (p *pgDB) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return p.conn.SendBatch(ctx, b)
}

func (p *pgDB) Stat() *pgxpool.Stat {
	return p.conn.Stat()
}

func (p *pgDB) ReadCommitted(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := p.conn.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	if err := fn(ctx); err != nil {
		log.Println("failed to execute transaction: " + err.Error())
		return tx.Rollback(ctx)
	}
	return tx.Commit(ctx)
}

func (p *pgDB) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return p.conn.Exec(ctx, query, args...)
}

func (p *pgDB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	log.Debugf("query: %s with args: %v\n", query, args)
	return p.conn.Query(ctx, query, args...)
}

func (p *pgDB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	log.Debugf("query row: %s with args: %v\n", query, args)
	return p.conn.QueryRow(ctx, query, args...)
}

func (p *pgDB) Ping(ctx context.Context) error {
	log.Debugf("ping")
	return p.conn.Ping(ctx)
}

func (p *pgDB) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return p.conn.BeginTx(ctx, opts)
}

func (p *pgDB) Close() {
	log.Debugf("close db ")
	p.conn.Close()
}

type client struct {
	db PgxPoolIface
}

func (c *client) DB() PgxPoolIface {
	return c.db
}

func (c *client) Close() error {
	c.db.Close()
	return nil
}
