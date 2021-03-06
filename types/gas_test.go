package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGasMeter(t *testing.T) {
	cases := []struct {
		limit Gas
		usage []Gas
	}{
		{10, []Gas{1, 2, 3, 4}},
		{1000, []Gas{40, 30, 20, 10, 900}},
		{100000, []Gas{99999, 1}},
		{100000000, []Gas{50000000, 40000000, 10000000}},
		{65535, []Gas{32768, 32767}},
		{65536, []Gas{32768, 32767, 1}},
	}

	for tcnum, tc := range cases {
		meter := NewGasMeter(tc.limit)
		used := uint64(0)

		for unum, usage := range tc.usage {
			used += usage
			require.NotPanics(t, func() { meter.ConsumeGas(usage, "") }, "Not exceeded limit but panicked. tc #%d, usage #%d", tcnum, unum)
			require.Equal(t, used, meter.GasConsumed(), "Gas consumption not match. tc #%d, usage #%d", tcnum, unum)
			require.Equal(t, used, meter.GasConsumedToLimit(), "Gas consumption (to limit) not match. tc #%d, usage #%d", tcnum, unum)
			require.False(t, meter.IsPastLimit(), "Not exceeded limit but got IsPastLimit() true")
			if unum < len(tc.usage)-1 {
				require.False(t, meter.IsOutOfGas(), "Not yet at limit but got IsOutOfGas() true")
			} else {
				require.True(t, meter.IsOutOfGas(), "At limit but got IsOutOfGas() false")
			}
		}

		require.Panics(t, func() { meter.ConsumeGas(1, "") }, "Exceeded but not panicked. tc #%d", tcnum)
		require.Equal(t, meter.GasConsumedToLimit(), meter.Limit(), "Gas consumption (to limit) not match limit")
		require.Equal(t, meter.GasConsumed(), meter.Limit()+1, "Gas consumption not match limit+1")
		break

	}
}

func TestGasMeterWithLog(t *testing.T) {
	cases := []struct {
		limit  Gas
		usage  Gas
		base   float64
		shift  uint64
		desc   string
		expect Gas
	}{
		{1000, 100, 10, 0, GasWritePerByteDesc, 2},
		{1000, 285, 1.02, 285, GasWritePerByteDesc, 285},
		{1000, 286, 1.02, 285, GasWritePerByteDesc, 285},
		{1000, 288, 1.02, 285, GasWritePerByteDesc, 285},
		{1000, 289, 1.02, 285, GasWritePerByteDesc, 286},
		{1000, 100, 10, 100, GasWritePerByteDesc, 100},
		{1000, 100, 10, 0, GasReadPerByteDesc, 2},
		{1000, 100, 10, 0, "", 100},
		{1000, 100, 1, 0, GasWritePerByteDesc, 100},
		{1000, 100, 0, 0, GasWritePerByteDesc, 100},
		{1000, 100, 0.99, 0, GasWritePerByteDesc, 100},
		{1000, 100, -0.1, 0, GasWritePerByteDesc, 100},
	}

	for tcnum, tc := range cases {
		meter := NewGasMeterWithBase(tc.limit, tc.base, tc.shift)

		meter.ConsumeGas(tc.usage, tc.desc)
		require.Equal(t, tc.expect, meter.GasConsumed(), "Gas consumption not match. tc #%d", tcnum)
	}
}
