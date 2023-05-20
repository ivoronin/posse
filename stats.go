package main

import (
	"log"
	"time"
)

type Stats struct {
	rdBlk uint64
	rdErr uint64
	wrBlk uint64
	wrErr uint64
	rxPkt uint64
	rxErr uint64
	txPkt uint64
	txErr uint64
}

var stats Stats

func reportStats(st *time.Ticker) {
	var prevStats Stats
	prevTime := time.Now()

	for currTime := range st.C {
		sec := currTime.Sub(prevTime).Seconds()
		log.Printf("status:%s "+
			"rdBlk:%d rdErr:%d rdBlk/s:%.2f "+
			"wrBlk:%d wrErr:%d wrBlk/s:%.2f "+
			"rxPkt:%d rxErr:%d rxPkt/s:%.2f "+
			"txPkt:%d txErr:%d txPkt/s:%.2f",
			peerStatus,
			stats.rdBlk-prevStats.rdBlk,
			stats.rdErr-prevStats.rdErr,
			float64(stats.rdBlk-prevStats.rdBlk)/sec,
			stats.wrBlk-prevStats.wrBlk,
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