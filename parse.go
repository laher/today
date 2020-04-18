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
	md := blackfriday.New()
	node := md.Parse(b)
	return tasks{node: node}, nil
}
