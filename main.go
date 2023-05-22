package main

import (
	"log"
	"time"

	"github.com/alecthomas/kingpin/v2"
)

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

	go diskRx(disk, rt, peer.RxFSM, rxq)
	go diskTx(disk, wt, peer.TxFSM, txq)
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
