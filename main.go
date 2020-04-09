package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"github.com/laher/today/md"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Please specify a subcommand")
		os.Exit(1)
	}
	var err error
	switch args[0] {
	case "init":
		err = initialise(args)
	case "rollover":
		err = rollover(args)
	case "days":
		err = days(args)
	default:
		err = errors.New("Unrecognised subcommand")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error handling [%s]: %v\n", args[0], err)
		os.Exit(1)
	}
}

type task struct {
	Description string
	Status      string
	Tags        []string
	Created     time.Time
	Updated     time.Time
	Completed   time.Time

	Subtasks []tasks

	RecurType recurType
	From      time.Time
	Until     time.Time
}

type recurType string

const (
	custom   recurType = "custom"
	hourly   recurType = "hourly"
	daily    recurType = "daily"
	weekly   recurType = "weekly"
	weekdays recurType = "weekdays"
)

type tasks struct {
	node ast.Node
}

func (t tasks) Tasks() []task {
	// TODO use AST
	return []task{}
}

func getBaseDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "today"), nil
}

func getTodayFilename() (string, error) {
	base, err := getBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "today.md"), nil
}

func getRecurringFilename() (string, error) {
	base, err := getBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "recurring.today.md"), nil
}

func getArchiveFilename(forTime time.Time) (string, error) {
	base, err := getBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, forTime.Format(filepath.Join("2006", "01", "02"))+".today.md"), nil
}

func loadToday() (tasks, error) {
	file, err := getTodayFilename()
	if err != nil {
		return tasks{}, err
	}
	return parseFile(file)
}

func parseFile(file string) (tasks, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return tasks{}, err
	}
	return parse(b)
}

func parse(b []byte) (tasks, error) {

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)
	node := parser.Parse(b)
	return tasks{node: node}, nil
}

func loadRecurring() (tasks, error) {
	file, err := getRecurringFilename()
	if err != nil {
		return tasks{}, err
	}
	return parseFile(file)
}

func archiveToday() error {
	return nil
}

func newToday(current tasks, recurring tasks) error {
	f, err := getTodayFilename()
	if err != nil {
		return err
	}
	return newFile(f, current)
}

func newRecurring(recurring tasks) error {
	f, err := getRecurringFilename()
	if err != nil {
		return err
	}
	return newFile(f, recurring)
}

func newFile(f string, t tasks) error {
	d, err := getBaseDir()
	if err != nil {
		return err
	}
	err = os.MkdirAll(d, 0700)
	if err != nil {
		return err
	}
	fh, err := os.Create(f)
	if err != nil {
		return err
	}
	b := markdown.Render(t.node, md.NewRenderer())
	if _, err := fh.Write(b); err != nil {
		return err
	}
	return fh.Close()
}

func headingNode(parent ast.Node, level int, text string) ast.Node {
	h := &ast.Heading{Level: level, Container: ast.Container{Parent: parent}}
	h.Container.Content = []byte(text)
	h.Container.Literal = []byte(text)
	return h
}

func todayNode(parent ast.Node) ast.Node {
	t := time.Now()
	return headingNode(parent, 1, fmt.Sprintf("%s", t.Format("2006-01-02, Mon")))
}

func rollover(args []string) error {
	// load today
	today, err := loadToday()
	if err != nil {
		return err
	}
	// load recurring
	recurring, err := loadRecurring()
	if err != nil {
		return err
	}
	if err := archiveToday(); err != nil {
		return err
	}
	return newToday(today, recurring)
}

func initialise(args []string) error {
	f, err := getTodayFilename()
	if err != nil {
		return err
	}

	today := tasks{node: &ast.Document{}}
	children := today.node.GetChildren()
	children = append(children, todayNode(today.node))
	today.node.SetChildren(children)

	recurring := tasks{node: &ast.Document{}}
	// TODO add tasks
	children = recurring.node.GetChildren()
	children = append(children, headingNode(recurring.node, 1, "Recurring tasks"))
	children = append(children, headingNode(recurring.node, 2, "Daily"))
	children = append(children, headingNode(recurring.node, 2, "Weekly"))
	children = append(children, headingNode(recurring.node, 2, "Weekdays"))
	recurring.node.SetChildren(children)

	if _, err := os.Stat(f); os.IsNotExist(err) {
		err = newToday(today, recurring)
		if err != nil {
			return err
		}
	}

	f, err = getRecurringFilename()
	if err != nil {
		return err
	}
	if _, err := os.Stat(f); os.IsNotExist(err) {

		return newRecurring(recurring)
	}
	return nil
}

func days(args []string) error {
	// print some days
	for i := 0; i < 5; i++ {
		fmt.Println(time.Now().Add(time.Duration(i) * time.Hour * 24).Format("2006-01-02, Mon"))
	}
	return nil
}
