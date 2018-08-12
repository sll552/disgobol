package disgobol

import (
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
)

type args struct {
	match string
}

type testdef struct {
	name     string
	args     args
	input    string
	mentions []string
	want     bool
}

func createMessageContext(content string) *MsgContext {
	return &MsgContext{
		Message: &discordgo.Message{
			Content: content,
		},
	}
}

func createMessageContextWithMentions(content string, usernames []string) *MsgContext {
	ctx := createMessageContext(content)
	for _, username := range usernames {
		ctx.Mentions = append(ctx.Mentions, &discordgo.User{Username: username})
	}
	return ctx
}

func TestMatchStart(t *testing.T) {
	tests := []testdef{
		{name: "match1", args: args{match: "simple"}, input: "simple text starting with", want: true},
		{name: "match2", args: args{match: "simple"}, input: "sample text starting with simple", want: false},
		{name: "newline1", args: args{match: "simple"}, input: "simple text \nstarting with", want: true},
		{name: "newline2", args: args{match: "simple"}, input: "sample \ntext starting with simple", want: false},
		{name: "case1", args: args{match: "simple"}, input: "Simple text starting with simple", want: false},
		{name: "case2", args: args{match: "Simple"}, input: "Simple text starting with simple", want: true},
		{name: "endswith", args: args{match: "simple"}, input: "text ending with simple", want: false},
		{name: "whitespace", args: args{match: "simple"}, input: "   simple text ending with simple", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchStart(tt.args.match)(createMessageContext(tt.input)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MatchStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchContains(t *testing.T) {
	tests := []testdef{
		{name: "match1", args: args{match: "simple"}, input: "simple text contains with", want: true},
		{name: "match2", args: args{match: "simple"}, input: "sample text contains with", want: false},
		{name: "case1", args: args{match: "simple"}, input: "Simple text contains with", want: false},
		{name: "case2", args: args{match: "Simple"}, input: "Simple text contains with", want: true},
		{name: "endswith", args: args{match: "simple"}, input: "text contains with simple", want: true},
		{name: "middle", args: args{match: "simple"}, input: "text simple contains with", want: true},
		{name: "inword1", args: args{match: "simple"}, input: "text simplecontains with", want: true},
		{name: "inword2", args: args{match: "simple"}, input: "testsimplecontains with", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchContains(tt.args.match)(createMessageContext(tt.input)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MatchContains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchContainsWord(t *testing.T) {
	tests := []testdef{
		{name: "match1", args: args{match: "simple"}, input: "simple text starting with", want: true},
		{name: "match2", args: args{match: "simple"}, input: "sample text starting with", want: false},
		{name: "case1", args: args{match: "simple"}, input: "Simple text starting with", want: false},
		{name: "case2", args: args{match: "Simple"}, input: "Simple text starting with", want: true},
		{name: "endswith", args: args{match: "simple"}, input: "text ending with simple", want: true},
		{name: "middle", args: args{match: "simple"}, input: "text simple \nending with", want: true},
		{name: "inword1", args: args{match: "simple"}, input: "text simpleending with", want: false},
		{name: "inword2", args: args{match: "simple"}, input: "textsim\npleending with", want: false},
		{name: "regex1", args: args{match: "[a-z]+"}, input: "textsimpleending with", want: false},
		{name: "regex2", args: args{match: "[a-z]+"}, input: "text \n[a-z]+ with", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchContainsWord(tt.args.match)(createMessageContext(tt.input)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MatchContainsWord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchMentioned(t *testing.T) {
	tests := []testdef{
		{
			name:     "match1",
			args:     args{match: "testuser1"},
			input:    "test",
			mentions: []string{"testuser1", "testuser2"},
			want:     true,
		},
		{
			name:     "match2",
			args:     args{match: "testuser"},
			input:    "test",
			mentions: []string{"testuser1", "testuser2"},
			want:     false,
		},
		{
			name:     "empty",
			args:     args{match: "testuser"},
			input:    "test",
			mentions: []string{},
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchMentioned(tt.args.match)(createMessageContextWithMentions(tt.input, tt.mentions))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MatchMentioned() = %v, want %v", got, tt.want)
			}
		})
	}
}
