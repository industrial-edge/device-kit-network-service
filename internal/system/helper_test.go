package system

import (
	"testing"
)

func TestHelperMethods(t *testing.T) {

	var sut = Helper{}

	if sut.DaemonReload("true") != true {
		t.FailNow()
	}
	if sut.DaemonReload("false") != false {
		t.FailNow()
	}
}
