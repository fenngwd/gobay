package entext
//
//import (
//	"context"
//	"fmt"
//	"github.com/facebook/ent/dialect/sql/schema"
//	"reflect"
//)
//
//type EntTx interface {
//	Commit() error
//	Rollback() error
//}
//
//type EntClient interface {
//	Tx(ctx context.Context) (*EntTx, error)
//}
//
//func WithTx(ctx context.Context, client *EntClient, fn func(tx *EntTx) error) error {
//	tx, err := (*client).Tx(ctx)
//	if err != nil {
//		return err
//	}
//	defer func() {
//		if v := recover(); v != nil {
//			_ = (*tx).Rollback()
//			panic(v)
//		}
//	}()
//	if err := fn(tx); err != nil {
//		if rerr := (*tx).Rollback(); rerr != nil {
//			err = fmt.Errorf("%w rolling back transaction: %v", err, rerr)
//		}
//		return err
//	}
//	if err := (*tx).Commit(); err != nil {
//		return fmt.Errorf("committing transaction: %w", err)
//	}
//	return nil
//}
//
//func TxDecorator(fn interface{}) func(...interface{}) (interface{}, error) {
//	if reflect.TypeOf(fn).Kind() != reflect.Func {
//		panic("protected item is not a function")
//	}
//	if reflect.TypeOf(fn).NumOut() != 2 {
//		panic("protected item length of outputs is not equal 2")
//	}
//	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
//	if reflect.TypeOf(fn).Out(1).Kind() != reflect.Interface || !reflect.TypeOf(fn).Out(1).Implements(errorInterface) {
//		panic("第二个参数不为 error")
//	}
//
//	return func(args ...interface{}) (interface{}, error) {
//		vargs := make([]reflect.Value, len(args))
//		for n, v := range args {
//			vargs[n] = reflect.ValueOf(v)
//		}
//		actx := vargs[0].Interface().(context.Context)
//
//		out := make([]reflect.Value, 2)
//		if args[1] == nil {
//			if err := WithTx(actx, zapp.EntClient, func(tx *schema.Tx) error {
//				vargs[1] = reflect.ValueOf(tx)
//				out = reflect.ValueOf(fn).Call(vargs)
//				if out[1].IsNil() {
//					return nil
//				}
//				return out[1].Interface().(error)
//			}); err != nil {
//				return nil, err
//			}
//			if out[1].IsNil() {
//				return out[0].Interface(), nil
//			}
//			return out[0].Interface(), out[1].Interface().(error)
//		}
//
//		out = reflect.ValueOf(fn).Call(vargs)
//		if out[1].IsNil() {
//			return out[0].Interface(), nil
//		}
//		return out[0].Interface(), out[1].Interface().(error)
//	}
//}
