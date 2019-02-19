package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

var opts struct {
	LogLevel string `long:"log-level" description:"Log level" default:"info"`
	BindAddr string `long:"bind-address" description:"address for binding checks to" default:":8080"`

	ReadyDelay time.Duration `long:"ready-delay" env:"READY_DELAY" description:"Duration to wait before becoming ready"`
}

func main() {
	sigs := make(chan os.Signal)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		// If the error was from the parser, then we can simply return
		// as Parse() prints the error already
		if _, ok := err.(*flags.Error); ok {
			os.Exit(1)
		}
		logrus.Fatalf("Error parsing flags: %v", err)
	}

	// Use log level
	level, err := logrus.ParseLevel(opts.LogLevel)
	if err != nil {
		logrus.Fatalf("Unknown log level %s: %v", opts.LogLevel, err)
	}
	logrus.SetLevel(level)

	// Set the log format to have a reasonable timestamp
	formatter := &logrus.TextFormatter{
		FullTimestamp: true,
	}
	logrus.SetFormatter(formatter)

	logrus.Infof("Starting")

	var ready bool

	l, err := net.Listen("tcp", opts.BindAddr)
	if err != nil {
		logrus.Fatalf("Error binding: %v", err)
	}

	go func() {
		http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
			if !ready {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		})
		// TODO: log error?
		http.Serve(l, http.DefaultServeMux)
	}()

	go func() {
		time.Sleep(opts.ReadyDelay)
		logrus.Infof("marking as ready")
		ready = true
	}()

	// Wait for kill signal
	for {
		select {
		case sig := <-sigs:
			switch sig {
			case syscall.SIGTERM, syscall.SIGINT:
				logrus.Infof("Got signal to stop, stopping")
				return
			}
		}
	}
}
