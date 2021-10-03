package container

import (
	"fmt"
	"strings"
	"testing"
)

func parsePattern(pattern string) []string {
	s := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, part := range s {
		if part != "" {
			parts = append(parts, part)
			if part[0] == '*' {
				break
			}
		}
	}
	return parts
}

func insertNodes(paths []string) *RadixNode {
	root_GET := &RadixNode{
		path: "/",
	}

	for _, path := range paths {
		parts := parsePattern(path)
		root_GET.Insert(parts)
	}

	return root_GET
}

func TestInsert(t *testing.T) {
	var paths = []string{
		"/",
		"/:user",
		"/:user/profile",
		"/:user/viewcount",
		"/admin",
		"/view",
		"/view/:id",
		"/view/:id/:user",
		"/static/*css",
	}

	root_GET := insertNodes(paths)

	list := make([]*RadixNode, 0)
	root_GET.Travel(&list)
	for _, l := range list {
		t.Log(fmt.Println(l.String()))
	}
}

func TestFind(t *testing.T) {
	var paths = []string{
		"/",
		"/:user",
		"/:user/profile",
		"/:user/viewcount",
		"/admin",
		"/view",
		"/view/:id",
		"/view/:id/:user",
		"/static/*css",
	}
	root_GET := insertNodes(paths)

	var find = []string{
		"/",
		"/abc",
		"/abc/profile",
		"/abc/viewcount",
		"/admin",
		"/view",
		"/view/123",
		"/view/456",
		"/view/456/abc",
		"/static/js",
		"/static/css/abc.css",
	}

	var ans = []string{
		"/",
		"/:user",
		"/:user/profile",
		"/:user/viewcount",
		"/admin",
		"/view",
		"/view/:id",
		"/view/:id",
		"/view/:id/:user",
		"/static/*css",
		"/static/*css",
	}

	for idx, path := range find {
		parts := parsePattern(path)
		node := root_GET.Find(parts)
		if node == nil {
			t.Fatal("shouldn't return nil")
		}
		if node.path != ans[idx] {
			t.Fatal("incorrect matching")
		}
	}
}
