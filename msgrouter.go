package disgobol

import (
	"errors"
)

type MsgRouter struct {
	routes       []*MsgRoute
	DefaultRoute *MsgRoute
}
type MsgRoute struct {
	ID      string
	Matches func(*MsgContext) bool
	Action  func(MsgContext) error
}

const (
	ErrNoRoutes           = "No routes defined"
	ErrNoRoute            = "No route matched"
	ErrRouteAlreadyExsits = "Route already exists"
)

func (rt *MsgRouter) GetRoute(id string) *MsgRoute {
	for _, r := range rt.routes {
		if r.ID == id {
			return r
		}
	}
	return nil
}

func (rt *MsgRouter) AddRoute(nr MsgRoute) (*MsgRoute, error) {
	if r := rt.GetRoute(nr.ID); r != nil {
		return r, errors.New(ErrRouteAlreadyExsits)
	}
	rt.routes = append(rt.routes, &nr)
	return &nr, nil
}

func (rt *MsgRouter) Route(msg MsgContext) error {
	if len(rt.routes) == 0 {
		return errors.New(ErrNoRoutes)
	}
	for _, r := range rt.routes {
		if r.Matches(&msg) {
			return r.Action(msg)
		}
	}
	if rt.DefaultRoute != nil && rt.DefaultRoute.Matches(&msg) {
		return rt.DefaultRoute.Action(msg)
	}
	return errors.New(ErrNoRoute)
}
