// Code generated by entc, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"reflect"
)

func WithTx(ctx context.Context, client *Client, fn func(tx *Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w rolling back transaction: %v", err, rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

func TxDecorator(client *Client, fn interface{}) func(...interface{}) (interface{}, error) {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("decorated object is not a function")
	}

	if fnType.NumIn() < 2 {
		panic("the size of decorated function's parameter is less than 2")
	}

	contextInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
	if fnType.In(0).Kind() != reflect.Interface || !fnType.In(0).Implements(contextInterface) {
		panic("the 1st parameter of decorated function is not context.Context")
	}

	txInterface := reflect.TypeOf((*Tx)(nil)).Elem()
	if fnType.In(1).Kind() != reflect.Ptr || fnType.In(1).Elem() != txInterface {
		panic("the 2nd parameter of decorated function is not *Tx")
	}

	if fnType.NumOut() != 2 {
		panic("the size of decorated function's output is not equal to 2")
	}

	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if fnType.Out(1).Kind() != reflect.Interface || !fnType.Out(1).Implements(errorInterface) {
		panic("the 2nd output of decorated function is not error")
	}

	return func(args ...interface{}) (interface{}, error) {
		vargs := make([]reflect.Value, len(args))
		for n, v := range args {
			vargs[n] = reflect.ValueOf(v)
		}
		// the size of wrapper function should be equal to the decorated function
		out := make([]reflect.Value, 2)
		// first parameter should be Context
		actx := vargs[0].Interface().(context.Context)

		// second parameter should be *Tx
		// caller called with nil tx
		if args[1] == nil {
			if err := WithTx(actx, client, func(tx *Tx) error {
				// modified tx to new value
				vargs[1] = reflect.ValueOf(tx)
				out = reflect.ValueOf(fn).Call(vargs)
				if out[1].IsNil() {
					return nil
				}
				return out[1].Interface().(error)
			}); err != nil {
				return nil, err
			}

			// pass the outputs of decorated function to caller
			if out[1].IsNil() {
				return out[0].Interface(), nil
			}
			return out[0].Interface(), out[1].Interface().(error)
		}

		// caller called with non-nil tx
		out = reflect.ValueOf(fn).Call(vargs)

		// pass the outputs of decorated function to caller
		if out[1].IsNil() {
			return out[0].Interface(), nil
		}
		return out[0].Interface(), out[1].Interface().(error)
	}
}
