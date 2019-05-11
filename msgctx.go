package disgobol

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
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

// RespondError responds to the message in this context with an error message that is specially formated
// the error object can be nil if not the message will be printed too
// Returns a MsgContext instance for the error message created
func (msgCtx *MsgContext) RespondError(msg string, srcerr error) (*MsgContext, error) {
	msg = "```LDIF\nError: " + msg
	if srcerr != nil {
		msg += ": " + srcerr.Error()
	}
	msg += "```"
	rmsg, err := msgCtx.Session.ChannelMessageSend(msgCtx.ChannelID, msg)
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
func (msgCtx *MsgContext) ParseArgs(args *[]CommandArg) error {
	// Dont do anything if no arguments are defined
	if len(*args) == 0 {
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
	if len(tmp) != len(*args) {
		return errors.New(ErrArgCntMismatch)
	}

	// Cast args to their required type
	for idx, arg := range *args {
		switch arg.Example.(type) {
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
	for idx, arg := range *args {
		if len(arg.Name) == 0 {
			msgCtx.Args[strconv.Itoa(idx)] = tmp[idx]
		} else {
			msgCtx.Args[arg.Name] = tmp[idx]
		}
	}

	return err
}
