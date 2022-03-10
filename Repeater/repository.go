package Repeater

import (
	"github.com/jackc/pgx"
)

type RepoPostgres struct {
	db *pgx.ConnPool
}

func NewRepoPostgres(db *pgx.ConnPool) *RepoPostgres {
	return &RepoPostgres{
		db: db,
	}
}

func (r RepoPostgres) GetRequests() ([]Request, error) {
	var reqs []Request
	rows, err := r.db.Query(`SELECT "id", "method", "scheme", "host", "path", "add_time" FROM "request"`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var req Request
		err = rows.Scan(
			&req.Id,
			&req.Method,
			&req.Scheme,
			&req.Host,
			&req.Path,
			&req.AddTime,
		)
		if err != nil {
			return nil, err
		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}

func (r RepoPostgres) GetRequest(id int) (Request, error) {
	var req Request
	var header string

	err := r.db.QueryRow(`SELECT "id", "method", "scheme", "host", "path", "header", "body", "add_time"
		FROM "request" WHERE "id" = $1;`, id).Scan(
		&req.Id,
		&req.Method,
		&req.Scheme,
		&req.Host,
		&req.Path,
		&header,
		&req.Body,
		&req.AddTime,
	)

	req.Header = strToHeader(header)

	return req, err
}
