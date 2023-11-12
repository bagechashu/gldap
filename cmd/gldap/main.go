package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/arl/statsviz"
	"github.com/bagechashu/gldap/internal/monitoring"
	"github.com/bagechashu/gldap/internal/toml"
	"github.com/bagechashu/gldap/internal/version"
	"github.com/bagechashu/gldap/pkg/config"
	"github.com/bagechashu/gldap/pkg/frontend"
	"github.com/bagechashu/gldap/pkg/logging"
	"github.com/bagechashu/gldap/pkg/server"
	"github.com/bagechashu/gldap/pkg/stats"
	docopt "github.com/docopt/docopt-go"
	"github.com/fsnotify/fsnotify"
	"github.com/jinzhu/copier"
)

// const programName = "gldap"

var usage = `glauth: securely expose your LDAP for external auth

Usage:
  glauth [options] -c <file|s3 url>
  glauth -h --help
  glauth --version

Options:
  -c, --config <file>       Config file.
  -K <aws_key_id>           AWS Key ID.
  -S <aws_secret_key>       AWS Secret Key.
  -r <aws_region>           AWS Region [default: us-east-1].
  --aws_endpoint_url <url>  Custom S3 endpoint.
  --ldap <address>          Listen address for the LDAP server.
  --ldaps <address>         Listen address for the LDAPS server.
  --ldaps-cert <cert-file>  Path to cert file for the LDAPS server.
  --ldaps-key <key-file>    Path to key file for the LDAPS server.
  --check-config            Check configuration file and exit.
  -h, --help                Show this screen.
  --version                 Show version.
`

var (
	args map[string]interface{}

	activeConfig = &config.Config{}
)

func main() {

	if err := parseArgs(); err != nil {
		fmt.Println("Could not parse command-line arguments")
		fmt.Println(err)
		os.Exit(1)
	}
	checkConfig := false
	if cc, ok := args["--check-config"]; ok {
		if cc == true {
			checkConfig = true
		}
	}

	cfg, err := toml.NewConfig(checkConfig, getConfigLocation(), args)

	if err != nil {
		fmt.Println("Configuration file error")
		fmt.Println(err)
		os.Exit(1)
	}

	if checkConfig {
		fmt.Println("Config file seems ok (but I am not checking much at this time)")
		return
	} else {
		logging.InitSlogDefault(cfg.Debug)
		slog.Debug("slog Debug enabled")
		// slog.Info("slog Info")
	}

	if err := copier.Copy(activeConfig, cfg); err != nil {
		slog.Error("Could not save reloaded config. Holding on to old config", err)
	}

	slog.Info("AP start")

	startService()
}

func startService() {
	// stats
	stats.General.Set("version", stats.Stringer(version.Version))

	// web API
	if activeConfig.API.Enabled {
		slog.Info("Web API enabled")

		if activeConfig.API.Internals {
			statsviz.Register(
				http.DefaultServeMux,
				statsviz.Root("/internals"),
				statsviz.SendFrequency(1000*time.Millisecond),
			)
		}

		go frontend.RunAPI(
			frontend.Config(&activeConfig.API),
		)
	}

	monitor := monitoring.NewMonitor()

	startConfigWatcher()

	s, err := server.NewServer(
		server.Config(activeConfig),
		server.Monitor(monitor),
	)

	if err != nil {
		slog.Error("could not create server", err)
		os.Exit(1)
	}

	if activeConfig.LDAP.Enabled {
		go func() {
			if err := s.ListenAndServe(); err != nil {
				slog.Error("could not start LDAP server", err)
				os.Exit(1)
			}
		}()
	}

	if activeConfig.LDAPS.Enabled {
		go func() {
			if err := s.ListenAndServeTLS(); err != nil {
				slog.Error("could not start LDAPS server", err)
				os.Exit(1)
			}
		}()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until we receive our signal.
	<-c

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	s.Shutdown()

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	slog.Info("AP exit")
	os.Exit(0)
}

func startConfigWatcher() {
	configFileLocation := getConfigLocation()
	if !activeConfig.WatchConfig || strings.HasPrefix(configFileLocation, "s3://") {
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("could not start config-watcher", err)
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		isChanged, isRemoved := false, false
		for {
			select {
			case event := <-watcher.Events:
				slog.Info("watcher got event", "e", event.Op.String())
				if event.Op&fsnotify.Write == fsnotify.Write {
					isChanged = true
				} else if event.Op&fsnotify.Remove == fsnotify.Remove { // vim edit file with rename/remove
					isChanged, isRemoved = true, true
				}
			case err := <-watcher.Errors:
				slog.Error("watcher error", err)
			case <-ticker.C:
				// wakeup, try finding removed config
			}
			if _, err := os.Stat(configFileLocation); !os.IsNotExist(err) && (isRemoved || isChanged) {
				if isRemoved {
					slog.Info("rewatching config", "file", configFileLocation)
					watcher.Add(configFileLocation) // overwrite
					isChanged, isRemoved = true, false
				}
				if isChanged {

					cfg, err := toml.NewConfig(false, configFileLocation, args)

					if err != nil {
						slog.Error("Could not reload config. Holding on to old config", err)
					} else {
						slog.Info("Config was reloaded")

						if err := copier.Copy(activeConfig, cfg); err != nil {
							slog.Error("Could not save reloaded config. Holding on to old config", err)
						}
					}
					isChanged = false
				}
			}
		}
	}()

	watcher.Add(configFileLocation)
}

func parseArgs() error {
	var err error

	if args, err = docopt.Parse(usage, nil, true, version.GetVersion(), false); err != nil {
		return err
	}

	return nil
}

func getConfigLocation() string {
	return args["--config"].(string)
}
