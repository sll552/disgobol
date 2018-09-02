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
	Action  func(MsgContext)
}

const (
	ErrNoRoutes           = "No routes defined"
	ErrNoRoute            = "No route matched"
	ErrRouteAlreadyExsits = "Route already exists"
)

// GetRoute gets a route by its ID, default route does not count as route in this case
// if no route is found with the given ID nil is returned
func (rt *MsgRouter) GetRoute(id string) *MsgRoute {
	for _, r := range rt.routes {
		if r.ID == id {
			return r
		}
	}
	return nil
}

// AddRoute adds a route to this router, returns the route added or if a route already exists
// the existing route and an error
func (rt *MsgRouter) AddRoute(nr MsgRoute) (*MsgRoute, error) {
	if r := rt.GetRoute(nr.ID); r != nil {
		return r, errors.New(ErrRouteAlreadyExsits)
	}
	rt.routes = append(rt.routes, &nr)
	return &nr, nil
}

// Route finds all matching routes for the given MsgContext and execute their Action function,
// if no route matches the default route is checked and executed if it matches
// returns errors if no routes are defined or no routes match
func (rt *MsgRouter) Route(msg MsgContext) error {
	if len(rt.routes) == 0 {
		return errors.New(ErrNoRoutes)
	}
	rtMatch := false
	for _, r := range rt.routes {
		if r.Matches(&msg) {
			r.Action(msg)
			rtMatch = true
		}
	}
	if rtMatch {
		return nil
	} else if rt.DefaultRoute != nil && rt.DefaultRoute.Matches(&msg) {
		rt.DefaultRoute.Action(msg)
		return nil
	}
	return errors.New(ErrNoRoute)
}
