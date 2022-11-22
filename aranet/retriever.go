package aranet

import (
	"fmt"
	"sync"

	"sbinet.org/x/aranet4"
)

type Retriever struct {
	ID    string
	data  aranet4.Data
	mutex *sync.Mutex
}

func (r *Retriever) Read() aranet4.Data {
	return r.data
}

func (r *Retriever) Room() string {
	return "WIP"
}

func (r *Retriever) Update() error {
	device, err := aranet4.New(r.ID)
	if err != nil {
		return fmt.Errorf("failed connecting to aranet4: %v\n", err)
	}
	data, err := device.Read()
	if err != nil {
		return fmt.Errorf("failed updating data from aranet4: %v\n", err)
	}

	r.data = data
	_ = device.Close()
	return nil
}
