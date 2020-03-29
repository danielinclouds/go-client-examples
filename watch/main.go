package main

import (
	"fmt"
	"log"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	watchPodPhase("run=server", "default", clientset)

}

func watchPodPhase(label string, namespace string, clientset *kubernetes.Clientset) {

	w, err := clientset.CoreV1().Pods("default").Watch(metav1.ListOptions{LabelSelector: "run=server"})
	if err != nil {
		return
	}

	go func() {
		for event := range w.ResultChan() {

			p, ok := event.Object.(*v1.Pod)
			if !ok {
				log.Fatal("Not a Pod Object")
				return
			}

			switch event.Type {
			case watch.Added:
				fmt.Printf("Event: Added %s\n", p.Name)
			case watch.Deleted:
				fmt.Printf("Event: Deleted %s\n", p.Name)
			case watch.Modified:
				fmt.Printf("Event: Modified %s\n", p.Name)
			}
		}
	}()
	time.Sleep(time.Second * 60)
}
