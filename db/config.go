package agent

import "encore.dev/config"

type Config struct {
	TemporalServer    string
	TemporalNamespace string
}

var cfg = config.Load[*Config]()
