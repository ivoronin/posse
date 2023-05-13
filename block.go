package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
)

type Block struct {
	id      uint64
	payload []byte
}

const BlockSize = 512
const IDSize = 8
const PayloadSize = BlockSize - IDSize

func NewBlockFromBytes(buf []byte) *Block {
	var block Block
	if len(buf) != BlockSize {
		panic(fmt.Sprintf("invalid block size: %d", len(buf)))
	}

	block.id = binary.BigEndian.Uint64(buf)
	block.payload = buf[IDSize:]

	return &block
}

func NewBlockWithUniqueId(payload []byte) Block {
	if len(payload) > PayloadSize {
		panic(fmt.Sprintf("payload size is too big: %d", len(payload)))
	}
	id := rand.Uint64()
	block := Block{
		id:      id,
		payload: payload,
	}
	return block
}

func (block *Block) ToBytes() []byte {
	buf := make([]byte, IDSize)
	binary.BigEndian.PutUint64(buf, block.id)
	buf = append(buf, block.payload...)

	if len(buf) > BlockSize {
		panic(fmt.Sprintf("invalid block size: %d", len(buf)))
	}

	return buf
}
