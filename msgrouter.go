package disgobol

import (
	"errors"
	"sync"
)

type MsgRouter struct {
	routes       []*MsgRoute
	DefaultRoute *MsgRoute
	rtListMutex  sync.Mutex
}
type MsgRoute struct {
	ID      string
	Matches func(*MsgContext) bool
	Action  func(MsgContext)
}

const (
	ErrNoRoutes           = "no routes defined"
	ErrNoRoute            = "no route matched"
	ErrRouteAlreadyExsits = "route already exists"
)

// GetRoute gets a route by its ID, default route does not count as route in this case
// if no route is found with the given ID nil is returned
func (rt *MsgRouter) GetRoute(id string) *MsgRoute {
	rt.rtListMutex.Lock()
	defer rt.rtListMutex.Unlock()
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
	rt.rtListMutex.Lock()
	rt.routes = append(rt.routes, &nr)
	rt.rtListMutex.Unlock()
	return &nr, nil
}

// Route finds all matching routes for the given MsgContext and executes their Action function,
// if no route matches the default route is checked and executed if it matches
// returns errors if no routes are defined or no routes match
func (rt *MsgRouter) Route(msg MsgContext) error {
	rt.rtListMutex.Lock()
	rtCnt := len(rt.routes)
	rt.rtListMutex.Unlock()

	if rtCnt == 0 {
		return errors.New(ErrNoRoutes)
	}
	rtMatched := false
	rtMatch := []*MsgRoute{}

	// first determine the routes that match
	rt.rtListMutex.Lock()
	for _, r := range rt.routes {
		if r.Matches(&msg) {
			rtMatch = append(rtMatch, r)
			rtMatched = true
		}
	}
	if !rtMatched && rt.DefaultRoute != nil && rt.DefaultRoute.Matches(&msg) {
		rtMatch = append(rtMatch, rt.DefaultRoute)
	}
	rt.rtListMutex.Unlock()

	// then execute their actions
	for _, r := range rtMatch {
		r.Action(msg)
	}
	if len(rtMatch) > 0 {
		return nil
	}

	return errors.New(ErrNoRoute)
}
