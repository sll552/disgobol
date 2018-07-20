package disgobol

import (
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
	return func(msg *MsgContext) bool {
		return strings.Contains(msg.Content, " "+match+" ")
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
