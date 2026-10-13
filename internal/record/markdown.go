package record

import "strings"

const blockTag = "noether"

type line struct {
	number  int
	text    string
	section string
	fenced  bool
	info    string
	block   int
}

// scan reads markdown line by line, noting the heading each line sits under and whether it
// is inside a fenced code block. Fence markers themselves are left out.
func scan(markdown string) []line {
	var (
		lines           []line
		section, open   string
		info            string
		block, fenceLen int
	)
	for i, text := range strings.Split(markdown, "\n") {
		text = strings.TrimRight(text, "\r")
		run, rest := fence(text)
		switch {
		case open == "" && run != "":
			open, fenceLen, info = run[:1], len(run), firstWord(rest)
			block++
			continue
		case open != "" && run != "" && run[:1] == open && len(run) >= fenceLen && strings.TrimSpace(rest) == "":
			open, info = "", ""
			continue
		case open == "" && strings.HasPrefix(text, "#"):
			if title, ok := heading(text); ok {
				section = title
				continue
			}
		}
		lines = append(lines, line{number: i + 1, text: text, section: section, fenced: open != "", info: info, block: block})
	}
	return lines
}

// Block returns the body of the first code block fenced with the info string "noether".
func Block(markdown string) (string, bool) {
	var body []string
	found := 0
	for _, l := range scan(markdown) {
		if !l.fenced || l.info != blockTag || (found != 0 && l.block != found) {
			continue
		}
		found = l.block
		body = append(body, l.text)
	}
	return strings.Join(body, "\n"), found != 0
}

func fence(text string) (run, rest string) {
	trimmed := strings.TrimLeft(text, " ")
	if len(text)-len(trimmed) > 3 || trimmed == "" {
		return "", ""
	}
	marker := trimmed[0]
	if marker != '`' && marker != '~' {
		return "", ""
	}
	n := 0
	for n < len(trimmed) && trimmed[n] == marker {
		n++
	}
	if n < 3 {
		return "", ""
	}
	return trimmed[:n], trimmed[n:]
}

func heading(text string) (string, bool) {
	level := 0
	for level < len(text) && text[level] == '#' {
		level++
	}
	if level > 6 || (level < len(text) && text[level] != ' ') {
		return "", false
	}
	return strings.TrimSpace(strings.TrimRight(text[level:], "# ")), true
}

func firstWord(s string) string {
	if fields := strings.Fields(s); len(fields) > 0 {
		return fields[0]
	}
	return ""
}
