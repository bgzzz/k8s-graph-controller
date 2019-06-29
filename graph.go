package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
)

const (
	InternetVerge = "internet"

	IngressType = "ingress"
	PodType     = "pod"
	ServiceType = "service"

	maxObjOfOneKindAttached = 10
)

type Verge struct {
	Name      string      `json:"name"`
	VergeType string      `json:"type"`
	Meta      interface{} `json:"meta"`
}

type Edge struct {
	VergeA string `json:"verge_a"`
	VergeB string `json:"verge_b"`
}

var possibleSvcTypes = []string{
	"nodePort",
	"clusterIp",
}

type SvcMeta struct {
	SvcType string `json:"type"`
}

type Graph struct {
	Verges []Verge `json:"verges"`
	Edges  []Edge  `json:"edges"`

	Hash string `json:"hash"`
}

type ClusterGraph struct {
	ClusterMtx        sync.RWMutex
	vergesFacilitator map[string]Verge
	edgesFacilitator  map[string]map[string]struct{}
}

func NewClusterGraph() *ClusterGraph {
	return &ClusterGraph{
		vergesFacilitator: map[string]Verge{},
		// [vB]map[vA]vB
		edgesFacilitator: map[string]map[string]struct{}{},
	}
}

func (cg *ClusterGraph) AddEdge(vA, vB string) {
	cg.ClusterMtx.Lock()
	defer cg.ClusterMtx.Unlock()

	_, ok := cg.vergesFacilitator[vA]
	if !ok {
		fmt.Println("Warn: there is no vergeA " + vA)
		return
	}

	_, ok = cg.vergesFacilitator[vB]
	if !ok {
		fmt.Println("Warn: there is no vergeB " + vB)
		return
	}

	v, ok := cg.edgesFacilitator[vA]

	if ok {
		// check vB
		_, vBOK := v[vB]
		if vBOK {
			return
		}

	}

	v, ok = cg.edgesFacilitator[vB]

	if ok {
		// check vB
		_, vBOK := v[vA]
		if vBOK {
			return
		}
	}

	base, ok := cg.edgesFacilitator[vA]
	if ok {
		base[vB] = struct{}{}
		cg.edgesFacilitator[vA] = base
	} else {
		cg.edgesFacilitator[vA] = map[string]struct{}{
			vB: struct{}{},
		}
	}
}

func (cg *ClusterGraph) RemoveEdge(vA, vB string) {
	cg.ClusterMtx.Lock()
	defer cg.ClusterMtx.Unlock()

	v, ok := cg.edgesFacilitator[vA]
	if ok {
		delete(v, vB)
	}

	v, ok = cg.edgesFacilitator[vB]
	if ok {
		delete(v, vA)
	}
}

func (cg *ClusterGraph) GetEdges(verge string) []Edge {
	cg.ClusterMtx.RLock()
	defer cg.ClusterMtx.RUnlock()

	_, ok := cg.edgesFacilitator[verge]
	if !ok {
		return []Edge{}
	}

	edges := []Edge{}
	for k, _ := range cg.edgesFacilitator[verge] {
		edges = append(edges, Edge{
			VergeA: verge,
			VergeB: k,
		})
	}

	return edges
}

func (cg *ClusterGraph) AddVerge(name string, obj interface{}, objType string) {
	cg.ClusterMtx.Lock()
	defer cg.ClusterMtx.Unlock()

	if objType != InternetVerge {
		cg.vergesFacilitator[objType+"-"+name] = Verge{
			Name:      objType + "-" + name,
			VergeType: objType,
			// Meta:      obj,
		}
	} else {
		cg.vergesFacilitator[name] = Verge{
			Name:      name,
			VergeType: objType,
			// Meta:      obj,
		}
	}
}

func (cg *ClusterGraph) DeleteVerge(name string) {
	cg.ClusterMtx.Lock()
	defer cg.ClusterMtx.Unlock()

	delete(cg.vergesFacilitator, name)
	delete(cg.edgesFacilitator, name)
}

func (cg *ClusterGraph) GetVerges() []Verge {
	cg.ClusterMtx.RLock()
	defer cg.ClusterMtx.RUnlock()

	verges := []Verge{}

	for _, v := range cg.vergesFacilitator {
		verges = append(verges, v)
	}

	return verges
}

func (cl *ClusterGraph) GetGraph(s *Storage) *Graph {
	ingresses := s.GetIngresses()
	services := s.GetServices()
	pods := s.GetPods()

	// go through pods
	for i := range pods {
		tmpPod, ok := s.GetPod(pods[i])
		if !ok {
			fmt.Println("Err: there is already no pod " + pods[i])
			continue
		}

		fmt.Println("pod name: " + pods[i])
		pod := tmpPod.(*v1.Pod)
		cl.AddVerge(pod.GetName(), pod, PodType)

	}

	// go through services
	for i := range services {
		tmpSvc, ok := s.GetService(services[i])
		if !ok {
			continue
		}

		svc := tmpSvc.(*v1.Service)
		cl.AddVerge(svc.GetName(), svc, ServiceType)

		podsList := GetPodsBySelector(svc.Spec.Selector, s, pods)

		fmt.Println("AATTENTION")
		fmt.Println(svc.GetName())
		fmt.Println(podsList)

		for i := range podsList {
			fmt.Println(cl.GetVerges())
			fmt.Println(ServiceType + "-" + svc.GetName() + " : " + podsList[i])
			cl.AddEdge(ServiceType+"-"+svc.GetName(), podsList[i])
		}

		if svc.Spec.Type == v1.ServiceTypeNodePort || svc.Spec.Type == v1.ServiceTypeLoadBalancer {
			cl.AddEdge(ServiceType+"-"+svc.GetName(), InternetVerge)
		}
	}

	// go through ingresses

	for i := range ingresses {
		tmp, ok := s.GetIngress(ingresses[i])
		if !ok {
			fmt.Println("Err: there is already no ingress " + ingresses[i])
			continue
		}

		ingress := tmp.(*v1beta1.Ingress)
		cl.AddVerge(ingress.GetName(), ingress, IngressType)

		fmt.Println("ingress-name : " + ingress.GetName())
		cl.AddEdge(IngressType+"-"+ingress.GetName(), InternetVerge)

		fmt.Println(cl.GetEdges(IngressType + "-" + ingress.GetName()))

		for j := range ingress.Spec.Rules {
			for jj := range ingress.Spec.Rules[j].HTTP.Paths {
				svcName := ingress.Spec.Rules[j].HTTP.Paths[jj].Backend.ServiceName
				_, ok := s.GetService(svcName)
				if ok {
					cl.AddEdge(IngressType+"-"+ingress.GetName(), ServiceType+"-"+svcName)
				}
			}
		}
	}

	svcLink, err := GetSvcLinks()
	if err != nil {
		fmt.Println("Err: " + err.Error())
	} else {
		for k, v := range svcLink.InitConfig {
			for i := range v {
				cl.AddEdge("service-"+k, "service-"+v[i])
			}
		}
	}

	verges := cl.GetVerges()
	edges := []Edge{}
	for i := range verges {
		ed := cl.GetEdges(verges[i].Name)
		edges = append(edges, ed...)
	}

	graph := Graph{
		Verges: verges,
		Edges:  edges,
	}

	bytes, err := json.Marshal(graph)
	if err != nil {
		fmt.Printf("Err: %v\n", err.Error())
	}

	hash := md5.Sum([]byte(bytes))
	graph.Hash = hex.EncodeToString(hash[:])

	s.SetHash(graph.Hash)

	return &graph
}

func GetPodsBySelector(selector map[string]string,
	s *Storage, pods []string) []string {

	neededPods := []string{}

	svcL, ok := selector["app"]
	if !ok {
		svcL = "app"
	}

	for i := range pods {
		tmp, ok := s.GetPod(pods[i])
		if !ok {
			fmt.Println("Warn: is not able to get a pod: " + pods[i])
			continue
		}

		pod := tmp.(*v1.Pod)

		labels := pod.GetLabels()

		podL, ok := labels["app"]
		if !ok {
			podL = "podlabel"
		}

		if podL == svcL {
			neededPods = append(neededPods, PodType+"-"+pod.GetName())
		}

		// for l, lv := range labels {
		// 	if l == k && v == lv {
		// 		neededPods = append(neededPods, PodType+"-"+pod.GetName())
		// 	}
		// }
	}

	return neededPods
}

type SvcLink struct {
	InitConfig map[string][]string `json:"name"`
}

func GetSvcLinks() (*SvcLink, error) {
	svL := SvcLink{}

	err := json.Unmarshal([]byte(serviceLinks), &svL)
	if err != nil {
		return nil, err
	}

	return &svL, nil
}
