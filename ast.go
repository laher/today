package main

import (
	"fmt"
	"time"

	"github.com/russross/blackfriday/v2"
)

func breakNode(parent *blackfriday.Node) *blackfriday.Node {
	h := blackfriday.NewNode(blackfriday.Paragraph)
	parent.AppendChild(h)
	//h := &blackfriday.Paragraph{Container: blackfriday.Container{Parent: parent}}
	textNode := blackfriday.NewNode(blackfriday.Text)
	textNode.Literal = []byte("\n")
	h.AppendChild(textNode)
	return h
}

func headingNode(parent *blackfriday.Node, level int, text string) *blackfriday.Node {
	h := blackfriday.NewNode(blackfriday.Heading)
	h.Level = level
	parent.AppendChild(h)
	textNode := blackfriday.NewNode(blackfriday.Text)
	textNode.Literal = []byte(text + "\n")
	h.AppendChild(textNode)
	return h
}

func todayNode(parent *blackfriday.Node) *blackfriday.Node {
	t := time.Now()
	return headingNode(parent, 1, fmt.Sprintf("%s", t.Format("2006-01-02, Monday")))
}
