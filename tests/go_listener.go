package main

import (
	"fmt"
	"time"

	"github.com/nats-io/go-nats"
)

var count int
const server = "nats://nero:4222"
// const server = nats.DefaultURL

func on_subscribe(m *nats.Msg) {
    count = count + 1
    fmt.Printf("GET: %s %s\n", m.Subject, string(m.Data))
    time.Sleep(3 * time.Second)
    fmt.Printf("Finish: %s %s\n", m.Subject, string(m.Data))
}

func main() {
    nc, err := nats.Connect(server)
    if nil != err {
        fmt.Println(err)
    }
    fmt.Println("Connected to:", server)
    defer nc.Close()
    defer fmt.Print(count)

    forever := make(chan bool)

    nc.Subscribe(">", func(m *nats.Msg) {
        go on_subscribe(m)
    })

    <-forever

}
