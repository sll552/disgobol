package disgobol

import (
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestMsgRouter_GetRoute(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		rt   *MsgRouter
		args args
		want *MsgRoute
	}{
		{name: "empty", rt: &MsgRouter{}, args: args{id: "test"}, want: nil},
		{
			name: "single",
			rt:   &MsgRouter{routes: []*MsgRoute{{ID: "test"}}},
			args: args{id: "test"},
			want: &MsgRoute{ID: "test"},
		},
		{
			name: "multi",
			rt:   &MsgRouter{routes: []*MsgRoute{{ID: "test"}, {ID: "test123"}, {ID: "test1234"}}},
			args: args{id: "test123"},
			want: &MsgRoute{ID: "test123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rt.GetRoute(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MsgRouter.GetRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMsgRouter_AddRoute(t *testing.T) {
	type args struct {
		nr MsgRoute
	}
	tests := []struct {
		name    string
		rt      *MsgRouter
		args    args
		want    *MsgRoute
		wantErr bool
	}{
		{
			name: "empty",
			rt:   &MsgRouter{},
			args: args{
				nr: MsgRoute{
					ID: "test1",
				},
			},
			want: &MsgRoute{
				ID: "test1",
			},
			wantErr: false,
		},
		{
			name: "exists",
			rt:   &MsgRouter{routes: []*MsgRoute{{ID: "test1"}, {ID: "test123"}, {ID: "test1234"}}},
			args: args{
				nr: MsgRoute{
					ID:      "test1",
					Matches: MatchStart("test1"),
				},
			},
			want: &MsgRoute{
				ID: "test1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.rt.AddRoute(tt.args.nr)
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgRouter.AddRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) || !reflect.DeepEqual(tt.rt.GetRoute(tt.args.nr.ID), tt.want) {
				t.Errorf("MsgRouter.AddRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMsgRouter_Route(t *testing.T) {
	rt := MsgRouter{}
	mctx := MsgContext{
		Message: &discordgo.Message{
			Content: "testmessage asd def test simple",
		},
	}
	handler1Called := int32(0)
	handler2Called := int32(0)
	handler3Called := int32(0)
	handler1 := func(msg MsgContext) {
		atomic.AddInt32(&handler1Called, 1)
	}
	handler2 := func(msg MsgContext) {
		atomic.AddInt32(&handler2Called, 1)
	}
	handler3 := func(msg MsgContext) {
		atomic.AddInt32(&handler3Called, 1)
	}

	// Call on empty router should yield an error and no handlers called
	err := rt.Route(mctx)
	if err == nil || handler1Called != 0 || handler2Called != 0 || handler3Called != 0 {
		t.Errorf("MsgRouter.Route() error = %v, wantErr %v", err, true)
	}

	// Add a single route that should match
	// #nosec G104
	_, _ = rt.AddRoute(MsgRoute{
		Action:  handler1,
		Matches: MatchStart("testmessage"),
		ID:      "handler1",
	})
	err = rt.Route(mctx)
	if err != nil {
		t.Errorf("MsgRouter.Route() error = %v, wantErr %v", err, false)
		return
	}
	if handler1Called != 1 || handler2Called != 0 || handler3Called != 0 {
		t.Errorf("MsgRouter.Route() handler1Called = %v, expected %v", handler1Called, 1)
	}

	// Add a second route that should match
	// #nosec G104
	_, _ = rt.AddRoute(MsgRoute{
		Action:  handler2,
		Matches: MatchContains("test"),
		ID:      "handler2",
	})
	err = rt.Route(mctx)
	if err != nil {
		t.Errorf("MsgRouter.Route() error = %v, wantErr %v", err, false)
		return
	}
	if handler1Called != 2 || handler2Called != 1 || handler3Called != 0 {
		t.Errorf("MsgRouter.Route() handler1Called = %v, expected %v", handler2Called, 1)
	}

	// Clear Router
	rt = MsgRouter{}
	handler1Called = int32(0)
	handler2Called = int32(0)
	handler3Called = int32(0)
	// Add route that doesnt match
	// #nosec G104
	_, _ = rt.AddRoute(MsgRoute{
		Action:  handler1,
		Matches: MatchStart("simple"),
		ID:      "handler1",
	})
	err = rt.Route(mctx)
	if err == nil || handler1Called != 0 || handler2Called != 0 || handler3Called != 0 {
		t.Errorf("MsgRouter.Route() error = %v, wantErr %v", err, true)
	}
	// Add default route that matches
	rt.DefaultRoute = &MsgRoute{
		Action:  handler3,
		Matches: MatchStart("testmessage"),
		ID:      "default",
	}
	err = rt.Route(mctx)
	if err != nil {
		t.Errorf("MsgRouter.Route() error = %v, wantErr %v", err, false)
		return
	}
	if handler1Called != 0 || handler2Called != 0 || handler3Called != 1 {
		t.Errorf("MsgRouter.Route() handler1Called = %v, expected %v", handler3Called, 1)
	}
}
