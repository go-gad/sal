package foobar

type Name string
type Desc string

type CreateAuthorReq struct {
	Name
	Desc
}

func (cr *CreateAuthorReq) Query() string {
	return `INSERT INTO authors (name, desc, created_at) VALUES(@name, @desc, now()) RETURNING id, created_at`
}
