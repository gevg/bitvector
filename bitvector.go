package bitvector

import (
	"fmt"
)

type BitVector struct {
	Bytes   []byte
	Length  uint64
	large   []uint64
	small   []uint16
	rankMax uint64
}

const smallBlockBit = 0x80                                      // 128 (= 2^7)
const largeBlockBit = (2 * smallBlockBit) * (2 * smallBlockBit) // 65536(= 2^16)
const smallBlockByte = smallBlockBit / 8
const largeBlockByte = largeBlockBit / 8
const largeToSmall = largeBlockByte / smallBlockByte

func popCount(b byte) byte {
	return (1 & b) +
		(1 & (b >> 1)) +
		(1 & (b >> 2)) +
		(1 & (b >> 3)) +
		(1 & (b >> 4)) +
		(1 & (b >> 5)) +
		(1 & (b >> 6)) +
		(1 & (b >> 7))
}

func NewBitVector(bytes []byte, bitsLen uint64) *BitVector {

	var bv = BitVector{
		bytes,
		bitsLen,
		make([]uint64, (bitsLen/largeBlockBit)+1),
		make([]uint16, (bitsLen/smallBlockBit)+1),
		0,
	}

	var largeIndex uint64 = 1
	var largeCount uint64
	var largeValue uint64
	var smallIndex uint64 = 1
	var smallCount uint64
	var smallValue uint16
	var bytesLen = bitsLen / 8

	for i := uint64(0); i < bytesLen; i++ {
		if largeCount == largeBlockByte {
			largeValue += uint64(smallValue)
			bv.large[largeIndex] = largeValue
			largeCount = 0
			smallValue = 0
			largeIndex++
		}
		if smallCount == smallBlockByte {
			bv.small[smallIndex] = smallValue
			smallCount = 0
			smallIndex++
		}

		smallValue += uint16(popCount(bytes[i]))
		largeCount++
		smallCount++
	}
	bv.rankMax = largeValue + uint64(smallValue)

	return &bv
}

func (bv *BitVector) Rank1(pos uint64) (uint64, error) {
	if bv.Length < pos {
		return 0, fmt.Errorf("Length(=%d) < pos(=%d)", bv.Length, pos)
	}

	largePos := pos / largeBlockBit
	smallPos := pos / smallBlockBit
	startBit := (smallPos * smallBlockBit)
	startByte := startBit / 8
	endByte := pos / 8

	rank := bv.large[largePos]
	rank += uint64(bv.small[smallPos])

	for i := startByte; i < endByte; i++ {
		rank += uint64(popCount(bv.Bytes[i]))
	}

	modPos := pos % 8
	for i := byte(0); i < byte(modPos); i++ {
		rank += uint64(1 & (bv.Bytes[endByte] >> i))
	}

	return rank, nil
}

func (bv *BitVector) Select1(rank uint64) (uint64, error) {
	if bv.rankMax < rank+1 {
		return 0, fmt.Errorf("bv.rankMax(=%d) < rank+1(=%d)", bv.rankMax, rank+1)
	}

	largeLen := uint64(len(bv.large))
	min := uint64(0)
	max := largeLen - 1
	for {
		mid := (max - min) / 2
		if mid == 0 {
			break
		}
		mid += min

		if rank+1 <= bv.large[mid] {
			max = mid - 1
		} else {
			min = mid
		}
	}

	largePos := uint64(0)
	if rank+1 <= bv.large[max] {
		largePos = min
	} else {
		largePos = max
	}

	smallLen := uint64(len(bv.small))
	min = largePos * largeToSmall
	max = min + largeToSmall - 1
	if smallLen-1 < max {
		max = smallLen - 1
	}

	r := bv.large[largePos]
	for {
		mid := (max - min) / 2
		if mid == 0 {
			break
		}
		mid += min

		if rank+1 <= uint64(bv.small[mid])+r {
			max = mid - 1
		} else {
			min = mid
		}
	}

	smallPos := uint64(0)
	if rank+1 <= uint64(bv.small[max])+r {
		smallPos = min
	} else {
		smallPos = max
	}

	r += uint64(bv.small[smallPos])
	if r == rank+1 {
		return smallPos * smallBlockBit, nil
	}

	for n := uint64(0); n < smallBlockByte; n++ {
		b := bv.Bytes[n+smallPos*smallBlockByte]
		for i := uint64(0); i < 8; i++ {
			r += uint64(1 & (b >> i))
			if r == rank+1 {
				return i + 8*n + smallPos*smallBlockBit, nil
			}
		}
	}

	return 0, fmt.Errorf("over rank(=%d)", rank)
}

func Log2Floor(n uint64) uint64 {
	result := uint64(0)
	for {
		n >>= 1
		if n == 0 {
			return result
		}
		result++
	}
}

func Log2Ceil(n uint64) uint64 {
	result := uint64(0)
	if n != 0 {
		result = Log2Floor(n - 1)
	}
	return result
}
