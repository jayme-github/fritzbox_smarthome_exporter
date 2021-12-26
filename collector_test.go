package main

import (
	"fmt"
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

func Test_parseSwitchStrings(t *testing.T) {
	tests := []struct {
		arg     string
		want    float64
		wantErr bool
	}{
		{arg: "manuell", want: 1.0, wantErr: false},
		{arg: "1", want: 1.0, wantErr: false},
		{arg: "0", want: 0.0, wantErr: false},
		{arg: "auto", want: 0.0, wantErr: false},
		{arg: "", want: -1.0, wantErr: true},
		{arg: "9", want: -1.0, wantErr: true},
	}
	for idx, tt := range tests {
		t.Run(fmt.Sprint(idx), func(t *testing.T) {
			got, err := parseSwitchStrings(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSwitchStrings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseSwitchStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}
