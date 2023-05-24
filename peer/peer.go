package peer

import (
	"log"

	"github.com/ivoronin/posse/fsm"
	"github.com/ivoronin/posse/metrics"
)

const (
	StateRxInit fsm.State = iota
	StateRxDown
	StateRxUp
	StateRxError
)

const (
	StateTxUp fsm.State = iota
	StateTxIdle
	StateTxError
)

var stateRxNames = map[fsm.State]string{
	StateRxInit:  "init",
	StateRxDown:  "down",
	StateRxUp:    "up",
	StateRxError: "error",
}

var stateTxNames = map[fsm.State]string{
	StateTxUp:    "up",
	StateTxIdle:  "idle",
	StateTxError: "error",
}

const (
	EventBlockReadErr fsm.Event = iota
	EventBlockReadStale
	EventBlockReadNew
)

const (
	EventBlockWritten fsm.Event = iota
	EventBlockWriteErr
	EventBlockWriteSkipped
)

type Peer struct {
	RxFSM *fsm.FSM
	TxFSM *fsm.FSM
}

func rxStateChanged(from fsm.State, to fsm.State, evt fsm.Event) {
	if from == to {
		return
	}
	metrics.PeerRxState.Set(float64(to))
	log.Printf("peer rx status: %s -> %s", stateRxNames[from], stateRxNames[to])
}

func txStateChanged(from fsm.State, to fsm.State, evt fsm.Event) {
	if from == to {
		return
	}
	metrics.PeerTxState.Set(float64(to))
}

func NewPeer(maxStale uint64) *Peer {
	peer := new(Peer)
	AllRxStates := []fsm.State{StateRxInit, StateRxDown, StateRxUp, StateRxError}
	peer.RxFSM = fsm.NewFSM(
		StateRxInit,
		[]fsm.Transition{
			{
				Evt: EventBlockReadErr,
				Src: AllRxStates,
				Dst: StateRxError,
			},
			{
				Evt:      EventBlockReadStale,
				Src:      AllRxStates,
				Dst:      StateRxDown,
				MinTimes: uint(maxStale),
			},
			{
				Evt: EventBlockReadNew,
				Src: AllRxStates,
				Dst: StateRxUp,
			},
		},
		rxStateChanged,
	)

	AllTxStates := []fsm.State{StateTxUp, StateTxIdle, StateTxError}
	peer.TxFSM = fsm.NewFSM(
		StateTxUp,
		[]fsm.Transition{
			{
				Evt: EventBlockWritten,
				Src: AllTxStates,
				Dst: StateTxUp,
			},
			{
				Evt:      EventBlockWriteSkipped,
				Src:      AllTxStates,
				Dst:      StateTxIdle,
				MinTimes: uint(maxStale - 1),
			},
			{
				Evt: EventBlockWriteErr,
				Src: AllTxStates,
				Dst: StateTxError,
			},
		},
		txStateChanged,
	)
	return peer
}
