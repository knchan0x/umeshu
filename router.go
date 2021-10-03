package umeshu

import (
	"strings"

	"github.com/knchan0x/umeshu/container"
	"github.com/knchan0x/umeshu/log"
)

//-------------------------- Router --------------------------//

// Router is a multiplexer. It registers, directs and handles url path.
type Router interface {
	// addRoute registers pattern and handler
	addRoute(method string, pattern string, handlers ...HandlerFunc)

	// getRoute find registered pattern according to the http request,
	// it also returns route parameters parsed
	getRoute(method string, path string) (registeredPath string, params map[string]string)

	// applyMiddlewares adds middlewares to the handlerChain
	applyMiddlewares(method string, path string, handlers handlerChain)

	// allRoutes returns a slice of  all registered routes' RouteInfo
	allRoutes() []RouteInfo

	// handle handles the http request
	handle(*Context)
}

// RouteInfo contains information of a registered route like method and pattern.
type RouteInfo struct {
	Method  string
	Pattern string
}

// Default implementation of Router interface. It is thread-safe unless registering
// route after http.Server starts listening and serving http requests.
type router struct {
	// method trees
	// different tree for different http methods
	trees map[string]*routerNode

	// map registered pattern with handler
	handlers map[string]handlerChain
}

var _ Router = (*router)(nil) // interface check

// NewRouter returns an new router instance
func NewRouter() Router {
	router := &router{
		trees:    make(map[string]*routerNode),
		handlers: make(map[string]handlerChain),
	}
	return router
}

// SetRouter sets the global router.
func SetRouter(router Router) {
	GlobalRouter = router
}

// addRoute adds pattern and handler to relvent method tree.
func (r *router) addRoute(method string, pattern string, handlers ...HandlerFunc) {
	if _, ok := r.trees[method]; !ok {
		// use new(Node) will create a dummy head node
		// and will cause mismatch in levels when searching
		r.trees[method] = NewRootNode()
	}
	r.trees[method].Insert(pattern)

	fullPath := method + "-" + pattern
	r.handlers[fullPath] = handlers
	log.Info("Route: %s-%s", method, pattern)
}

// getRoute find registered pattern according to the path,
// it also returns route parameters parsed.
func (r *router) getRoute(method string, path string) (string, map[string]string) {
	// check is http method registered
	if root, ok := r.trees[method]; ok {
		// check is route exists
		if registeredPath := root.Find(path); registeredPath != "" {
			parts := parsePattern(path)
			keys := parsePattern(registeredPath)
			params := matchParams(keys, parts)
			return registeredPath, params
		}
	}

	return "", nil
}

// applyMiddlewares adds middlewares into the existing handlerChain.
// Those middlewares will be placed before exisiting handers.
func (r *router) applyMiddlewares(method string, path string, middlewares handlerChain) {
	fullPath := method + "-" + path
	log.Debug("Applying middlewares for route: %s", fullPath)
	oldHandlers := r.handlers[fullPath]
	if oldHandlers != nil {
		oldHandlers = append(middlewares, oldHandlers...)
		r.handlers[fullPath] = oldHandlers
	}
}

// Routes returns a slice of registered routes' registered info.
func (r *router) allRoutes() []RouteInfo {
	list := make([]RouteInfo, len(r.handlers))
	index := 0
	for fullPath := range r.handlers {

		parts := strings.Split(fullPath, "-")
		method := parts[0]
		route := parts[len(parts)-1]

		list[index] = RouteInfo{method, route}
		index++
	}

	return list
}

// handle handles the http request.
func (r *router) handle(c *Context) {

	route, params := r.getRoute(c.Method, c.Path)

	if route != "" {
		// if route found
		fullPath := c.Method + "-" + route
		handlers := r.handlers[fullPath]
		c.handlers = handlers
		c.RouteParams = params
	} else {
		// if route not found
		c.handlers = append(c.handlers, HTTP404Handler)
	}

	c.Next()
}
func matchParams(registered []string, url []string) map[string]string {
	params := make(map[string]string)
	for i, reg := range registered {
		if reg[0] == ':' {
			params[reg[1:]] = url[i]
		}
		if reg[0] == '*' && len(reg) > 1 {
			params[reg[1:]] = strings.Join(url[i:], "/")
		}
	}
	return params
}

//-------------------------- RouterNode --------------------------//

// RouterNode is the basis unit of a router tree.
type RouterNode interface {
	// Find searchs the path and returns registered pattern
	Find(path string) (pattern string)

	// Insert adds new pattern to router node
	Insert(pattern string)
}

// routerNode is a wrapper of container.RadixNode.
// It is the default implementation of RouteNode interface.
type routerNode struct {
	*container.RadixNode
}

var _ RouterNode = (*routerNode)(nil) // interface check

// NewRootNode creates and returns new *routerNode.
func NewRootNode() *routerNode {
	newNode := &routerNode{}
	newNode.RadixNode = container.NewRootNode()
	return newNode
}

// Find searchs the path and returns registered pattern.
func (n *routerNode) Find(path string) (reg string) {
	parts := parsePattern(path)
	node := n.RadixNode.Find(parts)
	if node != nil {
		reg = node.GetPath()
	}
	return reg
}

// Insert adds new pattern to router node.
func (n *routerNode) Insert(pattern string) {
	parts := parsePattern(pattern)
	n.RadixNode.Insert(parts)
}
