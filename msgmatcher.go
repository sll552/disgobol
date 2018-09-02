package disgobol

import (
	"regexp"
	"strings"
)

func MatchStart(match string) func(*MsgContext) bool {
	return func(msg *MsgContext) bool {
		return strings.HasPrefix(msg.Content, match)
	}
}

func MatchContains(match string) func(*MsgContext) bool {
	return func(msg *MsgContext) bool {
		return strings.Contains(msg.Content, match)
	}
}

func MatchContainsWord(match string) func(*MsgContext) bool {
	// set multiline flag and match for whitespace to be able to match nonword words (which wouldn't work with \b)
	r := regexp.MustCompile(`(?m)(\s|^)` + regexp.QuoteMeta(match) + `(\s|$)`)
	return func(msg *MsgContext) bool {
		return r.MatchString(msg.Content)
	}
}

func MatchStartWord(match string) func(*MsgContext) bool {
	// set multiline flag and match for whitespace to be able to match nonword words (which wouldn't work with \b)
	r := regexp.MustCompile(`^` + regexp.QuoteMeta(match) + `(\s|$)`)
	return func(msg *MsgContext) bool {
		return r.MatchString(msg.Content)
	}
}

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
