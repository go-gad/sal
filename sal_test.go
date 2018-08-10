package sal_test

import (
	"testing"

	"github.com/go-gad/sal"
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
		query, args := sal.QueryArgs(tc.QueryNamed)
		assert.Equal(t, tc.QueryPg, query)
		assert.Equal(t, tc.NamedArgs, args)
	}
}

func TestUnderscoreToCamelCase(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		param    string
		expected string
	}{
		{"a_b_c", "ABC"},
		{"my_func", "MyFunc"},
		{"1ab_cd", "1abCd"},
	}
	for _, test := range tests {
		actual := sal.UnderscoreToCamelCase(test.param)
		if actual != test.expected {
			t.Errorf("Expected UnderscoreToCamelCase(%q) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}

func TestCamelCaseToUnderscore(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		param    string
		expected string
	}{
		{"MyFunc", "my_func"},
		{"ABC", "a_b_c"},
		{"1B", "1_b"},
		{"foo_bar", "foo_bar"},
		{"FooV2Bar", "foo_v2_bar"},
	}
	for _, test := range tests {
		actual := sal.CamelCaseToUnderscore(test.param)
		if actual != test.expected {
			t.Errorf("Expected CamelCaseToUnderscore(%q) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}
