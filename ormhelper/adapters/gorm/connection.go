package gorm

import (
	"errors"
	"net/url"
	"strings"

	"github.com/dtalabs/go-packages/ormhelper/adapters/gorm/drivers/d1"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Open(databaseURL string, options ...gorm.Option) (*gorm.DB, error) {
	databaseURL = strings.TrimSpace(databaseURL)
	scheme, _, ok := strings.Cut(databaseURL, "://")

	if !ok {
		return nil, errors.New("gorm: database URL must use scheme://")
	}

	if scheme == "postgresql" || scheme == "postgres" {
		return OpenPostgres(databaseURL, options...)
	} else if scheme == "sqlite" {
		return OpenSQLite(databaseURL, options...)
	} else if scheme == "d1" {
		return OpenD1(databaseURL, options...)
	}

	return nil, errors.New("gorm: unsupported database URL scheme")
}

func OpenPostgres(databaseURL string, options ...gorm.Option) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(databaseURL), options...)
}

func OpenSQLite(databaseURL string, options ...gorm.Option) (*gorm.DB, error) {
	dsn := strings.TrimPrefix(databaseURL, "sqlite://")
	if dsn == "" {
		return nil, errors.New("gorm: SQLite database path is required")
	}
	return gorm.Open(sqlite.Open(dsn), options...)
}

func OpenD1(databaseURL string, options ...gorm.Option) (*gorm.DB, error) {
	credentials, databaseID, hasDatabase := strings.Cut(strings.TrimPrefix(databaseURL, "d1://"), "/")
	accountID, apiToken, hasToken := strings.Cut(credentials, ":")
	if !hasDatabase || !hasToken || strings.ContainsAny(databaseID, "/?#") || strings.ContainsAny(credentials, "@?#") {
		return nil, errors.New("gorm: D1 URL must use d1://<account-id>:<api-token>/<database-id>")
	}
	accountID, accountErr := url.PathUnescape(accountID)
	apiToken, tokenErr := url.PathUnescape(apiToken)
	databaseID, databaseErr := url.PathUnescape(databaseID)
	if accountErr != nil || tokenErr != nil || databaseErr != nil || strings.TrimSpace(accountID) == "" || strings.TrimSpace(apiToken) == "" || strings.TrimSpace(databaseID) == "" || strings.Contains(databaseID, "/") {
		return nil, errors.New("gorm: invalid D1 account ID, API token, or database ID")
	}
	return gorm.Open(d1.Open(d1.Config{
		AccountID:  accountID,
		DatabaseID: databaseID,
		APIToken:   apiToken,
	}), options...)
}
