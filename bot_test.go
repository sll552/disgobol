package disgobol

import (
	"sync/atomic"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestBot_validate(t *testing.T) {
	tests := []struct {
		name    string
		bot     *Bot
		wantErr bool
	}{
		{name: "empty", bot: &Bot{}, wantErr: true},
		{name: "tokenInv", bot: &Bot{Token: "123123123", CommandPrefix: "!"}, wantErr: true},
		// This is just a random string generated, so don't try to connect to discord with it ;)
		{
			name: "valid",
			bot: &Bot{
				Token:         "GkJyfEUNRbEY3-p67QrBfEwLf.zApBkxyhpqGvZk-1Mg-J7TUE6VmMLsOgn",
				CommandPrefix: "!",
			},
			wantErr: false,
		},
		{
			name: "prefixInv1",
			bot: &Bot{
				Token:         "GkJyfEUNRbEY3-p67QrBfEwLf.zApBkxyhpqGvZk-1Mg-J7TUE6VmMLsOgn",
				CommandPrefix: "a",
			},
			wantErr: true,
		},
		{
			name: "prefixInv2",
			bot: &Bot{
				Token:         "GkJyfEUNRbEY3-p67QrBfEwLf.zApBkxyhpqGvZk-1Mg-J7TUE6VmMLsOgn",
				CommandPrefix: "!asd",
			},
			wantErr: true},
		{
			name: "prefixInv3",
			bot: &Bot{
				Token:         "GkJyfEUNRbEY3-p67QrBfEwLf.zApBkxyhpqGvZk-1Mg-J7TUE6VmMLsOgn",
				CommandPrefix: "asd",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.bot.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Bot.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBot_GetInviteURL(t *testing.T) {
	tests := []struct {
		name    string
		bot     *Bot
		want    string
		wantErr bool
	}{
		{name: "empty", bot: &Bot{}, want: "", wantErr: true},
		{
			name:    "invurl",
			bot:     &Bot{ClientID: "123123123123123123"},
			want:    "https://discordapp.com/api/v6/oauth2/authorize?client_id=123123123123123123&scope=bot",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.bot.GetInviteURL()
			if (err != nil) != tt.wantErr {
				t.Errorf("Bot.GetInviteURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Bot.GetInviteURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBot_AddCommand(t *testing.T) {
	bot := Bot{
		CommandPrefix: "!",
		messageRouter: MsgRouter{},
	}

	err := bot.AddCommand("testcmd", "a testcommand", func(msg MsgContext) string { return "" })
	containscmd := false
	for _, cmd := range bot.commands {
		if cmd.Name == "!testcmd" {
			containscmd = true
		}
	}
	if containscmd {
		cmdrt := bot.messageRouter.GetRoute("testcmd")
		if cmdrt == nil {
			containscmd = false
		}
	}
	if !containscmd || err != nil {
		t.Errorf("Bot.AddCommand() does not contain command, error: %v", err)
	}

}

func TestBot_messageCreateHandler(t *testing.T) {
	type args struct {
		s *discordgo.Session
		m *discordgo.MessageCreate
	}
	handlerCalled := int32(0)
	tests := []struct {
		name       string
		bot        *Bot
		args       args
		wantCalled int32
	}{
		{
			name: "handlerCalled1",
			bot: &Bot{
				messageRouter: MsgRouter{
					routes: []*MsgRoute{
						{
							ID: "test",
							Action: func(msg MsgContext) {
								atomic.AddInt32(&handlerCalled, 1)
							},
							Matches: MatchStart("test"),
						},
					},
				},
			},
			args: args{
				m: &discordgo.MessageCreate{
					Message: &discordgo.Message{
						Content: "test message to trigger command",
						Author: &discordgo.User{
							ID:       "123123",
							Username: "testuser",
							Bot:      false,
						},
					},
				},
				s: &discordgo.Session{},
			},
			wantCalled: 1,
		},
		{
			name: "handlerCalled2",
			bot: &Bot{
				messageRouter: MsgRouter{
					routes: []*MsgRoute{
						{
							ID: "test",
							Action: func(msg MsgContext) {
								atomic.AddInt32(&handlerCalled, 1)
							},
							Matches: MatchStart("test"),
						},
					},
				},
			},
			args: args{
				m: &discordgo.MessageCreate{
					Message: &discordgo.Message{
						Content: "test message to trigger command",
						Author: &discordgo.User{
							ID:       "123123",
							Username: "testuser",
							Bot:      true,
						},
					},
				},
				s: &discordgo.Session{},
			},
			wantCalled: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled = int32(0)
			tt.bot.messageCreateHandler(tt.args.s, tt.args.m)
			if tt.wantCalled != handlerCalled {
				t.Errorf("Bot.messageCreateHandler() handler called %v times, want: %v", handlerCalled, tt.wantCalled)
			}
		})
	}
}

func TestBot_Run(t *testing.T) {
	errAuthFailed := "websocket: close 4004: Authentication failed."
	emptySession, _ := discordgo.New()

	tests := []struct {
		name    string
		bot     *Bot
		wantErr bool
		errTxt  string
	}{
		{
			name: "invToken",
			bot: &Bot{
				ClientID:      "123123123123123123",
				Token:         "asdasd",
				CommandPrefix: "!",
			},
			wantErr: true,
			errTxt:  ErrTokenInvalid,
		},
		{
			name: "invClientID",
			bot: &Bot{
				ClientID:      "12312",
				Token:         "asdasdasdasdasdasdasdasdasdasdasdaasdasdsdasdasdasdasdasdas",
				CommandPrefix: "!",
			},
			wantErr: true,
			// Client ID is only required for the invite URL so we expect an Auth fail here because of the fake token
			errTxt: errAuthFailed,
		},
		{
			name: "invPrefix",
			bot: &Bot{
				ClientID:      "123123123123123123",
				Token:         "asdasdasdasdasdasdasdasdasdasdasdaasdasdsdasdasdasdasdasdas",
				CommandPrefix: "asd",
			},
			wantErr: true,
			errTxt:  ErrPrefixNotSupported,
		},
		{
			name: "invSession",
			bot: &Bot{
				ClientID:       "123123123123123123",
				Token:          "asdasdasdasdasdasdasdasdasdasdasdaasdasdsdasdasdasdasdasdas",
				CommandPrefix:  "!",
				DiscordSession: emptySession,
			},
			wantErr: true,
			errTxt:  errAuthFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bot.Run()
			if (err != nil) != tt.wantErr || tt.errTxt != err.Error() {
				t.Errorf("Bot.Run() error = %v, wantErr %v, expected %v", err, tt.wantErr, tt.errTxt)
				return
			}
		})
	}
}
