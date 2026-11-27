package suite // want: testconvention: test file "orphan_test.go" has no corresponding implementation file

import "testing"

func TestSuiteOrphanUnit(t *testing.T) {}
