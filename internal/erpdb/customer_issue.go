package erpdb

import (
	"context"
	"database/sql"
	"strings"
)

func (r *Reader) CustomerIssueSourceExists(ctx context.Context, marker string) (bool, error) {
	marker = strings.TrimSpace(marker)
	if marker == "" {
		return false, nil
	}

	var name string
	err := r.db.QueryRowContext(ctx, `
		SELECT name
		FROM `+"`tabDelivery Note`"+`
		WHERE LOCATE(?, COALESCE(remarks, '')) > 0
			AND COALESCE(docstatus, 0) < 2
		LIMIT 1`,
		marker,
	).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(name) != "", nil
}
