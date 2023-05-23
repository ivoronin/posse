package main

import (
	"errors"
	"log"
	"time"

	"github.com/ivoronin/posse/fsm"
)

// reads packets from disk and puts them into rx queue
func diskRx(disk *Disk, rt *time.Ticker, fsm *fsm.FSM, rxq chan<- []byte) {
	var prevID uint32

	for range rt.C {
		block, err := disk.ReadBlock()
		if err != nil {
			// ReadBlock can return ErrBlock because peer had not yet written
			// anything to it's wblk and it is containing some garbage at the moment.
			// Such errors must be silenced.
			if fsm.CurrentState == PeerRxStateInit && errors.Is(err, ErrBlock) {
				continue
			}
			log.Printf("error reading from disk: %s", err)
			fsm.Event(PeerRxEventBlockReadErr)
			continue
		}

		// Skip first packet, because it can be the old one
		if prevID == 0 {
			prevID = block.ID
			continue
		}

		// Block didn't changed since last read
		if block.ID == prevID {
			fsm.Event(PeerRxEventBlockReadStale)
			stats.rdBlkStale++
			continue
		}
		fsm.Event(PeerRxEventBlockReadNew)

		if block.ID-prevID > 1 {
			stats.rdBlkMiss += uint64(block.ID - prevID - 1)
		}

		prevID = block.ID

		if block.Type == Data {
			stats.rdBlkData++
			rxq <- block.Payload
		} else if block.Type == Keepalive {
			stats.rdBlkKeep++
		}
	}
}

// reads packets from tx queue and writes them to disk device
func diskTx(disk *Disk, wt *time.Ticker, fsm *fsm.FSM, txq <-chan []byte) {
	var blkSeq uint32
	var wrBlkStat *uint64

	for range wt.C {
		var block *Block

		if len(txq) == 0 {
			fsm.Event(PeerTxEventBlockSkipped)
			if fsm.CurrentState != PeerTxStateIdle {
				continue
			}
			block = NewBlock(nil, blkSeq, Keepalive)
			wrBlkStat = &stats.wrBlkKeep
		} else {
			payload := <-txq
			block = NewBlock(payload, blkSeq, Data)
			wrBlkStat = &stats.wrBlkData
		}

		err := disk.WriteBlock(block)
		if err != nil {
			log.Printf("error writing to disk: %s", err)
			fsm.Event(PeerTxEventBlockWriteErr)
			continue
		}
		fsm.Event(PeerTxEventBlockWritten)
		*wrBlkStat++
		blkSeq++
	}
}

// reads packets from tun device and puts them into tx qeueue
func tunRx(tun *TUN, txq chan<- []byte) {
	for {
		buf := make([]byte, PayloadMaxSize)
		_, err := tun.Read(buf)
		if err != nil {
			log.Printf("error reading from tun: %s", err)
			continue
		}
		txq <- buf
	}
}

// read packets from rx queue and writes them to tun device
func tunTx(tun *TUN, rxq <-chan []byte) {
	for buf := range rxq {
		_, err := tun.Write(buf)
		if err != nil {
			log.Printf("error writing to tun: %s", err)
			continue
		}
	}
}
