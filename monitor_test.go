package main

import (
	"testing"

	"github.com/macrat/landns/lib-landns/logger/logtest"
)

func TestHealthMonitor_test(t *testing.T) {
	suc := &HealthMonitor{
		Command: []string{"echo", "hello world"},
		Logger:  &logtest.DummyLogger{},
	}
	suc.test()

	if !suc.IsHealthy() {
		t.Errorf("expected success but failure")
	}

	fail := &HealthMonitor{
		Command: []string{"sh", "-c", "exit 1"},
		Logger:  &logtest.DummyLogger{},
	}
	fail.test()

	if fail.IsHealthy() {
		t.Errorf("expected failure but success")
	}

	env := &HealthMonitor{
		Command: []string{"sh", "-c", "test 1 -eq $ONE"},
		Env:     map[string]string{"ONE": "1"},
		Logger:  &logtest.DummyLogger{},
	}
	env.test()

	if !env.IsHealthy() {
		t.Errorf("expected failure but success")
	}
}
