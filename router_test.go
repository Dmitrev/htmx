package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type FakeFileReader struct {}
func (reader FakeFileReader) ReadFile(filename string) ([]byte, error) {
    return []byte("content"), nil
}

func TestGetRoute(t *testing.T) {
    expected := "Hello World"
    handler := func (w http.ResponseWriter, r RequestContext) {
	w.WriteHeader(200)
	w.Write([]byte(expected))
    }

    writer := httptest.NewRecorder()
    request := httptest.NewRequest("GET", "/", nil)

    router := CreateRouter(func(r *Router) {
	r.reader = FakeFileReader{}
    })
    router.Get("/", handler)

    router.ServeHTTP(writer, request)

    if writer.Body.String() != expected {
	t.Error("expected router not hit\n")
    }
}

func TestPostRoute(t *testing.T) {
    handler := func (w http.ResponseWriter, r RequestContext) {
	w.WriteHeader(200)
	r.Request.ParseForm()
	name := r.Request.FormValue("name")
	fmt.Printf("name: %s\n", name)
	result := fmt.Sprintf("My name is %s", name)
	w.Write([]byte(result))
    }

    writer := httptest.NewRecorder()
    formData := "name=dmitri"
    request := httptest.NewRequest("POST", "/", strings.NewReader(formData))
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    router := CreateRouter()
    router.Post("/", handler)

    router.ServeHTTP(writer, request)

    expected := "My name is dmitri"

    if writer.Code != 200 {
	t.Error("expected 200 code\n")
    }

    if writer.Body.String() != expected {
	t.Errorf("unexpected file body: %s", writer.Body.String())
    }
}

func TestGetRouteWithParam(t *testing.T) {
    expected := "Hello account 123"
    handler := func (w http.ResponseWriter, r RequestContext) {

	w.WriteHeader(200)
	w.Write(fmt.Appendf(nil, "Hello account %s", r.Params["id"]))
    }

    writer := httptest.NewRecorder()
    request := httptest.NewRequest("GET", "/accounts/123", nil)

    router := CreateRouter(func(r *Router) {
	r.reader = FakeFileReader{}
    })
    router.Get("/accounts/:id", handler)

    router.ServeHTTP(writer, request)

    if writer.Body.String() != expected {
	t.Error("expected router not hit\n")
    }
}
