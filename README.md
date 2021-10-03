# Umeshu
Umeshu is a mini web framework written by Golang.

## Purpose
Why do I reinvent the wheel? Just for learning. 😊

Building a mini web framework from scratch using Go standard library such as net/http, container/list, sync, pprof, runtime, etc.


## Features

Umeshu has the following features:

* radix tree based routing
	* parameter pattern
	* wildcard pattern
	* path parameter matching
* routes grouping
* middleware support
* crash-free
	* auto recovery when panic
	* internal server error response
	* trackback log
* multi-level loggers
* cache based session
* pprof support
* graceful shutdown


## References
* Gin Web Framework
	* [https://github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)
* echo
	* [https://github.com/labstack/echo](https://github.com/labstack/echo)
* 7 days golang programs from scratch
	* [https://github.com/geektutu/7days-golang](https://github.com/geektutu/7days-golang)
* Build Web Application with Golang
	* [https://github.com/astaxie/build-web-application-with-golang](https://github.com/astaxie/build-web-application-with-golang)

## License
MIT License