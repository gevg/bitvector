package bitvector

import (
	"fmt"
)

type BitVector struct {
	Bytes   []byte
	Length  uint64
	large   []uint64
	small   []uint16
	rank1Max uint64
}

const smallBlockBit = 0x80                                      // 128 (= 2^7)
const largeBlockBit = (2 * smallBlockBit) * (2 * smallBlockBit) // 65536(= 2^16)
const smallBlockByte = smallBlockBit / 8
const largeBlockByte = largeBlockBit / 8
const largeToSmall = largeBlockByte / smallBlockByte

func bitToByte(b uint64) uint64 {
	if b%8 == 0 {
		return b / 8
	} else {
		return b/8 + 1
	}
}

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
	var bytesLen = bitToByte(bitsLen)

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
	bv.rank1Max = largeValue + uint64(smallValue)

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
	if bv.rank1Max < rank+1 {
		return 0, fmt.Errorf("bv.rank1Max(=%d) < rank+1(=%d)", bv.rank1Max, rank+1)
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

	var largePos uint64
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

	var smallPos uint64
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

func (bv *BitVector) Select0(rank uint64) (uint64, error) {
	if (bv.Length - bv.rank1Max) < rank+1 {
		return 0, fmt.Errorf("bv.Length(%d) - bv.rank1Max(=%d) < rank+1(=%d)", bv.Length, bv.rank1Max, rank+1)
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

		if rank+1 <= (mid*largeBlockByte)-bv.large[mid] {
			max = mid - 1
		} else {
			min = mid
		}
	}

	var largePos uint64
	if rank+1 <= (max*largeBlockByte)-bv.large[max] {
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

	r := (largePos * largeBlockByte) - bv.large[largePos]
	for {
		mid := (max - min) / 2
		if mid == 0 {
			break
		}
		mid += min

		if rank+1 <= (mid*smallBlockByte)-uint64(bv.small[mid])+r {
			max = mid - 1
		} else {
			min = mid
		}
	}

	var smallPos uint64
	if rank+1 <= (max*smallBlockByte)-uint64(bv.small[max])+r {
		smallPos = min
	} else {
		smallPos = max
	}

	r += (smallPos * smallBlockByte) - uint64(bv.small[smallPos])
	if r == rank+1 {
		return smallPos * smallBlockBit, nil
	}

	for n := uint64(0); n < smallBlockByte; n++ {
		if uint64(len(bv.Bytes)) <= n+smallPos*smallBlockByte {
			break
		}
		b := bv.Bytes[n+smallPos*smallBlockByte]
		for i := uint64(0); i < 8; i++ {
			r += uint64((1 & (^b >> i)))
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
	if n == 0 {
		return 0
	} else {
		return Log2Floor(n-1) + 1
	}
}

type SparseBitVector struct {
	Bytes      []byte
	Length     uint64
	HighLength uint64
	LowLength  uint64
	weight     uint64
	rank1Max   uint64
	high       *BitVector
	low        []byte
}

func NewSparseBitVector(bytes []byte, bitsLen uint64) *SparseBitVector {

	weight := uint64(0)
	lastIndex := uint64(0)
	for i := uint64(0); i < bitsLen; i++ {
		bit := 1 & (bytes[i/8] >> (i % 8))
		if 1 == bit {
			weight++
			lastIndex = i
		}
	}
	lowLen := Log2Floor(bitsLen / weight)
	highLen := Log2Ceil(bitsLen) - lowLen
	lowMask := uint64((1 << lowLen) - 1)

	if lowLen == 0 {
		highBV := NewBitVector(bytes, bitsLen)
		bv := &SparseBitVector{
			bytes,
			bitsLen,
			highLen,
			lowLen,
			weight,
			highBV.rank1Max,
			highBV,
			nil,
		}
		return bv
	}

	lowSize := lowLen * weight
	highSize := weight
	highSize += uint64(lastIndex >> lowLen)

	low := make([]byte, bitToByte(lowSize))
	high := make([]byte, bitToByte(highSize))
	lowIndex := uint64(0)
	highIndex := uint64(0)
	prevHighValue := uint64(0)
	rank1 := uint64(0)
	for i := uint64(0); i < bitsLen; i++ {
		bit := 1 & (bytes[i/8] >> (i % 8))
		if bit != 1 {
			continue
		}
		rank1++

		lowVal := i & lowMask
		lowEnd := (lowIndex + lowLen - 1) / 8
		lowShift := lowIndex % 8
		for n := lowIndex / 8; n <= lowEnd; n++ {
			low[n] |= byte((lowVal << lowShift) & 0xFF)
			lowVal >>= (8 - lowShift)
			lowShift = 0
		}
		lowIndex += lowLen

		highValue := i >> lowLen
		inc := highValue - prevHighValue
		highIndex += inc
		high[highIndex/8] |= 1 << (highIndex % 8)
		highIndex++
		prevHighValue = highValue
	}

	highBV := NewBitVector(high, highSize)
	bv := &SparseBitVector{
		bytes,
		bitsLen,
		highLen,
		lowLen,
		weight,
		rank1,
		highBV,
		low,
	}
	return bv
}

func (bv *SparseBitVector) Rank1(pos uint64) (uint64, error) {
	if bv.Length < pos {
		return 0, fmt.Errorf("Length(=%d) < pos(=%d)", bv.Length, pos)
	}

	if bv.LowLength == 0 {
		return bv.high.Rank1(pos)
	}

	highRank0 := pos >> bv.LowLength
	if bv.high.Length-bv.high.rank1Max < highRank0 {
		return bv.rank1Max, nil
	}

	highPos := uint64(0)
	n := uint64(0)
	var err error
	if 1 <= highRank0 {
		highPos, err = bv.high.Select0(highRank0 - 1)
		if err != nil {
			return 0, fmt.Errorf("bv.high.Select0(%v) error: %v", highRank0-1, err)
		}
		highPos += 1
		n = highPos - highRank0
	}

	for i := highPos; 1 == 1&(bv.high.Bytes[i/8]>>(i%8)); i++ {

		lowIndex := bv.LowLength * n
		lowEnd := (lowIndex + bv.LowLength - 1) / 8
		leftShift := uint64(0)
		lowVal := uint64(0)
		for n := lowIndex / 8; n <= lowEnd; n++ {
			lowVal += uint64(bv.low[n]) << leftShift
			leftShift += 8
		}
		lowVal >>= lowIndex % 8
		lowVal &= uint64((1 << bv.LowLength) - 1)
		value := (highRank0 << bv.LowLength) + lowVal

		if pos <= value {
			return n, nil
		}
		n++
	}

	return n, nil

}

func (bv *SparseBitVector) Select1(rank uint64) (uint64, error) {
	result, err := bv.high.Select1(rank)
	if err != nil {
		return 0, err
	}
	if bv.LowLength == 0 {
		return result, err
	}
	result = (result - rank) << bv.LowLength
	if len(bv.low) != 0 {
		lowIndex := bv.LowLength * rank
		lowEnd := (lowIndex + bv.LowLength - 1) / 8
		leftShift := uint64(0)
		lowVal := uint64(0)
		for n := lowIndex / 8; n <= lowEnd; n++ {
			lowVal += uint64(bv.low[n]) << leftShift
			leftShift += 8
		}
		lowVal >>= lowIndex % 8
		lowVal &= uint64((1 << bv.LowLength) - 1)
		result += uint64(lowVal)
	}
	return result, nil
}

func (bv *BitVector) Dump() {
	fmt.Println("dupm high")
	for i, v := range bv.Bytes {
		fmt.Printf("Bytes[%d]: %x\n", i, v)
	}
	fmt.Printf("length %d\n", bv.Length)
	for i, v := range bv.large {
		fmt.Printf("large[%d]: %x\n", i, v)
	}
	for i, v := range bv.small {
		fmt.Printf("small[%d]: %x\n", i, v)
	}
	v, err := bv.Select0(0)
	fmt.Printf("select0 = %d, %v\n", v, err)
}
