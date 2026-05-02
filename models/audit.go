package models

import "time"

type AuditLog struct {
	ID        int64
	UserID    int
	Action    string
	Timestamp time.Time
}
