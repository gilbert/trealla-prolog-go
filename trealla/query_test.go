package trealla_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/trealla-prolog/go/trealla"
)

func TestQuery(t *testing.T) {
	pl, err := trealla.New(trealla.WithPreopenDir("testdata"))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		want trealla.Answer
		err  error
	}{
		{
			name: "true/0",
			want: trealla.Answer{
				Query:   `true.`,
				Answers: []trealla.Solution{{}},
			},
		},
		{
			name: "member/2",
			want: trealla.Answer{
				Query: `member(X, [1,foo(bar),4.2,"baz",'boop', [q, '"'], '\\', '\n']).`,
				Answers: []trealla.Solution{
					{"X": int64(1)},
					{"X": trealla.Compound{Functor: "foo", Args: []trealla.Term{"bar"}}},
					{"X": 4.2},
					{"X": "baz"},
					{"X": "boop"},
					{"X": []trealla.Term{"q", `"`}},
					{"X": `\`},
					{"X": "\n"},
				},
			},
		},
		{
			name: "false/0",
			want: trealla.Answer{
				Query: `false.`,
			},
			err: trealla.ErrFailure,
		},
		{
			name: "tak",
			want: trealla.Answer{
				Query:   "consult('testdata/tak'), run",
				Answers: []trealla.Solution{{}},
				Output:  "'<https://josd.github.io/eye/ns#tak>'([34,13,8],13).\n",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			ans, err := pl.Query(ctx, tc.want.Query)
			if tc.err == nil && err != nil {
				t.Fatal(err)
			} else if tc.err != nil && !errors.Is(err, tc.err) {
				t.Error("unexpected error:", err)
			}
			if !reflect.DeepEqual(ans, tc.want) {
				t.Errorf("bad answer. want: %#v got: %#v", tc.want, ans)
			}
		})
	}

}

func TestThrow(t *testing.T) {
	pl, err := trealla.New(trealla.WithPreopenDir("testdata"))
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = pl.Query(ctx, `throw(ball).`)

	var ex trealla.ErrThrow
	if !errors.As(err, &ex) {
		t.Fatal("unexpected error:", err, "want ErrThrow")
	}

	if ex.Ball != "ball" {
		t.Error(`unexpected error value. want: "ball" got:`, ex.Ball)
	}
}

func TestSyntaxError(t *testing.T) {
	pl, err := trealla.New(trealla.WithPreopenDir("testdata"))
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	_, err = pl.Query(ctx, `hello(`)

	var ex trealla.ErrThrow
	if !errors.As(err, &ex) {
		t.Fatal("unexpected error:", err, "want ErrThrow")
	}
	want := trealla.Compound{Functor: "error", Args: []trealla.Term{
		trealla.Compound{
			Functor: "syntax_error",
			Args:    []trealla.Term{"mismatched_parens_or_brackets_or_braces"},
		},
		trealla.Compound{
			Functor: "/",
			Args:    []trealla.Term{"read_term_from_chars", int64(3)},
		},
	}}

	if !reflect.DeepEqual(ex.Ball, want) {
		t.Error(`unexpected error value. want:`, want, `got:`, ex.Ball)
	}
}