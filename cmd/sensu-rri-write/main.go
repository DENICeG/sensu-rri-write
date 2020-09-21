package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/DENICeG/go-rriclient/pkg/rri"
	"github.com/danielb42/whiteflag"
	"github.com/gobuffalo/packr/v2"
)

var (
	timeBegin = time.Now()
	rriClient *rri.Client
	packrbox  = packr.New("box", "../../orderfile")
)

func main() {

	var err error
	log.SetOutput(os.Stderr)

	whiteflag.Alias("r", "regacc", "sets the regacc to use")
	whiteflag.Alias("p", "password", "sets the password to use")
	whiteflag.Alias("s", "server", "sets the RRI server to use")

	regacc := whiteflag.GetString("regacc")
	password := whiteflag.GetString("password")
	rriServer := whiteflag.GetString("server") + ":51131"

	rriClient, err = rri.NewClient(rriServer, nil)
	if err != nil {
		printFailMetricsAndExit("could not connect to RRI server:", err.Error())
	}

	err = rriClient.Login(regacc, password)
	if err != nil {
		printFailMetricsAndExit("login failed:", err.Error())
	}
	defer rriClient.Logout() // nolint:errcheck

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

}

func printFailMetricsAndExit(errors ...string) {

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
