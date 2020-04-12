package main

import (
	"fmt"
	"time"

	"github.com/gomarkdown/markdown/ast"
)

func breakNode(parent ast.Node) ast.Node {
	h := &ast.Paragraph{Container: ast.Container{Parent: parent}}
	textNode := &ast.Text{
		Leaf: ast.Leaf{
			Content: []byte("\n"),
			Literal: []byte("\n"),
		},
	}
	h.SetChildren([]ast.Node{textNode})
	parent.SetChildren(append(parent.GetChildren(), h))
	return h
}

func headingNode(parent ast.Node, level int, text string) ast.Node {
	h := &ast.Heading{Level: level, Container: ast.Container{Parent: parent}}
	textNode := &ast.Text{
		Leaf: ast.Leaf{
			Content: []byte(text + "\n"),
			Literal: []byte(text + "\n"),
		},
	}
	h.SetChildren([]ast.Node{textNode})
	parent.SetChildren(append(parent.GetChildren(), h))
	return h
}

func todayNode(parent ast.Node) ast.Node {
	t := time.Now()
	return headingNode(parent, 1, fmt.Sprintf("%s", t.Format("2006-01-02, Monday")))
}
