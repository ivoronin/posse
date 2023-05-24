package main

import (
	"log"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/ivoronin/posse/block"
	"github.com/ivoronin/posse/disk"
	"github.com/ivoronin/posse/metrics"
	"github.com/ivoronin/posse/peer"
	"github.com/ivoronin/posse/tun"
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

	tn, err := tun.NewTUN(*tunName, block.PayloadMaxSize, *localAddr, *remoteAddr)
	if err != nil {
		errx("error setting up tun device: %s", err)
	}

	dsk, err := disk.NewDisk(*diskPath, *rBlk, *wBlk)
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

	peer := peer.NewPeer(*maxStale)

	if *promAddr != "" {
		go metrics.Serve(*promAddr)
	}
	go diskRd(dsk, rt, peer.RxFSM, rxq)
	go diskWr(dsk, wt, peer.TxFSM, txq)
	go tunRx(tn, txq)
	go tunTx(tn, rxq)

	log.Printf("started up, running on %s", tn.Name())

	// block
	select {}
}
