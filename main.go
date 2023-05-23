package main

import (
	"log"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/ivoronin/posse/metrics"
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
		maxStale   = kingpin.Flag("maxstale", "Number of stale reads before declaring peer dead").Envar("MAXSTALE").Default("5").Uint64()
		promAddr   = kingpin.Flag("promaddr", "Addr:Port to listen for prometheus queries").Envar("PROMADDR").Default("").String()
	)

	kingpin.Parse()

	if *rBlk == *wBlk {
		errx("rblk and wblk values can't be equal")
	}

	if *maxStale == 0 {
		errx("maxstale must be greater than 0")
	}

	tun, err := NewTUN(*tunName, PayloadMaxSize, *localAddr, *remoteAddr)
	if err != nil {
		errx("error setting up tun device: %s", err)
	}

	disk, err := NewDisk(*diskPath, *rBlk, *wBlk)
	if err != nil {
		errx("error setting up disk device: %s", err)
	}

	if *promAddr != "" {
		go metrics.Serve(*promAddr)
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

	log.Printf("started up, running on %s", tun.Name())

	// block
	select {}
}
