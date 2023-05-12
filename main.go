package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

type Packet struct {
	id      uint64
	payload [PayloadSize]byte
}

func sleepUntilNextTxTime(lastTxTime time.Time, interval time.Duration) {
	timeSinceLastTx := time.Now().Sub(lastTxTime)
	if timeSinceLastTx < interval {
		time.Sleep(interval - timeSinceLastTx)
	}
}

// reads packets from disk and puts them into rx queue
func diskRx(disk *Disk, interval time.Duration, rxq chan<- []byte) {
	var prevId uint64 = 0
	var lastTxTime time.Time
	for {
		sleepUntilNextTxTime(lastTxTime, interval)
		block, err := disk.ReadBlock()
		if err != nil {
			log.Printf("error reading block: %s", err)
			goto finish
		}
		if block.id != prevId {
			// skip first packet, because it can be the old one
			if prevId != 0 {
				rxq <- block.payload
			}
			prevId = block.id
		}
	finish:
		lastTxTime = time.Now()
	}
}

// reads packets from tx queue and writes them to disk device
func diskTx(disk *Disk, interval time.Duration, txq <-chan []byte) {
	var lastTxTime time.Time
	for payload := range txq {
		sleepUntilNextTxTime(lastTxTime, interval)
		block := NewBlockWithUniqueId(payload)
		err := disk.WriteBlock(block)
		if err != nil {
			log.Printf("error writing block to disk: %s", err)
		}
		lastTxTime = time.Now()
	}
}

// reads packets from tun device and puts them into tx qeueue
func tunRx(tun *TUN, txq chan<- []byte) {
	buf := make([]byte, PayloadSize)
	for {
		_, err := tun.Read(buf)
		if err != nil {
			log.Printf("error reading from network device: %s", err)
			continue
		}
		txq <- buf
	}
}

// read packets from rx queue and writes them to tun device
func tunTx(run *TUN, rxq <-chan []byte) {
	for buf := range rxq {
		_, err := run.Write(buf)
		if err != nil {
			log.Printf("error writing to network device: %s", err)
			continue
		}
	}
}

func die(fmts string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmts, args...)
	os.Exit(1)
}

func main() {
	diskPath := flag.String("disk", "", "disk path")
	tunName := flag.String("tun", "", "tun name")
	localAddr := flag.String("addr", "", "local address")
	remoteAddr := flag.String("peer", "", "remote address")
	rBlk := flag.Int64("rblk", -1, "disk block to read packets from")
	wBlk := flag.Int64("wblk", -1, "disk block to write packets to")
	txQLen := flag.Int("txqlen", 16, "tx queue length")
	rxQLen := flag.Int("rxqlen", 16, "rx queue length")
	hz := flag.Int("hz", 10, "polling and writing frequency in hz")
	flag.Parse()

	if *diskPath == "" {
		die("device path must be set")
	}

	if *localAddr == "" {
		die("local address must be set")
	}

	if *remoteAddr == "" {
		die("remote address must be set")
	}

	if *rBlk < 0 {
		die("device read offset must be set and be positive")
	}

	if *wBlk < 0 {
		die("device write offset must be set and be positive")
	}

	tun, err := NewTUN(*tunName, PayloadSize, *localAddr, *remoteAddr)
	if err != nil {
		die("error setting up tun device: %s", err)
	}

	disk, err := NewDisk(*diskPath, *rBlk, *wBlk)
	if err != nil {
		die("error setting up disk device: %s", err)
	}

	// disk -> tun queue
	rxq := make(chan []byte, *rxQLen)
	// tun -> disk queue
	txq := make(chan []byte, *txQLen)

	delay := time.Millisecond * time.Duration(1000 / *hz)

	go diskRx(disk, delay, rxq)
	go diskTx(disk, delay, txq)
	go tunRx(tun, txq)
	go tunTx(tun, rxq)

	// block
	select {}
}
