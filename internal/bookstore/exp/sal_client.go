// It is an example of code that should be generated.
package exp

import (
	"context"
	"database/sql"

	"github.com/go-gad/sal/internal/bookstore"
)

type salStoreClient struct {
	DB *sql.DB
}

func NewStoreClient(db *sql.DB) *salStoreClient {
	return &salStoreClient{DB: db}
}

func (s *salStoreClient) CreateAuthor(ctx context.Context, req *bookstore.CreateAuthorReq) (*bookstore.CreateAuthorResp, error) {
	args := []interface{}{
		&req.Name,
		&req.Desc,
	}
	rows, err := s.DB.Query(req.Query(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// sql.DB.QueryRow
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}

	var resp bookstore.CreateAuthorResp
	var mm = make(keysDest)
	mm["id"] = &resp.Id
	mm["created_at"] = &resp.CreatedAt
	var dest = make([]interface{}, 0, len(mm))
	for _, v := range cols {
		if intr, ok := mm[v]; ok {
			dest = append(dest, intr)
		}
	}

	if err = rows.Scan(dest...); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *salStoreClient) GetAuthors(ctx context.Context, req *bookstore.GetAuthorsReq) ([]*bookstore.GetAuthorsResp, error) {
	args := []interface{}{
		&req.Id,
	}
	rows, err := s.DB.Query(req.Query(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// sql.DB.Query

	var list = make([]*bookstore.GetAuthorsResp, 0)

	for rows.Next() {
		var resp bookstore.GetAuthorsResp
		var mm = make(keysDest)
		mm["id"] = &resp.Id
		mm["created_at"] = &resp.CreatedAt
		mm["name"] = &resp.Name
		mm["desc"] = &resp.Desc
		var dest = make([]interface{}, 0, len(mm))
		for _, v := range cols {
			if intr, ok := mm[v]; ok {
				dest = append(dest, intr)
			}
		}

		if err = rows.Scan(dest...); err != nil {
			return nil, err
		}
		list = append(list, &resp)
	}

	return list, nil
}

func (s *salStoreClient) UpdateAuthor(ctx context.Context, req *bookstore.UpdateAuthorReq) error {
	// sql.DB.Exec
	args := []interface{}{
		&req.Name,
		&req.Desc,
		&req.Id,
	}
	_, err := s.DB.Exec(req.Query(), args...)
	if err != nil {
		return err
	}
	return nil
}

type keysDest map[string]interface{}