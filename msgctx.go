package disgobol

import (
	"errors"
	"reflect"

	"github.com/bwmarrin/discordgo"
)

type MsgContext struct {
	*discordgo.Message
	EventType interface{}
	Session   *discordgo.Session
}

const (
	ErrNoCtxForEvt = "No MsgContext could be created for the given type"
)

// Creates a MsgContext instance from a discordgo event
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

func (msgCtx *MsgContext) EditSimple(newCont string) (*MsgContext, error) {
	rmsg, err := msgCtx.Session.ChannelMessageEdit(msgCtx.ChannelID, msgCtx.ID, newCont)
	return &MsgContext{Message: rmsg, Session: msgCtx.Session}, err
}
