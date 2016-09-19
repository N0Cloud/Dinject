package dinject_test

import (
	"context"
	"dinject"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func myServiceLoader() interface{} {
	return &MyStruct{v: "mytest"}
}

func contextLoader() interface{} {
	return context.WithValue(context.Background(), "test", "myvalue")
}

type SpecialType interface{}

type MyInterface interface {
	Test() string
}
type MyStruct struct {
	v string
}

func (st *MyStruct) Test() string {
	return st.v
}

func TestNew(t *testing.T) {
	inj := dinject.New()
	assert.NotNil(t, inj)
}

func TestInvoke(t *testing.T) {
	inj := dinject.New()

	s1 := "something"
	inj.AddService(s1)
	s2 := "another thing"
	inj.AddServiceTo(s2, (*SpecialType)(nil))
	inj.AddServiceLoader(contextLoader, context.Background())
	inj.AddServiceLoaderTo(myServiceLoader, (*MyInterface)(nil))

	assert.Equal(t, reflect.TypeOf(s1), reflect.TypeOf(s2))

	result, err := inj.Invoke(func(serv1 string, serv2 SpecialType) string {
		assert.Equal(t, serv1, s1)
		assert.Equal(t, serv2, s2)

		return "Coucou"
	})

	assert.Nil(t, err)
	assert.Equal(t, result[0].String(), "Coucou")

	//alloc test if alloc free slice work
	result, err = inj.Invoke(func(serv1 string) string {
		return "Test2"
	})
	assert.Nil(t, err)
	assert.Equal(t, result[0].String(), "Test2")

	result, err = inj.Invoke(func(i MyInterface) string {
		return i.Test()
	})
	assert.Nil(t, err)
	assert.Equal(t, result[0].String(), "mytest")

	result, err = inj.Invoke(func(ctx context.Context) string {
		return ctx.Value("test").(string)
	})
	assert.Nil(t, err)
	assert.Equal(t, result[0].String(), "myvalue")
}

func TestGet(t *testing.T) {
	inj := dinject.New()
	inj.AddService("something")

	assert.True(t, inj.GetService(reflect.TypeOf("string")).IsValid())
	assert.False(t, inj.GetService(reflect.TypeOf(123)).IsValid())
}

func TestGetServiceLoader(t *testing.T) {
	inj := dinject.New()
	inj.AddService("something")
	inj.AddServiceLoader(contextLoader, context.Background())
	inj.AddServiceLoaderTo(myServiceLoader, (*MyInterface)(nil))

	assert.True(t, inj.GetService(reflect.TypeOf("string")).IsValid())
	assert.True(t, inj.GetService(reflect.TypeOf(context.Background())).IsValid())
	assert.True(t, inj.GetService(reflect.TypeOf((*MyInterface)(nil)).Elem()).IsValid())
	assert.False(t, inj.GetService(reflect.TypeOf(123)).IsValid())
}

func TestParent(t *testing.T) {
	inj1 := dinject.New()
	inj1.AddService("something")

	inj2 := dinject.New()
	inj2.Parent(inj1)

	assert.True(t, inj2.GetService(reflect.TypeOf("string")).IsValid())
}

func TestImplement(t *testing.T) {
	inj := dinject.New()

	s1 := &MyStruct{"something"}
	inj.AddService(s1)

	assert.True(t, inj.GetService(reflect.TypeOf((*MyInterface)(nil)).Elem()).IsValid())
}

func TestCommonFunc(t *testing.T) {
	inj := dinject.New()

	s1 := &MyStruct{"something"}
	inj.AddService(s1)

	result, err := inj.Invoke(func() {

	})
	assert.Nil(t, err)
	assert.Nil(t, result)

	result, err = inj.Invoke(func() string {
		return "test"
	})
	assert.Nil(t, err)
	assert.Equal(t, result[0].String(), "test")

	write := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	inj.AddService(write)
	inj.AddService(req)
	result, err = inj.Invoke(func(w http.ResponseWriter, r *http.Request) {

	})
	assert.Nil(t, err)

}
func BenchmarkInvoke(b *testing.B) {
	inj := dinject.New()

	s1 := &MyStruct{"something"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inj.Reset()
		inj.AddService("")
		inj.AddService(1)
		inj.AddService('t')
		inj.AddService(float64(1.1))
		inj.AddService(float32(1.2))
		inj.AddService(int8(1))
		inj.AddService(uint8(1))
		inj.AddService(int16(1))
		inj.AddService(uint16(1))
		inj.AddService(s1)
		inj.Invoke(func(s *MyStruct) {
			//doSomething
		})
	}
}

func BenchmarkGet(b *testing.B) {
	inj := dinject.New()

	s1 := &MyStruct{"something"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inj.Reset()
		inj.AddService(1)
		inj.AddService(float64(1.1))
		inj.AddService(float32(1.2))
		inj.AddService(int8(1))
		inj.AddService(uint8(1))
		inj.AddService(int16(1))
		inj.AddService(uint16(1))
		inj.AddService(s1)
		inj.GetService(reflect.TypeOf(s1))
	}
}
