package models

import "time"

type AuditLog struct {
	ID        int64
	UserID    int64
	Action    string
	Timestamp time.Time
}
