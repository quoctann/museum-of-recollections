package articles

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestValidateInput(t *testing.T) {
	valid := Input{Title: "Ký ức", Slug: "ky-uc", Content: json.RawMessage(`{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Xin chào 🌿","marks":[{"type":"bold"},{"type":"link","attrs":{"href":"https://example.com"}}]}]}]}}`)}
	if err := ValidateInput(valid); err != nil {
		t.Fatal(err)
	}
	cases := []Input{
		{}, {Title: "x", Slug: "A"}, {Title: "x", Slug: "a_"}, {Title: "x", Slug: strings.Repeat("a", 121)},
	}
	for _, content := range []string{`null`, `{}`, `{"schemaVersion":2,"doc":{"type":"doc"}}`, `{"schemaVersion":1,"doc":{"type":"rawHTML"}}`, `{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"image","attrs":{"src":"https://example.com"}}]}}`, `{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"text","text":"x","marks":[{"type":"link","attrs":{"href":"javascript:alert(1)"}}]}]}}`, `{"schemaVersion":1,"doc":{"type":"doc","onclick":"x"}}`, `{"schemaVersion":1,"doc":{"type":"doc"}} {}`, `{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"paragraph","attrs":{"style":"bad"}}]}}`} {
		bad := valid
		bad.Content = json.RawMessage(content)
		cases = append(cases, bad)
	}
	for i, input := range cases {
		if !errors.Is(ValidateInput(input), ErrInvalid) {
			t.Errorf("case %d should be invalid", i)
		}
	}
}

func TestContentGrammar(t *testing.T) {
	for _, doc := range []string{
		`{"type":"doc","content":[]}`,
		`{"type":"doc","content":[{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Title"}]},{"type":"blockquote","content":[{"type":"paragraph"}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph"}]}]},{"type":"orderedList","attrs":{"start":1},"content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"hardBreak"}]}]}]}]}`,
	} {
		if err := ValidateInput(Input{Title: "x", Slug: "a", Content: json.RawMessage(`{"schemaVersion":1,"doc":` + doc + `}`)}); err != nil {
			t.Fatal(err)
		}
	}
	for _, href := range []string{"javascript:x", "data:text/html,x", "//example.com", "http://user:pass@example.com", "https://example.com\n", " /foo", "https://"} {
		if safeLink(href) {
			t.Errorf("unsafe link %q", href)
		}
	}
	for _, href := range []string{"https://example.com/path?a=1", "http://example.com", "mailto:hello@example.com", "/articles/ky-uc", "#memory"} {
		if !safeLink(href) {
			t.Errorf("safe link %q", href)
		}
	}
}

func TestContentRejectsAmbiguityAndLimits(t *testing.T) {
	for _, content := range []string{
		`{"schemaVersion":2,"schemaVersion":1,"doc":{"type":"doc"}}`,
		`{"SchemaVersion":1,"doc":{"type":"doc"}}`,
		`{"schemaVersion":1,"doc":{"type":"doc","type":"doc"}}`,
		`{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"a","marks":[{"type":"bold"},{"type":"bold"}]}]}]}}`,
		`{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"heading","attrs":{"level":1}}]}}`,
		`{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"orderedList","attrs":{"start":0},"content":[{"type":"listItem","content":[{"type":"paragraph"}]}]}]}}`,
		`{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"listItem","content":[{"type":"paragraph"}]}]}}`,
	} {
		in := testInput("invalid", "Bad")
		in.Content = json.RawMessage(content)
		if !errors.Is(ValidateInput(in), ErrInvalid) {
			t.Errorf("accepted %s", content)
		}
	}
	in := testInput("size", "Big")
	in.Content = json.RawMessage(strings.Repeat(" ", MaxContentBytes+1))
	if ValidateInput(in) == nil {
		t.Fatal("oversize accepted")
	}
	deep := `{"type":"paragraph"}`
	for range 34 {
		deep = `{"type":"blockquote","content":[` + deep + `]}`
	}
	in.Content = json.RawMessage(`{"schemaVersion":1,"doc":{"type":"doc","content":[` + deep + `]}}`)
	if ValidateInput(in) == nil {
		t.Fatal("deep content accepted")
	}
	in.Content = json.RawMessage(`{"schemaVersion":1,"doc":{"type":"doc","content":[` + strings.TrimSuffix(strings.Repeat(`{"type":"paragraph"},`, 10001), ",") + `]}}`)
	if ValidateInput(in) == nil {
		t.Fatal("too many nodes accepted")
	}
	if safeLink("/%2fexample.com") {
		t.Fatal("encoded network path accepted")
	}
}

func TestContentSurrogateEscapes(t *testing.T) {
	for _, text := range []string{`\ud800`, `\udfff`, `\ud800x`, `\ud800\u0041`, `\ud800\ud800`, `\udc00\ud800`} {
		in := testInput("unicode", "Unicode")
		in.Content = json.RawMessage(`{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"` + text + `"}]}]}}`)
		if !errors.Is(ValidateInput(in), ErrInvalid) {
			t.Errorf("accepted isolated surrogate %s", text)
		}
	}
	for _, text := range []string{`\ud83d\ude00`, `\uD83D\uDE00`, `\\ud800`, `🌿`} {
		in := testInput("unicode", "Unicode")
		in.Content = json.RawMessage(`{"schemaVersion":1,"doc":{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"` + text + `"}]}]}}`)
		if err := ValidateInput(in); err != nil {
			t.Errorf("rejected valid Unicode %s: %v", text, err)
		}
	}
}
