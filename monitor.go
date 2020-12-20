package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/alessio/shellescape"
	"github.com/macrat/landns/lib-landns/logger"
)

type HealthMonitor struct {
	sync.RWMutex

	Command  []string
	Env      map[string]string
	Interval time.Duration
	Logger   logger.Logger
	healthy  bool
}

func (h *HealthMonitor) setHealthy(ok bool) {
	h.Lock()
	defer h.Unlock()

	h.healthy = ok
}

func (h *HealthMonitor) makeCommand(stdout io.Writer, stderr io.Writer) *exec.Cmd {
	cmd := exec.Command(h.Command[0], h.Command[1:]...)

	for k, v := range h.Env {
		s := fmt.Sprintf("%s=%s", k, v)
		cmd.Env = append(cmd.Env, s)
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd
}

func (h *HealthMonitor) fields(cmd *exec.Cmd) logger.Fields {
	env := []string{}
	for k, v := range h.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, shellescape.Quote(v)))
	}

	return logger.Fields{
		"command": cmd.Path,
		"args":    shellescape.QuoteCommand(cmd.Args),
		"env":     strings.Join(env, " "),
	}
}

func (h *HealthMonitor) test() {
	stdout := bytes.NewBuffer([]byte{})
	stderr := bytes.NewBuffer([]byte{})
	cmd := h.makeCommand(stdout, stderr)

	f := h.fields(cmd)
	h.Logger.Debug("start health check", f)

	err := cmd.Run()
	h.setHealthy(err == nil)

	f["stdout"] = string(stdout.Bytes())
	f["stderr"] = string(stderr.Bytes())

	if err != nil {
		f["status"] = "unhealthy"
		h.Logger.Warn("service is down", f)
	} else {
		f["status"] = "healthy"
		h.Logger.Info("service is up", f)
	}
}

func (h *HealthMonitor) Run(ctx context.Context) {
	tick := time.Tick(h.Interval)

	for {
		h.test()

		select {
		case <-ctx.Done():
			break
		case <-tick:
		}
	}
}

func (h *HealthMonitor) IsHealthy() bool {
	h.RLock()
	defer h.RUnlock()

	return h.healthy
}
