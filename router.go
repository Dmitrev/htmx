package main

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)


type FileReader interface {
    ReadFile(filename string) ([]byte, error)
}

type ReadFileReader struct{}

func (r ReadFileReader) ReadFile(filename string) ([]byte, error) {
    return os.ReadFile(filename)
}

type RouteHandler func(w http.ResponseWriter, r RequestContext)
type Route struct {
    Method string
    Uri string
    Handler RouteHandler
}

type Router struct {
    reader FileReader
    routes []Route
}

type RequestContext struct {
    Params map[string]string
    Request *http.Request
}


func (r *Router) ServeHTTP (writer http.ResponseWriter, request *http.Request) {
    routeUriMatch := false

    for _, route := range r.routes {
	matches := routeMatches(request.RequestURI, route.Uri)

	if !matches {
	    fmt.Printf("no match\n")
	    continue
	}

	fmt.Printf("Found a match %s\n", route.Uri)
	routeUriMatch = true

	if route.Method != request.Method {
	    continue
	}

	// get parmas
	params := extractParams(route.Uri, request.RequestURI)
	context := RequestContext{
	    Params: params,
	    Request: request,
	}

	route.Handler(writer, context)
	return
    }

    if routeUriMatch {
	writer.WriteHeader(405)
	writer.Write([]byte("<p>Method not allowed</p>"))
	return
    }

    path := "./public" + request.RequestURI
    fileBytes, err := os.ReadFile(path)

    if err == nil {
	ext := filepath.Ext(path)
	contentType := mime.TypeByExtension(ext)

	writer.Header().Set("Content-Type", contentType)
	writer.WriteHeader(200)
	writer.Write(fileBytes)
	return
    }
    // check if valid file
    
    writer.Header().Set("Content-Type", "text/html")
    writer.WriteHeader(404)
    writer.Write([]byte("<p>Not Found</p>"))
}

func (r *Router) Get(uri string, handler RouteHandler) {
    r.routes = append(r.routes, Route{"GET", uri, handler})
}

func (r *Router) Post(uri string, handler RouteHandler) {
    r.routes = append(r.routes, Route{"POST", uri, handler})
}

func (r *Router) Delete(uri string, handler RouteHandler) {
    r.routes = append(r.routes, Route{"DELETE", uri, handler})
}

func extractParams(routeUri, requestUri string) map[string]string {
    requestUriParts := strings.Split(requestUri, "/")
    routeUriParts := strings.Split(routeUri, "/")

    params := make(map[string]string)

    for index, part := range routeUriParts {
	fmt.Printf("index: %d, part: %s requestPart: %s\n", index, part, requestUriParts[index])

	if !strings.HasPrefix(part, ":") {
	    continue;
	}

	name := strings.TrimLeft(part, ":")
	params[name] = requestUriParts[index]
    }

    return params
}

func routeMatches(requestUri, route string) bool {
	fmt.Printf("Check request uri: %s matches %s\n", requestUri, route)
	requestUriParts := strings.Split(requestUri, "/")
	routeUriParts := strings.Split(route, "/")

	fmt.Printf("requestUriParts: %#v\n\n", requestUriParts)
	fmt.Printf("routeUriParts: %#v\n\n", routeUriParts)

	if len(requestUriParts) != len(routeUriParts) {
	    fmt.Printf("lenght mismatch\n")
	    return false
	}

	length := len(requestUriParts)
     
	for i := range length {
	    fmt.Printf("%d\n", i)
	    if strings.HasPrefix(routeUriParts[i], ":") {
		fmt.Printf("has prefix\n")
		// wildcard
		continue
	    }

	    if requestUriParts[i] != routeUriParts[i] {
		fmt.Printf("mismatch route part %s && %s\n", routeUriParts[i], requestUriParts[i])
		return false
	    }
	}

    return true
}

func CreateRouter(options ...func(*Router)) *Router {
    router := &Router{
	reader: ReadFileReader{},
    }

    for _, option := range options {
	option(router)
    }

    return router
}

func startServer(host string, port int, router *Router) {
    serverHost := fmt.Sprintf("%s:%d", host, port)
    fmt.Printf("%s\n\n", serverHost)
    err := http.ListenAndServe(serverHost, router)
    check(err)
}
