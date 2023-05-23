package main

import (
	"log"

	"github.com/ivoronin/posse/fsm"
	"github.com/ivoronin/posse/metrics"
)

const (
	PeerRxStateInit fsm.State = iota
	PeerRxStateDown
	PeerRxStateUp
	PeerRxStateError
)

const (
	PeerTxStateUp fsm.State = iota
	PeerTxStateIdle
	PeerTxStateError
)

var PeerRxStateNames = map[fsm.State]string{
	PeerRxStateInit:  "init",
	PeerRxStateDown:  "down",
	PeerRxStateUp:    "up",
	PeerRxStateError: "error",
}

var PeerTxStateNames = map[fsm.State]string{
	PeerTxStateUp:    "ip",
	PeerTxStateIdle:  "idle",
	PeerTxStateError: "error",
}

const (
	PeerRxEventBlockReadErr fsm.Event = iota
	PeerRxEventBlockReadStale
	PeerRxEventBlockReadNew

	PeerTxEventBlockWritten
	PeerTxEventBlockWriteErr
	PeerTxEventBlockSkipped
)

type Peer struct {
	RxFSM *fsm.FSM
	TxFSM *fsm.FSM
}

func peerRxStateChanged(from fsm.State, to fsm.State, evt fsm.Event) {
	if from == to {
		return
	}
	metrics.PeerRxState.Set(float64(to))
	log.Printf("peer rx status: %s -> %s", PeerRxStateNames[from], PeerRxStateNames[to])
}

func peerTxStateChanged(from fsm.State, to fsm.State, evt fsm.Event) {
	if from == to {
		return
	}
	metrics.PeerTxState.Set(float64(to))
}

func NewPeer(maxStale uint64) *Peer {
	peer := new(Peer)
	AllRxStates := []fsm.State{PeerRxStateInit, PeerRxStateDown, PeerRxStateUp, PeerRxStateError}
	peer.RxFSM = fsm.NewFSM(
		PeerRxStateInit,
		[]fsm.Transition{
			{
				Evt: PeerRxEventBlockReadErr,
				Src: AllRxStates,
				Dst: PeerRxStateError,
			},
			{
				Evt:      PeerRxEventBlockReadStale,
				Src:      AllRxStates,
				Dst:      PeerRxStateDown,
				MinTimes: uint(maxStale),
			},
			{
				Evt: PeerRxEventBlockReadNew,
				Src: AllRxStates,
				Dst: PeerRxStateUp,
			},
		},
		peerRxStateChanged,
	)

	AllTxStates := []fsm.State{PeerTxStateUp, PeerTxStateIdle, PeerTxStateError}
	peer.TxFSM = fsm.NewFSM(
		PeerTxStateUp,
		[]fsm.Transition{
			{
				Evt: PeerTxEventBlockWritten,
				Src: AllTxStates,
				Dst: PeerTxStateUp,
			},
			{
				Evt:      PeerTxEventBlockSkipped,
				Src:      AllTxStates,
				Dst:      PeerTxStateIdle,
				MinTimes: uint(maxStale - 1),
			},
			{
				Evt: PeerTxEventBlockWriteErr,
				Src: AllTxStates,
				Dst: PeerTxStateError,
			},
		},
		peerTxStateChanged,
	)
	return peer
}
