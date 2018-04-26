package main

import (
	"fmt"
    "time"

	"github.com/nats-io/go-nats"
    "net/http"
    "io/ioutil"
)

var count int
var err error
const server = "nats://nero:4222"
const web_server = "nats://nero:2030"
// const server = nats.DefaultURL
var nc *nats.Conn

func on_subscribe(m *nats.Msg) {
    fmt.Printf("GET: %s %s\n", m.Subject, m.Reply)
    nc.Publish(m.Reply, []byte("found"))

    res, err := http.Get("http://nero:2030/api_get_events_scheduled_master?date=2018-04-26")
    if nil != err {
        fmt.Printf("ERROR: %s %s %s\n", m.Subject, m.Reply, err)
        return
    }
    response, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if nil != err {
        fmt.Printf("ERROR: %s %s %s\n", m.Subject, m.Reply, err)
        return
    }
    nc.Publish(m.Reply, []byte(response))
    fmt.Printf("Finish: %s %s\n", m.Subject, m.Reply)
    count = count + 1
}

func counter() {
    for true {
        fmt.Println(count)
        time.Sleep(5 * time.Second)
    }
}

func main() {
    nc, err = nats.Connect(server)
    if nil != err {
        fmt.Println(err)
    }
    fmt.Println("Connected to:", server)
    defer nc.Close()

    forever := make(chan bool)

    nc.Subscribe("api.>", func(m *nats.Msg) {
        go on_subscribe(m)
    })

    go counter()

    <-forever

}
