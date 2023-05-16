package main

import (
	"encoding/binary"
	"hash/crc32"
	"math/rand"
	"unsafe"
)

type Block struct {
	id      uint64
	crc     uint32
	payload []byte
}

const BlockSize = 512
const IDSize = 8
const CRCSize = crc32.Size
const HeaderSize = IDSize + CRCSize
const PayloadSize = BlockSize - HeaderSize

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
	block.id = binary.BigEndian.Uint64(buf)
	block.crc = binary.BigEndian.Uint32(buf[IDSize:])
	block.payload = buf[HeaderSize:]

	return block
}

func NewBlockWithUniqueId(payload []byte) *Block {
	if len(payload) > PayloadSize {
		panicf("payload size is too big: %d", len(payload))
	}
	block := new(Block)
	block.id = rand.Uint64()
	block.crc = crc32.ChecksumIEEE(payload)
	block.payload = payload

	return block
}

func (block *Block) ToBytes() []byte {
	buf := alignedByteSlice(BlockSize, BlockSize)
	binary.BigEndian.PutUint64(buf, block.id)
	binary.BigEndian.PutUint32(buf[IDSize:], block.crc)
	copy(buf[HeaderSize:], block.payload)
	return buf
}

func (block *Block) VerifyCRC() bool {
	return crc32.ChecksumIEEE(block.payload) == block.crc
}
