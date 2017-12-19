package config

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Descriptions returns the default configurations and their
// descriptions.
func Descriptions() []Item {
	keys := Keys()
	sort.Strings(keys)

	cached.RLock()
	defer cached.RUnlock()

	items := make([]Item, 0, len(keys))
	for _, k := range keys {
		items = append(items, defaultConf[k].export(k))
	}

	return items
}

// Item describes a configuration key and its default value.
type Item struct {
	Category     string
	Name         string
	DefaultValue string
	Label        string
	Description  string
}

func (item *configItem) export(name string) Item {
	return Item{
		Name:         name,
		Category:     item.category,
		DefaultValue: item.defaultValue,
		Label:        item.label,
		Description:  item.description,
	}
}

// Argument returns a representation of the configuration key name as
// a command line argument.
func (item Item) Argument() string {
	return "--" + strings.Replace(item.Name, "_", "-", -1)
}

// Describe returns a string representation of the description of
// configuration key as a command line descripiton, wrapped in the
// width with indented lines.
func (item Item) Describe(indent, width int) string {
	argument := fmt.Sprintf(
		"%s%s=%s",
		strings.Repeat(" ", indent),
		item.Argument(),
		item.Label,
	)
	defaultValue := fmt.Sprintf(
		"%sdefault: %s",
		strings.Repeat(" ", indent*2),
		item.DefaultValue,
	)
	description := indentLines(
		indent*2,
		wrapLines(
			width-indent*2,
			stripMarkdown(item.Description),
		),
	)

	length := len(argument) + len(defaultValue) + len(description) + 2
	buf := bytes.NewBuffer(make([]byte, 0, length))
	fmt.Fprintln(buf, argument)
	if len(item.DefaultValue) > 0 {
		fmt.Fprintln(buf, defaultValue)
	}
	buf.WriteString(description)

	return buf.String()
}

func indentLines(n int, s string) string {
	buf := bytes.NewBuffer(make([]byte, 0, len(s)))
	indent := strings.Repeat(" ", n)

	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		buf.WriteString(indent)
		fmt.Fprintln(buf, scanner.Text())
	}

	return buf.String()
}

func wrapLines(width int, s string) string {
	buf := bytes.NewBuffer(make([]byte, 0, len(s)))

	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		fmt.Fprintln(buf, wrapLine(width, scanner.Text()))
	}

	return buf.String()
}

func wrapLine(width int, s string) string {
	if len(s) <= width {
		return s
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(s)))

	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanWords)

	count := 0
	length := 0
	for scanner.Scan() {
		token := scanner.Text()
		len := len(token)
		if length+1+len > width {
			fmt.Fprintln(buf, "")
			buf.WriteString(token)
			length = len
		} else {
			if count > 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(token)
			length += 1 + len
		}
		count++
	}

	return buf.String()
}

func stripMarkdown(s string) string {
	s = tags.ReplaceAllLiteralString(s, "")
	s = links.ReplaceAllString(s, "$1")
	s = strings.Replace(s, "`", "", -1)
	return s
}

var (
	tags  = regexp.MustCompile("<[a-zA-Z0-9'\" /._-]+>")
	links = regexp.MustCompile("\\[([^\\]]+)\\](?:\\[[^\\]]*\\]|\\([^\\)]*\\))")
)
