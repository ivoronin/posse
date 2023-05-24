package main

import (
	"errors"
	"log"
	"time"

	"github.com/ivoronin/posse/block"
	"github.com/ivoronin/posse/disk"
	"github.com/ivoronin/posse/fsm"
	"github.com/ivoronin/posse/metrics"
	"github.com/ivoronin/posse/peer"
	"github.com/ivoronin/posse/tun"
	"github.com/prometheus/client_golang/prometheus"
)

// reads packets from disk and puts them into rx queue
func diskRd(dsk *disk.Disk, rt *time.Ticker, fsm *fsm.FSM, rxq chan<- []byte) {
	var prevID uint32

	for range rt.C {
		blk, err := dsk.ReadBlock()
		if err != nil {
			// ReadBlock can return ErrBlock because peer had not yet written
			// anything to it's wblk and it is containing some garbage at the moment.
			// Such errors must be silenced.
			if fsm.CurrentState == peer.StateRxInit && errors.Is(err, disk.ErrBlock) {
				continue
			}
			log.Printf("error reading from disk: %s", err)
			fsm.Event(peer.EventBlockReadErr)
			continue
		}

		// Skip first packet, because it can be the old one
		if prevID == 0 {
			prevID = blk.ID
			continue
		}

		// Block didn't changed since last read
		if blk.ID == prevID {
			fsm.Event(peer.EventBlockReadStale)
			metrics.RdBlkStale.Inc()
			continue
		}
		fsm.Event(peer.EventBlockReadNew)

		if blk.ID-prevID > 1 {
			metrics.RdBlkMiss.Add(float64(blk.ID - prevID - 1))
		}

		prevID = blk.ID

		if blk.Type == block.TypeData {
			metrics.RdBlkData.Inc()
			rxq <- blk.Payload
		} else if blk.Type == block.TypeKeepalive {
			metrics.RdBlkKeep.Inc()
		}
	}
}

// reads packets from tx queue and writes them to disk device
func diskWr(dsk *disk.Disk, wt *time.Ticker, fsm *fsm.FSM, txq <-chan []byte) {
	var blkSeq uint32
	var wrBlkMetric prometheus.Counter

	for range wt.C {
		var blk *block.Block

		if len(txq) == 0 {
			fsm.Event(peer.EventBlockWriteSkipped)
			if fsm.CurrentState != peer.StateTxIdle {
				continue
			}
			blk = block.NewBlock(nil, blkSeq, block.TypeKeepalive)
			wrBlkMetric = metrics.WrBlkKeep
		} else {
			payload := <-txq
			blk = block.NewBlock(payload, blkSeq, block.TypeData)
			wrBlkMetric = metrics.WrBlkData
		}

		err := dsk.WriteBlock(blk)
		if err != nil {
			log.Printf("error writing to disk: %s", err)
			fsm.Event(peer.EventBlockWriteErr)
			continue
		}
		fsm.Event(peer.EventBlockWritten)
		wrBlkMetric.Inc()
		blkSeq++
	}
}

// reads packets from tun device and puts them into tx qeueue
func tunRx(tn *tun.TUN, txq chan<- []byte) {
	for {
		buf := make([]byte, block.PayloadMaxSize)
		_, err := tn.Read(buf)
		if err != nil {
			log.Printf("error reading from tun: %s", err)
			continue
		}
		txq <- buf
	}
}

// read packets from rx queue and writes them to tun device
func tunTx(tn *tun.TUN, rxq <-chan []byte) {
	for buf := range rxq {
		_, err := tn.Write(buf)
		if err != nil {
			log.Printf("error writing to tun: %s", err)
			continue
		}
	}
}
