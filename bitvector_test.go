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
			t.Errorf("bv.Rank1(%d) != %d", i, r)
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
			t.Errorf("bv.Rank1(%d) != %d", i, (i+1)/2)
		}
	}

    i := uint64(17)
	r, err := bv.Rank1(i)
	if err == nil {
		t.Errorf("Over Length error")
	}
	if r != 0 {
		t.Errorf("bv.Rank1(%d) != %d", i, 0)
	}
}

