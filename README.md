# sal
Toolkit for working with database

## Install

```
go get -u github.com/go-gad/sal/...
```

## Usage

Define interface
```go
type Store interface {
	CreateAuthor(context.Context, CreateAuthorReq) (*CreateAuthorResp, error)
}
```
First argument should be `context.Context`. Second is your request struct. 
First output value is struct that describes a response. Second output value is an error.

Define request and response structs:
```go
type CreateAuthorReq struct {
	Name string
	Desc string
}

type CreateAuthorResp struct {
	ID        int64
	CreatedAt time.Time
}
``` 
Field names are used to match named arguments for request and columns of output for response.

Define the sql query as method on request struct that should be associated with the call:
```go
func (cr *CreateAuthorReq) Query() string {
	return `INSERT INTO authors (Name, Desc, CreatedAt) VALUES(@Name, @Desc, now()) RETURNING ID, CreatedAt`
}
``` 

Put `go generate` instruction for your interface:
```go
//go:generate salgen -destination=./actsal/sal_client.go -package=actsal github.com/go-gad/sal/examples/bookstore1 Store
```

- flag `destination` describes the output file.
- flag `package` describes the package of the generated code.
- first argument describes the import path of package that contains the interfaces.
- second argument describes the list of interfaces that should be used to generate clients.

Run `go generate ./...`. Your client based on interface would be generated. You can used it like:
```go
db, err := sql.Open("postgres", connStr)

client := NewStore(db)
req := bookstore1.CreateAuthorReq{Name: "foo", Desc: "Bar"}
resp, err := client.CreateAuthor(context.Background(), req)

```

See `examples/bookstore1`.

### Use custom datatypes

If you want to use custom database types, like Array, you can use ProcessRower interface for that. 
Your request or response should contain method `ProcessRow(rowMap sal.RowMap)`.

```go
func (r GetAuthorsReq) ProcessRow(rowMap sal.RowMap) {
	rowMap["tags"] = pq.Array(r.Tags)
}

func (r *GetAuthorsResp) ProcessRow(rowMap sal.RowMap) {
	rowMap["tags"] = pq.Array(&r.Tags)
}
``` 

### Transaction

To open transaction use:
```go
tx, err := actsal.NewStoreManager().Begin(client)
```
`tx` is a Store implementation that contains `*sql.Tx` handler instead of `*sql.DB`.

To commit or rollback opened transaction use:
```go
err = actsal.NewStoreManager().Commit(tx)
```
