package disgobol

import (
	"errors"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/kballard/go-shellquote"
)

type MsgContext struct {
	*discordgo.Message
	EventType interface{}
	Session   *discordgo.Session
	Args      map[string]interface{}
}

type CommandArg struct {
	Name        string
	Description string
	Example     interface{}
}

const (
	ErrNoCtxForEvt    = "no MsgContext could be created for the given type"
	ErrArgCntMismatch = "argument count does not match expected number of arguments"
)

// NewMsgContext creates a MsgContext instance from a discordgo event
func NewMsgContext(evt interface{}, session *discordgo.Session) (MsgContext, error) {
	switch evt.(type) {
	case discordgo.MessageCreate:
		e := evt.(discordgo.MessageCreate)
		return MsgContext{Message: e.Message, EventType: reflect.TypeOf(e), Session: session}, nil
	}

	return MsgContext{}, errors.New(ErrNoCtxForEvt)
}

// RespondSimple is a convenience function to respond to the message in this context
// Returns a MsgContext instance of the created response
func (msgCtx *MsgContext) RespondSimple(msg string) (*MsgContext, error) {
	rmsg, err := msgCtx.Session.ChannelMessageSend(msgCtx.ChannelID, msg)
	return &MsgContext{Message: rmsg, Session: msgCtx.Session}, err
}

func generateErrorString(msg string, srcerr error) string {
	var genmsg strings.Builder

	genmsg.WriteString("```LDIF\nError: ")

	if len(msg) != 0 {
		genmsg.WriteString(msg)
		genmsg.WriteString(": ")
	}

	if srcerr != nil {
		genmsg.WriteString(srcerr.Error())
	} else {
		genmsg.WriteString("an undefined error occurred")
	}

	genmsg.WriteString("```")

	return genmsg.String()
}

// RespondError responds to the message in this context with an error message that is specially formated
// Returns a MsgContext instance for the error message created
func (msgCtx *MsgContext) RespondError(msg string, srcerr error) (*MsgContext, error) {
	rmsg, err := msgCtx.Session.ChannelMessageSend(msgCtx.ChannelID, generateErrorString(msg, srcerr))
	return &MsgContext{Message: rmsg, Session: msgCtx.Session}, err
}

// EditSimple edits the message in this context with the content provided (replacing the original content)
// Returns a MsgContext instance of the edited message
func (msgCtx *MsgContext) EditSimple(newCont string) (*MsgContext, error) {
	rmsg, err := msgCtx.Session.ChannelMessageEdit(msgCtx.ChannelID, msgCtx.ID, newCont)
	return &MsgContext{Message: rmsg, Session: msgCtx.Session}, err
}

// ParseArgs parses arguments from the current message and fills the MsgContext.Args array
// according to the given argument map.
// An Error is returned if argument parsing fails or the types do not match
func (msgCtx *MsgContext) ParseArgs(args []CommandArg) error {
	// Dont do anything if no arguments are defined
	if len(args) == 0 {
		return nil
	}

	whitespaceIndx := strings.IndexAny(msgCtx.Content, " \n\t")
	if whitespaceIndx < 0 {
		whitespaceIndx = len(msgCtx.Content)
	}
	// remove the actual command
	argstr := strings.TrimSpace(strings.TrimPrefix(msgCtx.Content, msgCtx.Content[:whitespaceIndx]))
	parsedArgs, err := shellquote.Split(argstr)

	if err != nil {
		return err
	}

	tmp := make([]interface{}, len(parsedArgs))

	for i, v := range parsedArgs {
		tmp[i] = v
	}

	if len(tmp) != len(args) {
		return errors.New(ErrArgCntMismatch)
	}

	// Cast args to their required type
	for idx := range args {
		switch args[idx].Example.(type) {
		case int:
			tmp[idx], err = strconv.ParseInt(tmp[idx].(string), 0, 64)
		case bool:
			tmp[idx], err = strconv.ParseBool(tmp[idx].(string))
		case float64:
			tmp[idx], err = strconv.ParseFloat(tmp[idx].(string), 64)
		}

		if err != nil {
			return err
		}
	}
	// build the resulting map and use index for arguments without a name
	msgCtx.Args = make(map[string]interface{})

	for idx := range args {
		if len(args[idx].Name) == 0 {
			msgCtx.Args[strconv.Itoa(idx)] = tmp[idx]
		} else {
			msgCtx.Args[args[idx].Name] = tmp[idx]
		}
	}

	return err
}

// GetChannel returns the discordgo Channel of this message
func (msgCtx *MsgContext) GetChannel() (*discordgo.Channel, error) {
	return msgCtx.Session.State.Channel(msgCtx.ChannelID)
}

// GetGuild returns the discordgo Guild of this message
func (msgCtx *MsgContext) GetGuild() (*discordgo.Guild, error) {
	c, err := msgCtx.GetChannel()
	if err != nil {
		return nil, err
	}

	return msgCtx.Session.State.Guild(c.GuildID)
}

// JoinVoiceChannel joins the voice channel with the given id and returns a discordgo.VoiceConnection on success
func (msgCtx *MsgContext) JoinVoiceChannel(chanid string) (*discordgo.VoiceConnection, error) {
	var guid string

	guild, err := msgCtx.GetGuild()

	if err != nil {
		return nil, err
	}

	guid = guild.ID

	return msgCtx.Session.ChannelVoiceJoin(guid, chanid, false, true)
}

// JoinUserVoiceChannel joins the voice channel of in which the user with the given userid resides
// and returns a discordgo.VoiceConnection on success
func (msgCtx *MsgContext) JoinUserVoiceChannel(userid string) (*discordgo.VoiceConnection, error) {
	var chanid string

	guild, err := msgCtx.GetGuild()

	if err != nil {
		return nil, err
	}

	for _, vs := range guild.VoiceStates {
		if vs.UserID == userid {
			chanid = vs.ChannelID
		}
	}

	return msgCtx.JoinVoiceChannel(chanid)
}

// PlayDCA streams the given dca data to the given discordgo.VoiceConnection
func (msgCtx *MsgContext) PlayDCA(data io.Reader, channel *discordgo.VoiceConnection) error {
	decoder := dca.NewDecoder(data)
	finished := make(chan error)
	// TODO: lock voice channel globally to avoid overlay playing
	dca.NewStream(decoder, channel, finished)
	err := <-finished

	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

// PlayDCAFile reads a given dca file from disk and streams it to the given discordgo.VoiceConnection
func (msgCtx *MsgContext) PlayDCAFile(path string, channel *discordgo.VoiceConnection) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	return msgCtx.PlayDCA(file, channel)
}
