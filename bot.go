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
	Name           string
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
	Args        []CommandArg
	Function    func(MsgContext) string
}

const (
	ErrClientIDInvalid    = "the provided Client ID is invalid"
	ErrTokenInvalid       = "the provided Token is invalid"
	ErrPrefixNotSupported = "the provided Prefix is not supported, " +
		"please choose a non-word character or string (regexp: '\\W+')"
	TokenLength    = 59
	ClientIDLength = 18
)

func (bot *Bot) validate() error {
	//Token is always 59 chars long
	if len(bot.Token) != TokenLength {
		return errors.New(ErrTokenInvalid)
	}

	r := regexp.MustCompile(`\w+`)

	if !(len(bot.CommandPrefix) > 0) || r.MatchString(bot.CommandPrefix) {
		return errors.New(ErrPrefixNotSupported)
	}

	return nil
}

// GetInviteURL returns the invite URL for this bot if a valid client id was provided
func (bot *Bot) GetInviteURL() (string, error) {
	if len(bot.ClientID) != ClientIDLength {
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
		// #nosec G104
		Action: func(msg MsgContext) {
			var cont strings.Builder
			cont.WriteString("**These are the available commands for this Bot:**\n```YAML\n")
			for _, cmd := range bot.commands {
				// Cut the prefix for YAML to work
				cont.WriteString(strings.TrimPrefix(cmd.Name, bot.CommandPrefix) + ": ")
				cont.WriteString(cmd.Description + "\n")
				if len(cmd.Args) > 0 {
					cont.WriteString("  Arguments:\n")
					for argidx := range cmd.Args {
						carg := &cmd.Args[argidx]
						cont.WriteString("    - " + carg.Name + ": ")
						cont.WriteString(carg.Description + "\n")
					}
				}
			}
			// Always add help command
			cont.WriteString("help: Display this message\n")
			cont.WriteString("```Invoke the commands with prefix `" +
				bot.CommandPrefix + "`" + " (e.g. `" + bot.CommandPrefix + "somecmd`)")
			_, _ = msg.RespondSimple(cont.String())
		}}
}

// AddCommand adds a new command to the bot
func (bot *Bot) AddCommand(bcmd BotCommand) error {
	brt := MsgRoute{
		ID:      bcmd.Name,
		Matches: MatchStartWord(bot.CommandPrefix + bcmd.Name),
		// #nosec G104
		Action: func(msg MsgContext) {
			err := msg.ParseArgs(bcmd.Args)
			if err != nil {
				_, _ = msg.RespondError("Parsing arguments failed", err)
				return
			}
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

// Run starts the bot and does not return until an os signal is received (e.g. SIGTERM)
// it should be called last in your bot implementation
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
	_, err = bot.messageRouter.AddRoute(helprt)

	if err != nil {
		return err
	}
	// This will match only if no other route matches, so match for the prefix to catch all commands that don't exist
	helprt.Matches = MatchStart(bot.CommandPrefix)
	bot.messageRouter.DefaultRoute = &helprt

	// Add Handlers
	bot.DiscordSession.AddHandler(bot.readyHandler)
	bot.DiscordSession.AddHandler(bot.messageCreateHandler)
	bot.DiscordSession.AddHandler(bot.guildCreateHandler)

	// Handle OS signals
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	err = bot.DiscordSession.Open()
	defer bot.DiscordSession.Close()

	if err != nil {
		return err
	}

	// Wait for Signal
	<-sc

	return nil
}

func (bot *Bot) readyHandler(s *discordgo.Session, r *discordgo.Ready) {
	_ = s.UpdateStatus(0, bot.CommandPrefix+"help")
}

func (bot *Bot) guildCreateHandler(s *discordgo.Session, g *discordgo.GuildCreate) {
	if g.Guild.Unavailable || len(bot.Name) == 0 {
		return
	}

	for _, channel := range g.Guild.Channels {
		if channel.ID == g.Guild.ID {
			var greetstr strings.Builder

			greetstr.WriteString(bot.Name)
			greetstr.WriteString(" is ready on this server!")
			greetstr.WriteString("\nWrite ")
			greetstr.WriteString("`" + bot.CommandPrefix + "help` to see a list of available commands")

			_, _ = s.ChannelMessageSend(channel.ID, greetstr.String())

			return
		}
	}
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
	// #nosec G104
	_ = bot.messageRouter.Route(msgctx)
}
