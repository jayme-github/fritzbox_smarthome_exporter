package main

import (
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/bpicode/fritzctl/fritz"
	"github.com/bpicode/fritzctl/mock"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCollector(t *testing.T) {
	m := mock.New()
	m.DeviceList = "test/devicelist.xml"
	m.Start()
	defer m.Close()

	fbURL, err := url.Parse(m.Server.URL)
	if err != nil {
		t.Fatalf("Failed to parse mock server url: %v", err)
	}

	fritzClient = NewClient(fritz.URL(fbURL))
	fc := NewFritzCollector()

	fixture := "test.metrics"
	exp, err := os.Open(path.Join("test", fixture))
	if err != nil {
		t.Fatalf("Error opening fixture file: %q: %v", fixture, err)
	}
	if err := testutil.CollectAndCompare(fc, exp); err != nil {
		t.Fatal("Unexpected metrics returned:", err)
	}
}
