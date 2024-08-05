package main

import (
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"sync"
)

type Bank struct {
	Items []client.SimpleItemSchema
	mux   sync.Mutex
}

func (b *Bank) Update(newItems []client.SimpleItemSchema) {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.Items = newItems
}

func (b *Bank) AsMap() map[string]int {
	bankItemsMap := map[string]int{}
	for _, bankItem := range b.Items {
		bankItemsMap[bankItem.Code] += bankItem.Quantity
	}
	return bankItemsMap
}
