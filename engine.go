package umeshu

import (
	"context"
	"net/http"
	"strings"

	"github.com/knchan0x/umeshu/log"
	"github.com/knchan0x/umeshu/session"
	"github.com/knchan0x/umeshu/view"
)

// Engine is the core of Umeshu. It contains the mux, middlewares, session
// manager and view render.
// Use New() or Default() to create it.
type Engine struct {
	*routerGroup
	groups   []*routerGroup
	shutdown chan struct{}
}

// HandlerFunc defines the request handler.
type HandlerFunc func(*Context)

// HandlerChain defines a slice of HandlerFunc for internal use.
type handlerChain []HandlerFunc

// FuncMap is a wrapper of map[string]interface{}, it use to pass
// FuncMap to HTML template render.
type FuncMap map[string]interface{}

// New returns a new blank Engine instance without any middleware attached.
// It is also act as the first routerGroup with empty prefix.
func New() *Engine {
	e := &Engine{
		groups: []*routerGroup{},
	}
	e.routerGroup = newRouterGroup("", e)
	return e
}

// Default returns an Engine instance with Recovery middleware already attached.
// Internally, it calls (*Engine).New() and attaches Recovery middleware.
func Default() *Engine {
	e := New()
	e.Use(Recovery())
	return e
}

// ServeHTTP conforms to http.Handler interface.
func (e *Engine) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	context := NewContext(rw, r)
	e.routerGroup.router.handle(context)
}

// Run sets up a http server and starts listening and serving HTTP requests.
func (e *Engine) Run(addr string) {
	srv := e.prepareServer(addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Panic("unable to run Umeshu engine")
	}
	log.Info("Umeshu is listening and serving HTTP on %s\n", addr)
}

// Run sets up a http server and starts listening and serving HTTPS requests.
func (e *Engine) RunTLS(addr, certFile, keyFile string) {
	srv := e.prepareServer(addr)
	if err := srv.ListenAndServeTLS(certFile, keyFile); err != http.ErrServerClosed {
		log.Panic("unable to run Umeshu engine")
	}
	log.Info("Umeshu is listening and serving HTTP on %s\n", addr)
}

// Shutdown sends a message to http.Server to shut it down.
func (e *Engine) Shutdown() {
	if e.shutdown == nil {
		log.Error("Umeshu engine is not running, unable to close it.")
		return
	}
	log.Info("Shutting down Uneshu engine...")
	close(e.shutdown)
}

// prepareServer creates and returns *http.Server instance.
//
// Internally, it will start a new goroutine to monitoring
// shutdown signal.
//
// It will also apply middlewares to all registered routes
// i.e. automaticlly calls (*Engine).ApplyMiddleware()
func (e *Engine) prepareServer(addr string) *http.Server {
	srv := &http.Server{
		Addr:    addr,
		Handler: e,
	}

	// apply middlewares
	e.ApplyMiddleware()

	e.shutdown = make(chan struct{}, 1)
	go func() {
		for {
			if _, ok := <-e.shutdown; !ok {
				err := srv.Shutdown(context.Background())
				if err != nil {
					log.Error("shutdown error: %s", err.Error())
				}
				log.Info("Uneshu engine is shutted down.")
			}
		}
	}()

	return srv
}

// ApplyMiddleware apply middlewares on all registered routes.
//
// Warning: this function must be invoked before http.Server starts
// Listening and serving http requests. Otherwise, no middlewares
// will be attached to routes.
func (e *Engine) ApplyMiddleware() {
	routes := e.router.allRoutes()
	for _, route := range routes {
		e.applyMiddleware(route.Method, route.Pattern)
	}
}

// applyMiddleware combines middlewares and actual HandlerFunc
// to create a handlerChain and stores it to the route registered.
func (e *Engine) applyMiddleware(method string, path string) {
	// add middlewares
	var chain handlerChain
	for _, group := range e.groups {
		if strings.HasPrefix(path, group.basePath) {
			chain = append(chain, group.middlewares...)
		}
	}

	e.router.applyMiddlewares(method, path, chain)
}

// Group creates a new router group.
func (e *Engine) Group(prefix string) *routerGroup {
	prefix = cleanPrefix(prefix)
	return newRouterGroup(prefix, e)
}

// LoadHTMLTemplates loads the templates from folder and stores it
// in ViewManager, it also stores FuncMap.
func (e *Engine) LoadHTMLTemplates(folder string, funcMap FuncMap) {
	pattern := folder + "/*"
	view.NewManager(pattern, view.FuncMap(funcMap))
}

// EnableSession starts session.Manager with settings provided.
// Use session.DefaultSession for default settings.
func (e *Engine) EnableSession(settings session.SessionSettings) {
	session.NewManager(settings)
}

// EnablePprof adds pprof related handlers to router.
// Default index page for debug is "/debug/pprof/"
func (e *Engine) EnablePprof() {
	for _, r := range pprofRouters {
		e.addRoute(r.Method, r.Path, r.Handler)
	}
}

func (e *Engine) isDeplicate(prefix string) bool {
	for _, group := range e.groups {
		if prefix == group.basePath {
			return true
		}
	}
	return false
}
