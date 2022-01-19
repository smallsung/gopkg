package rpc

import (
	"context"
	"go/token"
	"reflect"
	"runtime"
	"strings"

	"github.com/smallsung/gopkg/errors"
	strings2 "github.com/smallsung/gopkg/strings"
)

type Ins struct {
	Position   []reflect.Type
	Name       map[string]reflect.Type
	HasContext bool
}

type outs struct {
	Position []reflect.Type
	ErrPos   int
}

type callback struct {
	fn       reflect.Value
	receiver reflect.Value
	ins      Ins
	outs     outs
}

func (cb *callback) call(ctx context.Context, args []reflect.Value) (interface{}, error) {
	fullArgs := make([]reflect.Value, 0, 2+len(args))
	if cb.receiver.IsValid() {
		fullArgs = append(fullArgs, cb.receiver)
	}
	if cb.ins.HasContext {
		fullArgs = append(fullArgs, reflect.ValueOf(ctx))
	}
	fullArgs = append(fullArgs, args...)

	results := cb.fn.Call(fullArgs)

	if len(results) == 0 {
		return nil, nil
	}
	if cb.outs.ErrPos >= 0 && !results[cb.outs.ErrPos].IsNil() {
		return nil, results[cb.outs.ErrPos].Interface().(error)
	}
	return results[0].Interface(), nil
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// Does t satisfy the error interface?
func isErrorType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Implements(errorType)
}

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

func makeCallback(fn, receiver reflect.Value) *callback {
	cb := new(callback)
	typ := fn.Type()

	// make callback all input
	numIn := typ.NumIn()
	ins := Ins{Position: make([]reflect.Type, 0, numIn), HasContext: false}
	for i := 0; i < numIn; i++ {
		ins.Position = append(ins.Position, typ.In(i))
	}
	if receiver.IsValid() {
		cb.receiver = receiver
		ins.Position = ins.Position[1:]
	}
	if len(ins.Position) > 0 && ins.Position[0] == contextType {
		ins.HasContext = true
		ins.Position = ins.Position[1:]
	}
	for _, in := range ins.Position {
		if !isExportedOrBuiltinType(in) {
			return nil
		}
	}

	// make callback all output
	numOut := typ.NumOut()
	outs := outs{Position: make([]reflect.Type, 0, numOut), ErrPos: -1}
	if numOut > 2 {
		return nil
	}
	for i := 0; i < numOut; i++ {
		outs.Position = append(outs.Position, typ.Out(i))
		if !isExportedOrBuiltinType(outs.Position[i]) {
			return nil
		}
		if isErrorType(outs.Position[i]) {
			outs.ErrPos = i
		}
	}
	if numOut == 2 && (outs.ErrPos != 1 || isErrorType(outs.Position[0])) {
		return nil
	}

	cb.fn, cb.ins, cb.outs = fn, ins, outs
	return cb
}

func suitableCallbacks(receiver reflect.Value) (map[string]*callback, error) {
	cbs := suitableMethods(receiver)
	if receiver.Kind() != reflect.Func {
		return cbs, nil
	}

	// self
	if name := runtime.FuncForPC(receiver.Pointer()).Name(); name != "" {
		if names := strings.Split(name, "."); len(names) == 2 {
			name = strings2.LowerFirst(names[1])
			if _, exist := cbs[name]; exist {
				return nil, errors.Trace(ErrCallbackNameExist)
			}
			cbs[name] = cbs["func"]
			delete(cbs, "func")
		}
	}
	return cbs, nil
}

func suitableMethods(receiver reflect.Value) map[string]*callback {
	cbs := make(map[string]*callback)
	typ := receiver.Type()
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		if method.PkgPath != "" {
			continue
		}
		var cb *callback
		if cb = makeCallback(method.Func, receiver); cb == nil {
			continue
		}
		cbs[strings2.LowerFirst(method.Name)] = cb
	}
	return cbs
}
