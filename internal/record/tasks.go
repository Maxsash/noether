package record

import (
	"regexp"
	"strings"
)

type Task struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Section string `json:"section"`
	Text    string `json:"text"`
	Done    bool   `json:"done"`
}

var (
	checkbox = regexp.MustCompile(`^\s*[-*+]\s+\[([ xX])\]\s+(.*)$`)
	listItem = regexp.MustCompile(`^\s*([-*+]|\d+[.)])\s`)
)

// Tasks lists the markdown checkboxes in a file, skipping any inside code blocks. A task
// written over several indented lines is joined into one.
func Tasks(file, markdown string) []Task {
	var tasks []Task
	open, indent, previous := -1, 0, 0
	for _, l := range scan(markdown) {
		switch m := checkbox.FindStringSubmatch(l.text); {
		case l.fenced:
			open = -1
		case m != nil:
			tasks = append(tasks, Task{
				File:    file,
				Line:    l.number,
				Section: l.section,
				Text:    strings.TrimSpace(m[2]),
				Done:    m[1] != " ",
			})
			open, indent = len(tasks)-1, leadingSpaces(l.text)
		case open >= 0 && l.number == previous+1 && continues(l.text, indent):
			tasks[open].Text += " " + strings.TrimSpace(l.text)
		default:
			open = -1
		}
		previous = l.number
	}
	return tasks
}

func continues(text string, indent int) bool {
	return strings.TrimSpace(text) != "" && leadingSpaces(text) > indent && !listItem.MatchString(text)
}

func leadingSpaces(text string) int {
	return len(text) - len(strings.TrimLeft(text, " \t"))
}
