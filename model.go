package unitofwork

import "time"

type User struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	CreatedAt time.Time
}

type AuditLogs struct {
	ID        int
	UserID    int
	Action    string
	Timestamp time.Time
}
