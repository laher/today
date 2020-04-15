package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/laher/today/md"
)

const (
	usage = `today
Usage:
	today init     - initialise todo directory with today.md (and recurring.md)	
	today rollover - archive the current file and archive completed/cancelled tasks 
	today days     - list a few days (for fzf inputs) 
`
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
	case "days":
		err = days(args)
	case "headings":
		err = printHeadings(args)
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
	c, err := json.Marshal(map[string]string{"base": baseDir, "today": filepath.Join(baseDir, todayBase), "recurring": filepath.Join(baseDir, recurringBase)})
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

var doneMarkers = []string{"[x]", "[X]", "[C]", "[c]"} // x = done, c = cancelled

func isDone(content string) bool {
	for _, d := range doneMarkers {
		if strings.Contains(content, d) {
			return true
		}
	}
	return false
}

// 2 passes - first to find, second to remove
func filterDone(nodes []ast.Node) []ast.Node {
	//return nodes

	filtered := []ast.Node{}

	f := func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch t := node.(type) {
			case *ast.ListItem:
				//fmt.Printf("list item content: %s. %+v\n", t.Content, t)
			case *ast.Text:
				//fmt.Printf("node: %T>%T>%T\n", node.GetParent().GetParent(), node.GetParent(), node)
				if isDone(string(t.Literal)) {
					//fmt.Printf("doesnt contain: %s\n", t.Content)
					filtered = append(filtered, t.GetParent().GetParent())
				}
				//t.GetParent().GetParent().SetChildren()
			}
		}
		return ast.GoToNext
	}
	for _, node := range nodes {
		ast.Walk(node, ast.NodeVisitorFunc(f))
	}
	fmt.Printf("should be filtered: %d\n", len(filtered))
	fCount := 0
	for _, f := range filtered {
		fmt.Printf("filtered node: %T: %#v\n", f, f)
		p := f.GetParent()
		filteredChildren := []ast.Node{}
		for _, ch := range p.GetChildren() {
			if ch != f {
				filteredChildren = append(filteredChildren, ch)
			} else {
				fCount++
			}
		}
		p.SetChildren(filteredChildren)
	}
	fmt.Printf("filtered: %d/%d\n", fCount, len(filtered))
	return nodes

}

func newToday(current tasks, recurring tasks, old tasks) error {
	f, err := getTodayFilename()
	if err != nil {
		return err
	}
	breakNode(current.node)
	headingNode(current.node, 2, "Inbox")

	breakNode(current.node)
	headingNode(current.node, 2, "Rolled Over")

	i := old.ByHeader("Inbox")
	fmt.Printf("Old inbox: %d nodes\n", len(i))
	i = filterDone(i)
	for _, ti := range i {
		ti.SetParent(current.node)
	}
	current.node.SetChildren(append(current.node.GetChildren(), i...))

	i = old.ByHeader("Rolled Over")
	fmt.Printf("Old Rolled over: %d nodes\n", len(i))
	i = filterDone(i)
	for _, ti := range i {
		ti.SetParent(current.node)
	}
	current.node.SetChildren(append(current.node.GetChildren(), i...))

	breakNode(current.node)
	headingNode(current.node, 2, "Daily")
	// get recurring events
	t := recurring.ByHeader("Daily")
	fmt.Printf("Daily: %d nodes\n", len(t))
	for _, ti := range t {
		ti.SetParent(current.node)
	}
	current.node.SetChildren(append(current.node.GetChildren(), t...))
	return newFile(f, current)
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
	b := markdown.Render(t.node, md.NewRenderer())
	if _, err := fh.Write(b); err != nil {
		return err
	}
	return fh.Close()
}

func rollover(args []string) error {
	if err := archiveToday(); err != nil {
		return err
	}

	// new today
	today := tasks{node: &ast.Document{}}
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
	a, err :=parseFile(file)
	if err != nil {
		return err
	}
	f := func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch t := node.(type) {
			case *ast.Text:
				switch h := node.GetParent().(type) {
				case *ast.Heading:
					text := ""
					for i:=0 ; i<h.Level;i++ {
						text += "#"
					}
					text += " " + string(t.Literal)
					fmt.Println(text)
				}
			}
		}
		return ast.GoToNext
	}
	ast.Walk(a.node,ast.NodeVisitorFunc(f))
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

	today := tasks{node: &ast.Document{}}
	todayNode(today.node)

	recurring := tasks{node: &ast.Document{}}
	if !recurringExists || forceNewRecurring {
		children := recurring.node.GetChildren()
		children = append(children, headingNode(recurring.node, 1, "Recurring tasks"))
		children = append(children, headingNode(recurring.node, 2, "Daily"))
		children = append(children, headingNode(recurring.node, 2, "Weekly"))
		children = append(children, headingNode(recurring.node, 2, "Weekdays"))
		recurring.node.SetChildren(children)
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
		err = newToday(today, recurring, tasks{node: &ast.Document{}}) // nothing rolled over
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
