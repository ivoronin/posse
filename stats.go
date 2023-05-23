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
	rdSvcTime  uint64
	wrBlk      uint64
	wrBlkData  uint64
	wrBlkKeep  uint64
	wrErr      uint64
	wrSvcTime  uint64
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
		blkRead := stats.rdBlk - prevStats.rdBlk
		blkWritten := stats.wrBlk - prevStats.wrBlk
		log.Printf("stats: rx:%s tx:%s "+
			"rdBlk:%d rdBlkErr:%d rdBlkMiss:%d "+
			"rdBlkStale:%d rdBlkData:%d rdBlkKeep:%d "+
			"rdErr:%d rdBlk/s:%.2f rdSvcTimeAvg:%.02f "+
			"wrBlk:%d wrBlkData:%d wrBlkKeep:%d "+
			"wrErr:%d wrBlk/s:%.2f wrSvcTimeAvg:%.02f "+
			"rxPkt:%d rxErr:%d rxPkt/s:%.2f "+
			"txPkt:%d txErr:%d txPkt/s:%.2f",
			peer.RxFSM.CurrentState, peer.TxFSM.CurrentState,
			blkRead,
			stats.rdBlkErr-prevStats.rdBlkErr,
			stats.rdBlkMiss-prevStats.rdBlkMiss,
			stats.rdBlkStale-prevStats.rdBlkStale,
			stats.rdBlkData-prevStats.rdBlkData,
			stats.rdBlkKeep-prevStats.rdBlkKeep,
			stats.rdErr-prevStats.rdErr,
			float64(blkRead)/sec,
			float64(stats.rdSvcTime-prevStats.rdSvcTime)/float64(blkRead)/1000,
			blkWritten,
			stats.wrBlkData-prevStats.wrBlkData,
			stats.wrBlkKeep-prevStats.wrBlkKeep,
			stats.wrErr-prevStats.wrErr,
			float64(blkWritten)/sec,
			float64(stats.wrSvcTime-prevStats.wrSvcTime)/float64(blkWritten)/1000,
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
