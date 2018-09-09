package disgobol

import (
	"regexp"
	"strings"
)

// MatchStart returns a function that matches a MsgContext that starts with the given string
func MatchStart(match string) func(*MsgContext) bool {
	return func(msg *MsgContext) bool {
		return strings.HasPrefix(msg.Content, match)
	}
}

// MatchContains returns a function that matches a MsgContext that contains the given string
func MatchContains(match string) func(*MsgContext) bool {
	return func(msg *MsgContext) bool {
		return strings.Contains(msg.Content, match)
	}
}

// MatchContainsWord returns a function that matches a MsgContext that contains the given word
func MatchContainsWord(match string) func(*MsgContext) bool {
	// set multiline flag and match for whitespace to be able to match nonword words (which wouldn't work with \b)
	r := regexp.MustCompile(`(?m)(\s|^)` + regexp.QuoteMeta(match) + `(\s|$)`)
	return func(msg *MsgContext) bool {
		return r.MatchString(msg.Content)
	}
}

// MatchStartWord returns a function that matches a MsgContext that starts with the given word
func MatchStartWord(match string) func(*MsgContext) bool {
	// set multiline flag and match for whitespace to be able to match nonword words (which wouldn't work with \b)
	r := regexp.MustCompile(`^` + regexp.QuoteMeta(match) + `(\s|$)`)
	return func(msg *MsgContext) bool {
		return r.MatchString(msg.Content)
	}
}

// MatchMentioned returns a function that matches a MsgContext that contains a mention for the given username
func MatchMentioned(username string) func(*MsgContext) bool {
	return func(msg *MsgContext) bool {
		for _, m := range msg.Mentions {
			if m.Username == username {
				return true
			}
		}
		return false
	}
}
