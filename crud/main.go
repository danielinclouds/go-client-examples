package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

var clientset *kubernetes.Clientset

func main() {

	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// getPodByLabel("owner=developer", "default")
	// printPodPhase("server1", "default")
	// createPodFromJSONFile("../resources/server.json")
	// createPodFromYAMLFile("../resources/server.yaml")
	// printPodList("default")
	// printPodLogs("echo", "default")
	// streamPodLogs("echo", "default")
	// updateDeploymentImage("nginx", "default", "nginx:1.10.0")
	// watchPodPhase("run=server", "default")
	// updatePodAnnotations("server", "default")
	portforwardToPodFor10Seconds("server", "default", []string{"8080:80"}, config)

}

func getPodByLabel(label string, namespace string) {
	podList, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: label})
	if err != nil {
		panic(err.Error())
	}

	for _, item := range podList.Items {
		fmt.Println(item.Name)
	}
}

func printPodPhase(name string, namespace string) {
	pod, err := clientset.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(pod.Status.Phase)
}

func createPodFromJSONFile(filepath string) {
	var server v1.Pod

	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err.Error())
	}

	json.Unmarshal(b, &server)
	clientset.CoreV1().Pods("default").Create(&server)
}

func createPodFromYAMLFile(filepath string) {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	yaml, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err.Error())
	}

	obj, _, err := decode([]byte(yaml), nil, nil)
	if err != nil {
		panic(err.Error())
	}

	spec := obj.(*v1.Pod)

	clientset.CoreV1().Pods("default").Create(spec)
}

func printPodList(namespace string) {
	podList, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, item := range podList.Items {
		fmt.Println(item.Name)
	}
}

func printPodLogs(name string, namespace string) {
	req := clientset.CoreV1().Pods(namespace).GetLogs(name, &v1.PodLogOptions{})

	b, err := req.Do().Raw()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(string(b))
}

func streamPodLogs(name string, namespace string) {
	ls, err := clientset.CoreV1().Pods(namespace).GetLogs(name, &v1.PodLogOptions{}).Stream()
	if err != nil {
		panic(err)
	}

	defer ls.Close()

	go func() {
		sc := bufio.NewScanner(ls)

		for sc.Scan() {
			fmt.Println(sc.Text())
		}
	}()

	time.Sleep(time.Second * 10)
}

func updateDeploymentImage(name string, namespace string, image string) {
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	deployment.Spec.Template.Spec.Containers[0].Image = image

	deployment, err = clientset.AppsV1().Deployments(namespace).Update(deployment)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(deployment.Spec.Template.Spec.Containers[0].Image)

}

func watchPodPhase(label string, namespace string) {

	watch, err := clientset.CoreV1().Pods("default").Watch(metav1.ListOptions{LabelSelector: "run=server"})
	if err != nil {
		return
	}

	go func() {
		for event := range watch.ResultChan() {
			p, ok := event.Object.(*v1.Pod)
			if !ok {
				log.Fatal("Not a Pod Object")
				return
			}

			fmt.Println(p.Status.Phase)
		}
	}()
	time.Sleep(time.Second * 10)
}

func updatePodAnnotations(name string, namespace string) {
	pod, err := clientset.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	pod.Annotations = make(map[string]string)
	pod.Annotations["sample"] = "true"

	_, err = clientset.CoreV1().Pods(namespace).Update(pod)
	if err != nil {
		panic(err)
	}
}

func portforwardToPodFor10Seconds(podName string, namespace string, ports []string, config *rest.Config) {

	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		panic(err)
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, podName)
	hostIP := strings.TrimLeft(config.Host, "htps:/")
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	forwarder, err := portforward.New(dialer, ports, stopChan, readyChan, out, errOut)
	if err != nil {
		panic(err)
	}

	go func() {
		for range readyChan { // Kubernetes will close this channel when it has something to tell us.
		}
		if len(errOut.String()) != 0 {
			panic(errOut.String())
		} else if len(out.String()) != 0 {
			fmt.Println(out.String())
		}
	}()

	go func() {
		time.Sleep(time.Second * 10)
		close(stopChan)
	}()

	if err = forwarder.ForwardPorts(); err != nil { // Locks until stopChan is closed.
		panic(err)
	}

}
