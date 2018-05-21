// It is an example of code that should be generated.
package bookstore

import (
	"context"
	"database/sql"
)

type salStoreClient struct {
	DB *sql.DB
}

func New(db *sql.DB) *salStoreClient {
	return &salStoreClient{DB: db}
}

func (s *salStoreClient) CreateAuthor(ctx context.Context, req *CreateAuthorReq) (*CreateAuthorResp, error) {
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

	var resp CreateAuthorResp
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

func (s *salStoreClient) GetAuthors(ctx context.Context, req *GetAuthorsReq) ([]*GetAuthorsResp, error) {
	args := []interface{}{
		&req,
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

	var list = make([]*GetAuthorsResp, 0)

	for rows.Next() {
		var resp GetAuthorsResp
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

type keysDest map[string]interface{}
