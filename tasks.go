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
	var candidates = []*blackfriday.Node{}
	inLevel := -1 // grab everything
	for topLevelNode := t.node.FirstChild; topLevelNode != nil; topLevelNode = topLevelNode.Next {
		if inLevel == -1 {
			if topLevelNode.Type == blackfriday.Heading {

				for child := topLevelNode.FirstChild; child != nil; child = child.Next {
					if strings.Contains(string(child.Literal), s) {
						inLevel = topLevelNode.Level
						//log.Printf("Found %s (%s)", s, child.Literal)
					}
				}
			}
		} else {
			if topLevelNode.Type == blackfriday.Heading {
				if topLevelNode.Level >= inLevel {
					break
				}
			}
			candidates = append(candidates, topLevelNode)
		}
	}
	/*
		f := func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
			switch node.Type {
			case blackfriday.Document: // ignore
			case blackfriday.Heading:
				if entering {
					//fmt.Printf("Heading, l %d: '%s'\n", h.Level, h.Content)
					if inLevel > -1 { // reset
						if node.Level <= inLevel {
							inLevel = -1
						}
					}
					for child := node.FirstChild; child != nil; child = child.Next {
						if strings.Contains(string(child.Literal), s) {
							firstChild = child
							return blackfriday.Terminate
						}
					}
					//fmt.Printf("Heading children: %d, %d\n", len(h.Children), len(node.GetChildren()))
				} else {
					if node.Level <= inLevel {
						inLevel = -1
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
		t.node.Walk(f)
	*/
	return candidates
}
