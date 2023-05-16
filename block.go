package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math/rand"
	"unsafe"
)

type Block struct {
	version uint8
	id      uint32
	crc     uint32
	payload []byte
}

const BlockVersion = 1
const BlockSize = 512

const versionSize = 1
const idSize = 8
const crcSize = crc32.Size

const idOffset = versionSize
const crcOffset = idOffset + idSize
const payloadOffset = crcOffset + crcSize

const payloadMaxSize = BlockSize - payloadOffset

func alignedByteSlice(size uint, align uint) []byte {
	bytes := make([]byte, size+align)
	if align == 0 {
		return bytes
	}
	gap := uintptr(unsafe.Pointer(&bytes[0])) & uintptr(align-1)
	offset := align - uint(gap)
	bytes = bytes[offset : offset+size]
	return bytes
}

func NewBlockFromBytes(buf []byte) *Block {
	if len(buf) != BlockSize {
		panicf("invalid block size: %d", len(buf))
	}

	block := new(Block)
	block.version = buf[0]
	block.id = binary.BigEndian.Uint32(buf[idOffset:])
	block.crc = binary.BigEndian.Uint32(buf[crcOffset:])
	block.payload = buf[payloadOffset:]

	return block
}

func NewBlockWithUniqueId(payload []byte) *Block {
	if len(payload) > payloadMaxSize {
		panicf("payload size is too big: %d", len(payload))
	}
	block := new(Block)
	block.version = BlockVersion
	block.id = rand.Uint32()
	block.crc = crc32.ChecksumIEEE(payload)
	block.payload = payload

	return block
}

func (block *Block) ToBytes() []byte {
	buf := alignedByteSlice(BlockSize, BlockSize)
	buf[0] = block.version
	binary.BigEndian.PutUint32(buf[idOffset:], block.id)
	binary.BigEndian.PutUint32(buf[crcOffset:], block.crc)
	copy(buf[payloadOffset:], block.payload)
	return buf
}

func (block *Block) Validate() error {
	if block.version != BlockVersion {
		return fmt.Errorf("wrong block version")
	}
	if block.crc != crc32.ChecksumIEEE(block.payload) {
		return fmt.Errorf("wrong block crc")
	}
	return nil
}
