package rpckit_db

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type DB struct {
DSN string
NoCache bool
Logger
 // OnNotice is a callback function called when a notice response is received.
    OnNotice NoticeHandler

    // OnNotification is a callback function called when a notification from the LISTEN/NOTIFY system is received.
    OnNotification NotificationHandler

}

// Connect makes connect to DB and returns ins handle
func (mig *Migrator) Connect(dsn string, opts ...Option) (*pgx.Conn, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	//	config.Logger = logrusadapter.NewLogger(l)
//https://github.com/jackc/pgx/tree/master/log/zapadapter
	config.OnNotice = func(c *pgconn.PgConn, n *pgconn.Notice) {
		mig.ProcessNotice(n.Code, n.Message, n.Detail)
	}

	// TODO: statement_cache_mode = "describe"
	config.BuildStatementCache = nil // disable stmt cache for `reinit pgmig`
	ctx := context.Background()
	return pgx.ConnectConfig(ctx, config)
}
