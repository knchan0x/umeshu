package umeshu

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/knchan0x/umeshu/log"
	"github.com/knchan0x/umeshu/session"
	"github.com/knchan0x/umeshu/view"
)

// Context represents the context of the current HTTP request. It contains
// path, route parameters, session and registered handlers. It also holds
// request and response objects.
type Context struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request

	// middlewares and handlers
	handlers []HandlerFunc
	index    int
	session  session.Session

	// provide direct access info extracts from request for convenience
	Path        string
	Method      string
	RouteParams map[string]string

	StatusCode int // status code for response
}

// JSONData is a map[string]interface{}.
type JSONData map[string]interface{}

// ctxPool is the context pool for re-using context object.
var ctxPool = sync.Pool{
	New: func() interface{} {
		return new(Context)
	},
}

var (
	HTTP404Handler func(c *Context)
	HTTP500Handler func(c *Context)
)

func init() {
	HTTP404Handler = func(c *Context) {
		c.Fail(http.StatusNotFound, fmt.Sprintf("404 NOT FOUND: %s\n", c.Path))
	}
	HTTP500Handler = func(c *Context) {
		c.Fail(http.StatusInternalServerError, "Internal Server Error")
	}
}

// NewContext return a new context instance from context pool.
func NewContext(rw http.ResponseWriter, r *http.Request) *Context {
	c := ctxPool.Get().(*Context)
	c.Init(rw, r)
	return c
}

// Init sets the initinal values for new context.
func (c *Context) Init(rw http.ResponseWriter, r *http.Request) {
	c.ResponseWriter = rw
	c.Request = r

	c.Method = r.Method
	c.Path = r.URL.Path
}

// Free frees the context object and put it into context pool.
func (c *Context) Free() {
	c.ResponseWriter = nil
	c.Request = nil
	c.handlers = nil
	c.index = 0
	c.session = nil
	c.Method = ""
	c.Path = ""
	c.RouteParams = nil
	c.StatusCode = 0
	ctxPool.Put(c)
}

// Next runs the next middleware/handler.
func (c *Context) Next() {
	c.index++
	for c.index <= len(c.handlers) {
		c.handlers[c.index-1](c)
		c.index++
	}
}

// Exit skips the remaining middlewares/handlers and executes exit handler.
//
// Note: It will still execute those code after c.Next() of those middlewares
// and those defer functions.
//
// Warning: It will also skip the handler registered. Re-define handler in exit
// handler to handle it if it is necessary.
func (c *Context) Exit(exitHandler HandlerFunc) {
	if exitHandler != nil {
		c.handlers = append(c.handlers, exitHandler)
		c.handlers[len(c.handlers)-1](c)
	}
}

// StartSession returns existing session or starts new session if no one exists.
func (c *Context) StartSession() {
	c.session = session.Manager.StartSession(c.ResponseWriter, c.Request)
}

// EndSession ends session.
func (c *Context) EndSession() {
	session.Manager.EndSession(c.ResponseWriter, c.Request)
}

// GetSession gets session object, will implicitly call (c *Context).StartSession()
// if there is no session object exists.
func (c *Context) GetSession() session.Session {
	if c.session == nil {
		// start session when first read
		c.StartSession()
	}
	return c.session
}

// GetRouteParam returns route parameters.
func (c *Context) GetRouteParam(key string) string {
	value := c.RouteParams[key]
	return value
}

// Wrapper function of (*http.Request).FormValue(key string) string.
func (c *Context) FormValue(key string) string {
	return c.Request.FormValue(key)
}

// Wrapper function of (url.Values).Get(key string) string.
func (c *Context) GetQuery(key string) string {
	return c.Request.URL.Query().Get(key)
}

func (c *Context) SetStatus(code int) {
	c.StatusCode = code
	c.ResponseWriter.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.ResponseWriter.Header().Set(key, value)
}

// Data responses http request by returning data specified
func (c *Context) Data(code int, data []byte) {
	c.SetStatus(code)
	if _, err := c.ResponseWriter.Write(data); err != nil {
		log.Error("http response writer error: %s", err)
	}
}

// HTML responses http request by returning a static html page
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.StatusCode = code
	c.Data(code, []byte(html))
}

// HTMLTemplate responses http request by returning a html page according to template specified.
func (c *Context) HTMLTemplate(code int, name string, data interface{}) {
	if view.Manager == nil {
		log.Error("view manager not exists")
	}

	c.SetHeader("Content-Type", "text/html")
	c.StatusCode = code
	if err := view.Manager.ExecuteTemplate(c.ResponseWriter, name, data); err != nil {
		log.Error("unable to execute %s", err)

	}
}

// String responses http request by returning plain text
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Data(code, []byte(fmt.Sprintf(format, values...)))
}

// Fail responses http request by returning error message specified
func (c *Context) Fail(code int, err string) {
	c.String(code, "%d %s", code, err)
}

// JSON responses http request by returning JSON object
func (c *Context) JSON(code int, object interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.SetStatus(code)
	encoder := json.NewEncoder(c.ResponseWriter)
	if err := encoder.Encode(object); err != nil {
		panic(err)
	}
}

// Redirect redicects route to another path.
func (c *Context) Redirect(code int, to string) {
	http.Redirect(c.ResponseWriter, c.Request, to, code)
}
