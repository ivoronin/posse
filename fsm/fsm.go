package fsm

type Handler func(fromState State, toState State, evt Event)

type FSM struct {
	CurrentState   State
	transitions    []Transition
	handler        Handler
	lastEvent      Event
	lastEventTimes uint
}

type State string

type Event string

type Transition struct {
	Evt      Event
	Src      []State
	Dst      State
	MinTimes uint
}

func NewFSM(initialState State, transitions []Transition, handler Handler) *FSM {
	fsm := new(FSM)
	fsm.CurrentState = initialState
	fsm.transitions = transitions
	fsm.handler = handler
	return fsm
}

func (f *FSM) makeTransit(trans Transition) {
	fromState := f.CurrentState
	f.CurrentState = trans.Dst
	if f.handler != nil {
		f.handler(fromState, f.CurrentState, trans.Evt)
	}
}

func (f *FSM) Event(evt Event) {
	if f.lastEvent != evt {
		f.lastEvent = evt
		f.lastEventTimes = 0
	}
	f.lastEventTimes++
	for _, t := range f.transitions {
		for _, s := range t.Src {
			if s == f.CurrentState && t.Evt == evt {
				if t.MinTimes != 0 {
					if f.lastEventTimes >= t.MinTimes {
						f.makeTransit(t)
					}
				} else {
					f.makeTransit(t)
				}
				return
			}
		}
	}
	panic("No valid transition found")
}
