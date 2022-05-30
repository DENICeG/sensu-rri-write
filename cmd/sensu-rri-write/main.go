package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/DENICeG/go-rriclient/pkg/rri"
	"github.com/danielb42/whiteflag"
	"github.com/gobuffalo/packr/v2"
)

var (
	timeBegin time.Time
	rriClient *rri.Client
	packrbox  = packr.New("box", "../../orderfile")
	fails     int
	regacc    string
	password  string
	rriServer string
)

func main() {
	whiteflag.Alias("r", "regacc", "sets the regacc to use")
	whiteflag.Alias("p", "password", "sets the password to use")
	whiteflag.Alias("s", "server", "sets the RRI server to use")

	regacc = whiteflag.GetString("regacc")
	password = whiteflag.GetString("password")
	rriServer = whiteflag.GetString("server") + ":51131"

	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Intn(15)) * time.Second)

	timeBegin = time.Now()

	run()
}

func run() {
	var err error
	log.SetOutput(os.Stderr)
	log.SetPrefix("UTC ")
	log.SetFlags(log.Ltime | log.Lmsgprefix | log.LUTC)

	if rriClient != nil {
		rriClient.Logout() // nolint:errcheck
	}

	rriClient, err = rri.NewClient(rriServer, &rri.ClientConfig{})
	if err != nil {
		printFailMetricsAndExit("could not connect to RRI server:", err.Error())
	}

	err = rriClient.Login(regacc, password)
	if err != nil {
		printFailMetricsAndExit("login failed:", err.Error())
	}

	timeLoginDone := time.Now()

	rriQuery, err := packrbox.FindString("order.rri")
	if err != nil {
		panic(err)
	}

	rriResponse, err := rriClient.SendRaw(rriQuery)
	if err != nil {
		printFailMetricsAndExit("SendRaw() failed:", err.Error())
	}

	if !strings.Contains(rriResponse, "RESULT: success") {
		printFailMetricsAndExit(rriResponse)
	}

	durationLogin := timeLoginDone.Sub(timeBegin).Milliseconds()
	durationOrder := time.Now().Sub(timeLoginDone).Milliseconds() // nolint:gosimple
	durationTotal := durationLogin + durationOrder

	log.Printf("OK: RRI login + order: %dms + %dms = %dms\n\n", durationLogin, durationOrder, durationTotal)
	fmt.Printf("extmon,service=%s,ordertype=%s %s=%d,%s=%d,%s=%d,%s=%d %d\n",
		"rri",
		"WRITE",
		"available", 1,
		"login", durationLogin,
		"order", durationOrder,
		"total", durationTotal,
		timeBegin.Unix())

	rriClient.Logout() // nolint:errcheck
	os.Exit(0)
}

func printFailMetricsAndExit(errors ...string) {

	if fails < 3 {
		fails++
		run()
	}

	errStr := "ERROR:"

	for _, err := range errors {
		errStr += " " + err
	}

	log.Printf("%s\n\n", errStr)

	fmt.Printf("extmon,service=%s,ordertype=%s %s=%d,%s=%d,%s=%d,%s=%d %d\n",
		"rri",
		"WRITE",
		"available", 0,
		"login", 0,
		"order", 0,
		"total", 0,
		timeBegin.Unix())

	if rriClient != nil {
		rriClient.Logout() // nolint:errcheck
	}

	os.Exit(2)
}
