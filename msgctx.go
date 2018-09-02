package disgobol

import (
	"errors"
	"reflect"
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
	ErrNoCtxForEvt = "No MsgContext could be created for the given type"
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

func (msgCtx *MsgContext) RespondSimple(msg string) (*MsgContext, error) {
	rmsg, err := msgCtx.Session.ChannelMessageSend(msgCtx.ChannelID, msg)
	return &MsgContext{Message: rmsg, Session: msgCtx.Session}, err
}

func (msgCtx *MsgContext) RespondError(msg string, srcerr error) (*MsgContext, error) {
	msg = "```LDIF\nError: " + msg + ": " + srcerr.Error() + "```"
	rmsg, err := msgCtx.Session.ChannelMessageSend(msgCtx.ChannelID, msg)
	return &MsgContext{Message: rmsg, Session: msgCtx.Session}, err
}

func (msgCtx *MsgContext) EditSimple(newCont string) (*MsgContext, error) {
	rmsg, err := msgCtx.Session.ChannelMessageEdit(msgCtx.ChannelID, msgCtx.ID, newCont)
	return &MsgContext{Message: rmsg, Session: msgCtx.Session}, err
}

func (msgCtx *MsgContext) ParseArgs(args *map[string]interface{}) error {
	whitespaceIndx := strings.IndexAny(msgCtx.Content, " \n\t")
	if whitespaceIndx < 0 {
		whitespaceIndx = len(msgCtx.Content)
	}
	// remove the actual command
	argstr := strings.TrimSpace(strings.TrimPrefix(msgCtx.Content, msgCtx.Content[:whitespaceIndx]))
	parsedArgs, err := shellquote.Split(argstr)
	tmp := make([]interface{}, len(parsedArgs))
	for i, v := range parsedArgs {
		tmp[i] = v
	}
	msgCtx.Args = tmp

	return err
}
