package copier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRegularCopy(t *testing.T) {
	type str struct {
		Str     string
		Array   []string
		Ints    []int
		Integer int
		Empty   string
		Nested  struct {
			Str []string
		}
	}
	base := str{
		Str:     "hello string",
		Array:   []string{"user", "role", "admin"},
		Ints:    []int{1, 2, 3, 4, 5},
		Integer: 100000000,
		Nested: struct {
			Str []string
		}{
			Str: []string{"hello world", "oh my god"},
		},
	}
	var new str
	err := Copy(&new, &base)

	assert.NoError(t, err)
	assert.EqualValues(t, base, new)
}
func TestIgnoreEmpty(t *testing.T) {
	type str struct {
		Str     string
		Array   []string
		Ints    []int
		Integer int
		Empty   string
		Nested  struct {
			Str []string
		}
	}
	base := str{
		Str:     "hello string",
		Array:   []string{"user", "role", "admin"},
		Ints:    []int{1, 2, 3, 4, 5},
		Integer: 100000000,
		Nested: struct {
			Str []string
		}{
			Str: []string{"hello world", "oh my god"},
		},
	}
	new := str{Empty: "256256"}
	err := Copy(&new, &base, WithIgnoreEmptyFields)

	assert.NoError(t, err)
	assert.Equal(t, base.Array, new.Array)
	assert.Equal(t, "256256", new.Empty)
	assert.Empty(t, base.Empty)
}
func TestPrimitiveToString(t *testing.T) {
	str := struct {
		Id    primitive.ObjectID
		Ids   []primitive.ObjectID
		Empty primitive.ObjectID
	}{
		Id:  primitive.NewObjectID(),
		Ids: []primitive.ObjectID{primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()},
	}
	new := struct {
		Id    string
		Ids   []string
		Empty string
	}{Empty: "not empty"}

	err := Copy(&new, &str, WithIgnoreEmptyFields, WithPrimitiveToStringConverter)

	assert.NoError(t, err)

	assert.Equal(t, str.Id.Hex(), new.Id)
	assert.Equal(t, str.Ids[0].Hex(), new.Ids[0])
	assert.Equal(t, str.Ids[1].Hex(), new.Ids[1])
	assert.Equal(t, str.Ids[2].Hex(), new.Ids[2])
	assert.Equal(t, "not empty", new.Empty)
	assert.Empty(t, str.Empty)

	str = struct {
		Id    primitive.ObjectID
		Ids   []primitive.ObjectID
		Empty primitive.ObjectID
	}{}

	new.Empty = ""
	err = Copy(&str, &new, WithIgnoreEmptyFields, WithPrimitiveToStringConverter)

	assert.NoError(t, err)

	assert.Equal(t, new.Id, str.Id.Hex())
	assert.Equal(t, new.Ids[0], str.Ids[0].Hex())
	assert.Equal(t, new.Ids[1], str.Ids[1].Hex())
	assert.Equal(t, new.Ids[2], str.Ids[2].Hex())
	assert.Equal(t, primitive.ObjectID{}, str.Empty)
}
func TestPrimitiveError(t *testing.T) {
	str := struct {
		Str  string
		StrA []string
	}{
		Str:  "not a primitive id",
		StrA: []string{"not", "a", "primitive", "id"},
	}
	new := struct {
		Str  primitive.ObjectID
		StrA []primitive.ObjectID
	}{}
	err := Copy(&new, &str, WithPrimitiveToStringConverter)
	assert.Error(t, err)
}
