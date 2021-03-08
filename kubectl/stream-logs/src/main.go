// This is the main class.
// Where you will extract the inputs asked on the config.json file and call the formula's method(s).

package main

import (
	"formula/pkg/formula"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func main() {

	formula.Formula{
		LabelSelector: "app.kubernetes.io/instance in (dennis-runner, dennis-builder-wrapper, dennis-gateway)",
		SinceTime:     metav1.NewTime(time.Now().Add(-1 * time.Duration(0))),
		LogLevel: map[string]bool{
			"info":  true,
			"error": true,
			"panic": true,
			"fatal": true,
			"warn":  false,
			"debug": false,
		},
	}.Run()
}
