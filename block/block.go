package block

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"unsafe"
)

type BlockType uint8

const (
	TypeData BlockType = iota
	TypeKeepalive
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
		panic(fmt.Sprintf("invalid block size: %d", len(buf)))
	}

	blk := new(Block)
	blkCrc := binary.BigEndian.Uint32(buf[crcOffset:])
	crc := crc32.ChecksumIEEE(buf[0:crcOffset])

	if blkCrc != crc {
		return nil, fmt.Errorf("wrong crc")
	}

	magic := buf[0]
	if magic != blockMagic {
		return nil, fmt.Errorf("wrong magic")
	}

	blk.ID = binary.BigEndian.Uint32(buf[idOffset:])

	blk.Type = BlockType(buf[typeOffset])
	if (blk.Type != TypeData) && (blk.Type != TypeKeepalive) {
		return nil, fmt.Errorf("wrong type")
	}

	pLen := binary.BigEndian.Uint16(buf[lenOffset:])
	if pLen > PayloadMaxSize {
		return nil, fmt.Errorf("payload length too big")
	}

	blk.Payload = make([]byte, pLen)
	copy(blk.Payload, buf[payloadOffset:payloadOffset+pLen])

	return blk, nil
}

func NewBlock(payload []byte, id uint32, typ BlockType) *Block {
	pLen := len(payload)
	if pLen > PayloadMaxSize {
		panic(fmt.Sprintf("payload size is too big: %d", len(payload)))
	}
	blk := new(Block)
	blk.Type = typ
	blk.ID = id
	blk.Payload = make([]byte, pLen)
	copy(blk.Payload, payload)

	return blk
}

func (b *Block) ToBytes() []byte {
	buf := alignedByteSlice(BlockSize, BlockSize)
	buf[0] = blockMagic
	binary.BigEndian.PutUint32(buf[idOffset:], b.ID)
	buf[typeOffset] = uint8(b.Type)
	binary.BigEndian.PutUint16(buf[lenOffset:], uint16(len(b.Payload)))
	copy(buf[payloadOffset:payloadOffset+PayloadMaxSize], b.Payload)
	crc := crc32.ChecksumIEEE(buf[0:crcOffset])
	binary.BigEndian.PutUint32(buf[crcOffset:], crc)

	return buf
}
