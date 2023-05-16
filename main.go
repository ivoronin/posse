package main

import (
	"flag"
	"log"
	"time"
)

// reads packets from disk and puts them into rx queue
func diskRx(disk *Disk, rt *time.Ticker, rxq chan<- []byte) {
	var prevId uint32 = 0
	for range rt.C {
		block, err := disk.ReadBlock()
		if err != nil {
			log.Printf("error reading from disk: %s", err)
			continue
		}
		if block.id != prevId {
			// skip first packet, because it can be the old one
			if prevId != 0 {
				err = block.Validate()
				if err != nil {
					log.Printf("block validation error: %s", err)
					continue
				}
				rxq <- block.payload
			}
			prevId = block.id
		}
	}
}

// reads packets from tx queue and writes them to disk device
func diskTx(disk *Disk, wt *time.Ticker, txq <-chan []byte) {
	for range wt.C {
		if len(txq) == 0 {
			continue
		}
		payload := <-txq
		block := NewBlockWithUniqueId(payload)
		err := disk.WriteBlock(block)
		if err != nil {
			log.Printf("error writing to disk: %s", err)
			continue
		}
	}
}

// reads packets from tun device and puts them into tx qeueue
func tunRx(tun *TUN, txq chan<- []byte) {
	for {
		buf := make([]byte, payloadMaxSize)
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
	diskPath := flag.String("disk", "", "disk path")
	tunName := flag.String("tun", "", "tun name")
	localAddr := flag.String("addr", "", "local address")
	remoteAddr := flag.String("peer", "", "remote address")
	rBlk := flag.Uint64("rblk", 0, "disk block to read packets from")
	wBlk := flag.Uint64("wblk", 0, "disk block to write packets to")
	txQLen := flag.Int("txqlen", 16, "tx queue length")
	rxQLen := flag.Int("rxqlen", 16, "rx queue length")
	hz := flag.Int("hz", 10, "polling and writing frequency in hz")
	flag.Parse()

	mandatoryFlags := []string{"disk", "addr", "peer", "rblk", "wblk"}
	seenFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seenFlags[f.Name] = true })
	for _, flagName := range mandatoryFlags {
		if !seenFlags[flagName] {
			errx("%s must be set", flagName)
		}
	}

	if rBlk == wBlk {
		errx("rblk and wblk values can't be equal")
	}

	tun, err := NewTUN(*tunName, payloadMaxSize, *localAddr, *remoteAddr)
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

	// block
	select {}
}
