package container

import (
	"fmt"
	"strings"

	"github.com/knchan0x/umeshu/log"
)

// RadixNode is a radix node.
type RadixNode struct {
	// self
	pattern string // pattern registered
	path    string // path registered
	isParam bool   // is parameter pattern
	isAny   bool   // is wildcard pattern

	// child node
	children      []*RadixNode
	hasParamChild bool // only one param child is allowed
	hasAnyChild   bool // "*"
}

// NewRootNode returns a new *RadixNode.
func NewRootNode() *RadixNode {
	return &RadixNode{
		path: "/",
	}
}

// Find searchs radix node according to parts
func (self *RadixNode) Find(parts []string) *RadixNode {
	return self.findChild(parts, 0)
}

func (self *RadixNode) findChild(parts []string, height int) *RadixNode {
	if len(parts) == height || strings.HasPrefix(self.pattern, "*") {
		return self
	}

	part := parts[height]
	children := self.matchChildren(part)
	for _, child := range children {
		result := child.findChild(parts, height+1)
		return result
	}
	return nil
}

func (self *RadixNode) matchChildren(part string) []*RadixNode {
	nodes := make([]*RadixNode, 0)
	for _, child := range self.children {
		if child.pattern == part || child.isParam || child.isAny || part[0] == '*' {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (self *RadixNode) matchChild(part string) *RadixNode {
	for _, child := range self.children {
		if child.pattern == part {
			return child
		}
	}
	return nil
}

// Insert adds new node to radix tree.
func (self *RadixNode) Insert(parts []string) {
	self.insertChild(parts, 0)
}

func (self *RadixNode) insertChild(parts []string, height int) {
	if len(parts) == height {
		return
	}

	part := parts[height]
	child := self.matchChild(part)
	if child == nil {
		path := "/" + strings.Join(parts[:height+1], "/")

		// unable to add new pattern if wildcard pattern exists
		if part[0] != '*' && self.hasAnyChild {
			log.Panic("fail to add %s, duplicated with existing wildcard pattern", path)
			return
		}
		// only one parameter pattern is allowed in the same level
		if part[0] == ':' && self.hasParamChild {
			log.Panic("fail to add %s, dynamic patterns already exists", path)
		}

		child = &RadixNode{
			path:    path,
			pattern: part,
			isParam: part[0] == ':',
			isAny:   part[0] == '*',
		}

		if child.isParam || child.isAny {
			if child.isParam {
				self.children = append(self.children, child)
				self.hasParamChild = true
			}
			if child.isAny {
				if len(self.children) != 0 {
					if self.hasAnyChild {
						log.Warning("adding %s will replace %s", path, self.children[0].path)
						self.children[0] = child
					} else {
						log.Warning("%s will replace all child nodes of %s", path, "/"+strings.Join(parts[:height], "/"))
						self.children = []*RadixNode{child}
					}
				} else {
					self.children = append(self.children, child)
				}

				self.hasAnyChild = true
				if len(part) == 1 {
					log.Warning("%s is unnamed wildcard", child.path)
				}
			}
		} else {
			self.children = append([]*RadixNode{child}, self.children...)
		}
	}

	child.insertChild(parts, height+1)
}

// String returns formatted string of a node's data.
func (self *RadixNode) String() string {
	return fmt.Sprintf("pattern: %s, path: %s, isParam: %t, isAny: %t, no of children: %d, hasParamChild: %t, hasAnyChild: %t",
		self.pattern, self.path, self.isParam, self.isAny, len(self.children), self.hasParamChild, self.hasAnyChild)
}

// Travel returns a slice contains all nodes.
func (self *RadixNode) Travel(list *([]*RadixNode)) {
	*list = append(*list, self)
	for _, child := range self.children {
		child.Travel(list)
	}
}

// GetPath returns the path value.
func (self *RadixNode) GetPath() string {
	return self.path
}
