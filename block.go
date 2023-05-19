package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"math/rand"
	"unsafe"
)

type BlockType uint8

const (
	Data BlockType = iota
	Keepalive
)

type Block struct {
	Type    BlockType
	ID      uint32
	Payload []byte
}

/*
* On-disk format:
* off	type		desc
* 0 	uint8 		magic
* 1 	uint32 		block ID
* 5 	uint8 		type
* 1 	uint16 		payload length
* 8		[500]byte 	payload data
* 508 	uint32 		crc
 */

const blockMagic = uint8(42)

const BlockSize = 512
const PayloadMaxSize = 500

const (
	idOffset      = 1
	typeOffset    = 5
	lenOffset     = 6
	payloadOffset = 8
	crcOffset     = 508
)

var ErrBlock = errors.New("block validation error")

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
	blockCrc := binary.BigEndian.Uint32(buf[crcOffset:])
	crc := crc32.ChecksumIEEE(buf[0:crcOffset])

	if blockCrc != crc {
		return nil, fmt.Errorf("%w: %s", ErrBlock, "wrong crc")
	}

	magic := buf[0]
	if magic != blockMagic {
		return nil, fmt.Errorf("%w: %s", ErrBlock, "wrong magic")
	}

	block.ID = binary.BigEndian.Uint32(buf[idOffset:])

	block.Type = BlockType(buf[typeOffset])
	if (block.Type != Data) && (block.Type != Keepalive) {
		return nil, fmt.Errorf("%w: %s", ErrBlock, "wrong type")
	}

	pLen := binary.BigEndian.Uint16(buf[lenOffset:])
	if pLen > PayloadMaxSize {
		return nil, fmt.Errorf("%w: %s", ErrBlock, "payload length too big")
	}

	block.Payload = make([]byte, pLen)
	copy(block.Payload, buf[payloadOffset:payloadOffset+pLen])

	return block, nil
}

func NewBlock(payload []byte, typ BlockType) *Block {
	pLen := len(payload)
	if pLen > PayloadMaxSize {
		panicf("payload size is too big: %d", len(payload))
	}
	block := new(Block)
	block.Type = typ
	block.ID = rand.Uint32()
	block.Payload = make([]byte, pLen)
	copy(block.Payload, payload)

	return block
}

func (block *Block) ToBytes() []byte {
	buf := alignedByteSlice(BlockSize, BlockSize)
	buf[0] = blockMagic
	binary.BigEndian.PutUint32(buf[idOffset:], block.ID)
	buf[typeOffset] = uint8(block.Type)
	binary.BigEndian.PutUint16(buf[lenOffset:], uint16(len(block.Payload)))
	copy(buf[payloadOffset:PayloadMaxSize], block.Payload)
	crc := crc32.ChecksumIEEE(buf[0:crcOffset])
	binary.BigEndian.PutUint32(buf[crcOffset:], crc)

	return buf
}
