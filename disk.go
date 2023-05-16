package main

import (
	"os"
	"syscall"
)

type Disk struct {
	file *os.File
	rOff int64
	wOff int64
}

func NewDisk(diskPath string, rOff uint64, wOff uint64) (*Disk, error) {
	var err error

	disk := Disk{
		rOff: int64(rOff * BlockSize),
		wOff: int64(wOff * BlockSize),
	}

	disk.file, err = os.OpenFile(diskPath, os.O_RDWR|syscall.O_DIRECT, 0)
	if err != nil {
		return nil, err
	}

	return &disk, nil
}

func (disk *Disk) ReadBlock() (*Block, error) {
	buf := make([]byte, BlockSize)

	_, err := disk.file.ReadAt(buf, disk.rOff)
	if err != nil {
		return nil, err
	}

	block := NewBlockFromBytes(buf)

	return block, nil
}

func (disk *Disk) WriteBlock(block *Block) error {
	buf := block.ToBytes()
	_, err := disk.file.WriteAt(buf, disk.wOff)
	if err != nil {
		return err
	}

	return nil
}
