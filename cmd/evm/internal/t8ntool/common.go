package t8ntool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
)

// a big.Int wrapper which marshals/unmarshals into byte arrays
type BigInt struct {
	*big.Int
}

func (x *BigInt) SetUint64(value uint64) *BigInt {
	if x.Int == nil {
		x.Int = new(big.Int)
	}

	_ = x.Int.SetUint64(value)
	return x
}

func (x *BigInt) MarshalText() (text []byte, err error) {
	if x == nil {
		return []byte("[<nil>]"), nil
	}

	slice := []byte("\"0x")

	slice = append(slice, []byte(x.Int.Text(16))...)
	slice = append(slice, []byte("\"")...)

	return slice, nil
}

func (z *BigInt) UnmarshalText(text []byte) error {
	// ignore the opening and end braces
	if z.Int == nil {
		z.Int = new(big.Int)
	}

	text = text[1 : len(text)-1]
	if _, success := z.Int.SetString(string(text), 0); !success {
		return fmt.Errorf("failed to unmarshal text")
	}

	return nil
}

func (x *BigInt) MarshalJSON() ([]byte, error) {
	return x.MarshalText()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (z *BigInt) UnmarshalJSON(text []byte) error {
	return z.UnmarshalText(text)
}

// implements rlp.Decoder
func (z *BigInt) DecodeRLP(s *rlp.Stream) error {
	if z.Int == nil {
		z.Int = new(big.Int)
	}
	err := decodeBigInt(s, z.Int)
	if err != nil {
		return err
	}
	return nil
}

func decodeBigInt(s *rlp.Stream, val *big.Int) error {
	dst := uint256.Int{}
	err := s.ReadUint256(&dst)
	if err != nil {
		return fmt.Errorf("%v, %v", err, val)
	}

	if val != nil {
		val.Set(new(big.Int).SetBytes(dst.Bytes()))
	} else {
		return fmt.Errorf("val is nil, can't set big.Int value")
	}
	return nil
}
