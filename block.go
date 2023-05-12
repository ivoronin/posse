package main

import (
	"encoding/binary"
	"fmt"

	"github.com/sony/sonyflake"
)

type Block struct {
	id      uint64
	payload []byte
}

var sf = sonyflake.NewSonyflake(sonyflake.Settings{})

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
	id, err := sf.NextID()
	if err != nil {
		panic("unable to generate unique id")
	}
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
