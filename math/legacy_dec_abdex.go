package math

import (
	"bytes"
	"errors"
	"math/big"
)

// Dec is an alias for LegacyDec
type Dec = LegacyDec

// Marshal implements the gogo proto custom type interface.
func (d LegacyDec) Marshal() ([]byte, error) {
	i := d.i
	if i == nil {
		i = new(big.Int)
	}

	// Use the ordered encoding instead of the default MarshalText
	result := new(bytes.Buffer)

	// Get the absolute value bytes
	absI := new(big.Int).Abs(i)
	rawBytes := absI.Bytes()

	// Write length (fixed 1 byte, supports integers up to 255 bytes)
	if len(rawBytes) > 255 {
		return nil, errors.New("integer too large, exceeds 255 bytes")
	}

	if i.Sign() < 0 {
		// For negative numbers: sign=0, inverted length, inverted content
		result.WriteByte(0) // negative sign
		// Length needs to be inverted: larger absolute values should come first (smaller in sort order)
		result.WriteByte(255 - byte(len(rawBytes)))
		// Content also needs to be inverted for proper ordering
		for _, b := range rawBytes {
			result.WriteByte(255 - b)
		}
	} else {
		// For positive or zero: sign=1, normal length, normal content
		result.WriteByte(1)                   // positive sign
		result.WriteByte(byte(len(rawBytes))) // normal length
		result.Write(rawBytes)
	}

	return result.Bytes(), nil
}

// MarshalTo implements the gogo proto custom type interface.
func (d *LegacyDec) MarshalTo(data []byte) (n int, err error) {
	i := d.i
	if i == nil {
		i = new(big.Int)
	}

	// For zero value, we still encode it properly using our ordered encoding
	bz, err := d.Marshal()
	if err != nil {
		return 0, err
	}

	if len(bz) > len(data) {
		return 0, errors.New("buffer too small")
	}

	copy(data, bz)
	return len(bz), nil
}

// Unmarshal implements the gogo proto custom type interface.
func (d *LegacyDec) Unmarshal(data []byte) error {
	if len(data) == 0 {
		d = nil
		return nil
	}

	if d.i == nil {
		d.i = new(big.Int)
	}

	if len(data) < 2 {
		return errors.New("invalid encoding: too short")
	}

	// Read sign byte
	sign := data[0]

	// Read length
	var length int
	if sign == 0 {
		// Negative number: length was inverted during encoding
		length = int(255 - data[1])
	} else {
		// Positive number: length is normal
		length = int(data[1])
	}

	if len(data) < 2+length {
		return errors.New("invalid encoding: content length mismatch")
	}

	// Read content
	content := data[2 : 2+length]

	// Construct big.Int
	if sign == 0 {
		// Negative number: content was inverted during encoding
		inverted := make([]byte, len(content))
		for i, b := range content {
			inverted[i] = 255 - b
		}
		d.i.SetBytes(inverted)
		d.i.Neg(d.i)
	} else {
		// Positive or zero: content is normal
		d.i.SetBytes(content)
	}

	if !d.IsInValidRange() {
		return errors.New("decimal out of range")
	}
	return nil
}
