package disk

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/ivoronin/posse/block"
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

	dsk := new(Disk)
	dsk.rOff = int64(rOff * block.BlockSize)
	dsk.wOff = int64(wOff * block.BlockSize)

	dsk.file, err = os.OpenFile(diskPath, os.O_RDWR|os.O_SYNC|syscall.O_DIRECT, 0)
	if err != nil {
		return nil, err
	}

	return dsk, nil
}

func (d *Disk) ReadBlock() (*block.Block, error) {
	buf := make([]byte, block.BlockSize)

	t := prometheus.NewTimer(metrics.RdSvcTime)
	_, err := d.file.ReadAt(buf, d.rOff)
	if err != nil {
		t.ObserveDuration()
		metrics.RdErr.Inc()
		return nil, err
	}
	t.ObserveDuration()

	blk, err := block.NewBlockFromBytes(buf)
	if err != nil {
		metrics.RdBlkErr.Inc()
		return nil, fmt.Errorf("%w: %s", ErrBlock, err)
	}

	metrics.RdBlk.Inc()
	return blk, nil
}

func (d *Disk) WriteBlock(blk *block.Block) error {
	buf := blk.ToBytes()
	t := prometheus.NewTimer(metrics.WrSvcTime)
	_, err := d.file.WriteAt(buf, d.wOff)
	if err != nil {
		t.ObserveDuration()
		metrics.WrErr.Inc()
		return err
	}
	t.ObserveDuration()
	metrics.WrBlk.Inc()
	return nil
}
