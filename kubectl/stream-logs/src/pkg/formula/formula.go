// This is the formula implementation class.
// Where you will code your methods and manipulate the inputs to perform the specific operation you wish to automate.

package formula

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gookit/color"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"time"
)

type Formula struct {
	LabelSelector string
	SinceTime     metav1.Time
	LogLevel      map[string]bool
}

func (f Formula) Run() {

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
		LabelSelector: f.LabelSelector,
	}

	podsList := make(map[string]v1.Pod, 0)

	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		pods, err := clientset.CoreV1().Pods("").List(ctx, listOptions)
		if err != nil {
			panic(err)
		}

		for _, pod := range pods.Items {
			if _, exist := podsList[pod.Name]; exist {
				continue
			}
			if pod.Status.Phase == "Running" {
				println("add pod", pod.Name)
				podsList[pod.Name] = pod
				go f.streamLogOfPod(ctx, pod, clientset)
			}
		}

	}

}

func (f Formula) streamLogOfPod(
	ctx context.Context,
	pod v1.Pod,
	clientset *kubernetes.Clientset,
) {
	opts := v1.PodLogOptions{
		Follow:    true,
		SinceTime: &f.SinceTime,
	}
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &opts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		panic(err.Error())
	}

	scanner := bufio.NewScanner(podLogs)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		msg := scanner.Text()
		err := f.printJson(msg, pod)
		if err != nil {
			f.printLine(msg, pod)
		}
	}
}

func (f Formula) printLine(msg string, pod v1.Pod) {
	fmt.Printf("[%s]: %s\n", pod.Name, msg)
}

func (f Formula) printJson(msg string, pod v1.Pod) error {
	jsonLine := make(map[string]interface{})
	err := json.Unmarshal([]byte(msg), &jsonLine)
	if err != nil {
		return err
	}

	level := fmt.Sprintf("%s", jsonLine["level"])
	if printLevel, exist := f.LogLevel[level]; !exist || !printLevel {
		return nil
	}

	podName := pod.Name
	colorOfMsg := getLevelColor(level)
	logTime := "-"
	if value, ok := jsonLine["time"].(int64); ok {
		logTime = fmt.Sprintf("%s", time.Unix(value, 0).Format("2 Jan 2006 15:04:05"))
	}
	jsonMsg := fmt.Sprintf("[%s][%s]", logTime, jsonLine["message"])
	fmt.Printf("(%s) - %s\n", podName, colorOfMsg.Render(jsonMsg))
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
	case "debug":
		return color.FgGray
	default:
		return color.FgWhite
	}
}
