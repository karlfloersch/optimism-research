package fees

import (
	"math"
	"testing"
)

type CalcGasPriceTestCase struct {
	name                     string
	avgGasPerSecondLastEpoch float64
	expectedNextGasPrice     float64
}

func returnConstFn(retVal float64) func() float64 {
	return func() float64 { return retVal }
}

func runCalcGasPriceTests(gp L2GasPricer, tcs []CalcGasPriceTestCase, t *testing.T) {
	for _, tc := range tcs {
		if tc.expectedNextGasPrice != gp.CalcNextGasPrice(tc.avgGasPerSecondLastEpoch) {
			t.Fatalf("failed on test: %s", tc.name)
		}
	}
}

func TestCalcGasPriceFarFromFloor(t *testing.T) {
	gp := L2GasPricer{
		curPrice:                 100,
		floorPrice:               1,
		getTargetGasPerSecond:    returnConstFn(10),
		maxPercentChangePerEpoch: 0.5,
	}
	tcs := []CalcGasPriceTestCase{
		// No change
		{
			name:                     "No change expected when already at target",
			avgGasPerSecondLastEpoch: 10,
			expectedNextGasPrice:     100,
		},
		// Price reduction
		{
			name:                     "Max % change bounds the reduction in price",
			avgGasPerSecondLastEpoch: 1,
			expectedNextGasPrice:     50,
		},
		{
			// We're half of our target, so reduce by half
			name:                     "Reduce fee by half if at 50% capacity",
			avgGasPerSecondLastEpoch: 5,
			expectedNextGasPrice:     50,
		},
		{
			name:                     "Reduce fee by 75% if at 75% capacity",
			avgGasPerSecondLastEpoch: 7.5,
			expectedNextGasPrice:     75,
		},
		// Price increase
		{
			name:                     "Max % change bounds the increase in price",
			avgGasPerSecondLastEpoch: 100,
			expectedNextGasPrice:     150,
		},
		{
			name:                     "Increase fee by 25% if at 125% capacity",
			avgGasPerSecondLastEpoch: 12.5,
			expectedNextGasPrice:     125,
		},
	}
	runCalcGasPriceTests(gp, tcs, t)
}

func TestCalcGasPriceAtFloor(t *testing.T) {
	gp := L2GasPricer{
		curPrice:                 100,
		floorPrice:               100,
		getTargetGasPerSecond:    returnConstFn(10),
		maxPercentChangePerEpoch: 0.5,
	}
	tcs := []CalcGasPriceTestCase{
		// No change
		{
			name:                     "No change expected when already at target",
			avgGasPerSecondLastEpoch: 10,
			expectedNextGasPrice:     100,
		},
		// Price reduction
		{
			name:                     "No change expected when at floorPrice",
			avgGasPerSecondLastEpoch: 1,
			expectedNextGasPrice:     100,
		},
		// Price increase
		{
			name:                     "Max % change bounds the increase in price",
			avgGasPerSecondLastEpoch: 100,
			expectedNextGasPrice:     150,
		},
	}
	runCalcGasPriceTests(gp, tcs, t)
}

func TestGasPricerUpdates(t *testing.T) {
	gp := L2GasPricer{
		curPrice:                 100,
		floorPrice:               100,
		getTargetGasPerSecond:    returnConstFn(10),
		maxPercentChangePerEpoch: 0.5,
	}
	gp.UpdateGasPrice(12.5)
	if gp.curPrice != 125 {
		t.Fatalf("gp.curPrice not updated correctly. Got: %v, expected: %v", gp.curPrice, 125)
	}
}

func TestGasPricerDynamicTarget(t *testing.T) {
	counter := float64(9)
	dynamicGetTarget := func() float64 {
		counter += 1
		return counter
	}
	gp := L2GasPricer{
		curPrice:                 100,
		floorPrice:               100,
		getTargetGasPerSecond:    dynamicGetTarget,
		maxPercentChangePerEpoch: 0.5,
	}
	gp.UpdateGasPrice(12.5)
	expectedPrice := math.Ceil(100 * 12.5 / counter)
	if gp.curPrice != expectedPrice {
		t.Fatalf("gp.curPrice not updated correctly. Got: %v expected: %v", gp.curPrice, expectedPrice)
	}
	gp.UpdateGasPrice(12.5)
	expectedPrice = math.Ceil(expectedPrice * 12.5 / counter)
	if gp.curPrice != expectedPrice {
		t.Fatalf("gp.curPrice not updated correctly. Got: %v expected: %v", gp.curPrice, expectedPrice)
	}
}
