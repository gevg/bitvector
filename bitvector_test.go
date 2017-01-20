package bitvector_test

import (
	"github.com/tobi-c/bitvector"
	"testing"
)

func TestRank1(t *testing.T) {
	b := []byte{0xFF, 0xFF}
	bv := bitvector.NewBitVector(b, 16)
	for i := uint64(0); i <= 16; i++ {
		r, err := bv.Rank1(i)
		if err != nil {
			t.Errorf(err.Error())
		}
		if r != i {
			t.Errorf("%d(=bv.Rank1(%d)) != %d", r, i, i)
		}
	}

	b = []byte{0x55, 0x55}
	bv = bitvector.NewBitVector(b, 16)
	for i := uint64(0); i <= 16; i++ {
		r, err := bv.Rank1(i)
		if err != nil {
			t.Errorf(err.Error())
		}
		if r != (i+1)/2 {
			t.Errorf("%d(=bv.Rank1(%d)) != %d", r, i, (i+1)/2)
		}
	}

	i := uint64(17)
	r, err := bv.Rank1(i)
	if err == nil {
		t.Errorf("Over Length error")
	}
	if r != 0 {
		t.Errorf("%d(=bv.Rank1(%d)) != %d", r, i, 0)
	}
}

func TestSelect1(t *testing.T) {
	b := []byte{0xFF, 0xFF}
	bv := bitvector.NewBitVector(b, 16)
	for i := uint64(0); i < 16; i++ {
		s, err := bv.Select1(i)
		if err != nil {
			t.Errorf(err.Error())
		}
		if s != i {
			t.Errorf("%d(=bv.Select1(%d)) != %d ", s, i, i)
		}
	}

	b = []byte{0x55, 0x55}
	bv = bitvector.NewBitVector(b, 16)
	for i := uint64(0); i < 8; i++ {
		s, err := bv.Select1(i)
		if err != nil {
			t.Errorf(err.Error())
		}
		if s != i*2 {
			t.Errorf("%d(=bv.Select1(%d)) != %d ", s, i, i*2)
		}
	}

	i := uint64(8)
	r, err := bv.Select1(i)
	if err == nil {
		t.Errorf("Over rank error")
	}
	if r != 0 {
		t.Errorf("%d(=bv.Select1(%d)) != %d", r, i, 0)
	}
}

var cacheBV = make(map[uint64](*bitvector.BitVector))

func newBitVectorForBench(length uint64) *bitvector.BitVector {
	bv := cacheBV[length]
	if bv == nil {
		s := make([]byte, length)
		for i := 0; i < len(s); i++ {
			s[i] = 0x55
		}
		bv = bitvector.NewBitVector(s, uint64(len(s)*8))
		cacheBV[length] = bv
	}
	return bv
}

func BenchmarkRank1_100000(b *testing.B) {
	bv := newBitVectorForBench(100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bv.Rank1(bv.Length - 1)
	}
}

func BenchmarkRank1_100000000(b *testing.B) {
	bv := newBitVectorForBench(100000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bv.Rank1(bv.Length - 1)
	}
}

func BenchmarkSelect1_100000(b *testing.B) {
	bv := newBitVectorForBench(100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bv.Select1(bv.Length / 2)
	}
}

func BenchmarkSelect1_100000000(b *testing.B) {
	bv := newBitVectorForBench(100000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bv.Select1(bv.Length / 2)
	}
}

func TestSparseBitVector(t *testing.T) {
	b := []byte{0xFF, 0xFF}
	bitvector.NewSparseBitVector(b, 16)
}

func TestSparseBitVectorSelect1(t *testing.T) {
	b := []byte{}
	for i := 0; i < 33; i++ {
		b = append(b, 0xFF)
	}
	bv := bitvector.NewSparseBitVector(b, uint64(len(b)*8))
	for r := uint64(1); r < uint64(len(b)*8); r++ {
		s, err := bv.Select1(r)
		if err != nil {
			t.Errorf("Over rank error")
		}
		if s != r {
			t.Errorf("err: %d %d ", s, r)
		}
	}
}
