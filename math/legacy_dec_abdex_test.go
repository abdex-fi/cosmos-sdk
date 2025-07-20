// gbr edit

package math

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewOverflowLimits tests that the new overflow limits are much higher
func TestNewOverflowLimits(t *testing.T) {
	// Test that we can create much larger numbers without overflow
	// Old limit was 2^256 * 10^18, new limit is 2^512 * 10^18

	// Create a large number that would have overflowed with old limits
	largeNum := new(big.Int).Exp(big.NewInt(2), big.NewInt(300), nil)
	largeNum.Mul(largeNum, precisionReuse)

	dec := LegacyNewDecFromBigIntWithPrec(largeNum, LegacyPrecision)
	require.True(t, dec.IsInValidRange(), "Large number should be within new limits")

	// Test arithmetic operations with large numbers
	dec2 := LegacyNewDecFromBigIntWithPrec(largeNum, LegacyPrecision)
	result := dec.Add(dec2)
	require.True(t, result.IsInValidRange(), "Addition of large numbers should not overflow")

	// Test that extremely large numbers still cause overflow
	extremeNum := new(big.Int).Exp(big.NewInt(2), big.NewInt(600), nil)
	extremeNum.Mul(extremeNum, precisionReuse)
	extremeDec := LegacyNewDecFromBigIntWithPrec(extremeNum, LegacyPrecision)
	require.False(t, extremeDec.IsInValidRange(), "Extremely large number should still overflow")
}

// TestOrderedSerialization tests that the new serialization maintains order
func TestOrderedSerialization(t *testing.T) {
	testCases := []struct {
		name string
		decs []LegacyDec
	}{
		{
			name: "positive numbers",
			decs: []LegacyDec{
				LegacyNewDec(1),
				LegacyNewDec(2),
				LegacyNewDec(10),
				LegacyNewDec(100),
				LegacyNewDec(1000),
			},
		},
		{
			name: "negative numbers",
			decs: []LegacyDec{
				LegacyNewDec(-1000),
				LegacyNewDec(-100),
				LegacyNewDec(-10),
				LegacyNewDec(-2),
				LegacyNewDec(-1),
			},
		},
		{
			name: "mixed positive and negative",
			decs: []LegacyDec{
				LegacyNewDec(-100),
				LegacyNewDec(-1),
				LegacyZeroDec(),
				LegacyNewDec(1),
				LegacyNewDec(100),
			},
		},
		{
			name: "decimal values",
			decs: []LegacyDec{
				LegacyNewDecWithPrec(-150, 2), // -1.50
				LegacyNewDecWithPrec(-100, 2), // -1.00
				LegacyNewDecWithPrec(-50, 2),  // -0.50
				LegacyZeroDec(),               // 0.00
				LegacyNewDecWithPrec(50, 2),   // 0.50
				LegacyNewDecWithPrec(100, 2),  // 1.00
				LegacyNewDecWithPrec(150, 2),  // 1.50
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Serialize all decimals
			serialized := make([][]byte, len(tc.decs))
			for i, dec := range tc.decs {
				bz, err := dec.Marshal()
				require.NoError(t, err)
				serialized[i] = bz
			}

			// Check that serialized bytes are in ascending order
			for i := 1; i < len(serialized); i++ {
				cmp := bytes.Compare(serialized[i-1], serialized[i])
				require.True(t, cmp <= 0,
					"Serialized bytes should be in ascending order: %v vs %v (decimals: %v vs %v)",
					serialized[i-1], serialized[i], tc.decs[i-1], tc.decs[i])
			}
		})
	}
}

// TestSerializationRoundTrip tests that serialization and deserialization work correctly
func TestSerializationRoundTrip(t *testing.T) {
	testCases := []LegacyDec{
		LegacyZeroDec(),
		LegacyNewDec(1),
		LegacyNewDec(-1),
		LegacyNewDec(123456789),
		LegacyNewDec(-123456789),
		LegacyNewDecWithPrec(123456789, 9),
		LegacyNewDecWithPrec(-123456789, 9),
		LegacySmallestDec(),
		LegacySmallestDec().Neg(),
	}

	for i, original := range testCases {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			// Test Marshal/Unmarshal
			bz, err := original.Marshal()
			require.NoError(t, err)

			var decoded LegacyDec
			err = decoded.Unmarshal(bz)
			require.NoError(t, err)
			require.True(t, original.Equal(decoded),
				"Round trip failed: original=%v, decoded=%v", original, decoded)

			// Test MarshalTo
			data := make([]byte, 1000) // Large buffer
			n, err := original.MarshalTo(data)
			require.NoError(t, err)

			var decoded2 LegacyDec
			err = decoded2.Unmarshal(data[:n])
			require.NoError(t, err)
			require.True(t, original.Equal(decoded2),
				"MarshalTo round trip failed: original=%v, decoded=%v", original, decoded2)
		})
	}
}

// TestJSONCompatibility tests that JSON serialization is unchanged
func TestJSONCompatibility(t *testing.T) {
	testCases := []LegacyDec{
		LegacyZeroDec(),
		LegacyNewDec(1),
		LegacyNewDec(-1),
		LegacyNewDec(123456789),
		LegacyNewDecWithPrec(123456789, 9),
	}

	for i, dec := range testCases {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			// JSON serialization should be unchanged
			jsonBz, err := json.Marshal(dec)
			require.NoError(t, err)

			var decoded LegacyDec
			err = json.Unmarshal(jsonBz, &decoded)
			require.NoError(t, err)
			require.True(t, dec.Equal(decoded),
				"JSON round trip failed: original=%v, decoded=%v", dec, decoded)

			// JSON should still be human readable
			require.Contains(t, string(jsonBz), dec.String())
		})
	}
}

// TestSerializationOrdering tests that serialized bytes maintain numerical order
func TestSerializationOrdering(t *testing.T) {
	// Create a list of decimals in random order
	decimals := []LegacyDec{
		LegacyNewDec(100),
		LegacyNewDec(-50),
		LegacyZeroDec(),
		LegacyNewDec(1),
		LegacyNewDec(-1),
		LegacyNewDec(50),
		LegacyNewDec(-100),
		LegacyNewDecWithPrec(150, 2), // 1.50
		LegacyNewDecWithPrec(-75, 2), // -0.75
	}

	// Sort by numerical value
	sort.Slice(decimals, func(i, j int) bool {
		return decimals[i].LT(decimals[j])
	})

	// Serialize all
	serialized := make([][]byte, len(decimals))
	for i, dec := range decimals {
		bz, err := dec.Marshal()
		require.NoError(t, err)
		serialized[i] = bz
	}

	// Check that serialized bytes are also in ascending order
	for i := 1; i < len(serialized); i++ {
		cmp := bytes.Compare(serialized[i-1], serialized[i])
		require.True(t, cmp < 0,
			"Serialized bytes should be in strict ascending order: %v vs %v (decimals: %v vs %v)",
			serialized[i-1], serialized[i], decimals[i-1], decimals[i])
	}
}

// TestSpecificNegativeOrdering tests the specific negative number ordering issue
func TestSpecificNegativeOrdering(t *testing.T) {
	// Test the specific case mentioned in the feedback
	decimals := []LegacyDec{
		LegacyNewDec(-100), // Should come first (most negative)
		LegacyNewDec(-10),  // Should come second
		LegacyNewDec(-1),   // Should come third
		LegacyZeroDec(),    // Should come fourth
		LegacyNewDec(1),    // Should come fifth
		LegacyNewDec(10),   // Should come sixth
		LegacyNewDec(100),  // Should come last (most positive)
	}

	// Serialize all
	serialized := make([][]byte, len(decimals))
	for i, dec := range decimals {
		bz, err := dec.Marshal()
		require.NoError(t, err)
		serialized[i] = bz
		t.Logf("Dec %v -> bytes %v", dec, bz)
	}

	// Check that serialized bytes are in strict ascending order
	for i := 1; i < len(serialized); i++ {
		cmp := bytes.Compare(serialized[i-1], serialized[i])
		require.True(t, cmp < 0,
			"Serialized bytes should be in strict ascending order: %v vs %v (decimals: %v vs %v)",
			serialized[i-1], serialized[i], decimals[i-1], decimals[i])
	}
}

// TestEdgeCases tests edge cases and boundary values
func TestEdgeCases(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		zero := LegacyZeroDec()
		bz, err := zero.Marshal()
		require.NoError(t, err)

		var decoded LegacyDec
		err = decoded.Unmarshal(bz)
		require.NoError(t, err)
		require.True(t, zero.Equal(decoded))
	})

	t.Run("nil value", func(t *testing.T) {
		var nilDec LegacyDec
		bz, err := nilDec.Marshal()
		require.NoError(t, err)

		var decoded LegacyDec
		err = decoded.Unmarshal(bz)
		require.NoError(t, err)
		// Both should represent zero
		require.True(t, decoded.IsZero())
	})

	t.Run("smallest decimal", func(t *testing.T) {
		smallest := LegacySmallestDec()
		bz, err := smallest.Marshal()
		require.NoError(t, err)

		var decoded LegacyDec
		err = decoded.Unmarshal(bz)
		require.NoError(t, err)
		require.True(t, smallest.Equal(decoded))
	})
}
