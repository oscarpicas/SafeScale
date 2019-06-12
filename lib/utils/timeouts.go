package utils

import (
	"os"
	"time"
)

const (
	// DefaultContextTimeout default timeout for grpc command invocation
	DefaultContextTimeout = 1 * time.Minute
	// HostTimeout timeout for grpc command relative to host creation
	HostTimeout = 5 * time.Minute

	// Long timeout
	LongHostOperationTimeout    = 10 * time.Minute

	DefaultSSHConnectionTimeout = 2 * time.Minute
	HostCleanupTimeout = 3 * time.Minute

	DefaultConnectionTimeout = 30 * time.Second
	DefaultExecutionTimeout  = 5 * time.Minute


	SmallDelay = 1*time.Second
	DefaultDelay = 5*time.Second
	BigDelay = 30*time.Second
)

func GetVariableTimeout(key string, duration time.Duration) time.Duration {
	defaultTimeout := duration

	if defaultTimeoutCandidate := os.Getenv(key); defaultTimeoutCandidate != "" {
		newTimeout, err := time.ParseDuration(defaultTimeoutCandidate)
		if err != nil {
			return defaultTimeout
		}
		return newTimeout
	}

	return defaultTimeout
}

func GetMinDelay() time.Duration {
	return GetVariableTimeout("SAFESCALE_MIN_DELAY", SmallDelay)
}

func GetDefaultDelay() time.Duration {
	return GetVariableTimeout("SAFESCALE_DEFAULT_DELAY", DefaultDelay)
}

func GetBigDelay() time.Duration {
	return GetVariableTimeout("SAFESCALE_BIG_DELAY", BigDelay)
}

// GetContextTimeout ...
func GetContextTimeout() time.Duration {
	return GetVariableTimeout("SAFESCALE_CONTEXT_TIMEOUT", DefaultContextTimeout)
}

// GetHostTimeout ...
func GetHostTimeout() time.Duration {
	return GetVariableTimeout("SAFESCALE_HOST_TIMEOUT", HostTimeout)
}

func GetHostCleanupTimeout() time.Duration {
	return GetVariableTimeout("SAFESCALE_HOST_CLEANUP_TIMEOUT", HostCleanupTimeout)
}

func GetConnectSSHTimeout() time.Duration {
	return GetVariableTimeout("SAFESCALE_SSH_CONNECT_TIMEOUT", DefaultSSHConnectionTimeout)
}

// GetHostTimeout ...
func GetLongOperationTimeout() time.Duration {
	return GetVariableTimeout("SAFESCALE_HOST_LONG_OPERATION_TIMEOUT", LongHostOperationTimeout)
}