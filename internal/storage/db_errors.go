package storage

import (
	"database/sql"
	"errors"
	"strings"
)

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func isMissingTable(err error, table string) bool {
	if err == nil || table == "" {
		return false
	}
	msg := strings.ToLower(err.Error())
	t := strings.ToLower(table)
	return strings.Contains(msg, "catalog error") && strings.Contains(msg, "table") && strings.Contains(msg, t) && strings.Contains(msg, "does not exist")
}
