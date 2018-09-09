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

func TestMsgContext_ParseArgs(t *testing.T) {
	type args struct {
		args *map[string]interface{}
	}

	// TODO: Add more testcases
	tests := []struct {
		name    string
		ctx     MsgContext
		args    args
		want    []interface{}
		wantErr bool
	}{
		{
			name: "simple",
			ctx: MsgContext{
				Message: &discordgo.Message{
					Content: "cmd 345 false 2.54",
				},
			},
			args: args{
				args: &map[string]interface{}{"int": 123, "bool": true, "float": 1.234},
			},
			want:    []interface{}{345, false, 2.54},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ctx.ParseArgs(tt.args.args)
			if (err != nil) != tt.wantErr && !reflect.DeepEqual(tt.ctx.Args, tt.want) {
				t.Errorf("MsgContext.ParseArgs() error = %v, wantErr %v, want: %v, got: %v", err, tt.wantErr, tt.want, tt.ctx.Args)
			}
		})
	}
}
