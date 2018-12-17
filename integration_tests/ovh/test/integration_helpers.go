package main

import (
	"errors"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/exec"
	"strings"
)

type HostInfo struct {
	ID         string
	Name       string
	CPU        int
	RAM        int
	Disk       int
	PUBLIC_IP  string
	PRIVATE_IP string
	State      int
	PrivateKey string
}

func isBrokerdLaunched() (bool, error) {
	cmd := "ps -ef | grep brokerd | grep -v grep"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(out), "brokerd"), nil
}

func canBeRun(command string) (bool, error) {
	cmd := "which " + command
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(out), command), nil
}

func getOutput(command string) (string, error) {
	logrus.Warningf("Running command: [%s]", command)
	out, err := exec.Command("bash", "-c", command).CombinedOutput()
	if err != nil {
		logrus.Warning(string(out))
		return string(out), err
	}

	if strings.Contains(strings.ToUpper(string(out)), strings.ToUpper("Error")) {
		logrus.Warning(string(out))
		return string(out), errors.New(string(out))
	}

	logrus.Warning(string(out))
	return string(out), nil
}

func runOnlyInIntegrationTest(key string) {
	if tenant_override := os.Getenv(key); tenant_override == "" {
		panic("This only runs as an integration test")
	}
}

func tearDown() {
	log.Printf("Starting cleanup...")
	_, _ = getOutput("broker volume detach volumetest easyvm")
	_, _ = getOutput("broker volume delete volumetest")
	_, _ = getOutput("broker host delete easyvm")
	_, _ = getOutput("broker host delete complexvm")
	_, _ = getOutput("broker nas delete bnastest")
	_, _ = getOutput("broker host delete easyvm")
	_, _ = getOutput("broker host delete complexvm")
	_, _ = getOutput("broker network delete crazy")
	log.Printf("Finishing cleanup...")
}