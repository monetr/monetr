package migrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		innerQuery := `
			SELECT bank_account.bank_account_id,
				   bank_account.account_id,
				   bank_account.current_balance             AS current,
				   bank_account.available_balance           AS available,
				   bank_account.available_balance - SUM(COALESCE(expense.current_amount, 0)) -
				   SUM(COALESCE(goal.current_amount, 0))    AS safe,
				   SUM(COALESCE(expense.current_amount, 0)) AS expenses,
				   SUM(COALESCE(goal.current_amount, 0))    AS goals
			FROM bank_accounts AS bank_account
			LEFT JOIN spending AS expense 
				ON expense.bank_account_id = bank_account.bank_account_id AND
				   expense.account_id = bank_account.account_id AND
				   expense.spending_type = 0 /* 0 means its an expense */
			LEFT JOIN spending AS goal 
				ON goal.bank_account_id = bank_account.bank_account_id AND
				   goal.account_id = bank_account.account_id AND
				   goal.spending_type = 1 /* 1 means its a goal */
			GROUP BY bank_account.bank_account_id, bank_account.account_id
		`
		innerQuery = strings.ReplaceAll(innerQuery, "\n", " ")
		innerQuery = strings.ReplaceAll(innerQuery, "\t", "")
		for i := 0; i < 10; i++ {
			innerQuery = strings.ReplaceAll(innerQuery, "  ", " ")
		}

		var createView string
		switch db.Dialect().Name() {
		case dialect.PG, dialect.MySQL:
			createView = fmt.Sprintf(`CREATE VIEW balances AS (%s);`, strings.TrimSpace(innerQuery))
		case dialect.SQLite:
			createView = fmt.Sprintf(`CREATE VIEW balances AS %s;`, strings.TrimSpace(innerQuery))
		}

		_, err := db.ExecContext(ctx, createView)
		return errors.Wrapf(err, "failed to create balances view")
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `DROP VIEW balances;`)
		return errors.Wrap(err, "failed to drop balances view")
	})
}
