package testsample

import (
	"testing"
)

func TestMissingSubtestRunner(t *testing.T) {
	testCases := []struct {
		inputNumber int
	}{
		{inputNumber: 1},
		{inputNumber: 2},
	}

	for _, currentCase := range testCases { // want `table-driven test loop in 'TestMissingSubtestRunner' must execute cases using t.Run`
		if currentCase.inputNumber == 0 {
			t.Fatalf("number cannot be zero")
		}
	}
}

func TestValidSubtestRunner(t *testing.T) {
	testCases := []struct {
		caseName    string
		inputNumber int
	}{
		{caseName: "positive", inputNumber: 1},
	}

	for _, currentCase := range testCases {
		t.Run(currentCase.caseName, func(subtest *testing.T) {
			if currentCase.inputNumber <= 0 {
				subtest.Errorf("expected positive number")
			}
		})
	}
}

func TestLoopWithBreakStatement(t *testing.T) {
	testNumbers := []int{1, 2, 3}

	for _, currentNumber := range testNumbers {
		if currentNumber == 2 {
			t.Errorf("found two")
			break
		}
	}
}
