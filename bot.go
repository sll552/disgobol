package disgobol

import (
	"errors"
	"net/url"
	"os"
	"os/signal"
	"regexp"
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
	ErrClientIDInvalid    = "The provided Client ID is invalid"
	ErrTokenInvalid       = "The provided Token is invalid"
	ErrPrefixNotSupported = "The provided Prefix is not supported, " +
		"please choose a non-word character or string (regexp: '\\W+')"
)

func (bot *Bot) validate() error {
	//Token is always 59 chars long
	if len(bot.Token) != 59 {
		return errors.New(ErrTokenInvalid)
	}
	r := regexp.MustCompile(`\w+`)
	if !(len(bot.CommandPrefix) > 0) || r.MatchString(bot.CommandPrefix) {
		return errors.New(ErrPrefixNotSupported)
	}
	return nil
}

func (bot *Bot) GetInviteURL() (string, error) {
	if len(bot.ClientID) != 18 {
		return "", errors.New(ErrClientIDInvalid)
	}
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
		ID:      "help",
		Matches: MatchStart(bot.CommandPrefix + "help"),
		Action: func(msg MsgContext) {
			var cont strings.Builder
			cont.WriteString("**These are the available commands for this Bot:**\n```YAML\n")
			for _, cmd := range bot.commands {
				// Cut the prefix for YAML to work
				cont.WriteString(strings.TrimPrefix(cmd.Name, bot.CommandPrefix) + ": ")
				cont.WriteString(cmd.Description + "\n")
			}
			// Always add help command
			cont.WriteString("help: Display this message\n")
			cont.WriteString("```Invoke the commands with prefix `" +
				bot.CommandPrefix + "`" + "  (e.g. `" + bot.CommandPrefix + "somecmd`)")
			_, _ = msg.RespondSimple(cont.String())
		}}
}

// TODO: Arguments for commands
func (bot *Bot) AddCommand(bcmd BotCommand) error {
	brt := MsgRoute{
		ID:      bcmd.Name,
		Matches: MatchStartWord(bot.CommandPrefix + bcmd.Name),
		Action: func(msg MsgContext) {
			resp := bcmd.Function(msg)
			if len(resp) > 0 {
				_, _ = msg.RespondSimple(resp)
			}
		}}
	cmdrt, err := bot.messageRouter.AddRoute(brt)
	if err != nil {
		return err
	}
	bcmd.Route = cmdrt
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
	helprt := bot.generateHelpCommand()
	_, _ = bot.messageRouter.AddRoute(helprt)
	// This will match only if no other route matches, so match for the prefix to catch all commands that don't exist
	helprt.Matches = MatchStart(bot.CommandPrefix)
	bot.messageRouter.DefaultRoute = &helprt

	// Add Handlers
	bot.DiscordSession.AddHandler(bot.messageCreateHandler)

	// Handle OS signals
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// TODO: Set gamestate to help command
	err = bot.DiscordSession.Open()
	defer bot.DiscordSession.Close()
	if err != nil {
		return err
	}

	_ = bot.DiscordSession.UpdateStatus(0, bot.CommandPrefix+"help")

	// Wait for Signal
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
