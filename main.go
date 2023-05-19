package main

import (
	"log"
	"time"

	"github.com/alecthomas/kingpin/v2"
)

// reads packets from disk and puts them into rx queue
func diskRx(disk *Disk, rt *time.Ticker, rxq chan<- []byte) {
	var prevID uint32 = 0
	for range rt.C {
		block, err := disk.ReadBlock()
		if err != nil {
			stats.rdErr++
			log.Printf("error reading from disk: %s", err)
			continue
		}
		stats.rdBlk++

		// block doesn't changed since last read, nothing to do
		if block.ID == prevID {
			continue
		}

		// skip first packet, because it can be the old one
		if stats.rdBlk == 1 {
			continue
		}
		if stats.rdBlk == 2 {
			log.Printf("received first block")
		}

		prevID = block.ID

		err = block.Validate()
		if err != nil {
			log.Printf("block validation error: %s", err)
			continue
		}
		rxq <- block.Payload
	}
}

// reads packets from tx queue and writes them to disk device
func diskTx(disk *Disk, wt *time.Ticker, txq <-chan []byte) {
	for range wt.C {
		if len(txq) == 0 {
			continue
		}
		payload := <-txq
		block := NewBlockWithPayload(payload)
		err := disk.WriteBlock(block)
		if err != nil {
			stats.wrErr++
			log.Printf("error writing to disk: %s", err)
			continue
		}
		stats.wrBlk++
	}
}

// reads packets from tun device and puts them into tx qeueue
func tunRx(tun *TUN, txq chan<- []byte) {
	for {
		buf := make([]byte, PayloadMaxSize)
		_, err := tun.Read(buf)
		if err != nil {
			stats.rxErr++
			log.Printf("error reading from tun: %s", err)
			continue
		}
		stats.rxPkt++
		txq <- buf
	}
}

// read packets from rx queue and writes them to tun device
func tunTx(tun *TUN, rxq <-chan []byte) {
	for buf := range rxq {
		_, err := tun.Write(buf)
		if err != nil {
			stats.txErr++
			log.Printf("error writing to tun: %s", err)
			continue
		}
		stats.txPkt++
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

	go diskRx(disk, rt, rxq)
	go diskTx(disk, wt, txq)
	go tunRx(tun, txq)
	go tunTx(tun, rxq)

	if *statsInt != 0 {
		st := time.NewTicker(*statsInt)
		go reportStats(st)
	}

	log.Printf("started up, running on %s", tun.Name())

	// block
	select {}
}
