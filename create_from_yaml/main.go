package main

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var yaml = `
apiVersion: v1
kind: Pod
metadata:
  labels:
    run: server
  name: server
spec:
  containers:
  - image: server
    name: nginx
    ports:
    - containerPort: 80
    resources: {}
  dnsPolicy: ClusterFirst
  restartPolicy: Never
`

func main() {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode([]byte(yaml), nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}

	pod := obj.(*v1.Pod)

	fmt.Printf("%#v\n", pod)
}
