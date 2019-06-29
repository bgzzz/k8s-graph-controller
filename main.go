package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func main() {
	k, err := GetKubectl()
	if err != nil {
		panic(err)
	}

	s := NewStorage()
	gr := NewClusterGraph()

	gr.AddVerge(InternetVerge, nil, InternetVerge)

	podWatchlist := cache.NewListWatchFromClient(k.CoreV1().RESTClient(), "pods", v1.NamespaceAll,
		fields.Everything())

	_, podController := cache.NewInformer(
		podWatchlist,
		&v1.Pod{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				go func(o interface{}) {
					// fmt.Printf("Info: pod added: %s \n", o)
					pod := o.(*v1.Pod)
					s.SetPod(pod.GetName(), pod)
				}(obj)
			},
			DeleteFunc: func(obj interface{}) {
				go func(o interface{}) {
					// fmt.Printf("Info: pod deleted: %s \n", o)
					pod := o.(*v1.Pod)
					s.SetPod(pod.GetName(), nil)
				}(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				go func(o interface{}, oOld interface{}) {
					// fmt.Printf("Info: pod added: %s \n", o)
					pod := o.(*v1.Pod)
					podOld := oOld.(*v1.Pod)
					s.SetPod(podOld.GetName(), nil)
					s.SetPod(pod.GetName(), pod)
				}(newObj, oldObj)
			},
		},
	)
	podControllerStop := make(chan struct{})
	go podController.Run(podControllerStop)

	//--------------------------------------------------------------------------

	svcWatchlist := cache.NewListWatchFromClient(k.CoreV1().RESTClient(), "services", v1.NamespaceAll,
		fields.Everything())

	_, svcController := cache.NewInformer(
		svcWatchlist,
		&v1.Service{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				go func(o interface{}) {
					// fmt.Printf("Info: svc added: %s \n", o)
					pod := o.(*v1.Service)
					s.SetService(pod.GetName(), pod)
				}(obj)
			},
			DeleteFunc: func(obj interface{}) {
				go func(o interface{}) {
					// fmt.Printf("Info: svc deleted: %s \n", o)
					pod := o.(*v1.Service)
					s.SetService(pod.GetName(), nil)
				}(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				go func(o interface{}, oOld interface{}) {
					// fmt.Printf("Info: svc added: %s \n", o)
					pod := o.(*v1.Service)
					podOld := oOld.(*v1.Service)
					s.SetPod(podOld.GetName(), nil)
					s.SetPod(pod.GetName(), pod)
				}(newObj, oldObj)
			},
		},
	)
	svcControllerStop := make(chan struct{})
	go svcController.Run(svcControllerStop)

	//--------------------------------------------------------------------------

	ingressWatchlist := cache.NewListWatchFromClient(k.ExtensionsV1beta1().RESTClient(), "ingresses", v1.NamespaceAll,
		fields.Everything())

	_, ingressController := cache.NewInformer(
		ingressWatchlist,
		&v1beta1.Ingress{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				go func(o interface{}) {
					// fmt.Printf("Info: ingress added: %s \n", o)
					pod := o.(*v1beta1.Ingress)
					s.SetIngress(pod.GetName(), pod)
				}(obj)
			},
			DeleteFunc: func(obj interface{}) {
				go func(o interface{}) {
					// fmt.Printf("Info: ingress deleted: %s \n", o)
					pod := o.(*v1beta1.Ingress)
					s.SetIngress(pod.GetName(), nil)
				}(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				go func(o interface{}, oOld interface{}) {
					// fmt.Printf("Info: ingress added: %s \n", o)
					pod := o.(*v1beta1.Ingress)
					podOld := oOld.(*v1beta1.Ingress)
					s.SetIngress(podOld.GetName(), nil)
					s.SetIngress(pod.GetName(), pod)
				}(newObj, oldObj)
			},
		},
	)
	ingressControllerStop := make(chan struct{})
	go ingressController.Run(ingressControllerStop)

	// http.HandleFunc("/graph", func(w http.ResponseWriter, r *http.Request) {
	//
	// 	graph := gr.GetGraph(s)
	// 	fmt.Printf("Info: Graph Served: %+v\n", graph)
	// 	json.NewEncoder(w).Encode(graph)
	// })

	fmt.Println("Listening :8080")

	rtr := mux.NewRouter()
	rtr.HandleFunc("/graph/{hash:[a-z0-9]+}", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)

		hash, ok := params["hash"]
		if !ok {
			fmt.Println("Warn: there should be a hash provided")
			return
		}
		graph := &Graph{Hash: hash}

		// fmt.Println("hash")
		fmt.Println("hash: " + hash)

		if hash == "" || s.GetHash() != hash {

			graph = gr.GetGraph(s)
		}

		fmt.Printf("Info: Graph Served: %+v\n", graph)
		json.NewEncoder(w).Encode(graph)

	}).Methods("GET")
	rtr.HandleFunc(`/meta/{name:[a-z0-9\-]+}`, func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		name, ok := params["name"]
		if !ok {
			fmt.Println("Warn: there should be a name provided")
			return
		}

		fmt.Println("name_param : " + name)

		answer := &Verge{}
		if name == InternetVerge {
			answer.Name = InternetVerge
			answer.VergeType = InternetVerge
			json.NewEncoder(w).Encode(answer)
			return
		}

		str := strings.Split(name, "-")
		if len(str) < 2 {
			fmt.Println("Warn: there should be a different format ")
			return
		}
		reqType := str[0]
		objName := strings.Join(str[1:], "-")
		var objMeta interface{}
		if reqType == PodType {
			objMeta, _ = s.GetPod(objName)
		} else if reqType == ServiceType {
			objMeta, _ = s.GetService(objName)
		} else if reqType == IngressType {
			objMeta, _ = s.GetService(objName)
		} else {
			fmt.Println("Warn: there should be a different format ")
			return
		}

		answer.Name = name
		answer.VergeType = reqType
		answer.Meta = objMeta
		json.NewEncoder(w).Encode(answer)

	}).Methods("GET")

	http.Handle("/", rtr)

	log.Println("Listening...")
	http.ListenAndServe(":8080", nil)

}
