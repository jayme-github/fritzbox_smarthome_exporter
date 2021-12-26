package main

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bpicode/fritzctl/fritz"
	"github.com/bpicode/fritzctl/mock"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCollector(t *testing.T) {
	deviceLists, _ := filepath.Glob(path.Join("test", "*.xml"))
	for _, devicelistPath := range deviceLists {
		t.Run(devicelistPath, func(t *testing.T) {
			metricsPath := strings.TrimSuffix(devicelistPath, filepath.Ext(devicelistPath)) + ".metrics"
			exp, err := os.Open(metricsPath)
			if err != nil {
				t.Skipf("Error opening fixture file: %q: %v", metricsPath, err)
			}

			m := mock.New()
			m.DeviceList = devicelistPath
			m.Start()
			defer m.Close()

			fbURL, err := url.Parse(m.Server.URL)
			if err != nil {
				t.Errorf("Failed to parse mock server url: %v", err)
			}

			fritzClient = NewClient(fritz.URL(fbURL))
			fc := NewFritzCollector()
			if err := testutil.CollectAndCompare(fc, exp); err != nil {
				t.Error("Unexpected metrics returned:", err)
			}
		})
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
