package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/snookish/unit-of-work/services/usersvc"
	"github.com/snookish/unit-of-work/uow"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/uow_demo?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	unit := uow.NewUnitOfWork(db)
	svc := usersvc.NewUserService(unit)

	user, err := svc.CreateUser(context.Background(), "Hello", "World", "hello@world.com")
	if err != nil {
		log.Fatalf("create user: %v", err)
	}

	fmt.Printf("created user id=%d email=%s\n", user.ID, user.Email)

	if err := svc.UpdateUserEmail(context.Background(), user.ID, "alice@newdomain.com"); err != nil {
		log.Fatalf("update email: %v", err)
	}

	fmt.Printf("updated email for user id=%d\n", user.ID)
}
