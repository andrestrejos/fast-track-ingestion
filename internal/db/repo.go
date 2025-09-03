package db

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/andrestrejos/fast-track-ingestion/internal/domain"

	mysql "github.com/go-sql-driver/mysql"
)

var ErrDuplicate = errors.New("duplicate entry")

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS payment_events (
			user_id INT NOT NULL,
			payment_id INT NOT NULL PRIMARY KEY,
			deposit_amount INT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS skipped_messages (
			user_id INT NOT NULL,
			payment_id INT NOT NULL PRIMARY KEY,
			deposit_amount INT NOT NULL
		)`,
	}
	for _, s := range stmts {
		if _, err := r.db.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) InsertPayment(ctx context.Context, e domain.PaymentEvent) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO payment_events(user_id, payment_id, deposit_amount) VALUES (?,?,?)`,
		e.UserID, e.PaymentID, e.DepositAmount,
	)
	if err != nil {
		// 1) detectar por tipo de error
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return ErrDuplicate
		}
		// 2) fallback: inspeccionar el mensaje
		if strings.Contains(strings.ToLower(err.Error()), "duplicate entry") {
			return ErrDuplicate
		}
		return err
	}
	return nil
}

func (r *Repo) InsertSkipped(ctx context.Context, e domain.PaymentEvent) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO skipped_messages(user_id, payment_id, deposit_amount) VALUES (?,?,?)`,
		e.UserID, e.PaymentID, e.DepositAmount,
	)
	return err
}
