package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func TestFileTrip(t *testing.T) {
	req := http.Request{}
	req.URL, _ = url.Parse("file://etc/files/data.json")
	res, _ := FileTrip(&req)
	data1, _ := ioutil.ReadAll(res.Body)
	data2, _ := ioutil.ReadFile("etc/files/data.json")
	if string(data1) != string(data2) {
		t.Error("File tripper does not work according to plan")
	}
}
