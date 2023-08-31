package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/ppc64le-cloud/kube-rtas/pkg/dbreader"
	"github.com/ppc64le-cloud/kube-rtas/pkg/knode"
	"github.com/ppc64le-cloud/kube-rtas/pkg/utils"
)

var (
	wg          sync.WaitGroup
	cfgFilePath string
)

func main() {
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	klogFlags.Set("logtostderr", "true")

	pflag.CommandLine.AddGoFlagSet(klogFlags)
	pflag.StringVarP(&cfgFilePath, "config", "c", "config.json", "Path to KubeRTAS config file")
	pflag.Parse()

	klog.Infof("setting up KubeRTAS, reading config from %s", cfgFilePath)
	serviceConfig, err := utils.ReadConfig(cfgFilePath)
	if err != nil {
		klog.Fatalf("error while reading config. Error: %v", err)
	}

	// Check if the servicelog.db file is present in the path set in config file.
	if _, err := os.Stat(serviceConfig.ServicelogPath); err != nil {
		klog.Fatalf("Error finding the servicelog.db file. %v", err)
	}

	// Create an instance of NewReader to retieve entries from the servicelog.db file.
	svLogReader := dbreader.NewReader(serviceConfig.ServicelogPath, serviceConfig.Severity)

	// Set up notifier to post events to the Kube API server.
	notifier := knode.NewNotifier()
	err = notifier.InitializeNotifier()
	if err != nil {
		klog.Fatalf("cannot initialize the k8s apiserver notifier. %v", err)
	}

	// Set up a ticker to periodically check the servicelog.db file for new entries.
	ticker := time.NewTicker(serviceConfig.PollInterval * time.Second)
	done := make(chan struct{}, 1)
	initRun := make(chan struct{}, 1)
	wg.Add(2)

	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				klog.V(1).Info("Ticker stopped.")
				wg.Done()
			case <-ticker.C:
			case <-initRun:
			}
			if err = svLogReader.ParseServiceLogDB(notifier); err != nil {
				klog.Fatalf("error while processing servicelogs. %v", err)
				return
			}
		}
	}()

	// Capture any OS signals to perform graceful exit.
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigchan
		done <- struct{}{}
		klog.Infof("Captured %s, Shutting down KubeRTAS", sig.String())
		wg.Done()
		os.Exit(0)
	}()
	// Trigger to collect the servicelogs soon after the service has started.
	initRun <- struct{}{}
	wg.Wait()
}
