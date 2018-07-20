package disgobol

import (
	"errors"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	CommandPrefix  string
	Token          string
	ClientID       string
	DiscordSession *discordgo.Session
	messageRouter  MsgRouter
	commands       []*BotCommand
}

type BotCommand struct {
	Name        string
	Description string
	Route       *MsgRoute
	Function    func(MsgContext) string
}

const (
	ErrTokenInvalid = "The provided Token is invalid"
)

func (bot *Bot) validate() error {
	//Token is always 59 chars long
	if len(bot.Token) != 59 {
		return errors.New(ErrTokenInvalid)
	}
	return nil
}

func (bot *Bot) GetInviteURL() (string, error) {
	u, err := url.Parse(discordgo.EndpointOauth2 + "authorize")
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("client_id", bot.ClientID)
	q.Add("scope", "bot")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (bot *Bot) generateHelpCommand() MsgRoute {
	return MsgRoute{
		ID:      bot.CommandPrefix + "help",
		Matches: MatchStart(bot.CommandPrefix + "help"),
		Action: func(msg MsgContext) error {
			var cont strings.Builder
			cont.WriteString("**These are the available commands for this Bot:**\n```YAML\n")
			for _, cmd := range bot.commands {
				// Cut the prefix for YAML to work
				cont.WriteString(strings.TrimPrefix(cmd.Name, bot.CommandPrefix) + ": ")
				cont.WriteString(cmd.Description + "\n")
			}
			cont.WriteString("```Invoke the commands with prefix `" +
				bot.CommandPrefix + "`" + "  (e.g. `" + bot.CommandPrefix + "somecmd`)")
			_, err := msg.RespondSimple(cont.String())
			return err
		}}
}

// TODO: Arguments for commands
func (bot *Bot) AddCommand(cmd string, desc string, handler func(MsgContext) string) error {
	bcmd := BotCommand{
		Name:        bot.CommandPrefix + cmd,
		Description: desc,
		Function:    handler}
	brt := MsgRoute{
		ID:      cmd,
		Matches: MatchStart(bcmd.Name),
		Action: func(msg MsgContext) error {
			resp := bcmd.Function(msg)
			if len(resp) > 0 {
				_, err := msg.RespondSimple(resp)
				return err
			}
			return nil
		}}
	_, err := bot.messageRouter.AddRoute(brt)
	if err != nil {
		return err
	}
	bot.commands = append(bot.commands, &bcmd)
	return nil
}

func (bot *Bot) Run() error {
	err := bot.validate()
	if err != nil {
		return err
	}

	if bot.DiscordSession == nil {
		bot.DiscordSession, err = discordgo.New("Bot " + bot.Token)
		if err != nil {
			return err
		}
	}

	// Generate default route
	defrt, _ := bot.messageRouter.AddRoute(bot.generateHelpCommand())
	// This will match only if no other route matches, so match for the prefix to catch all commands that don't exist
	defrt.Matches = MatchStart(bot.CommandPrefix)
	bot.messageRouter.DefaultRoute = defrt

	// Add Handlers
	bot.DiscordSession.AddHandler(bot.messageCreateHandler)

	err = bot.DiscordSession.Open()
	defer bot.DiscordSession.Close()
	if err != nil {
		return err
	}

	// Handle OS signals
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	return nil
}

func (bot *Bot) messageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from other bots
	if m.Author.Bot {
		return
	}
	msgctx, err := NewMsgContext(*m, s)
	if err != nil {
		return
	}
	_ = bot.messageRouter.Route(msgctx)
}
