package porcelain

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type scanner struct {
	lines []string
	nul   bool
	i     int
}

func newScanner(data []byte) *scanner {
	nul := bytes.IndexByte(data, 0) >= 0
	term := "\n"
	if nul {
		term = "\x00"
	}
	s := string(data)
	if s == "" {
		return &scanner{nul: nul}
	}
	return &scanner{lines: strings.Split(strings.TrimSuffix(s, term), term), nul: nul}
}

func (s *scanner) more() bool { return s.i < len(s.lines) }

func (s *scanner) next() string {
	line := s.lines[s.i]
	s.i++
	return line
}

func (s *scanner) path(field string) (string, error) {
	if s.nul || !strings.HasPrefix(field, `"`) {
		return field, nil
	}
	return unquote(field)
}

var escapes = map[byte]byte{
	'a':  '\a',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'v':  '\v',
	'"':  '"',
	'\\': '\\',
}

func unquote(field string) (string, error) {
	if len(field) < 2 || !strings.HasPrefix(field, `"`) || !strings.HasSuffix(field, `"`) {
		return "", fmt.Errorf("malformed quoted path: %s", field)
	}
	body := field[1 : len(field)-1]

	var out strings.Builder
	out.Grow(len(body))
	for i := 0; i < len(body); i++ {
		if body[i] != '\\' {
			out.WriteByte(body[i])
			continue
		}
		i++
		if i >= len(body) {
			return "", fmt.Errorf("truncated escape in quoted path: %s", field)
		}
		c := body[i]
		if decoded, ok := escapes[c]; ok {
			out.WriteByte(decoded)
			continue
		}

		if c < '0' || c > '7' || i+2 >= len(body) {
			return "", fmt.Errorf("unknown escape %q in quoted path: %s", c, field)
		}
		b, err := strconv.ParseUint(body[i:i+3], 8, 8)
		if err != nil {
			return "", fmt.Errorf("bad octal escape in quoted path %s: %w", field, err)
		}
		out.WriteByte(byte(b))
		i += 2
	}
	return out.String(), nil
}

func keyword(line string) (key, value string) {
	if k, v, ok := strings.Cut(line, " "); ok {
		return k, v
	}
	return line, ""
}

func ellipsis(line string, limit int) string {
	runes := []rune(line)
	if len(runes) <= limit {
		return line
	}
	return string(runes[:limit]) + "…"
}
