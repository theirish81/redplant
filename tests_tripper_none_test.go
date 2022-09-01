package main

import (
	"net/http"
	"net/url"
	"testing"
)

func TestNoneTrip(t *testing.T) {
	req := http.Request{}
	req.URL, _ = url.Parse("none://foobar")
	res, _ := NoneTrip(&req)
	if res.StatusCode != 200 {
		t.Error("NoneTrip not working")
	}
}
