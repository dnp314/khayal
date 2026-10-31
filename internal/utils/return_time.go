package utils

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

func StartReturnTimeChecker(db *pgx.Conn) {

	for {
		slog.Debug("checking question return times")
		time.Sleep(1 * time.Minute)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, err := db.Exec(
			ctx,
			`update questions
			 set answered = false
			 where return_at <= current_timestamp
			 and answered = true
			`,
		)
		cancel()
		if err != nil {
			slog.Error("update question return times", "error", err)
		}

	}

}
