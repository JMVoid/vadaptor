package utils

import (
	"testing"

)

func TestGetLoadAvg(t *testing.T) {
	out := GetLoadAvg()
	t.Log(out)

}
