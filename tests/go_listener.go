package main

import (
	"fmt"
    "time"
    "strings"
    "strconv"

    "net/http"
    "io/ioutil"
    "encoding/json"

	"github.com/nats-io/go-nats"
)

var count int
var err error
const server = "nats://nero:4222"
const web_server = "http://127.0.0.1:2030/"
// const server = nats.DefaultURL
var nc *nats.Conn

type Packet struct {
    Method string
    Data map[string]string
    Header map[string]string
    Raw string
}

func on_subscribe(m *nats.Msg) {
    fmt.Printf("GET: %s %s\n", m.Subject, m.Reply)
    nc.Publish(m.Reply, []byte("found"))

    var p Packet

    err = json.Unmarshal(m.Data, &p)
    if nil != err {
        fmt.Printf("ERROR: %s %s\n", m.Subject, err)
        nc.Publish(m.Reply, []byte("400"))
        return
    }

    method := p.Method
    arr := strings.Split(method, ".")
    path := strings.Join(arr[:len(arr)-1], "_")
    method = arr[len(arr)-1]
    method = strings.ToUpper(method)

    uri := web_server + path

    req, _ := http.NewRequest(method, uri, nil)
    q := req.URL.Query()
    for k, v := range p.Data {
        q.Add(k, v)
    }
    req.URL.RawQuery = q.Encode()

    for k, v := range p.Header {
        req.Header.Add(k, v)
    }
    // fmt.Println(m.Subject, req.URL.String())

    res, err := http.DefaultClient.Do(req)

    if nil != err {
        fmt.Printf("ERROR: %s %s\n", m.Subject, err)
        nc.Publish(m.Reply, []byte("502"))
        return
    }
    response, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if nil != err {
        fmt.Printf("ERROR: %s %s\n", m.Subject, err)
        nc.Publish(m.Reply, []byte("502"))
        return
    }
    if res.StatusCode == 200 {
        nc.Publish(m.Reply, []byte(response))
        fmt.Printf("Done: %s %d\n", m.Subject, res.StatusCode)
        // count = count + 1
    } else {
        nc.Publish(m.Reply, []byte(strconv.Itoa(res.StatusCode)))
        fmt.Printf("Finish: %s %d %s\n", m.Subject, res.StatusCode, string(response))
    }
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

    // go counter()

    <-forever

}
