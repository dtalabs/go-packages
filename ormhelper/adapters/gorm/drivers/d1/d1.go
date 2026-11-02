package d1

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var ErrTransactionsUnsupported = errors.New("d1: transactions spanning HTTP requests are unsupported")

type Config struct {
	AccountID  string
	DatabaseID string
	APIToken   string
	HTTPClient *http.Client
}

func Open(config Config) gorm.Dialector {
	return &d1Dialector{config: config}
}

type d1Dialector struct {
	gorm.Dialector
	config Config
}

func (d *d1Dialector) Name() string { return "d1" }
func (d *d1Dialector) Initialize(db *gorm.DB) error {
	c := d.config
	if strings.TrimSpace(c.AccountID) == "" || strings.TrimSpace(c.DatabaseID) == "" || strings.TrimSpace(c.APIToken) == "" {
		return errors.New("d1: account ID, database ID, and API token are required")
	}
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	conn := sql.OpenDB(&connector{client: client, token: c.APIToken,
		endpoint: "https://api.cloudflare.com/client/v4/accounts/" + url.PathEscape(c.AccountID) + "/d1/database/" + url.PathEscape(c.DatabaseID) + "/raw"})
	db.Config.SkipDefaultTransaction = true
	d.Dialector = sqlite.New(sqlite.Config{Conn: conn})
	if err := d.Dialector.Initialize(db); err != nil {
		conn.Close()
		return err
	}
	return nil
}

func (*d1Dialector) SavePoint(*gorm.DB, string) error  { return ErrTransactionsUnsupported }
func (*d1Dialector) RollbackTo(*gorm.DB, string) error { return ErrTransactionsUnsupported }

type connector struct {
	client          *http.Client
	endpoint, token string
}

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &connection{c}, nil
}
func (c *connector) Driver() driver.Driver { return sqlDriver{} }

type sqlDriver struct{}

func (sqlDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("d1: use d1.Open with a Config")
}

type connection struct{ *connector }

func (*connection) Close() error              { return nil }
func (*connection) Begin() (driver.Tx, error) { return nil, ErrTransactionsUnsupported }
func (*connection) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return nil, ErrTransactionsUnsupported
}
func (c *connection) Prepare(query string) (driver.Stmt, error) { return &statement{c, query}, nil }
func (c *connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return c.Prepare(query)
}
func (c *connection) Ping(ctx context.Context) error {
	_, err := c.execute(ctx, "SELECT 1", nil)
	return err
}
func (c *connection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	r, err := c.execute(ctx, query, args)
	if err != nil {
		return nil, err
	}
	return r.Meta, nil
}
func (c *connection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	r, err := c.execute(ctx, query, args)
	if err != nil {
		return nil, err
	}
	return &rows{columns: r.Results.Columns, values: r.Results.Rows}, nil
}

type statement struct {
	conn  *connection
	query string
}

func (*statement) Close() error  { return nil }
func (*statement) NumInput() int { return -1 }
func (s *statement) Exec(args []driver.Value) (driver.Result, error) {
	return s.ExecContext(context.Background(), named(args))
}
func (s *statement) Query(args []driver.Value) (driver.Rows, error) {
	return s.QueryContext(context.Background(), named(args))
}
func (s *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return s.conn.ExecContext(ctx, s.query, args)
}
func (s *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return s.conn.QueryContext(ctx, s.query, args)
}
func named(args []driver.Value) []driver.NamedValue {
	result := make([]driver.NamedValue, len(args))
	for i, v := range args {
		result[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	return result
}

var _ gorm.Dialector = (*d1Dialector)(nil)
var _ driver.Connector = (*connector)(nil)
var _ driver.ConnBeginTx = (*connection)(nil)
var _ driver.ExecerContext = (*connection)(nil)
var _ driver.QueryerContext = (*connection)(nil)
var _ driver.ConnPrepareContext = (*connection)(nil)
var _ driver.Pinger = (*connection)(nil)
var _ driver.StmtExecContext = (*statement)(nil)
var _ driver.StmtQueryContext = (*statement)(nil)
