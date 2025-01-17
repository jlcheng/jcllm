package extract

import (
	"strings"
	"unicode"
)

// MentionsFromEnd looks for mentions (strings starting with "@")
// at the end of `text`. It returns two things:
//  1. The prefix (up to fireIdx) that was not consumed
//  2. The slice of mentions found in left-to-right order
//
// Mentions must appear at the end of the string, possibly separated by spaces,
// newlines, or other whitespace. Each mention is collected until we reach
// whitespace or the start of the line. For example:
//
// Input:
//
//	"Hello World.\nThis is a test message. @test @send\n@store"
//
// Output:
//
//	unburnt: "Hello World.\nThis is a test message."
//	mentions: ["store", "send", "test"]
func MentionsFromEnd(text string) (string, []string) {
	// We'll use a "fire pointer" that starts at the end.
	i := len(text)

	mentions := make([]string, 0)

	// Helper function to move backwards over whitespace.
	skipTrailingWhitespace := func() {
		for i > 0 && unicode.IsSpace(rune(text[i-1])) {
			i--
		}
	}

	// We repeatedly skip whitespace and then see if there's a mention
	// chunk we can parse.
	for {
		// 1) Skip trailing whitespace
		skipTrailingWhitespace()
		if i <= 0 {
			break
		}

		// If the next chunk does not end with '@', we can't parse a mention.
		// Actually, we parse "mention" from the end. So let's see if
		// there's a mention pattern: @xxxx
		// We proceed backwards looking for '@' and not crossing whitespace.
		j := i - 1

		// Move backwards until we hit whitespace or the beginning.
		for j >= 0 && (unicode.IsLetter(rune(text[j])) || rune(text[j]) == '@') {
			j--
		}

		// Now j+1 is the start of a token at the end of the string, i is the end of it.
		tokenStart := j + 1
		token := text[tokenStart:i]

		// Check if it starts with '@'
		if strings.HasPrefix(token, "@") && len(token) > 1 {
			// We found a mention. We'll record it (minus the '@') and
			// shrink the "fire pointer" to j.
			mentionName := token[1:] // remove leading '@'
			mentions = append(mentions, mentionName)
			i = j + 1 // move the fire pointer to the start of this token
		} else {
			// Not a mention: we are done
			break
		}
	}

	// Reverse the mentions so the leftmost mention is first
	// (because we collected them from right to left).
	for k := 0; k < len(mentions)/2; k++ {
		opp := len(mentions) - 1 - k
		mentions[k], mentions[opp] = mentions[opp], mentions[k]
	}

	// text[:i] is the unburnt prefix that remains.
	return text[:i], mentions
}
