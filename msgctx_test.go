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
		args *[]CommandArg
	}
	var empty map[string]interface{}

	// TODO: Add more testcases
	tests := []struct {
		name    string
		ctx     MsgContext
		args    args
		want    map[string]interface{}
		wantErr bool
		errTxt  string
	}{
		{
			name: "noargs",
			ctx: MsgContext{
				Message: &discordgo.Message{
					Content: "cmd 345 false 2.54 asdasd",
				},
			},
			args: args{
				args: &[]CommandArg{},
			},
			want:    empty,
			wantErr: false,
		},
		{
			name: "simple",
			ctx: MsgContext{
				Message: &discordgo.Message{
					Content: "cmd 345 false 2.54 asdasd",
				},
			},
			args: args{
				args: &[]CommandArg{
					{
						Name:    "int",
						Example: 123,
					},
					{
						Name:    "bool",
						Example: true,
					},
					{
						Name:    "float",
						Example: 1.234,
					},
					{
						Name:    "string",
						Example: "test",
					},
				},
			},
			want:    map[string]interface{}{"int": int64(345), "bool": false, "float": 2.54, "string": "asdasd"},
			wantErr: false,
		},
		{
			name: "cntmismatch",
			ctx: MsgContext{
				Message: &discordgo.Message{
					Content: "cmd 345 false asdasd",
				},
			},
			args: args{
				args: &[]CommandArg{
					{
						Name:    "int",
						Example: 123,
					},
					{
						Name:    "bool",
						Example: true,
					},
					{
						Name:    "float",
						Example: 1.234,
					},
					{
						Name:    "string",
						Example: "test",
					},
				},
			},
			want:    empty,
			wantErr: true,
			errTxt:  ErrArgCntMismatch,
		},
		{
			name: "intparse",
			ctx: MsgContext{
				Message: &discordgo.Message{
					Content: "cmd 3.45 false asdasd",
				},
			},
			args: args{
				args: &[]CommandArg{
					{
						Name:    "int",
						Example: 123,
					},
					{
						Name:    "bool",
						Example: true,
					},
					{
						Name:    "string",
						Example: "test",
					},
				},
			},
			want:    empty,
			wantErr: true,
			errTxt:  "strconv.ParseInt: parsing \"3.45\": invalid syntax",
		},
		{
			name: "boolparse",
			ctx: MsgContext{
				Message: &discordgo.Message{
					Content: "cmd 345 333 asdasd",
				},
			},
			args: args{
				args: &[]CommandArg{
					{
						Name:    "int",
						Example: 123,
					},
					{
						Name:    "bool",
						Example: true,
					},
					{
						Name:    "string",
						Example: "test",
					},
				},
			},
			want:    empty,
			wantErr: true,
			errTxt:  "strconv.ParseBool: parsing \"333\": invalid syntax",
		},
		{
			name: "floatparse",
			ctx: MsgContext{
				Message: &discordgo.Message{
					Content: "cmd 345 asdasd asdasd",
				},
			},
			args: args{
				args: &[]CommandArg{
					{
						Name:    "int",
						Example: 123,
					},
					{
						Name:    "float",
						Example: 1.234,
					},
					{
						Name:    "string",
						Example: "test",
					},
				},
			},
			want:    empty,
			wantErr: true,
			errTxt:  "strconv.ParseFloat: parsing \"asdasd\": invalid syntax",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ctx.ParseArgs(tt.args.args)
			if ((err != nil) != tt.wantErr || (err != nil && tt.errTxt != err.Error())) || !reflect.DeepEqual(tt.ctx.Args, tt.want) {
				t.Errorf("MsgContext.ParseArgs() error = %v, wantErr %v, expectedErr %v, want: %v, got: %v", err, tt.wantErr, tt.errTxt, tt.want, tt.ctx.Args)
			}
		})
	}
}
