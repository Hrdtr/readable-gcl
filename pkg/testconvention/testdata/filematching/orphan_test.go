package filematching // want `test file "orphan_test.go" has no corresponding implementation file 'orphan.go' and does not match directory name "filematching"`

import "testing"

func TestOrphan(t *testing.T) {
	_ = Service{Name: "orphan"}
}
