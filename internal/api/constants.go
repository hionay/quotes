package api

import "time"

const (
	apiReadTimeout     = 10 * time.Second
	apiWriteTimeout    = 30 * time.Second
	apiIdleTimeout     = 120 * time.Second
	apiShutdownTimeout = 30 * time.Second
)

const defaultLimit = 10
