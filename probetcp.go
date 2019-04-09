package main

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
)

const usageHeader = `%s probes TCP endpoints.

Usage:
  %s [OPTION...] TARGET

Options:
`

const usageFooter = `
Examples:
  Probe a MySQL database over TCP (10 attempts every 2 seconds):
    probecli -r 9 -i 2000 tcp://localhost:3306
`

const tcpTargetRegexp = "^(?:tcp:\\/\\/)?(.+)$"

func main() {

	var interval int
	var retries int
	var quiet bool
	var timeout int

	flag.CommandLine.SetOutput(os.Stdout)

	flag.IntVarP(&interval, "interval", "i", 1e3, "Time to wait between probe retries in milliseconds")
	flag.IntVarP(&retries, "retries", "r", 0, "Number of times to retry to probe the target if it fails (default 0)")
	flag.BoolVarP(&quiet, "quiet", "q", false, "Do not print anything (default false)")
	flag.IntVarP(&timeout, "timeout", "t", 60e3, "TCP connection timeout in milliseconds")

	flag.Usage = func() {
		fmt.Printf(usageHeader, os.Args[0], os.Args[0])
		flag.PrintDefaults()
		fmt.Print(usageFooter)
	}

	flag.Parse()

	target := flag.Arg(0)

	if interval < 0 {
		fail(quiet, "the \"interval\" option must be greater than or equal to zero")
	} else if retries < 0 {
		fail(quiet, "the \"retries\" option must be greater than or equal to zero")
	} else if timeout <= 0 {
		fail(quiet, "the \"timeout\" option must be greater than zero")
	} else if target == "" {
		fail(quiet, "a target to probe must be given as an argument (e.g. \"tcp://localhost:3306\")")
	}

	tcpRegexp := regexp.MustCompile(tcpTargetRegexp)

	config := &TCPProbeConfig{}
	config.Address = tcpRegexp.ReplaceAllString(target, "$1")
	config.Interval = time.Duration(interval * 1e6)
	config.Retries = retries
	config.Timeout = time.Duration(timeout * 1e6)

	config.OnAttempt = func(attempt int, config *TCPProbeConfig, _ *error) {
		if config.Retries != 0 && !quiet {
			fmt.Printf("Probing %s (%d)...\n", config.Address, attempt+1)
		}
	}

	result, err := probeTCPEndpoint(config)
	if err != nil {
		fail(quiet, "probe error: %s", err)
	} else if !result.Success {
		fail(quiet, "Probe failed after %fs", result.Duration.Seconds())
	}

	succeed(quiet, "Probe succeeded after %fs", result.Duration.Seconds())
}

func fail(quiet bool, format string, values ...interface{}) {
	if !quiet {
		fmt.Fprintf(os.Stderr, color.RedString("Error: "+format+"\n"), values...)
	}

	os.Exit(1)
}

func succeed(quiet bool, format string, values ...interface{}) {
	if !quiet {
		fmt.Printf(color.GreenString(format+"\n", values...))
	}
}