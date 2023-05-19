package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math/rand"
	"unsafe"
)

type blockFlags uint8

const (
	keepalive blockFlags = 1 << iota
)

type Block struct {
	version uint8
	flags   blockFlags
	len     uint16
	ID      uint32
	crc     uint32
	Payload []byte
}

const blockVersion = 2
const BlockSize = 512

const lenOffset = 2
const idOffset = 4
const crcOffset = 8
const payloadOffset = 12

const PayloadMaxSize = BlockSize - payloadOffset

type BlockValidationError struct {
	msg string
}

func NewBlockValidationError(fmts string, args ...interface{}) *BlockValidationError {
	e := new(BlockValidationError)
	e.msg = fmt.Sprintf(fmts, args...)
	return e
}

func (e *BlockValidationError) Error() string {
	return e.msg
}

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

func NewBlockFromBytes(buf []byte) (*Block, error) {
	if len(buf) != BlockSize {
		panicf("invalid block size: %d", len(buf))
	}

	block := new(Block)
	block.version = buf[0]
	block.flags = blockFlags(buf[1])
	block.len = binary.BigEndian.Uint16(buf[lenOffset:])
	block.ID = binary.BigEndian.Uint32(buf[idOffset:])
	block.crc = binary.BigEndian.Uint32(buf[crcOffset:])

	if block.version != blockVersion {
		return nil, NewBlockValidationError("block version is not supported: %d", block.version)
	}
	if block.len > PayloadMaxSize {
		return nil, NewBlockValidationError("payload size is too big: %d", block.len)
	}

	block.Payload = buf[payloadOffset : payloadOffset+block.len]

	if block.crc != crc32.ChecksumIEEE(block.Payload) {
		return nil, NewBlockValidationError("block has wrong payload crc")
	}

	return block, nil
}

func NewBlock(payload []byte, flags blockFlags) *Block {
	if len(payload) > PayloadMaxSize {
		panicf("payload size is too big: %d", len(payload))
	}
	block := new(Block)
	block.version = blockVersion
	block.flags = flags
	block.len = uint16(len(payload))
	block.ID = rand.Uint32()
	block.crc = crc32.ChecksumIEEE(payload)
	block.Payload = make([]byte, PayloadMaxSize)
	copy(block.Payload, payload)

	return block
}

func (block *Block) ToBytes() []byte {
	buf := alignedByteSlice(BlockSize, BlockSize)
	buf[0] = block.version
	buf[1] = uint8(block.flags)
	binary.BigEndian.PutUint16(buf[lenOffset:], block.len)
	binary.BigEndian.PutUint32(buf[idOffset:], block.ID)
	binary.BigEndian.PutUint32(buf[crcOffset:], block.crc)
	copy(buf[payloadOffset:], block.Payload)
	return buf
}

func (block *Block) IsKeepalive() bool {
	return block.flags&keepalive != 0
}
