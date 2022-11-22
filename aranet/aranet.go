package aranet

import (
	"context"
	"log"
	"time"

	"github.com/brutella/hap/accessory"
	"sbinet.org/x/aranet4"
)

type AranetData interface {
	Read() aranet4.Data
	Room() string
}

type Aranet struct {
	accessory *Accessory
	id        string
	room      string
	context   context.Context
	ticker    time.Ticker
	retriever Retriever
}

func New(context context.Context, id string, room string) *Aranet {
	retriever := Retriever{ID: id}
	acc := NewAranetAccessory(accessory.Info{Name: "Aranet4"})
	return &Aranet{
		accessory: acc,
		context:   context,
		id:        id,
		retriever: retriever,
		room:      room,
	}
}

func (a *Aranet) RunUpdateLoop(verbose bool) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	log.Printf("Monitoring aranet %s as room=%s", a.id, a.room)

	data := a.retriever.Read()

	for {
		select {
		case <-ticker.C:
			if verbose {
				log.Println("tick")
			}

			if time.Since(data.Time) < data.Interval {
				continue
			}

			if verbose {
				log.Println("updating")
			}

			if err := a.retriever.Update(); err != nil {
				log.Printf("failed update (%s): %v", a.id, err)
				continue
			}
			data = a.retriever.Read()
			if verbose {
				log.Printf("got: %#v\n", data)
			}

			a.accessory.Update(data)
		case <-a.context.Done():
			log.Println("Stopped updating loop")
			return
		}
	}
}

func (a *Aranet) Read() aranet4.Data {
	return a.retriever.Read()
}

func (a *Aranet) Room() string {
	return a.room
}

func (a *Aranet) Accessory() *Accessory {
	return a.accessory
}
