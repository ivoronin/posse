package main

import "log"

type PeerStatus int

const (
	Init PeerStatus = iota
	Down
	Up
)

var peerStatus PeerStatus

func (p PeerStatus) String() string {
	status := []string{"init", "down", "up"}
	return status[p]
}

func (p PeerStatus) Log() {
	log.Printf("peer status: %s", p)
}
