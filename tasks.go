package main

import (
	"strings"

	"github.com/russross/blackfriday/v2"
)

type tasks struct {
	node *blackfriday.Node
}

func (t tasks) Tasks() []task {
	// TODO use AST
	return []task{}
}

func (t tasks) GetFirstHeadingText() string {
	var (
		state = 0
		ret   = ""
	)

	f := func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		switch node.Type {
		case blackfriday.Heading:
			if state == 0 {
				state = 1
			} else {
				state = 2
			}
		case blackfriday.Text:
			if state == 1 {
				ret = string(node.Literal)
				return blackfriday.Terminate
			}
		}
		return blackfriday.GoToNext
	}
	t.node.Walk(blackfriday.NodeVisitor(f))
	return ret
}

func (t tasks) ByHeader(s string) []*blackfriday.Node {
	tasks := []*blackfriday.Node{}
	inLevel := -1
	f := func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		switch node.Type {
		case blackfriday.Document: // ignore
		case blackfriday.List:
			if entering {
				if inLevel > -1 {
					tasks = append(tasks, node)
				}
			}
		case blackfriday.Heading:
			if entering {
				//fmt.Printf("Heading, l %d: '%s'\n", h.Level, h.Content)
				if inLevel > -1 { // reset
					if node.Level <= inLevel {
						inLevel = -1
					}
				}
				//fmt.Printf("Heading children: %d, %d\n", len(h.Children), len(node.GetChildren()))
			} else {
				if node.Level == inLevel {
					//inLevel = -1
				}
			}
		case blackfriday.Text:
			//fmt.Printf("literal: %v, leaf: %#v\n", string(node.AsLeaf().Literal), node.AsLeaf())
			if node.Parent.Type == blackfriday.Heading {
				if strings.Contains(string(node.Literal), s) {
					inLevel = node.Level
					//tasks = append(tasks, node)
				}
			}
		case blackfriday.Item:
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
			//			if node.AsContainer() != nil {
			//				fmt.Printf("inLevel %d, type: %T, content: %s\n", inLevel, node, node.AsContainer().Content)
			//			}
		}
		return blackfriday.GoToNext
	}
	t.node.Walk(blackfriday.NodeVisitor(f))
	return tasks
}
