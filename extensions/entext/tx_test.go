package entext

import (
	"context"
	"fmt"
	"github.com/shanbay/gobay/testdata/ent"

	//"git.17bdc.com/words/backend/words-learning/app"
	//schema "git.17bdc.com/words/backend/words-learning/gen/entschema"
	"testing"

	"github.com/stretchr/testify/assert"
)

func fixture() {
	setup()
	ctx := context.Background()
	// migrate
	if err := entclient.Schema.Create(ctx); err != nil {
		panic(err)
	}
	// create
	_, err := entclient.User.Create().SetUsername("jeff").Save(ctx)
	if err != nil {
		panic(err)
	}
}

func ExampleWithTx() {
	setup()

	ctx := context.Background()
	// migrate
	if err := entclient.Schema.Create(ctx); err != nil {
		panic(err)
	}
	// create
	_, err := entclient.User.Create().SetUsername("jeff").Save(ctx)
	if err != nil {
		panic(err)
	}

	ent.WithTx(ctx, entclient, func(tx *ent.Tx) error {
		users := tx.User.Query().AllX(ctx)
		for _, user := range users {
			fmt.Println(user)
		}
		return nil
	})
}

func TestWithTx(t *testing.T) {
	assert := assert.New(t)

	fixture()
	ctx := context.Background()
	defer entclient.User.Delete().ExecX(ctx)

	t.Run("failed_case_will_rollback", func(t *testing.T) {
		ent.WithTx(ctx, entclient, func(tx *ent.Tx) error {
			_, err := tx.User.Create().SetUsername("black_jeff").Save(ctx)
			if err != nil {
				panic(err)
			}

			return fmt.Errorf("error not nil in WithTx func")
		})
		assert.EqualValues(1, entclient.User.Query().CountX(ctx))
	})

	t.Run("unexpected_panic_case_will_rollback", func(t *testing.T) {
		defer func() {
			v := recover()
			assert.NotNil(v, "not propagation panic")

			assert.EqualValues(1, entclient.User.Query().CountX(ctx))
		}()
		ent.WithTx(ctx, entclient, func(tx *ent.Tx) error {
			_, err := tx.User.Create().SetUsername("black_jeff").Save(ctx)
			if err != nil {
				panic(err)
			}

			panic(fmt.Errorf("error not nil in WithTx func"))
		})
	})

	t.Run("succeed_case_will_commit", func(t *testing.T) {
		defer func() {
			assert.Nil(recover())
		}()
		ent.WithTx(ctx, entclient, func(tx *ent.Tx) error {
			_, err := tx.User.Create().SetUsername("black_jeff").Save(ctx)
			if err != nil {
				panic(err)
			}

			return nil
		})
		assert.EqualValues(2, entclient.User.Query().CountX(ctx))
	})
}

func TestTxDecorator(t *testing.T) {
	ctx := context.Background()
	setup()
	fixture()
	defer entclient.User.Delete().ExecX(ctx)

	assert := assert.New(t)

	t.Run("invalid_decorated_objects", func(t *testing.T) {
		t.Run("not_function", func(t *testing.T) {
			defer func() {
				v := recover()
				assert.NotNil(v)

				assert.Contains(v, "decorated object is not a function")
			}()
			dFn := ent.TxDecorator(entclient, 1)
			_, _ = dFn()
		})

		t.Run("function_parameter_size_less_2", func(t *testing.T) {
			defer func() {
				v := recover()
				assert.NotNil(v)

				assert.Contains(v, "the size of decorated function's parameter is less than 2")
			}()
			dFn := ent.TxDecorator(entclient, func() {})
			_, _ = dFn()
		})

		t.Run("function_1st_parameter_not_context", func(t *testing.T) {
			defer func() {
				v := recover()
				assert.NotNil(v)

				assert.Contains(v, "the 1st parameter of decorated function is not context.Context")
			}()
			dFn := ent.TxDecorator(entclient, func(i, a int) {})
			_, _ = dFn()
		})

		t.Run("function_2nd_parameter_not_Tx", func(t *testing.T) {
			defer func() {
				v := recover()
				assert.NotNil(v)

				assert.Contains(v, "the 2nd parameter of decorated function is not *Tx")
			}()
			dFn := ent.TxDecorator(entclient, func(ctx context.Context, a int) {})
			_, _ = dFn()
		})

		t.Run("function_output_size_not_equal_2", func(t *testing.T) {
			defer func() {
				v := recover()
				assert.NotNil(v)

				assert.Contains(v, "the size of decorated function's output is not equal to 2")
			}()
			dFn := ent.TxDecorator(entclient, func(ctx context.Context, tx *ent.Tx) error {
				return nil
			})
			_, _ = dFn()
		})

		t.Run("function_2nd_output_not_error", func(t *testing.T) {
			defer func() {
				v := recover()
				assert.NotNil(v)

				assert.Contains(v, "the 2nd output of decorated function is not error")
			}()
			dFn := ent.TxDecorator(entclient, func(ctx context.Context, tx *ent.Tx) (a error, b int) {
				return nil, 0
			})
			_, _ = dFn()
		})
	})

	t.Run("valid_decorated_objects", func(t *testing.T) {

		fn := func(ctx context.Context, tx *ent.Tx) (interface{}, error) {
			return tx, nil
		}

		t.Run("tx_is_nil", func(t *testing.T) {
			dFN := ent.TxDecorator(entclient, fn)

			tx, err := dFN(ctx, nil)

			assert.Nil(err)
			assert.NotNil(tx)
		})
		t.Run("tx_not_nil", func(t *testing.T) {
			dFN := ent.TxDecorator(entclient, fn)

			tx, err := entclient.Tx(ctx)

			rtx, err := dFN(ctx, tx)

			assert.Nil(err)
			assert.Equal(tx, rtx)
		})
	})
}

func BenchmarkTxDecorator(b *testing.B) {
	ctx := context.Background()
	setup()
	fixture()
	defer entclient.User.Delete().ExecX(ctx)

	a := func(ctx context.Context, tx *ent.Tx, userID int) (interface{}, error) {
		return tx.User.Get(ctx, userID)
	}

	b.Run("tx_nil", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ent.TxDecorator(entclient, a)(ctx, nil, 1)
		}
	})

	b.Run("tx_not_nil", func(b *testing.B) {
		b.ResetTimer()
		tx, err := entclient.Tx(ctx)
		if err != nil {
			panic(err)
		}
		for i := 0; i < b.N; i++ {
			_, _ = ent.TxDecorator(entclient, a)(ctx, tx, 1)
		}
	})

}

func BenchmarkTxDirect(b *testing.B) {
	ctx := context.Background()
	setup()
	fixture()
	defer entclient.User.Delete().ExecX(ctx)

	a := func(ctx context.Context, tx *ent.Tx, userID int) (interface{}, error) {
		return tx.User.Get(ctx, userID)
	}
	da := func(ctx context.Context, tx *ent.Tx, userID int) (out interface{}, err error) {
		if tx == nil {
			if err = ent.WithTx(ctx, entclient, func(tx *ent.Tx) error {
				out, err = a(ctx, tx, userID)
				return err
			}); err != nil {
				return out, err
			}
			return out, nil
		}
		return a(ctx, tx, userID)
	}

	b.Run("tx_nil", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = da(ctx, nil, 1)
		}
	})

	b.Run("tx_not_nil", func(b *testing.B) {
		tx, err := entclient.Tx(ctx)
		if err != nil {
			panic(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = da(ctx, tx, 1)
		}
	})
}
