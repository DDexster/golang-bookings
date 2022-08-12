package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNoSurve(t *testing.T) {
	var mh myHandler

	h := NoSurve(&mh)

	switch v := h.(type) {
	case http.Handler:
	//	Do nothing
	default:
		t.Error(fmt.Sprintf("Type is not an http.Handler, but it is %T", v))
	}
}

func TestSessionLoad(t *testing.T) {
	var mh myHandler

	h := SessionLoad(&mh)

	switch v := h.(type) {
	case http.Handler:
	//	Do nothing
	default:
		t.Error(fmt.Sprintf("Type is not an http.Handler, but it is %T", v))
	}
}
