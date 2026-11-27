package ignored

import "testing"

var userId int // want: namingclarity: disallowed initialism casing "Id"

func TestOther_bad(t *testing.T) {
	for range []int{1} {
		t.Error("test convention intentionally ignored here")
	}
}
