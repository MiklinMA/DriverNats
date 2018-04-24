package main

import (
	"log"
	"fmt"

    "html"
    "strings"
    // "reflect"
	"math/rand"
    "net/http"
    "io/ioutil"
    "encoding/json"

	"github.com/nats-io/go-nats"
)

const TimeoutQueue = 1e+9
const TimeoutWork = 30e+9

var con *nats.Conn

type Packet struct {
    Method string
    Data map[string]string
    Raw string
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func request_mq(subject string, data []byte) (code int, result string) {

    subject_resp := subject + "." + randomString(10)
    code = 200

    sub, err := con.SubscribeSync(subject_resp)
    defer sub.Unsubscribe()
    if nil != err {
        return 500, err.Error()
    }

    err = con.PublishRequest(subject, subject_resp, data)

    res, err := sub.NextMsg(TimeoutQueue)
    if nil != err {
        return 404, err.Error()
    }

    result = string(res.Data)

    for result == "found" {
        res, err := sub.NextMsg(TimeoutWork)
        if nil != err {
            return 504, err.Error()
        }

        result = string(res.Data)
    }
    return
}

func parse_request(r *http.Request) (p Packet, err error) {
    method := ""
    method = fmt.Sprintf("%s.%s", html.EscapeString(r.URL.Path), r.Method)
    method = strings.ToLower(method)
    method = strings.Trim(method, "/")
    method = strings.Replace(method, "/", ".", -1)
    method = strings.Replace(method, "_", ".", -1)

    p = Packet{}
    p.Method = method

    r.ParseForm()
    p.Data = make(map[string]string)
    for k, v := range r.Form {
        p.Data[k] = v[0]
    }

    defer r.Body.Close()
    body, err := ioutil.ReadAll(r.Body)

    if nil != err {
        fmt.Println("body err", err)
        return
    }
    p.Raw = string(body[:])
    return
}

func http_handler(w http.ResponseWriter, r *http.Request) {

    p, err := parse_request(r)
    if nil != err {
        fmt.Println("parse error", err)
        return
    }

    log.Printf("API: %s", p.Method)

    data, err := json.Marshal(p)
    if nil != err {
        fmt.Println("json error", err)
        return
    }

    if p.Method[:8] == "api.async" {
        err = con.Publish(p.Method, []byte(data))
        fmt.Fprintf(w, "OK")
    } else {
        code, response := request_mq(p.Method, []byte(data))
        w.WriteHeader(code)
        fmt.Fprintf(w, "%s", response)
        log.Printf("API: %s %s", p.Method, response)
    }
}

func main() {
    var err error
    nats_host := nats.DefaultURL

    con, err = nats.Connect(nats_host)
    if nil != err {
        log.Println("Connection to " + nats_host + " failed")
        return
    }

	log.Println("Connected to " + nats_host)

	defer func() {
        con.Close()
        log.Println("Disconnected from " + nats_host)
    }()

    http.HandleFunc("/", http_handler)

    log.Println("Listening on " + ":8888")
    log.Fatal(http.ListenAndServe(":8888", nil))
}


