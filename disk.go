package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/ivoronin/posse/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type Disk struct {
	file *os.File
	rOff int64
	wOff int64
}

var ErrBlock = errors.New("block error")

func NewDisk(diskPath string, rOff uint64, wOff uint64) (*Disk, error) {
	var err error

	disk := Disk{
		rOff: int64(rOff * BlockSize),
		wOff: int64(wOff * BlockSize),
	}

	disk.file, err = os.OpenFile(diskPath, os.O_RDWR|os.O_SYNC|syscall.O_DIRECT, 0)
	if err != nil {
		return nil, err
	}

	return &disk, nil
}

func (disk *Disk) ReadBlock() (*Block, error) {
	buf := make([]byte, BlockSize)

	t := prometheus.NewTimer(metrics.RdSvcTime)
	_, err := disk.file.ReadAt(buf, disk.rOff)
	if err != nil {
		t.ObserveDuration()
		metrics.RdErr.Inc()
		return nil, err
	}
	t.ObserveDuration()

	block, err := NewBlockFromBytes(buf)
	if err != nil {
		metrics.RdBlkErr.Inc()
		return nil, fmt.Errorf("%w: %s", ErrBlock, err)
	}

	metrics.RdBlk.Inc()
	return block, nil
}

func (disk *Disk) WriteBlock(block *Block) error {
	buf := block.ToBytes()
	t := prometheus.NewTimer(metrics.WrSvcTime)
	_, err := disk.file.WriteAt(buf, disk.wOff)
	if err != nil {
		t.ObserveDuration()
		metrics.WrErr.Inc()
		return err
	}
	t.ObserveDuration()
	metrics.WrBlk.Inc()
	return nil
}
