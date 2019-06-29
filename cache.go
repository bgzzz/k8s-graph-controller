package main

/*
import (
	"fmt"
	"sync"

	"github.com/r3labs/diff"
)

type GraphDiff struct {
	VergesToDelete []Verge `json:"verges_to_delete"`
	VergesToAdd    []Verge `json:"verges_to_add"`
	EdgesToDelete  []Edge  `json:"edges_to_delete"`
	EdgesToAdd     []Edge  `json:"edges_to_add"`
	NewHash        string  `json:"hash"`
}

type Cache struct {
	cacheLock sync.RWMutex

	graphMap map[string]Graph
}

func NewCache() *Cache {
	return &Cache{
		graphMap: map[string]Graph{},
	}
}

func (c *Cache) SetToCache(g Graph) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	c.graphMap[g.Hash] = g
}

func (c *Cache) GetDiff(g Graph, h string) *GraphDiff {

	c.cacheLock.RLock()
	oldGraph := c.graphMap[h]
	c.cacheLock.RUnlock()

	// check verges

	toAdd := map[string]Verge{}
	toDelete := map[string]Verge{}

	change, err := diff.Diff(oldGraph, g)
	if err != nil {
		fmt.Println("Err: " + err.Error())
		return nil
	}

  change.Filter("path")

	// for
	return nil
}
*/
