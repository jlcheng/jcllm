package extract_test

import (
	"reflect"
	"testing"

	"github.com/jlcheng/jcllm/extract"
)

func TestParseMentionsFromEnd(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantPrefix   string
		wantMentions []string
	}{
		{
			name:         "No mentions",
			input:        "Hello world",
			wantPrefix:   "Hello world",
			wantMentions: []string{},
		},
		{
			name:         "One mention at the end",
			input:        "Hello world @test",
			wantPrefix:   "Hello world",
			wantMentions: []string{"test"},
		},
		{
			name:         "Multiple mentions, same line",
			input:        "Hello world @test @send",
			wantPrefix:   "Hello world",
			wantMentions: []string{"test", "send"},
		},
		{
			name:         "Multiple mentions with new lines and spaces",
			input:        "Hello world.\nThis is a test message. @test @send\n@store",
			wantPrefix:   "Hello world.\nThis is a test message.",
			wantMentions: []string{"test", "send", "store"},
		},
		{
			name:         "Mention in the middle (not at the end)",
			input:        "Hello @world here",
			wantPrefix:   "Hello @world here",
			wantMentions: []string{},
		},
		{
			name:         "Empty string",
			input:        "",
			wantPrefix:   "",
			wantMentions: []string{},
		},
		{
			name:         "Mention with punctuation after it (not valid mention at end)",
			input:        "Some text @excited!",
			wantPrefix:   "Some text @excited!",
			wantMentions: []string{},
		},
		{
			name:         "Only mention",
			input:        "@onlymention",
			wantPrefix:   "",
			wantMentions: []string{"onlymention"},
		},
		{
			name:         "Mention followed by whitespace and punctuation",
			input:        "Something @user   , next",
			wantPrefix:   "Something @user   , next",
			wantMentions: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrefix, gotMentions := extract.MentionsFromEnd(tt.input)
			if gotPrefix != tt.wantPrefix {
				t.Errorf("MentionsFromEnd(%q) prefix = %q; want %q",
					tt.input, gotPrefix, tt.wantPrefix)
			}
			if !reflect.DeepEqual(gotMentions, tt.wantMentions) {
				t.Errorf("MentionsFromEnd(%q) mentions = %v; want %v",
					tt.input, gotMentions, tt.wantMentions)
			}
		})
	}
}
