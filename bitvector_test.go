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
		if s != i * 2 {
			t.Errorf("%d(=bv.Select1(%d)) != %d ", s, i, i*2)
		}
	}
}

