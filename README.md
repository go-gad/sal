# sal
Toolkit for working with database

## Install

```
go get -u github.com/go-gad/sal/...
```

## Usage

Define interface
```go
type StoreClient interface {
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
//go:generate salgen -destination=./actsal/sal_client.go -package=actsal github.com/go-gad/sal/examples/bookstore1 StoreClient
```

- flag `destination` describes the output file.
- flag `package` describes the package of the generated code.
– first argument describes the import path of package that contains the interfaces.
- second argument describes the list of interfaces that should be used to generate clients.

Run `go generate ./...`. Your client based on interface would be generated. You can used it like:
```go
db, err := sql.Open("postgres", connStr)

client := NewStoreClient(db)
req := bookstore1.CreateAuthorReq{Name: "foo", Desc: "Bar"}
resp, err := client.CreateAuthor(context.Background(), req)

```

See `examples/bookstore1`.