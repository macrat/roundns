package main

import (
	"github.com/macrat/landns/lib-landns"
)

type Host struct {
	Record landns.Record
	Health *HealthMonitor
}
