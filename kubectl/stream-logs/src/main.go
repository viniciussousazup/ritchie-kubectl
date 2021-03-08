// This is the main class.
// Where you will extract the inputs asked on the config.json file and call the formula's method(s).

package main

import (
	"formula/pkg/formula"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
	"time"
)

func main() {

	podLogLevel := os.Getenv("POD_LOG_LEVEL")

	logLevels := strings.Split(podLogLevel, "|")

	logLevel := map[string]bool{
		"info":  false,
		"error": false,
		"panic": false,
		"fatal": false,
		"warn":  false,
		"debug": false,
	}

	for _, log := range logLevels {
		logLevel[log] = true
	}

	formula.Formula{
		LabelSelector: "app.kubernetes.io/instance in (dennis-runner, dennis-builder-wrapper, dennis-gateway)",
		SinceTime:     metav1.NewTime(time.Now().Add(-1 * time.Duration(0))),
		LogLevel:      logLevel,
	}.Run()
}
