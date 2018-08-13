package disgobol

import (
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestNewMsgContext(t *testing.T) {
	inputevt := discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content: "testmsg",
			ID:      "123",
		},
	}
	inputsess := &discordgo.Session{
		State: &discordgo.State{
			Ready: discordgo.Ready{},
		},
		Debug: true,
	}

	want := MsgContext{
		EventType: reflect.TypeOf(inputevt),
		Message:   inputevt.Message,
		Session:   inputsess,
	}

	got, err := NewMsgContext(inputevt, inputsess)
	if !reflect.DeepEqual(got, want) || err != nil {
		t.Errorf("NewMsgContext() = %v, want %v", got, want)
	}

	// Unsupported event should error and give empty ctx
	inputevt2 := discordgo.MessageSend{}
	got, err = NewMsgContext(inputevt2, inputsess)
	if err == nil || !reflect.DeepEqual(got, MsgContext{}) {
		t.Errorf("NewMsgContext() error = %v, wantErr %v", err, true)
		return
	}
}
