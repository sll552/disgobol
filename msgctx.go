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
	Args      []interface{}
}

const (
	ErrNoCtxForEvt      = "No MsgContext could be created for the given type"
	ErrTooManyArguments = "Too many arguments provided"
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

// ParseArgs parses arguments from the current message and filling the MsgContext.Args array
// according to the given argument map.
// An Error is returned if argument parsing fails or the types do not match
func (msgCtx *MsgContext) ParseArgs(args *map[string]interface{}) error {
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
		return errors.New(ErrTooManyArguments)
	}

	// Cast args to their required type
	i := 0
	for _, arg := range *args {
		switch arg.(type) {
		case int:
			tmp[i], err = strconv.ParseInt(tmp[i].(string), 0, 64)
			if err != nil {
				return err
			}
		case bool:
			tmp[i], err = strconv.ParseBool(tmp[i].(string))
			if err != nil {
				return err
			}
		case float32:
			tmp[i], err = strconv.ParseFloat(tmp[i].(string), 32)
			if err != nil {
				return err
			}
		case float64:
			tmp[i], err = strconv.ParseFloat(tmp[i].(string), 64)
			if err != nil {
				return err
			}
		case uint:
			tmp[i], err = strconv.ParseUint(tmp[i].(string), 0, 64)
			if err != nil {
				return err
			}
		}
		i++
	}
	msgCtx.Args = tmp

	return err
}
