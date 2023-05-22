package main

import (
	"errors"
	"log"
	"time"

	"github.com/alecthomas/kingpin/v2"
)

// reads packets from disk and puts them into rx queue
func diskRx(disk *Disk, rt *time.Ticker, peer *Peer, rxq chan<- []byte) {
	var prevID uint32

	for range rt.C {
		block, err := disk.ReadBlock()
		if err != nil {
			// ReadBlock can return ErrBlock because peer had not yet written
			// anything to it's wblk and it is containing some garbage at the moment.
			// Such errors must be silenced.
			if peer.State() == PeerStateInit && errors.Is(err, ErrBlock) {
				continue
			}
			log.Printf("error reading from disk: %s", err)
			peer.Event(PeerEventBlockReadErr)
			continue
		}

		// Skip first packet, because it can be the old one
		if prevID == 0 {
			prevID = block.ID
			continue
		}

		// Block didn't changed since last read
		if block.ID == prevID {
			peer.Event(PeerEventBlockReadStale)
			stats.rdBlkStale++
			continue
		}
		peer.Event(PeerEventBlockReadNew)

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
func diskTx(disk *Disk, wt *time.Ticker, maxStale uint64, txq <-chan []byte) {
	var missedWrites uint64
	var blkSeq uint32
	for range wt.C {
		var block *Block
		if len(txq) == 0 {
			missedWrites++
			if missedWrites*2 < maxStale {
				continue
			}
			block = NewBlock(nil, blkSeq, Keepalive)
			stats.wrBlkKeep++
		} else {
			payload := <-txq
			block = NewBlock(payload, blkSeq, Data)
			stats.wrBlkData++
		}

		missedWrites = 0

		err := disk.WriteBlock(block)
		if err != nil {
			log.Printf("error writing to disk: %s", err)
			continue
		}
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

func main() {
	var (
		diskPath   = kingpin.Flag("disk", "Disk path").Short('d').Envar("DISK").Required().String()
		tunName    = kingpin.Flag("tun", "Tun name").Short('t').Envar("TUN").String()
		localAddr  = kingpin.Flag("addr", "Local address").Short('a').Envar("ADDR").Required().String()
		remoteAddr = kingpin.Flag("peer", "Peer address").Short('p').Envar("PEER").Required().String()
		rBlk       = kingpin.Flag("rblk", "Disk block to read packets from").Short('r').Envar("RBLK").Required().Uint64()
		wBlk       = kingpin.Flag("wblk", "Disk block to write packets to").Short('w').Envar("WBLK").Required().Uint64()
		txQLen     = kingpin.Flag("txqlen", "TX queue length").Envar("TXQLEN").Default("16").Uint()
		rxQLen     = kingpin.Flag("rxqlen", "RX queue length").Envar("RXQLEN").Default("16").Uint()
		hz         = kingpin.Flag("hz", "Disk polling and writing frequency in hz").Short('f').Envar("HZ").Default("10").Uint()
		statsInt   = kingpin.Flag("stats", "Interval between periodic stats reports").Short('i').Envar("STATS").Default("60s").Duration()
		maxStale   = kingpin.Flag("maxstale", "Number of stale reads before declaring peer dead").Envar("MAXSTALE").Default("5").Uint64()
	)

	kingpin.Parse()

	if *rBlk == *wBlk {
		errx("rblk and wblk values can't be equal")
	}

	tun, err := NewTUN(*tunName, PayloadMaxSize, *localAddr, *remoteAddr)
	if err != nil {
		errx("error setting up tun device: %s", err)
	}

	disk, err := NewDisk(*diskPath, *rBlk, *wBlk)
	if err != nil {
		errx("error setting up disk device: %s", err)
	}

	// disk -> tun queue
	rxq := make(chan []byte, *rxQLen)
	// tun -> disk queue
	txq := make(chan []byte, *txQLen)

	tickDuration := time.Millisecond * time.Duration(1000 / *hz)
	rt := time.NewTicker(tickDuration)
	wt := time.NewTicker(tickDuration)

	peer := NewPeer(*maxStale)

	go diskRx(disk, rt, peer, rxq)
	go diskTx(disk, wt, *maxStale, txq)
	go tunRx(tun, txq)
	go tunTx(tun, rxq)

	if *statsInt != 0 {
		st := time.NewTicker(*statsInt)
		go reportStats(st, peer)
	}

	log.Printf("started up, running on %s", tun.Name())

	// block
	select {}
}
