package slug

import (
	"strings"
	"testing"
)

func TestMake(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "basic", text: "How should I learn Go?", want: "how-should-i-learn-go"},
		{name: "punctuation", text: "  Hello,   world!!  ", want: "hello-world"},
		{name: "underscores", text: "one_two_three", want: "one-two-three"},
		{name: "digits", text: "Top 10 books", want: "top-10-books"},
		{name: "empty", text: "!!!", want: "item"},
		{name: "non ascii", text: "こんにちは", want: "item"},
		{name: "long", text: strings.Repeat("word ", 20), want: strings.TrimRight(strings.Repeat("word-", 12), "-")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Make(tt.text); got != tt.want {
				t.Fatalf("Make(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}
