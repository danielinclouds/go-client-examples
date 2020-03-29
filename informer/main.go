package main

import (
	"log"
	"os"

	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	log.Print("Informer app started")
	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	watchList := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "namespaces", v1.NamespaceAll, fields.Everything())
	_, controller := cache.NewInformer(
		watchList,
		&corev1.Namespace{},
		time.Second*10,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    onAdd,
			DeleteFunc: onDelete,
			UpdateFunc: onUpdate,
		},
	)

	stopper := make(chan struct{})
	defer close(stopper)
	defer runtime.HandleCrash()

	go controller.Run(stopper)
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
