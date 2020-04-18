package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/laher/markdownfmt/markdown"
	"github.com/russross/blackfriday/v2"
)

const (
	usage = `today
Usage:
	today init     - initialise todo directory with today.md (and recurring.md)	
	today config   - print config variables 
	today rollover - archive the current file and archive completed/cancelled tasks 
	today days     - list a few days (for fzf inputs) 
	today headings - list the headings in a file
	today statuses - list the statuses
`
)

var (
	doneMarkers = []string{"[x]", "[X]", "[C]", "[c]"}
	statuses    = map[string]string{" ": "Todo", "i": "In progress", "x": "Done", "p": "Postponed", "c": "Cancelled"}
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Please specify a subcommand")
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}
	var (
		err        error
		printUsage = false
	)
	switch args[0] {
	case "init":
		err = initialise(args)
	case "config":
		err = printConfig(args)
	case "rollover":
		err = rollover(args)
	case "rollover-dryrun":
		err = rolloverDryRun(args)
	case "days":
		err = days(args)
	case "headings":
		err = printHeadings(args)
	case "statuses":
		err = printStatuses(args)
	default:
		err = errors.New("Unrecognised subcommand")
		printUsage = true
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error handling [%s]: %v\n", args[0], err)
		if printUsage {
			fmt.Fprintln(os.Stderr, usage)
		}
		os.Exit(1)
	}
}

func printConfig(args []string) error {
	baseDir, err := getBaseDir()
	if err != nil {
		return err
	}
	c, err := json.Marshal(map[string]interface{}{"base": baseDir, "today": filepath.Join(baseDir, todayBase), "recurring": filepath.Join(baseDir, recurringBase), "states": statuses})
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, string(c))
	return nil
}

func loadToday() (tasks, error) {
	file, err := getTodayFilename()
	if err != nil {
		return tasks{}, err
	}
	return parseFile(file)
}

func loadRecurring() (tasks, error) {
	file, err := getRecurringFilename()
	if err != nil {
		return tasks{}, err
	}
	return parseFile(file)
}

func archiveToday() error {

	ft, err := getTodayFilename()
	if err != nil {
		return err
	}

	input, err := ioutil.ReadFile(ft)
	if err != nil {
		return err
	}
	tasks, err := parse(input)
	if err != nil {
		return err
	}
	h := tasks.GetFirstHeadingText()
	var fa string
	var archiveTime time.Time
	if h != "" {
		if len(h) > 10 {
			h = h[:10]
		}
		t, err := time.Parse("2006-01-02", h)
		if err != nil {
			return err
		}
		archiveTime = t
	}
	if archiveTime.IsZero() {
		// nope - use current day - 1
		// can be off by a day, occasionally
		archiveTime = time.Now().Add(-time.Hour * 24)
	}

	fa, err = getArchiveFilename(archiveTime)
	if err != nil {
		return err
	}
	d := filepath.Dir(fa)
	if err = os.MkdirAll(d, 0755); err != nil {
		return err
	}

	// TODO check if file exists. If so, change the name of this one
	// In the meantime, just append
	f, err := os.OpenFile(fa, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	if _, err := f.Write(input); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func isDone(content string) bool {
	for _, d := range doneMarkers {
		if strings.Contains(content, d) {
			return true
		}
	}
	return false
}

// 2 passes - first to find, second to remove
func filterDone(nodes []*blackfriday.Node) []*blackfriday.Node {
	//return nodes
	nodesToUnlink := []*blackfriday.Node{}

	for _, node := range nodes {
		node.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
			if entering {
				switch node.Type {
				case blackfriday.Item:
					if node.FirstChild != nil {
						c := node.FirstChild
						log.Printf("found c: %s", string(c.Literal))
						t := node.FirstChild.FirstChild
						if t != nil {
							log.Printf("found c.c: %s", string(t.Literal))
							if isDone(string(t.Literal)) {
								log.Printf("c.c is done: %s", string(t.Literal))
								nodesToUnlink = append(nodesToUnlink, node)
								return blackfriday.SkipChildren
							}
						}
					}
					//fmt.Printf("list item content: %s. %+v\n", t.Content, t)
					/*
						case blackfriday.Text:
							// just direct children of items
							if node.Parent.Parent.Type == blackfriday.Item {
								fmt.Printf("node: %v>%v>%v\n", node.Parent.Parent.Type, node.Parent.Type, node.Type)
								if isDone(string(node.Literal)) {

									//fmt.Printf("doesnt contain: %s\n", t.Content)
									//doneNodes = append(doneNodes, node.Parent.Parent)
									node.Parent.Parent.Unlink()
								}
							}
							//t.GetParent().GetParent().SetChildren()
					*/
				}
			}
			return blackfriday.GoToNext
		})
	}
	for _, n := range nodesToUnlink {
		n.Unlink()
	}
	//fmt.Printf("filtered: %d\n", len(doneNodes))
	return nodes

}

func newToday(current tasks, recurring tasks, old tasks) error {
	f, err := getTodayFilename()
	if err != nil {
		return err
	}
	c, err := buildToday(current, recurring, old)
	if err != nil {
		return err
	}
	return newFile(f, c)
}

func buildToday(current tasks, recurring tasks, old tasks) (tasks, error) {
	headingNode(current.node, 2, "Inbox") // empty

	headingNode(current.node, 2, "Rolled Over")
	//para := paraNode(current.node)

	unfiltered := old.ByHeader("Inbox")
	filtered := filterDone(unfiltered)
	log.Printf("unfiltered/filtered: %d/%d", len(unfiltered), len(filtered))
	for _, f := range filtered {
		current.node.AppendChild(f)
	}
	//current.node.SetChildren(append(current.node.GetChildren(), i...))

	unfiltered = old.ByHeader("Rolled Over")
	filtered = filterDone(unfiltered)
	log.Printf("unfiltered/filtered: %d/%d", len(unfiltered), len(filtered))
	for _, f := range filtered {
		current.node.AppendChild(f)
	}
	//current.node.SetChildren(append(current.node.GetChildren(), i...))

	headingNode(current.node, 2, "Daily")
	// get recurring events
	d := recurring.ByHeader("Daily")
	log.Printf("daily: %d", len(d))
	for _, n := range d {
		current.node.AppendChild(n)
	}
	//current.node.SetChildren(append(current.node.GetChildren(), t...))
	return current, nil
}

func newRecurring(recurring tasks) error {
	f, err := getRecurringFilename()
	if err != nil {
		return err
	}
	return newFile(f, recurring)
}

func newFile(filename string, t tasks) error {
	d, err := getBaseDir()
	if err != nil {
		return err
	}
	err = os.MkdirAll(d, 0755)
	if err != nil {
		return err
	}
	fh, err := os.Create(filename)
	if err != nil {
		return err
	}

	r := markdown.NewRenderer(&markdown.Options{Terminal: false, HashHeaders: true})
	render(r, fh, t.node)
	return fh.Close()
}

func rolloverDryRun(args []string) error {
	// new today
	doc := blackfriday.NewNode(blackfriday.Document)
	today := tasks{node: doc}
	todayNode(today.node)

	// load today
	old, err := loadToday()
	if err != nil {
		return err
	}
	fmt.Println("Before:")
	//printAST(os.Stdout, old.node)
	if false {
		return nil
	}
	r := markdown.NewRenderer(&markdown.Options{Terminal: false, HashHeaders: true})
	render(r, os.Stdout, old.node)

	// load recurring
	recurring, err := loadRecurring()
	if err != nil {
		return err
	}
	c, err := buildToday(today, recurring, old)
	if err != nil {
		return err

	}
	_ = c
	_ = r

	fmt.Println("\nAfter:")
	//printAST(os.Stdout, c.node)
	render(r, os.Stdout, c.node)
	return nil
}

func render(r blackfriday.Renderer, w io.Writer, ast *blackfriday.Node) {
	r.RenderHeader(w, ast)
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return r.RenderNode(w, node, entering)
	})
	r.RenderFooter(w, ast)
}

func rollover(args []string) error {
	if err := archiveToday(); err != nil {
		return err
	}

	// new today
	doc := blackfriday.NewNode(blackfriday.Document)
	today := tasks{node: doc}
	todayNode(today.node)

	// load today
	old, err := loadToday()
	if err != nil {
		return err
	}
	// load recurring
	recurring, err := loadRecurring()
	if err != nil {
		return err
	}
	return newToday(today, recurring, old)
}

func printStatuses(args []string) error {
	for s, d := range statuses {
		fmt.Println(s, d)
	}
	return nil
}

func printHeadings(args []string) error {
	file := ""
	if len(args) > 1 {
		file = args[1]
	} else {
		var err error
		file, err = getTodayFilename()
		if err != nil {
			return err
		}
	}
	a, err := parseFile(file)
	if err != nil {
		return err
	}
	f := func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if entering {
			switch node.Type {
			case blackfriday.Text:
				switch node.Parent.Type {
				case blackfriday.Heading:
					text := ""
					for i := 0; i < node.Level; i++ {
						text += "#"
					}
					text += " " + string(node.Literal)
					fmt.Println(text)
				}
			}
		}
		return blackfriday.GoToNext
	}
	a.node.Walk(blackfriday.NodeVisitor(f))
	return nil
}

const (
	force             = true
	forceNewRecurring = false
)

func initialise(args []string) error {
	tf, err := getTodayFilename()
	if err != nil {
		return err
	}
	rf, err := getRecurringFilename()
	if err != nil {
		return err
	}
	recurringExists := false
	if _, err := os.Stat(rf); !os.IsNotExist(err) {
		recurringExists = true
	}

	doc := blackfriday.NewNode(blackfriday.Document)
	today := tasks{node: doc}
	todayNode(today.node)

	doc = blackfriday.NewNode(blackfriday.Document)
	recurring := tasks{node: doc}
	if !recurringExists || forceNewRecurring {
		recurring.node.AppendChild(headingNode(recurring.node, 1, "Recurring tasks"))
		recurring.node.AppendChild(headingNode(recurring.node, 2, "Daily"))
		recurring.node.AppendChild(headingNode(recurring.node, 2, "Weekly"))
		recurring.node.AppendChild(headingNode(recurring.node, 2, "Weekdays"))
		err := newRecurring(recurring)
		if err != nil {
			return err
		}
	} else {
		recurring, err = parseFile(rf)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(tf); os.IsNotExist(err) || force {
		doc := blackfriday.NewNode(blackfriday.Document)
		err = newToday(today, recurring, tasks{node: doc}) // nothing rolled over
		if err != nil {
			return err
		}
	} else {
	}

	return nil
}

func days(args []string) error {
	// print some days
	for i := 0; i < 5; i++ {
		if i == 0 {
			fmt.Print("Today, ")
		} else if i == 1 {
			fmt.Print("Tomorrow, ")
		}
		fmt.Println(time.Now().Add(time.Duration(i) * time.Hour * 24).Format("2006-01-02, Mon"))
	}
	return nil
}
