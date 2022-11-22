package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	hapLog "github.com/brutella/hap/log"
	"github.com/ryansouza/aranet4-exporter/aranet"
)

const incorrectArgumentExitCode = 2
const bluetoothMacPattern = "^([[:xdigit:]]{2}:){5}[[:xdigit:]]{2}$"

type AranetConfig struct {
	ID   string
	Name string
}

func (a *AranetConfig) Set(s string) error {
	splits := strings.Split(s, "=")
	if !isValidBluetoothMac(splits[0]) {
		return fmt.Errorf("not a valid Bluetooth MAC address")
	}

	if len(splits) > 2 {
		return fmt.Errorf("invalid Aranet device: must be of format {ID} or {ID}={name}")
	}
	if len(splits) > 1 {
		a.Name = splits[1]
	} else {
		a.Name = splits[0]
	}

	a.ID = splits[0]

	return nil
}

func isValidBluetoothMac(a string) bool {
	matched, err := regexp.MatchString(bluetoothMacPattern, a)
	if err != nil {
		panic(err)
	}
	return matched
}

var aranets []AranetConfig
var homekit bool
var listen string
var verbose bool
var stateDir string

func init() {
	flag.BoolVar(&verbose, "verbose", false, "verbose logging")
	flag.StringVar(&listen, "listen", ":9963", "address to expose Prometheus metrics on")
	flag.BoolVar(&homekit, "homekit", false, "enable HomeKit support")
	flag.StringVar(&stateDir, "state", getDefaultStateDir(), "directory to store persistent state")
	flag.Func("device", "monitor an Aranet4 with format {ID} or {ID}={name}. may be specified multiple "+
		"times. examples: -device D8:9B:67:AA:BB:CC=bedroom -device D8:9B:67:AA:BB:DD", parseAranet)

	flag.Parse()

	if len(aranets) == 0 {
		fmt.Fprintf(os.Stderr, "Error: -device flag is required.\n\n")
		printHelpAndExit()
	}
}

// getDefaultStateDir returns a suitable default location for the state that may need to persist between runs of this
// program.
func getDefaultStateDir() string {
	// STATE_DIRECTORY may be set by systemd.
	if v, ok := os.LookupEnv("STATE_DIRECTORY"); ok {
		return v
	}
	return "./"
}

func printHelpAndExit() {
	fmt.Fprintln(os.Stderr, "Usage:")
	flag.PrintDefaults()
	os.Exit(incorrectArgumentExitCode)
}

func parseAranet(s string) error {
	a := AranetConfig{}

	if err := a.Set(s); err != nil {
		return err
	}

	aranets = append(aranets, a)

	return nil
}

func googleQRCode(uri string) string {
	return "https://chart.googleapis.com/chart?cht=qr&chl=" + uri + "&chs=400x400"
}

func serveMetricsHTTP(shutdownContext context.Context, shutdownWait *sync.WaitGroup, reg *prometheus.Registry) {
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	server := http.Server{Addr: listen}

	go func() {
		<-shutdownContext.Done()
		log.Println("Stopping httpserver...")
		server.Shutdown(context.Background())
	}()

	shutdownWait.Add(1)
	go func() {
		defer shutdownWait.Done()

		log.Printf("Prometheus metrics: %s/metrics\n", server.Addr)
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http.Server closed prematurely: %T %v", err, err)
		}
		log.Println("Stopped httpserver.")
	}()
}

func serveHomekitServer(shutdownContext context.Context, shutdownWait *sync.WaitGroup, bridge *aranet.Bridge) {
	shutdownWait.Add(1)
	go func() {
		defer shutdownWait.Done()

		log.Printf("Running homekit server with pin: %s\n", bridge.Pin)
		log.Printf("Homekit setup QR code: %s\n", googleQRCode(bridge.SetupURI))
		e := bridge.Serve(shutdownContext)
		log.Printf("Stopped homekit server: %v\n", e)
	}()
}

func main() {
	shutdownContext, shutdown := context.WithCancel(context.Background())
	shutdownWait := sync.WaitGroup{}

	seenIds := map[string]bool{}
	seenNames := map[string]bool{}

	collectedAranets := []aranet.AranetData{}
	collectedAccessories := []*aranet.Accessory{}
	for _, config := range aranets {
		if _, found := seenIds[config.ID]; found {
			fmt.Fprintf(os.Stderr, "Error: duplicate Bluetooth IDs are not allowed: %s\n", config.ID)
			os.Exit(incorrectArgumentExitCode)
		}
		if _, found := seenNames[config.Name]; found {
			fmt.Fprintf(os.Stderr, "Error: duplicate device names are not allowed: %s\n", config.Name)
			os.Exit(incorrectArgumentExitCode)
		}
		a := aranet.New(shutdownContext, config.ID, config.Name)
		seenIds[config.ID] = true
		seenNames[config.Name] = true

		shutdownWait.Add(1)
		go func() {
			defer shutdownWait.Done()

			a.RunUpdateLoop(verbose)
		}()

		collectedAranets = append(collectedAranets, a)
		collectedAccessories = append(collectedAccessories, a.Accessory())
	}

	aranetCollector := &aranet.Collector{Aranets: collectedAranets}

	var bridge *aranet.Bridge
	if homekit {
		hapLog.Debug.Enable()
		// TODO
		bridge = aranet.NewBridge(filepath.Join(stateDir, "homekit"), "12344321", "RNDM", collectedAccessories...)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		signal.Stop(sigs)
		fmt.Println()
		log.Printf("Got signal %v, shutting down...\n", strings.ToUpper(sig.String()))
		shutdown()
	}()

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)
	reg.MustRegister(aranetCollector)

	serveMetricsHTTP(shutdownContext, &shutdownWait, reg)
	if bridge != nil {
		serveHomekitServer(shutdownContext, &shutdownWait, bridge)
	}

	waitForShutdown(shutdownContext, &shutdownWait)
}

func waitForShutdown(shutdownContext context.Context, shutdownWait *sync.WaitGroup) {
	<-shutdownContext.Done()

	shutdownWaitCtx, shutdownComplete := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	go func() {
		shutdownWait.Wait()
		shutdownComplete()
	}()

	<-shutdownWaitCtx.Done()
	err := shutdownWaitCtx.Err()
	if err == context.DeadlineExceeded {
		log.Printf("Graceful shutdown did not finish within timeout")
	}
}
