package suite

import "testing"

func TestSuiteWorksUnit(t *testing.T)            {}
func TestSuite_badUnit(t *testing.T)             {} // want: testconvention: test function name "TestSuite_badUnit" does not match configured pattern
func TestOtherWorksUnit(t *testing.T)            {} // want: testconvention: test function name in folder 'suite' must start with
func TestSuiteWrongTierIntegration(t *testing.T) {} // want: testconvention: test in unit test file must end with 'Unit'

func TestSuiteSubtestsUnit(t *testing.T) {
	for _, number := range []int{1} {
		t.Run("case", func(t *testing.T) {
			if number == 0 {
				t.Error("unexpected zero")
			}
		})
	}
}

func TestSuiteMissingSubtestsUnit(t *testing.T) {
	for _, number := range []int{1} { // want: testconvention: table-driven test loop in 'TestSuiteMissingSubtestsUnit' must execute cases using t.Run
		if number == 0 {
			t.Error("unexpected zero")
		}
	}
}
