package repositories

import (
	"database/sql"

	"github.com/snookish/unit-of-work/repositories/auditrepo"
	"github.com/snookish/unit-of-work/repositories/userrepo"
)

type Repos struct {
	Users     userrepo.Repository
	AuditLogs auditrepo.Repository
}

func NewRepos(tx *sql.Tx) *Repos {
	return &Repos{
		Users:     userrepo.NewRepository(tx),
		AuditLogs: auditrepo.NewRepository(tx),
	}
}
