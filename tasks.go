package main

import (
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

type tasks struct {
	node ast.Node
}

func (t tasks) Tasks() []task {
	// TODO use AST
	return []task{}
}

func (t tasks) ByHeader(s string) []ast.Node {
	tasks := []ast.Node{}
	inLevel := -1
	f := func(node ast.Node, entering bool) ast.WalkStatus {
		switch h := node.(type) {
		case *ast.Document: // ignore
		case *ast.List:
			if entering {
				if inLevel > -1 {
					tasks = append(tasks, node)
				}
			}
		case *ast.Heading:
			if entering {
				//fmt.Printf("Heading, l %d: '%s'\n", h.Level, h.Content)
				if inLevel > -1 { // reset
					if h.Level <= inLevel {
						inLevel = -1
					}
				}
				//fmt.Printf("Heading children: %d, %d\n", len(h.Children), len(node.GetChildren()))
			} else {
				if h.Level == inLevel {
					//inLevel = -1
				}
			}
		case *ast.Text:
			//fmt.Printf("literal: %v, leaf: %#v\n", string(node.AsLeaf().Literal), node.AsLeaf())
			if p, ok := h.Parent.(*ast.Heading); ok {
				if strings.Contains(string(h.Literal), s) {
					inLevel = p.Level
					//tasks = append(tasks, node)
				}
			}
		case *ast.ListItem:
			if entering {
				if inLevel > -1 {
					//fmt.Printf("Including List item: %v, inLevel: %v, container: %#v\n", string(node.AsContainer().Content), inLevel, node.AsContainer())
					//		tasks = append(tasks, node)
				}
			}
		default:
			if entering {
				//fmt.Printf("*** Other Type ***: %T, full: %#v\n", node, node)
			}
		}
		if inLevel > -1 {
			if node.AsContainer() != nil {
				//				fmt.Printf("inLevel %d, type: %T, content: %s\n", inLevel, node, node.AsContainer().Content)
			}
		}
		return ast.GoToNext
	}
	ast.Walk(t.node, ast.NodeVisitorFunc(f))
	return tasks
}
