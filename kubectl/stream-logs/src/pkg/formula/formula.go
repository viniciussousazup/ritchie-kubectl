// This is the formula implementation class.
// Where you will code your methods and manipulate the inputs to perform the specific operation you wish to automate.

package formula

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gookit/color"
	"io"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"sync"
	"time"
)

type Formula struct {
	Text     string
	List     string
	Boolean  bool
	Password string
}

func (f Formula) Run(writer io.Writer) {

	//inputs
	labelSelector := "app.kubernetes.io/instance=dennis-gateway"
	previous := false
	// if 0 see all logs
	sinceTimeSeconds := 10
	parseJson := true

	ctx := context.TODO()

	configFilePath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	fmt.Printf("configFilePath:%s\n", configFilePath)
	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		panic(err.Error())
	}

	// use the current context in kubeconfig
	clientConfig, err := clientcmd.NewClientConfigFromBytes(configFile)
	if err != nil {
		panic(err.Error())
	}

	config, err := clientConfig.ClientConfig()
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	pods, err := clientset.CoreV1().Pods("").List(ctx, listOptions)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Pods founded:%d\n", len(pods.Items))

	wg := sync.WaitGroup{}

	for i, pod := range pods.Items {
		wg.Add(1)
		currentPod := pod
		currentColor := color.FgRed
		if i%2 == 0 {
			currentColor = color.FgBlue
		}
		go func() {
			var sinceTime metav1.Time
			if sinceTimeSeconds > -1 {
				sinceTime = metav1.NewTime(time.Now().Add(-1 * time.Duration(sinceTimeSeconds)))
			}
			opts := v1.PodLogOptions{
				Follow:    true,
				Previous:  previous,
				SinceTime: &sinceTime,
			}
			req := clientset.CoreV1().Pods(currentPod.Namespace).GetLogs(currentPod.Name, &opts)
			podLogs, err := req.Stream(ctx)
			if err != nil {
				panic(err.Error())
			}

			scanner := bufio.NewScanner(podLogs)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				m := scanner.Text()
				if parseJson {
					err := f.printJson(m, currentColor, currentPod)
					if err != nil {
						f.printLine(m, currentColor, currentPod)
					}
				} else {
					f.printLine(m, currentColor, currentPod)
				}

			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func (f Formula) printLine(m string, currentColor color.Color, currentPod v1.Pod) {
	msgWithColor := currentColor.Render(m)
	fmt.Printf("[%s]: %s\n", currentPod.Name, msgWithColor)
}

func (f Formula) printJson(m string, currentColor color.Color, currentPod v1.Pod) error {
	jsonLine := make(map[string]interface{})
	err := json.Unmarshal([]byte(m), &jsonLine)
	if err != nil {
		return err
	}

	podName := currentColor.Render(currentPod.Name)
	colorOfMsg := getLevelColor(fmt.Sprintf("%s", jsonLine["level"]))
	logTime := "-"
	if value, ok := jsonLine["time"].(int64); ok {
		logTime = fmt.Sprintf("%d", value)
	}
	msg := fmt.Sprintf("[%s][%s]", logTime, jsonLine["message"])
	fmt.Printf("(%s) - %s\n", podName, colorOfMsg.Render(msg))
	return nil
}

func getLevelColor(level string) color.Color {
	switch level {
	case "info":
		return color.FgBlue
	case "error":
		return color.FgRed
	case "panic":
		return color.FgRed
	case "fatal":
		return color.FgRed
	case "warn":
		return color.FgYellow
	default:
		return color.FgBlack
	}
}
