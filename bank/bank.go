package bank

import (
	"context"
	"github.com/ahornerr/artifacts/httperror"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"sync"
)

type Bank struct {
	items map[string]int

	client  *client.ClientWithResponses
	mux     sync.Mutex
	updates chan<- map[string]int
}

func (b *Bank) Items() map[string]int {
	b.mux.Lock()
	defer b.mux.Unlock()

	return b.items
}

func NewBank(c *client.ClientWithResponses, updates chan<- map[string]int) *Bank {
	return &Bank{
		items:   map[string]int{},
		client:  c,
		updates: updates,
	}
}

func (b *Bank) Load(ctx context.Context) ([]client.SimpleItemSchema, error) {
	page := 1
	size := 100

	var bankItems []client.SimpleItemSchema

	for {
		resp, err := b.client.GetBankItemsMyBankItemsGetWithResponse(ctx, &client.GetBankItemsMyBankItemsGetParams{
			Page: &page,
			Size: &size,
		})
		if err != nil {
			return nil, err
		} else if resp.JSON200 == nil {
			return nil, httperror.NewHTTPError(resp.StatusCode(), resp.Body)
		}

		bankItems = append(bankItems, resp.JSON200.Data...)

		if len(resp.JSON200.Data) < size {
			break
		}

		page++
	}

	b.Update(bankItems)

	return bankItems, nil
}

func (b *Bank) Update(newItems []client.SimpleItemSchema) {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.items = map[string]int{}
	for _, bankItem := range newItems {
		b.items[bankItem.Code] += bankItem.Quantity
	}

	b.updates <- b.items
}
