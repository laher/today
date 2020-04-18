package main

import (
	"io/ioutil"

	"github.com/russross/blackfriday/v2"
)

func parseFile(file string) (tasks, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return tasks{}, err
	}
	return parse(b)
}

func parse(b []byte) (tasks, error) {
	//extensions := parser.CommonExtensions | parser.AutoHeadingIDs

	//p := parser.NewWithExtensions(extensions)

	const extensions = blackfriday.NoIntraEmphasis |
		blackfriday.Tables |
		blackfriday.FencedCode |
		blackfriday.Autolink |
		blackfriday.Strikethrough |
		blackfriday.SpaceHeadings |
		blackfriday.NoEmptyLineBeforeBlock
	md := blackfriday.New(blackfriday.WithExtensions(extensions), blackfriday.WithExtensions(blackfriday.CommonExtensions))

	node := md.Parse(b)
	return tasks{node: node}, nil
}
