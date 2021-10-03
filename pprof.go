package umeshu

import (
	"net/http/pprof"
)

var pprofRouters = []struct {
	Method  HTTPMethodType
	Path    string
	Handler HandlerFunc
}{
	{HTTP_GET, "/debug/pprof/", IndexHandler()},
	{HTTP_GET, "/debug/heap", HeapHandler()},
	{HTTP_GET, "/debug/goroutine", GoroutineHandler()},
	{HTTP_GET, "/debug/allocs", AllocsHandler()},
	{HTTP_GET, "/debug/block", BlockHandler()},
	{HTTP_GET, "/debug/threadcreate", ThreadCreateHandler()},
	{HTTP_GET, "/debug/cmdline", CmdlineHandler()},
	{HTTP_GET, "/debug/profile", ProfileHandler()},
	{HTTP_GET, "/debug/symbol", SymbolHandler()},
	{HTTP_POST, "/debug/symbol", SymbolHandler()},
	{HTTP_GET, "/debug/trace", TraceHandler()},
	{HTTP_GET, "/debug/mutex", MutexHandler()},
}

// IndexHandler will pass the call from /debug/pprof to pprof.
func IndexHandler() HandlerFunc {
	return func(c *Context) {
		pprof.Index(c.ResponseWriter, c.Request)
	}
}

// HeapHandler will pass the call from /debug/pprof/heap to pprof.
func HeapHandler() HandlerFunc {
	return func(c *Context) {
		pprof.Handler("heap").ServeHTTP(c.ResponseWriter, c.Request)
	}
}

// GoroutineHandler will pass the call from /debug/pprof/goroutine to pprof.
func GoroutineHandler() HandlerFunc {
	return func(c *Context) {
		pprof.Handler("goroutine").ServeHTTP(c.ResponseWriter, c.Request)
	}
}

// AllocsHandler will pass the call from /debug/pprof/allocs to pprof.
func AllocsHandler() HandlerFunc {
	return func(c *Context) {
		pprof.Handler("allocs").ServeHTTP(c.ResponseWriter, c.Request)
	}
}

// BlockHandler will pass the call from /debug/pprof/block to pprof.
func BlockHandler() HandlerFunc {
	return func(c *Context) {
		pprof.Handler("block").ServeHTTP(c.ResponseWriter, c.Request)
	}
}

// ThreadCreateHandler will pass the call from /debug/pprof/threadcreate to pprof.
func ThreadCreateHandler() HandlerFunc {
	return func(c *Context) {
		pprof.Handler("threadcreate").ServeHTTP(c.ResponseWriter, c.Request)
	}
}

// CmdlineHandler will pass the call from /debug/pprof/cmdline to pprof.
func CmdlineHandler() HandlerFunc {
	return func(c *Context) {
		pprof.Cmdline(c.ResponseWriter, c.Request)
	}
}

// ProfileHandler will pass the call from /debug/pprof/profile to pprof.
func ProfileHandler() HandlerFunc {
	return func(c *Context) {
		pprof.Profile(c.ResponseWriter, c.Request)
	}
}

// SymbolHandler will pass the call from /debug/pprof/symbol to pprof.
func SymbolHandler() HandlerFunc {
	return func(c *Context) {
		pprof.Symbol(c.ResponseWriter, c.Request)
	}
}

// TraceHandler will pass the call from /debug/pprof/trace to pprof.
func TraceHandler() HandlerFunc {
	return func(c *Context) {
		pprof.Trace(c.ResponseWriter, c.Request)
	}
}

// MutexHandler will pass the call from /debug/pprof/mutex to pprof.
func MutexHandler() HandlerFunc {
	return func(c *Context) {
		pprof.Handler("mutex").ServeHTTP(c.ResponseWriter, c.Request)
	}
}
