# sal [![Build Status](https://travis-ci.org/go-gad/sal.svg?branch=master)](https://travis-ci.org/go-gad/sal) [![GoDoc](https://godoc.org/github.com/go-gad/sal?status.svg)](https://godoc.org/github.com/go-gad/sal)

Generator client to the database on the Golang based on interface.

## Install

```
go get -u github.com/go-gad/sal/...
```

## Usage

Read article https://medium.com/@zaurio/generator-the-client-to-sql-database-on-golang-dccfeb4641c3

```sh
 salgen -h
Usage:
    salgen [options...] <import_path> <interface_name>

Example:
    salgen -destination=./client.go -package=github.com/go-gad/sal/examples/profile/storage github.com/go-gad/sal/examples/profile/storage Store

  <import_path>
        describes the complete package path where the interface is located.
  <interface_name>
        indicates the interface name itself.

Options:
  -build_flags string
        Additional flags for go build.
  -destination string
        Output file; defaults to stdout.
  -package string
        The full import path of the library for the generated implementation
```

With go generate:
```go
//go:generate salgen -destination=./client.go -package=github.com/go-gad/sal/examples/profile/storage github.com/go-gad/sal/examples/profile/storage Store
type Store interface {
	CreateUser(ctx context.Context, req CreateUserReq) (*CreateUserResp, error)
}

type CreateUserReq struct {
	Name  string `sql:"name"`
	Email string `sql:"email"`
}

func (r CreateUserReq) Query() string {
	return `INSERT INTO users(name, email, created_at) VALUES(@name, @email, now()) RETURNING id, created_at`
}

type CreateUserResp struct {
	ID        int64     `sql:"id"`
	CreatedAt time.Time `sql:"created_at"`
}
```

In your project run command
```
$ go generate ./...
```
File `client.go` will be generated.

## Command line args and options

* flag `-destination` determines in which file the generated code will be written.
* flag `-package` is the full import path of the library for the generated implementation.
* first arg describes the complete package path where the interface is located.
* second indicates the interface name itself.

## Possible definitions of methods

```go
type Store interface {
	CreateAuthor(ctx context.Context, req CreateAuthorReq) (CreateAuthorResp, error)
	GetAuthors(ctx context.Context, req GetAuthorsReq) ([]*GetAuthorsResp, error)
	UpdateAuthor(ctx context.Context, req *UpdateAuthorReq) error
	DeleteAuthors(ctx context.Context, req *DeleteAuthorsReq) (sql.Result, error)
}
```

* The number of arguments is always strictly two.
* The first argument is the context.
* The second argument contains the data to bind the variables and defines the query string.
* The first output parameter can be an object, an array of objects, `sql.Result` or missing.
* Last output parameter is always an error.

The second argument expects a parameter with a base type of  `struct` (or a pointer to a `struct`). The parameter must satisfy the following interface:
```go
type Queryer interface {
	Query() string
}
```
The string returned by method `Query` is used as a SQL query.

## Prepared statements

The generated code supports prepared statements.
Prepared statements are cached.
After the first preparation of the statement, it is placed in the cache.
The `database/sql` library itself ensures
that prepared statements are transparently applied to the desired database connection,
including the processing of closed connections.
In turn, the `go-gad/sal` library cares about reusing the prepared statement
in the context of a transaction.
When the prepared statement is executed, the arguments are passed using variable binding,
transparently to the developer.

## Map structs to response messages

The `go-gad/sal` library cares about linking database response lines with response structures, table columns with structure fields:
```go
type GetRubricsReq struct {}
func (r GetRubricReq) Query() string {
	return `SELECT * FROM rubrics`
}

type Rubric struct {
	ID       int64     `sql:"id"`
	CreateAt time.Time `sql:"created_at"`
	Title    string    `sql:"title"`
}
type GetRubricsResp []*Rubric

type Store interface {
	GetRubrics(ctx context.Context, req GetRubricsReq) (GetRubricsResp, error)
}
```
And if the database response is:
```sql
dev > SELECT * FROM rubrics;
 id |       created_at        | title
----+-------------------------+-------
  1 | 2012-03-13 11:17:23.609 | Tech
  2 | 2015-07-21 18:05:43.412 | Style
(2 rows)
```
Then the `GetRubricsResp` list will return to us,
elements of which will be pointers to `Rubric`,
where the fields are filled with values from the columns that correspond to the names of the tags.

## Value `in` list

```go
type GetIDsReq struct {
	IDs  pq.StringArray `sql:"ids"`
}

func (r *GetIDsReq) Query() string {
	return `SELECT * FROM rubrics WHERE id = ANY(@ids)`
}
```

## Multiple insert/update

```go
type AddBooksToShelfReq struct {
	ShelfID     int64 `sql:"shelf_id"`
	BookID      pq.Int64Array `sql:"book_ids"`
}

func (c *AddBooksToShelfReq) Query() string {
	return `INSERT INTO shelf (shelf_id, book_id)
		SELECT @shelf_id, unnest(@book_ids);`
}
```

## Non-standard data types

The `database/sql` package provides support for basic data types (strings, numbers).
In order to handle data types such as an `array` or `json` in a request or response.

```go
type DeleteAuthrosReq struct {
	Tags []int64 `sql:"tags"`
}

func (r *DeleteAuthorsReq) ProcessRow(rowMap sal.RowMap) {
	rowMap.Set("tags", pq.Array(r.Tags))
}

func (r *DeleteAuthorsReq) Query() string {
	return `DELETE FROM authors WHERE tags=ANY(@tags::UUID[])`
}
```

The same can be done with `sql` package predefined types

```go
type DeleteAuthrosReq struct {
	Tags sql.Int64Array `sql:"tags"`
}

func (r *DeleteAuthorsReq) Query() string {
	return `DELETE FROM authors WHERE tags=ANY(@tags::UUID[])`
}
```

##  Nested types

Here we don't use struct tages because we map it in ProcessRow func to prevent misunderstanding for the same field names (`id` and `name` for `Book` and `Author` types)
```go
type Author struct {
    ID   int64 
    Name string
}

type Book struct {
    ID   int64
    Name string
    Description string
    Author Author
}
type CreateBookReq struct {
    Book Book
}

func (r *CreateBookReq) ProcessRow(rowMap sal.RowMap) {
	rowMap.Set("author_id", r.Book.Author.ID)
	rowMap.Set("book_id",   r.Book.ID)
	rowMap.Set("book_name", r.Book.Name)
	rowMap.Set("book_descriprion", r.Book.Description)
}

func (r *CreateBookReq) Query() string {
	return `INSERT INTO books (id, author_id, name, description)
	VALUES (@book_id, @author_id, @book_name, @book_description)`
}
```

## Transactions

To support transactions, the interface (Store) must be extended with the following methods:
```go
type Store interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Store, error)
	sal.Txer
	...
```

The implementation of the methods will be generated. The `BeginTx` method uses the connection from the current `sal.QueryHandler` object and opens the transaction `db.BeginTx(...)`; returns a new implementation object of the interface `Store`, but uses the resulting `*sql.Tx` object as a handle.

```go
tx, err := client.BeginTx(ctx, nil)
_, err = tx.CreateAuthor(ctx, req1)
err = tx.UpdateAuthor(ctx, &req2)
err = tx.Tx().Commit(ctx)
```

## Middleware

Hooks are provided for embedding tools.

When the hooks are executed, the context is filled with service keys with the following values:
* `ctx.Value(sal.ContextKeyTxOpened)`, boolean indicates whether the method is called in the context of a transaction or not.
* `ctx.Value(sal.ContextKeyOperationType)`, the string value of the operation type, `"QueryRow"`, `"Query"`, `"Exec"`, `"Commit"`, etc.
* `ctx.Value(sal.ContextKeyMethodName)`, the string value of the interface method, for example, `"GetAuthors"`.

As arguments, the `BeforeQueryFunc` hook takes the sql string of the query and the argument `req` of the custom query method. The `FinalizerFunc` hook takes the variable` err` as an argument.

```go
	beforeHook := func(ctx context.Context, query string, req interface{}) (context.Context, sal.FinalizerFunc) {
		start := time.Now()
		return ctx, func(ctx context.Context, err error) {
			log.Printf(
				"%q > Opeartion %q: %q with req %#v took [%v] inTx[%v] Error: %+v",
				ctx.Value(sal.ContextKeyMethodName),
				ctx.Value(sal.ContextKeyOperationType),
				query,
				req,
				time.Since(start),
				ctx.Value(sal.ContextKeyTxOpened),
				err,
			)
		}
	}

	client := NewStore(db, sal.BeforeQuery(beforeHook))
```

## Limitations

Currently support only PostgreSQL.
