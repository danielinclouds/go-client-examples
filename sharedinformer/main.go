// Based on https://gianarb.it/blog/kubernetes-shared-informer
//
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	log.Print("Shared Informer app started")
	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	factory := informers.NewSharedInformerFactory(clientset, time.Second*5)
	informer := factory.Core().V1().Namespaces().Informer()
	stopper := make(chan struct{})
	defer close(stopper)
	defer runtime.HandleCrash()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		DeleteFunc: onDelete,
		UpdateFunc: onUpdate,
	})

	go informer.Run(stopper)

	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	<-stopper
}

func onAdd(obj interface{}) {
	namespace := obj.(*corev1.Namespace)
	log.Printf("Adding: %s", namespace.Name)
}

func onDelete(obj interface{}) {
	namespace := obj.(*corev1.Namespace)
	log.Printf("Deleting: %s", namespace.Name)
}

func onUpdate(oldObj, newObj interface{}) {
	oldNamespace := oldObj.(*corev1.Namespace)
	newNamespace := oldObj.(*corev1.Namespace)
	log.Printf("Updating: %s", oldNamespace.Name)
	log.Printf("Updating: %s", newNamespace.Name)
}
