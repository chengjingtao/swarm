package eventemitter

import (
	"fmt"
)

type EventHandler func(string, interface{})

var evtHanlders map[string][]EventHandler

func init() {
	evtHanlders = map[string][]EventHandler{}
}

//trigger one event
func Emit(evt string, args interface{}) {
	handlers, ok := evtHanlders[evt]
	//fmt.Println(ok)
	//fmt.Println(handlers)

	if ok == false || len(handlers) == 0 {
		return
	}

	go func() {
		for _, handle := range handlers {
			if handle == nil {
				continue
			}

			handle(evt, args)
		}
	}()
}

func On(evt string, handler EventHandler) {

	if _, ok := evtHanlders[evt]; ok == false {
		evtHanlders[evt] = []EventHandler{}
	}
	evtHanlders[evt] = append(evtHanlders[evt], handler)
}

func Off(evt string, handler EventHandler) {

	var p = fmt.Sprintf("%x", handler)
	if _, ok := evtHanlders[evt]; ok == false {
		return
	}

	var copyHandlers = []EventHandler{}

	var todel = -1
	for index, h := range evtHanlders[evt] {
		copyHandlers = append(copyHandlers, h)

		if todel == -1 && fmt.Sprintf("%x", handler) == p {
			todel = index
		}
	}

	if todel == -1 {
		return
	}

	evtHanlders[evt] = []EventHandler{}
	evtHanlders[evt] = append(evtHanlders[evt], copyHandlers[0:todel]...)
	evtHanlders[evt] = append(evtHanlders[evt], copyHandlers[1:]...)
}
