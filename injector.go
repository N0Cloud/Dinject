package dinject

import (
	"fmt"
	"net/http"
	"reflect"
)

type Injector interface {
	Invoke( /* fn */ interface{}) ([]reflect.Value, error)

	AddService(interface{})
	AddServiceTo(interface{}, interface{} /*type*/)
	AddServices(...interface{})
	AddServiceLoader(ServiceLoader, interface{})
	AddServiceLoaderTo(ServiceLoader, interface{})
	SetService(reflect.Type, reflect.Value)
	GetService(reflect.Type) reflect.Value
	Reset()
	NServices() int

	Parent(Injector)
}

var serviceLoaderType = reflect.TypeOf((*ServiceLoader)(nil)).Elem()

var httpResponseWriterType = reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()

var httpRequestType = reflect.TypeOf((*http.Request)(nil))

//ServiceLoader function to create a new service on request only
type ServiceLoader func() interface{}

type injector struct {
	services map[reflect.Type]reflect.Value
	args     []reflect.Value

	parent Injector
}

func New() Injector {
	return &injector{
		services: make(map[reflect.Type]reflect.Value),
	}
}

//Reset services
//parent isn't affected
func (i *injector) Reset() {
	if len(i.services) == 0 {
		return
	}

	for name := range i.services {
		delete(i.services, name)
	}
	i.args = i.args[0:0]
}

func (i *injector) AddService(v interface{}) {
	i.SetService(reflect.TypeOf(v), reflect.ValueOf(v))
}

//AddServices WARING, this function making alloc
//For the best performance, use AddService in a loop
func (i *injector) AddServices(vals ...interface{}) {
	for _, v := range vals {
		i.AddService(v)
	}
}

//InterfaceOf return the type of an interface
func InterfaceOf(typ interface{}) reflect.Type {
	t := reflect.TypeOf(typ)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Interface {
		panic("iface need to be an Interface. (*InterfaceType)(nil)")
	}
	return t
}

func (i *injector) AddServiceTo(v interface{}, typ interface{}) {
	i.SetService(InterfaceOf(typ), reflect.ValueOf(v))
}

func (i *injector) AddServiceLoader(s ServiceLoader, typ interface{}) {
	if s == nil {
		panic("nil ServiceLoader")
	}
	i.SetService(reflect.TypeOf(typ), reflect.ValueOf(s))
}

func (i *injector) AddServiceLoaderTo(s ServiceLoader, typ interface{}) {
	if s == nil {
		panic("nil ServiceLoader")
	}
	i.SetService(InterfaceOf(typ), reflect.ValueOf(s))
}

//SetService add a new service. It's recommanded the use AddService* methods
func (i *injector) SetService(t reflect.Type, v reflect.Value) {
	i.services[t] = v
}

//GetService return a service by type
func (i *injector) GetService(t reflect.Type) reflect.Value {
	v := i.services[t]

	if v.IsValid() {
		return returnService(v)
	}

	if t.Kind() == reflect.Interface {
		for tp, val := range i.services {
			if tp.Implements(t) {
				if val.IsValid() {
					return returnService(val)
				}
			}
		}
	}

	if i.parent != nil {
		v = i.parent.GetService(t)
	}

	return returnService(v)
}
func returnService(v reflect.Value) reflect.Value {
	if v.IsValid() {
		if v.Type() == serviceLoaderType {
			return reflect.ValueOf(v.Interface().(ServiceLoader)())
		}
	}
	return v
}

//Invoke Call a function using DI
func (i *injector) Invoke(fn interface{}) ([]reflect.Value, error) {
	t := reflect.TypeOf(fn)

	if t.Kind() != reflect.Func {
		return nil, fmt.Errorf("Unable to Invoke: %v is not a Func", fn)
	}

	nArgs := t.NumIn()
	if cap(i.args) < nArgs {
		i.args = make([]reflect.Value, nArgs)
	} else {
		i.args = i.args[0:nArgs]
	}
	for j := 0; j < nArgs; j++ {
		argTyp := t.In(j)
		v := i.GetService(argTyp)
		if !v.IsValid() {
			return nil, fmt.Errorf("Unable to Invoke: value of args[%d] not found for type %v", j, argTyp)
		}
		i.args[j] = v
	}

	return reflect.ValueOf(fn).Call(i.args), nil
}

func (i *injector) Parent(p Injector) {
	i.parent = p
}

//NServices return the total number of services (child + parent)
func (i *injector) NServices() int {
	if i.parent != nil {
		return len(i.services) + i.parent.NServices()
	}

	return len(i.services)
}
