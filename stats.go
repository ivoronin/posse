package main

import (
	"log"
	"time"
)

type Stats struct {
	rdBlk      uint64
	rdBlkErr   uint64
	rdBlkMiss  uint64
	rdBlkStale uint64
	rdBlkData  uint64
	rdBlkKeep  uint64
	rdErr      uint64
	wrBlk      uint64
	wrBlkData  uint64
	wrBlkKeep  uint64
	wrErr      uint64
	rxPkt      uint64
	rxErr      uint64
	txPkt      uint64
	txErr      uint64
}

var stats Stats

func reportStats(st *time.Ticker, peer *Peer) {
	var prevStats Stats
	prevTime := time.Now()

	for currTime := range st.C {
		sec := currTime.Sub(prevTime).Seconds()
		log.Printf("stats: rx:%s tx:%s "+
			"rdBlk:%d rdBlkErr:%d rdBlkMiss:%d rdBlkStale:%d rdBlkData:%d rdBlkKeep:%d "+
			"rdErr:%d rdBlk/s:%.2f "+
			"wrBlk:%d wrBlkData:%d wrBlkKeep:%d "+
			"wrErr:%d wrBlk/s:%.2f "+
			"rxPkt:%d rxErr:%d rxPkt/s:%.2f "+
			"txPkt:%d txErr:%d txPkt/s:%.2f",
			peer.RxFSM.CurrentState, peer.TxFSM.CurrentState,
			stats.rdBlk-prevStats.rdBlk,
			stats.rdBlkErr-prevStats.rdBlkErr,
			stats.rdBlkMiss-prevStats.rdBlkMiss,
			stats.rdBlkStale-prevStats.rdBlkStale,
			stats.rdBlkData-prevStats.rdBlkData,
			stats.rdBlkKeep-prevStats.rdBlkKeep,
			stats.rdErr-prevStats.rdErr,
			float64(stats.rdBlk-prevStats.rdBlk)/sec,
			stats.wrBlk-prevStats.wrBlk,
			stats.wrBlkData-prevStats.wrBlkData,
			stats.wrBlkKeep-prevStats.wrBlkKeep,
			stats.wrErr-prevStats.wrErr,
			float64(stats.wrBlk-prevStats.wrBlk)/sec,
			stats.rxPkt-prevStats.rxPkt,
			stats.rxErr-prevStats.rxErr,
			float64(stats.rxPkt-prevStats.rxPkt)/sec,
			stats.txPkt-prevStats.txPkt,
			stats.txErr-prevStats.txErr,
			float64(stats.txPkt-prevStats.txPkt)/sec,
		)
		prevStats = stats
		prevTime = currTime
	}
}
