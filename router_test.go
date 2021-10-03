package umeshu

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/knchan0x/umeshu/log"
)

var routes = []string{
	"/",
	"/hello/:name",
	"/hello/b/c",
	"/hi/:name",
	"/assets/*filepath",
}

var paths = [][]string{
	{"/", "", ""},
	{"/hello/umeshu", "name", "umeshu"},
	{"/hello/b/c", "", ""},
	{"/hi/umeshu", "name", "umeshu"},
	{"/assets/css/test.css", "filepath", "css/test.css"},
}

var warningroutes = []string{
	"/assets/*", // warning
}

var warningpaths = [][]string{
	{"/assets/css/test.css", "", "css/test.css"}, // warning
}

func newTestRouter() *router {
	r := GlobalRouter
	for _, route := range routes {
		r.addRoute("GET", route, nil)
	}
	return r.(*router)
}

func TestParsePattern(t *testing.T) {
	ok := reflect.DeepEqual(parsePattern("/p/:name"), []string{"p", ":name"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*"), []string{"p", "*"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*name/*"), []string{"p", "*name"})
	if !ok {
		t.Fatal("test parsePattern failed")
	}
}

func TestGetRoute(t *testing.T) {
	r := newTestRouter()

	for idx, path := range paths {
		n, ps := r.getRoute("GET", path[0])

		if n == "" {
			t.Fatal("nil shouldn't be returned")
		}

		if n != routes[idx] {
			t.Fatal(fmt.Sprintf("should match %s", routes[idx]))
		}

		if path[1] != "" && ps[path[1]] != path[2] {
			t.Fatal(fmt.Sprintf("%s should be equal to '%s'", path[1], path[2]))
		}

		fmt.Printf("matched path: %s, params['%s']: %s\n", n, path[1], path[2])
	}
}

func BenchmarkGetRoute_goroutine(b *testing.B) {
	r := newTestRouter()
	log.SetLevel(log.Disable)
	b.ResetTimer()
	wg := new(sync.WaitGroup)
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			for _, path := range paths {
				r.getRoute("GET", path[0])
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
