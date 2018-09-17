package sal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryArgs(t *testing.T) {
	//t.Skip("todo")
	var tt = []struct {
		QueryNamed string
		QueryPg    string
		NamedArgs  []string
	}{
		{
			QueryNamed: `UPDATE authors SET name=@name, desc=@desc WHERE id=@id`,
			QueryPg:    `UPDATE authors SET name=$1, desc=$2 WHERE id=$3`,
			NamedArgs:  []string{"name", "desc", "id"},
		}, {
			QueryNamed: `SELECT id, created_at, name, desc FROM authors WHERE id>@id`,
			QueryPg:    `SELECT id, created_at, name, desc FROM authors WHERE id>$1`,
			NamedArgs:  []string{"id"},
		},
	}
	for _, tc := range tt {
		query, args := QueryArgs(tc.QueryNamed)
		assert.Equal(t, tc.QueryPg, query)
		assert.Equal(t, tc.NamedArgs, args)
	}
}

func TestMapIndex_NextVal(t *testing.T) {
	ind := make(mapIndex)
	assert.Equal(t, 0, ind.NextVal("foo"))
	assert.Equal(t, 1, ind.NextVal("foo"))
	assert.Equal(t, 2, ind.NextVal("foo"))
	assert.Equal(t, 0, ind.NextVal("bar"))
	assert.Equal(t, 3, ind.NextVal("foo"))
	assert.Equal(t, 1, ind.NextVal("bar"))
}
