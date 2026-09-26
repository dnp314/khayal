package utils

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

func StartReturnTimeChecker(db *pgx.Conn) {

	for {
		log.Println("Checking...")
		time.Sleep(1 * time.Minute)
		_, err := db.Exec(
			context.TODO(),
			`update questions
			 set answered = false
			 where return_at <= current_timestamp
			 and answered = true
			`,
		)
		if err != nil {
			return
		}

	}

}
