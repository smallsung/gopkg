package rpc

import (
	"reflect"
	"strings"
	"sync"

	"github.com/smallsung/gopkg/errors"
)

type service struct {
	name      string
	callbacks map[string]*callback
}

type registry struct {
	services map[string]service
	mu       sync.Mutex
}

func (r *registry) register(namespaces string, receiver interface{}) (err error) {
	rv := reflect.ValueOf(receiver)
	if namespaces == "" {
		return errors.Format("%s namespace empty", rv.Type().String())
	}
	var cbs map[string]*callback
	if cbs, err = suitableCallbacks(rv); err != nil {
		return errors.Trace(err)
	}
	if len(cbs) == 0 {
		return errors.Format("%s doesn't have any suitable methods", rv.Type().String())
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.services == nil {
		r.services = make(map[string]service)
	}

	if svc, exist := r.services[namespaces]; !exist {
		svc = service{name: "", callbacks: make(map[string]*callback)}
		r.services[namespaces] = svc
	} else {
		for name, _ := range cbs {
			if _, exist := r.services[namespaces].callbacks[name]; exist {
				return errors.Annotatef(ErrCallbackNameExist, "%s.%s", namespaces, name)
			}
		}
	}
	for name, cb := range cbs {
		r.services[namespaces].callbacks[name] = cb
	}
	return nil
}

const MethodSeparator = "."

func (r *registry) callback(method string) *callback {
	elem := strings.SplitN(method, MethodSeparator, 2)
	if len(elem) != 2 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	return r.services[elem[0]].callbacks[elem[1]]
}
