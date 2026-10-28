package articles

import (
	"bytes"
	"encoding/json"
	"io"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const MaxContentBytes = 1 << 20

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func validSlug(slug string) bool {
	return len(slug) > 0 && len(slug) <= 120 && slugPattern.MatchString(slug)
}

// ValidateInput accepts only schema v1 and a deliberately small Tiptap grammar.
// Images, HTML, arbitrary attributes and unknown marks are deferred, not trusted.
func ValidateInput(in Input) error {
	if !validSlug(in.Slug) || !utf8.ValidString(in.Title) || !utf8.ValidString(in.Subtitle) || strings.TrimSpace(in.Title) == "" || utf8.RuneCountInString(in.Title) > 200 || utf8.RuneCountInString(in.Subtitle) > 500 || strings.ContainsRune(in.Title, 0) || strings.ContainsRune(in.Subtitle, 0) {
		return ErrInvalid
	}
	if len(in.Content) == 0 || len(in.Content) > MaxContentBytes || !utf8.Valid(in.Content) || !validSurrogateEscapes(in.Content) {
		return ErrInvalid
	}
	if !unambiguousJSON(json.NewDecoder(bytes.NewReader(in.Content)), 0) {
		return ErrInvalid
	}
	var envelope struct {
		SchemaVersion int  `json:"schemaVersion"`
		Doc           node `json:"doc"`
	}
	if !decodeStrict(in.Content, &envelope) || envelope.SchemaVersion != 1 || envelope.Doc.Type != "doc" {
		return ErrInvalid
	}
	remaining := 10000
	if !validNode(envelope.Doc, "", 0, &remaining) {
		return ErrInvalid
	}
	return nil
}

// JSON allows UTF-16 escape pairs, but Go's decoder silently replaces isolated
// surrogates while PostgreSQL JSONB rejects them. Validate the original bytes so
// both layers interpret precisely the same Unicode content.
func validSurrogateEscapes(raw []byte) bool {
	for i := 0; i < len(raw); i++ {
		if raw[i] != '\\' {
			continue
		}
		if i+1 >= len(raw) {
			return false
		}
		if raw[i+1] != 'u' {
			i++
			continue
		}
		if i+6 > len(raw) {
			return false
		}
		code, err := strconv.ParseUint(string(raw[i+2:i+6]), 16, 16)
		if err != nil {
			return false
		}
		i += 5
		if code >= 0xdc00 && code <= 0xdfff {
			return false
		}
		if code >= 0xd800 && code <= 0xdbff {
			if i+7 > len(raw) || raw[i+1] != '\\' || raw[i+2] != 'u' {
				return false
			}
			low, err := strconv.ParseUint(string(raw[i+3:i+7]), 16, 16)
			if err != nil || low < 0xdc00 || low > 0xdfff {
				return false
			}
			i += 6
		}
	}
	return true
}

type node struct {
	Type    string                     `json:"type"`
	Attrs   map[string]json.RawMessage `json:"attrs,omitempty"`
	Content []node                     `json:"content,omitempty"`
	Text    *string                    `json:"text,omitempty"`
	Marks   []mark                     `json:"marks,omitempty"`
}
type mark struct {
	Type  string                     `json:"type"`
	Attrs map[string]json.RawMessage `json:"attrs,omitempty"`
}

func decodeStrict(raw []byte, out any) bool {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(out) != nil {
		return false
	}
	return d.Decode(new(any)) == io.EOF
}

// encoding/json otherwise silently accepts duplicate keys and case-insensitive
// field aliases. Reject both before interpreting the editor document.
func unambiguousJSON(d *json.Decoder, depth int) bool {
	if depth > 70 {
		return false
	}
	token, err := d.Token()
	if err != nil {
		return false
	}
	delim, compound := token.(json.Delim)
	if !compound {
		return true
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			keyToken, err := d.Token()
			if err != nil {
				return false
			}
			key, ok := keyToken.(string)
			if !ok || seen[key] || !contentKey(key) {
				return false
			}
			seen[key] = true
			if !unambiguousJSON(d, depth+1) {
				return false
			}
		}
	case '[':
		for d.More() {
			if !unambiguousJSON(d, depth+1) {
				return false
			}
		}
	default:
		return false
	}
	_, err = d.Token()
	return err == nil
}

func contentKey(key string) bool {
	switch key {
	case "schemaVersion", "doc", "type", "attrs", "content", "text", "marks", "level", "start", "href":
		return true
	default:
		return false
	}
}

func validNode(n node, parent string, depth int, remaining *int) bool {
	*remaining--
	if depth > 32 || *remaining < 0 {
		return false
	}
	inline := parent == "paragraph" || parent == "heading"
	block := parent == "doc" || parent == "blockquote" || parent == "listItem"
	switch n.Type {
	case "doc":
		if parent != "" || n.Text != nil || len(n.Marks) > 0 || len(n.Attrs) > 0 {
			return false
		}
	case "paragraph", "heading":
		if !block || n.Text != nil || len(n.Marks) > 0 {
			return false
		}
		if n.Type == "heading" {
			var level int
			if len(n.Attrs) != 1 || !decodeStrict(n.Attrs["level"], &level) || level < 2 || level > 3 {
				return false
			}
		} else if len(n.Attrs) > 0 {
			return false
		}
	case "blockquote", "bulletList", "orderedList", "listItem":
		if n.Text != nil || len(n.Marks) > 0 || len(n.Content) == 0 {
			return false
		}
		if n.Type == "listItem" {
			if parent != "bulletList" && parent != "orderedList" {
				return false
			}
			if n.Content[0].Type != "paragraph" {
				return false
			}
		} else if !block {
			return false
		}
		if n.Type == "orderedList" && len(n.Attrs) > 0 {
			var start int
			if len(n.Attrs) != 1 || !decodeStrict(n.Attrs["start"], &start) || start < 1 || start > 10000 {
				return false
			}
		} else if len(n.Attrs) > 0 {
			return false
		}
	case "text":
		if !inline || n.Text == nil || *n.Text == "" || strings.ContainsRune(*n.Text, 0) || len(n.Content) > 0 || len(n.Attrs) > 0 {
			return false
		}
		seen := map[string]bool{}
		for _, m := range n.Marks {
			if seen[m.Type] || !validMark(m) {
				return false
			}
			seen[m.Type] = true
		}
	case "hardBreak":
		if !inline || n.Text != nil || len(n.Content) > 0 || len(n.Marks) > 0 || len(n.Attrs) > 0 {
			return false
		}
	default:
		return false
	}
	for _, child := range n.Content {
		if !validNode(child, n.Type, depth+1, remaining) {
			return false
		}
	}
	return true
}

func validMark(m mark) bool {
	switch m.Type {
	case "bold", "italic":
		return len(m.Attrs) == 0
	case "link":
		var href string
		return len(m.Attrs) == 1 && decodeStrict(m.Attrs["href"], &href) && safeLink(href)
	default:
		return false
	}
}

func safeLink(href string) bool {
	if href == "" || len(href) > 2048 || strings.ContainsAny(href, "\\") || strings.IndexFunc(href, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
		return false
	}
	u, err := url.Parse(href)
	if err != nil || u.User != nil {
		return false
	}
	if strings.ContainsAny(u.Path, "\\\x00\r\n") {
		return false
	}
	switch u.Scheme {
	case "http", "https":
		return u.Hostname() != "" && u.Opaque == ""
	case "mailto":
		address, err := mail.ParseAddress(u.Opaque)
		return err == nil && address.Address == u.Opaque && u.RawQuery == "" && u.Fragment == ""
	case "":
		return u.Host == "" && !strings.HasPrefix(u.Path, "//") && ((strings.HasPrefix(href, "/") && !strings.HasPrefix(href, "//")) || (strings.HasPrefix(href, "#") && len(href) > 1))
	default:
		return false
	}
}
