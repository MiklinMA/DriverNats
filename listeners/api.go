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

const TimeoutQueue = 3e+9
const TimeoutWork = 30e+9

var con *nats.Conn

type Packet struct {
    Method string
    Data map[string]string
    Raw string
}

func failOnError(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %s", msg, err)
        panic(fmt.Sprintf("%s: %s", msg, err))
    }
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

func request_mq(subject string, data []byte) (result string, err error) {

    // log.Println(reflect.TypeOf(con))

    subject_res := subject + "." + randomString(10)

    log.Print(subject_res)
    sub, err := con.SubscribeSync(subject_res)
    defer sub.Unsubscribe()

    err = con.PublishRequest(subject, subject_res, data)

    res, err := sub.NextMsg(TimeoutQueue)
    if err != nil {
        log.Print("Request timed out: ", err)
        result = "Not found"
        return
    }
    result = string(res.Data)

    /*
    // for len(result) >= 7 && result[:7] == "working" {
    log.Print("Response: ", result)
    for result == "working" {
        res, err := sub.NextMsg(TimeoutWork)

        if err != nil {
            result = "Working request timed out: " + err.Error()
            log.Print(result)
            break
        }

        result = string(res.Data)

        log.Print("Response: ", result)
    }
    */

    log.Print("Response: ", result)
    return
}

func on_response(w http.ResponseWriter, r *http.Request) {
    // log.Print("Method URL Scheme")
    // log.Printf("%s %s", r.Method, r.URL.Path, r.URL.Scheme)

    method := ""
    method = fmt.Sprintf("%s.%s", html.EscapeString(r.URL.Path), r.Method)
    method = strings.ToLower(method)
    method = strings.Trim(method, "/")
    method = strings.Replace(method, "/", ".", -1)
    method = strings.Replace(method, "_", ".", -1)

    p := Packet{}
    p.Method = method

    log.Printf("API method: %s", method)

    // log.Print("Header  ")
    // for k, v := range r.Header {
    //     log.Printf(" %-20s : %s", k, v)
    // }

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

    b, err := json.Marshal(p)
    if nil != err {
        fmt.Println("json err", err)
        return
    }

    response := "DONE"
    if method[:8] == "api.async" {
        err = con.Publish(method, []byte(b))
    } else {
        response, err = request_mq(method, []byte(b))
        if nil != err {
            fmt.Println("Request error:", err)
            http.NotFound(w, r)
            return
        }

        if response == "Not found" {
            http.NotFound(w, r)
            return
        }
    }

    /// w.WriteHeader(404)
    fmt.Fprintf(w, "%s", response)
}

func main() {

	con, _ = nats.Connect(nats.DefaultURL)
	log.Println("Connected to " + nats.DefaultURL)

	defer con.Close()

    http.HandleFunc("/", on_response)

    log.Println("Listening on " + "*:8888")
    log.Fatal(http.ListenAndServe(":8888", nil))
}


