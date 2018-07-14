package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/bpicode/fritzctl/fritz"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/namsral/flag"
)

type client struct {
	fritz.HomeAuto
	*sync.Mutex
}

var (
	fritzClient     client
	fbURL           *url.URL
	username        = flag.String("username", "", "FRITZ!Box username.")
	password        = flag.String("password", "", "FRITZ!Box password.")
	urlString       = flag.String("url", "https://fritz.box", "FRITZ!Box URL.")
	noVerify        = flag.Bool("noverify", false, "Omit TLS verification of the FRITZ!Box certificate.")
	certificatePath = flag.String("cert", "", "Path to the FRITZ!Box certificate.")
)

func validateFlags() {
	var err error
	flag.Parse()
	fbURL, err = url.Parse(*urlString)
	if err != nil {
		log.Fatalln(err)
	}
	if len(*username) == 0 {
		log.Fatalln("No username provided.")
	}
	if len(*password) == 0 {
		log.Fatalln("No password provided.")
	}
}

func main() {
	validateFlags()

	options := []fritz.Option{
		fritz.Credentials(*username, *password),
		fritz.URL(fbURL),
	}
	if *noVerify {
		options = append(options, fritz.SkipTLSVerify())
	}

	if !*noVerify && len(*certificatePath) > 0 {
		crt, err := ioutil.ReadFile(*certificatePath)
		if err == nil {
			options = append(options, fritz.Certificate(crt))
		} else {
			log.Fatalf("Unable to read certificate file: %v\n", err)
		}
	}

	fritzClient = client{
		HomeAuto: fritz.NewHomeAuto(options...),
		Mutex:    &sync.Mutex{},
	}

	// Refresh login every 10 minutes
	go func() {
		for {
			fritzClient.Lock()
			err := fritzClient.Login()
			if err != nil {
				log.Println("Login refresh failed:", err)
			}
			fritzClient.Unlock()
			time.Sleep(10 * time.Minute)
		}
	}()

	fc := NewFritzCollector()
	prometheus.MustRegister(fc)
	http.Handle("/metrics", promhttp.Handler())

	if err := http.ListenAndServe(":9103", nil); err != nil {
		log.Fatalln(err)
	}
}
