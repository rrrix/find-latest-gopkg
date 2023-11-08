package instance

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

const StandardGOPROXY = "https://proxy.golang.org,direct"

type CLIOptions struct {
	GoProxy      string // Go Module Mirror Endpoint (e.g. https://proxy.golang.org)
	LogLevelName string // Set the log level by name
	LogDebug     bool   // Force debug logging, overrides --log-level
	LogVerbose   bool   // Force Verbose logging, overrides --log-level
	Name         bool   // Print Module Names
	Version      bool   // Print Module Versions
	Time         bool   // Print Module Fetched/Updated Timesteamp
	Repo         bool   // Print Module Origin Repository URL
	Ref          bool   // Print Module Git Ref (branch, tag, etc)
	Hash         bool   // Print Module Hash
	Dump         bool   // Dump JSON response
}

type MainContext struct {
	Options        *CLIOptions
	ProxyEndpoints []string // Go Module Mirror Endpoint (e.g. https://proxy.golang.org)
	logger         *log.Logger
}

func (m *MainContext) BuildLogger() {
	var logLevel log.Level

	switch {
	case m.Options.LogDebug:
		logLevel = log.DebugLevel
	case m.Options.LogVerbose:
		logLevel = log.InfoLevel
	case m.Options.LogLevelName != "":
		logLevel = log.ParseLevel(m.Options.LogLevelName)
	default:
		envLevel := os.Getenv("LOG_LEVEL")
		if envLevel != "" {
			logLevel = log.ParseLevel(envLevel)
		}
	}
	m.logger = log.NewWithOptions(os.Stderr, log.Options{
		TimeFormat:      time.Kitchen,
		Level:           logLevel,
		ReportTimestamp: true,
		ReportCaller:    true,
		CallerFormatter: log.ShortCallerFormatter,
		CallerOffset:    0,
		Formatter:       log.TextFormatter,
	})
	switch {
	case m.Options.LogLevelName != "" && m.Options.LogDebug:
		log.Debugf("--debug enabled, ignoring --log-level=%s", m.Options.LogLevelName)
	case m.Options.LogLevelName != "" && m.Options.LogVerbose:
		log.Debugf("--verbose enabled, ignoring --log-level=%s", m.Options.LogLevelName)
	case m.Options.LogLevelName != "":
		log.Debugf("--log-level=%s", m.Options.LogLevelName)
	}
	log.Debugf("Configured Logger: %+v", m.logger)
	log.Infof("Main: %+v", m)
}

func (m *MainContext) SetProxyEndpoints() {
	osEnvGoProxy := strings.TrimSpace(os.Getenv("GOPROXY"))
	var goproxy string

	switch {
	case m.Options.GoProxy != "":
		goproxy = m.Options.GoProxy
		log.Debugf("Using --GoProxy=%s", goproxy)
	case osEnvGoProxy != "":
		goproxy = strings.TrimSpace(osEnvGoProxy)
		log.Debugf("Using os env GOPROXY: %s", goproxy)
	default:
		if envp := execGoEnvGOPROXY(); envp != nil && *envp != "" {
			goproxy = strings.TrimSpace(*envp)
			log.Infof("Using go env GOPROXY: %s", goproxy)
		} else {
			goproxy = StandardGOPROXY
			log.Infof("Using standard GOPROXY: %s", goproxy)
		}
	}

	for _, p := range strings.Split(goproxy, ",") {
		p := strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if p == "direct" {
			continue
		}
		if p == "off" {
			log.Fatalf("GOPROXY set to \"off\", aborting")
		}
		m.ProxyEndpoints = append(m.ProxyEndpoints, p)
	}
	log.Infof("Using GOPROXY: %v", m.ProxyEndpoints)
}

func execGoEnvGOPROXY() *string {
	cmd := exec.Command("go", "env", "GOPROXY")
	output, err := cmd.Output()
	if err != nil {
		log.Warnf("Failed to get GOPROXY from go env: %v", err)
		return nil
	}
	goproxy := string(output)
	return &goproxy
}
