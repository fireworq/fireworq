package config

import (
	"sort"
	"testing"
)

func TestDescriptions(t *testing.T) {
	items := Descriptions()

	keys := make([]string, 0)
	for _, item := range items {
		keys = append(keys, item.Name)
	}
	sort.Strings(keys)

	for i, k := range keys {
		if items[i].Name != k {
			t.Error("Items should be sorted by their names")
		}
	}
}

func TestArgument(t *testing.T) {
	item := Item{Name: "foo_bar"}
	if item.Argument() != "--foo-bar" {
		t.Error("Wrong format")
	}
}

func TestDescribe(t *testing.T) {
	item1 := Item{
		Name:  "foo_bar",
		Label: "<value>",
		Description: `
This is an example description: configuration of ` + "`foo_bar`" + `.

The description of a configuration is wrapped in a specified width with an indentation by <code>Describe()</code> method.  HTML tags and Markdown [link](http://example.com/)s are stripped.
`,
	}

	expected1 := `  --foo-bar=<value>
    
    This is an example description:
    configuration of foo_bar.
    
    The description of a configuration
    is wrapped in a specified width with
    an indentation by Describe() method.
    HTML tags and Markdown links are
    stripped.
`

	if item1.Describe(2, 40) != expected1 {
		t.Error("Description should be wrapped and indented")
	}

	item2 := Item{
		Name:         "baz_qux",
		Label:        "<value>",
		DefaultValue: "hoge",
		Description: `
This is an example description: configuration of ` + "`baz_qux`" + `.
`,
	}

	expected2 := `  --baz-qux=<value>
    default: hoge
    
    This is an example description:
    configuration of baz_qux.
`

	if item2.Describe(2, 40) != expected2 {
		t.Error("Description should be wrapped and indented")
	}
}
