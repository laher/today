package main

import (
	"fmt"
	"io"
	"time"

	"github.com/russross/blackfriday/v2"
)

func printAST(w io.Writer, node *blackfriday.Node) {
	indent := 0
	f := func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering {
			fmt.Print("\n")
			for p := node.Parent; p != nil; p = p.Parent {
				fmt.Print(" ")
			}
			//fmt.Print(strings.Repeat(" ", indent))
			//if node.Prev != nil {
			//	fmt.Print(",")
			//}

			fmt.Printf("%v%d: [%v]", node.Type, node.Level, string(node.Literal))

			if node.FirstChild != nil {
				indent += 1
			}
		} else {
			if node.FirstChild != nil {
				indent -= 1
			}

		}
		return blackfriday.GoToNext
	}
	node.Walk(f)
	fmt.Println()
}

func paraNode(parent *blackfriday.Node) *blackfriday.Node {
	p := blackfriday.NewNode(blackfriday.Paragraph)
	parent.AppendChild(p)
	//h := &blackfriday.Paragraph{Container: blackfriday.Container{Parent: parent}}
	//textNode := blackfriday.NewNode(blackfriday.Text)
	//textNode.Literal = []byte("\n")
	//h.AppendChild(textNode)
	return p
}

func headingNode(parent *blackfriday.Node, level int, text string) *blackfriday.Node {
	h := blackfriday.NewNode(blackfriday.Heading)
	h.Level = level
	parent.AppendChild(h)
	textNode := blackfriday.NewNode(blackfriday.Text)
	textNode.Literal = []byte(text)
	h.AppendChild(textNode)
	return h
}

func todayNode(parent *blackfriday.Node) *blackfriday.Node {
	t := time.Now()
	return headingNode(parent, 1, fmt.Sprintf("%s", t.Format("2006-01-02, Monday")))
}
