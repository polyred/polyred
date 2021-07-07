// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gui

type dispatcher struct {
	eventMap map[EventName][]subscription
}

func newDispatcher() *dispatcher {
	return &dispatcher{
		eventMap: map[EventName][]subscription{},
	}
}

type EventCallBack func(Event)

type subscription struct {
	id interface{}
	cb EventCallBack
}

func (d *dispatcher) Dispatch(eventName EventName, ev Event) int {
	subs := d.eventMap[eventName]
	nsubs := len(subs)
	if nsubs == 0 {
		return 0
	}

	for _, s := range subs {
		s.cb(ev)
	}
	return nsubs
}
func (d *dispatcher) Subscribe(eventName EventName, cb EventCallBack) {
	d.eventMap[eventName] = append(d.eventMap[eventName], subscription{
		id: eventName,
		cb: cb,
	})
}
