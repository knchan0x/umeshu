package umeshu

import (
	"net/http"
	"path"
	"strings"

	"github.com/knchan0x/umeshu/log"
)

// routerGroup is a group of pattern/path with same prefix,
// it contains a slice of middlewares which will be applied
// to the router group and its sub-group.
//
// routerGroup actually just extends functionality of Router
// and works at the time configuring routes. When request
// comes in, Umeshu will use router directly for handling
// request and will by pass routerGroup.
type routerGroup struct {
	// all routerGroups share same Router instance,
	// as Umeshu uses same router for handling all routes.
	router      Router
	basePath    string
	middlewares []HandlerFunc
	engine      *Engine
}

type HTTPMethodType int

const (
	HTTP_GET HTTPMethodType = iota
	HTTP_HEAD
	HTTP_POST
	HTTP_PUT
	HTTP_DELETE
	HTTP_TRACE
	HTTP_OPTIONS
	HTTP_CONNECT
	HTTP_PATCH
)

// GlobalRouter is the router instance shared by all routerGroup.
var GlobalRouter Router = NewRouter()

// newRouterGroup returns new routerGroup and stores in (*engine).groups.
// Essentially router group just adds a prefix to the pattern,
// all routerGroup shares same router instance and shares the same radix tree.
// It will panic if group name is deplicated.
func newRouterGroup(prefix string, e *Engine) *routerGroup {
	if e.isDeplicate(prefix) {
		log.Panic("duplicated group name: %s, Umeshu may not run as you expected", prefix)
	}

	newGroup := &routerGroup{
		basePath: prefix,
		router:   GlobalRouter,
		engine:   e,
	}
	e.groups = append(e.groups, newGroup)
	return newGroup
}

// BasePath returns the base path of routerGroup.
// For example, if v := (*engine).Group("/v1/api"), v.BasePath() is "/v1/api".
func (g *routerGroup) BasePath() string {
	return g.basePath
}

// SubGroup creates new sub-routerGroup.
func (g *routerGroup) SubGroup(prefix string) *routerGroup {
	prefix = cleanPrefix(prefix)
	return newRouterGroup(g.basePath+prefix, g.engine)
}

// SubGroupWithHander creates new sub-routerGroup and defines handler for
// the sub-routerGroup pattern, i.e. SubGroupWithHander("/v1", umeshu.HTTP_GET, handlerFunc)
// defining a "/v1" sub-routerGroup - "example.com/v1" and the corresponding handlerFunc for
// "example.com/v1".
func (g *routerGroup) SubGroupWithHander(prefix string, method HTTPMethodType, handler HandlerFunc) *routerGroup {
	// g.basePath will be added at g.addRoute and g.Subgroup later
	prefix = cleanPrefix(prefix)
	g.addRoute(method, prefix, handler)
	return g.SubGroup(prefix)
}

// Use attaches middlewares to the router group.
func (g *routerGroup) Use(middlewares ...HandlerFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}

// GET registers handler for GET request.
func (g *routerGroup) GET(pattern string, handler HandlerFunc) {
	g.addRoute(HTTP_GET, pattern, handler)
}

// HEAD registers handler for HEAD request.
func (g *routerGroup) HEAD(pattern string, handler HandlerFunc) {
	g.addRoute(HTTP_HEAD, pattern, handler)
}

// POST registers handler for POST request.
func (g *routerGroup) POST(pattern string, handler HandlerFunc) {
	g.addRoute(HTTP_POST, pattern, handler)
}

// PUT registers handler for PUT request.
func (g *routerGroup) PUT(pattern string, handler HandlerFunc) {
	g.addRoute(HTTP_PUT, pattern, handler)
}

// DELETE registers handler for DELETE request.
func (g *routerGroup) DELETE(pattern string, handler HandlerFunc) {
	g.addRoute(HTTP_DELETE, pattern, handler)
}

// TRACE registers handler for TRACE request.
func (g *routerGroup) TRACE(pattern string, handler HandlerFunc) {
	g.addRoute(HTTP_TRACE, pattern, handler)
}

// OPTIONS registers handler for OPTIONS request.
func (g *routerGroup) OPTIONS(pattern string, handler HandlerFunc) {
	g.addRoute(HTTP_OPTIONS, pattern, handler)
}

// CONNECT registers handler for CONNECT request.
func (g *routerGroup) CONNECT(pattern string, handler HandlerFunc) {
	g.addRoute(HTTP_CONNECT, pattern, handler)
}

// PATCH registers handler for PATCH request.
func (g *routerGroup) PATCH(pattern string, handler HandlerFunc) {
	g.addRoute(HTTP_PATCH, pattern, handler)
}

// Any registers a route that matches all the HTTP methods, i.e.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (g *routerGroup) Any(pattern string, handler HandlerFunc) {
	for i := HTTP_GET; i <= HTTP_PATCH; i++ {
		g.addRoute(i, pattern, handler)
	}
}

// Handle registers a new request handle with the given pattern and method.
func (g *routerGroup) addRoute(method HTTPMethodType, pattern string, handler HandlerFunc) {
	assert(len(pattern) > 0, "pattern cannot be empty")
	assert(pattern[0] == '/', "pattern must begin with '/'")
	assert(handler != nil, "handler must not be nil")

	var methodString string
	switch method {
	case HTTP_GET:
		methodString = http.MethodGet
	case HTTP_HEAD:
		methodString = http.MethodHead
	case HTTP_POST:
		methodString = http.MethodPost
	case HTTP_PUT:
		methodString = http.MethodPut
	case HTTP_DELETE:
		methodString = http.MethodDelete
	case HTTP_TRACE:
		methodString = http.MethodTrace
	case HTTP_OPTIONS:
		methodString = http.MethodOptions
	case HTTP_CONNECT:
		methodString = http.MethodConnect
	case HTTP_PATCH:
		methodString = http.MethodPatch
	default:
		log.Error("invalid method type, unable to add route: %s-%s", method, pattern)
	}

	pattern = g.basePath + pattern

	if len(pattern) > 1 && pattern[len(pattern)-1] == '/' {
		pattern = pattern[:len(pattern)-1]
	}

	g.router.addRoute(methodString, pattern, handler)
}

// Static serves static files from the given file system root
//
// For example:
// 		pattern: /static
// 		root: ./assets
//
// routerGroup prefix will automatically applied if it is place under
// routerGroup or sub-routerGroup route parameters will be stored
// in "filepath".
//
// Use (*Context).GetRouteParam("filepath") to get the value.
func (g *routerGroup) Static(pattern string, root string) {
	if strings.Contains(pattern, ":") || strings.Contains(pattern, "*") {
		panic("URL parameters can not be used when serving a static folder")
	}

	handler := g.staticHandler(pattern, http.Dir(root))
	cleanPattern := path.Join(pattern, "/*filepath")

	// register handlers
	g.GET(cleanPattern, handler)
	g.HEAD(cleanPattern, handler)
}

func (g *routerGroup) staticHandler(pattern string, fs http.FileSystem) HandlerFunc {
	cleanPattern := path.Join(g.basePath, pattern)
	fileServer := http.StripPrefix(cleanPattern, http.FileServer(fs))
	return func(c *Context) {
		file := c.GetRouteParam("filepath")

		f, err := fs.Open(file)
		// check if file exists and if we are premit to access it
		if err != nil {
			c.SetStatus(http.StatusNotFound)
			f.Close()
			return
		}
		f.Close()
		fileServer.ServeHTTP(c.ResponseWriter, c.Request)
	}
}
