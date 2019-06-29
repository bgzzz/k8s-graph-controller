package main

import (
	"fmt"
	"sync"
)

type Storage struct {
	PodLock     sync.RWMutex
	ServiceLock sync.RWMutex
	IngressLock sync.RWMutex
	HashLock    sync.RWMutex

	serviceMap map[string]interface{}
	ingressMap map[string]interface{}
	podMap     map[string]interface{}

	currentHash string
}

func NewStorage() *Storage {
	return &Storage{
		ingressMap: map[string]interface{}{},
		serviceMap: map[string]interface{}{},
		podMap:     map[string]interface{}{},
	}
}

func (s *Storage) SetHash(h string) {
	s.HashLock.Lock()
	defer s.HashLock.Unlock()

	s.currentHash = h
}

func (s *Storage) GetHash() string {
	s.HashLock.RLock()
	defer s.HashLock.RUnlock()

	return s.currentHash
}

func (s *Storage) GetPods() []string {
	s.PodLock.RLock()
	defer s.PodLock.RUnlock()

	pods := []string{}
	for k, _ := range s.podMap {
		pods = append(pods, k)
	}

	return pods
}

func (s *Storage) GetServices() []string {
	s.ServiceLock.RLock()
	defer s.ServiceLock.RUnlock()
	pods := []string{}
	for k, _ := range s.serviceMap {
		pods = append(pods, k)
	}

	return pods
}

func (s *Storage) GetIngresses() []string {
	s.IngressLock.RLock()
	defer s.IngressLock.RUnlock()
	pods := []string{}
	for k, _ := range s.ingressMap {
		pods = append(pods, k)
	}

	return pods
}

func (s *Storage) GetPod(name string) (interface{}, bool) {
	return getHelper(&s.PodLock, s.podMap, PodType, name)
}

func (s *Storage) GetService(name string) (interface{}, bool) {
	return getHelper(&s.ServiceLock, s.serviceMap, ServiceType, name)
}

func (s *Storage) GetIngress(name string) (interface{}, bool) {
	return getHelper(&s.IngressLock, s.ingressMap, ServiceType, name)
}

func (s *Storage) SetPod(name string, obj interface{}) {
	setHelper(&s.PodLock, s.podMap, PodType, name, obj)

	s.HashLock.Lock()
	defer s.HashLock.Unlock()

	fmt.Println("hash_change : " + name)
	s.currentHash = s.currentHash + "1"

}

func (s *Storage) SetService(name string, obj interface{}) {
	setHelper(&s.ServiceLock, s.serviceMap, ServiceType, name, obj)

	s.HashLock.Lock()
	defer s.HashLock.Unlock()
	fmt.Println("hash_change : " + name)
	s.currentHash = s.currentHash + "1"

}

func (s *Storage) SetIngress(name string, obj interface{}) {
	setHelper(&s.IngressLock, s.ingressMap, ServiceType, name, obj)

	s.HashLock.Lock()
	defer s.HashLock.Unlock()
	fmt.Println("hash_change : " + name)
	s.currentHash = s.currentHash + "1"
}

func getHelper(m *sync.RWMutex,
	mH map[string]interface{},
	objType string,
	name string) (interface{}, bool) {
	m.RLock()
	defer m.RUnlock()

	answer, ok := mH[name]

	return answer, ok
}

func setHelper(m *sync.RWMutex,
	mH map[string]interface{},
	objType string,
	name string, obj interface{}) {

	m.Lock()
	defer m.Unlock()

	if obj == nil {
		delete(mH, name)
	} else {
		mH[name] = obj
	}
}
