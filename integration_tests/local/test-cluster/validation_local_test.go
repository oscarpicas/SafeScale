package main

import (
	"testing"

	"github.com/CS-SI/SafeScale/integration_tests"
	"github.com/CS-SI/SafeScale/integration_tests/enums/Providers"
)

func Test_ClusterK8S(t *testing.T) {
	integration_tests.ClusterK8S(t, Providers.LOCAL)
}

func Test_ClusterSwarm(t *testing.T) {
	integration_tests.ClusterSwarm(t, Providers.LOCAL)
}