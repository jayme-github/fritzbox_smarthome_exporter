package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
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

func (c *client) SafeLogin() error {
	c.Lock()
	defer c.Unlock()
	return c.Login()
}

func (c *client) SafeList() (*fritz.Devicelist, error) {
	c.Lock()
	defer c.Unlock()
	return c.List()
}

func NewClient(options ...fritz.Option) *client {
	return &client{
		HomeAuto: fritz.NewHomeAuto(options...),
		Mutex:    &sync.Mutex{},
	}
}

var (
	fritzClient     *client
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

	// Deprecate special syntax variabled for username and password.
	// All flags can be set via environment variables with the same name (uppercase)
	// like USERNAME for -username and PASSWORD for -password.
	fritzboxUser := os.Getenv("FRITZBOX_USER")
	fritzboxPassword := os.Getenv("FRITZBOX_PASSWORD")
	if fritzboxUser != "" {
		fmt.Println("You are using the deprecated environment variable \"FRITZBOX_USER\", please use \"USERNAME\" instead.")
		if *username == "" {
			*username = fritzboxUser
		}
	}
	if fritzboxPassword != "" {
		fmt.Println("You are using the deprecated environment variable \"FRITZBOX_PASSWORD\", please use \"PASSWORD\" instead.")
		if *password == "" {
			*password = fritzboxPassword
		}
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
		if err != nil {
			log.Fatalln("Unable to read certificate file:", err)
		}
		options = append(options, fritz.Certificate(crt))
	}

	fritzClient = NewClient(options...)

	if err := fritzClient.SafeLogin(); err != nil {
		log.Fatalln("Login failed:", err)
	}

	// Refresh login every 10 minutes
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			err := fritzClient.SafeLogin()
			if err != nil {
				log.Println("Login refresh failed:", err)
			}
		}
	}()

	fc := NewFritzCollector()
	prometheus.MustRegister(fc)
	http.Handle("/metrics", promhttp.Handler())

	if err := http.ListenAndServe(":9103", nil); err != nil {
		log.Fatalln(err)
	}
}
