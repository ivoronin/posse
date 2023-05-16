package main

import (
	"fmt"
	"os"
	"syscall"
)

type Disk struct {
	file *os.File
	rOff int64
	wOff int64
}

func NewDisk(diskPath string, rOff int64, wOff int64) (*Disk, error) {
	var err error

	disk := Disk{
		rOff: rOff * BlockSize,
		wOff: wOff * BlockSize,
	}

	disk.file, err = os.OpenFile(diskPath, os.O_RDWR|syscall.O_DIRECT, 0)
	if err != nil {
		return nil, fmt.Errorf("error opening storage device: %s", err)
	}

	return &disk, nil
}

func (disk *Disk) ReadBlock() (*Block, error) {
	buf := make([]byte, BlockSize)

	_, err := disk.file.ReadAt(buf, disk.rOff)
	if err != nil {
		return nil, fmt.Errorf("error reading block from disk device: %s", err)
	}

	block := NewBlockFromBytes(buf)

	return block, nil
}

func (disk *Disk) WriteBlock(block *Block) error {
	buf := block.ToBytes()
	_, err := disk.file.WriteAt(buf, disk.wOff)
	if err != nil {
		return fmt.Errorf("error writing block to disk device: %s", err)
	}

	return nil
}
