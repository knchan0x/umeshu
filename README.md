# Umeshu
Umeshu is a mini web framework written by Golang.

## Purpose
Why do I reinvent the wheel? Just for learning. 😊

Building a mini web framework from scratch using Go standard library such as net/http, container/list, sync, pprof, runtime, etc.


## Features

Umeshu has the following features:

1. radix tree based routing
	- parameter pattern
	- wildcard pattern
	- path parameter matching
2. routes grouping
3. middleware support
4. crash-free
	- auto recovery when panic
	- internal server error response
	- trackback log
5. multi-level loggers
6. cache based session
7. pprof support
8. graceful shutdown


## References
1. Gin Web Framework
	* [https://github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)
2. echo
	* [https://github.com/labstack/echo](https://github.com/labstack/echo)
3. 7 days golang programs from scratch
	* [https://github.com/geektutu/7days-golang](https://github.com/geektutu/7days-golang)
4. Build Web Application with Golang
	* [https://github.com/astaxie/build-web-application-with-golang](https://github.com/astaxie/build-web-application-with-golang)

## License
MIT License