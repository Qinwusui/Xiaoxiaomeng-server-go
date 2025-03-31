package test

import (
	"Xiaoxiaomeng-server/location"
	"testing"
)

func TestGetLocation(t *testing.T) {
	l, err := location.GetLocation("114.431221,30.503456")
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(string(l))
}
