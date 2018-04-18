package main

import (
    "fmt"
    "net"
    "os"
    "strings"
    // "encoding/json"
    "github.com/nats-io/go-nats"
)

const (
    CONN_HOST = ""
    CONN_PORT = "2333"
    CONN_TYPE = "tcp"
)

func main() {
    nc, _ := nats.Connect(nats.DefaultURL)
    fmt.Println("Connected to " + nats.DefaultURL)
    defer nc.Close()

    l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    defer l.Close()
    fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
    for {
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }
        go handleRequest(conn, nc)
    }
}

func add_listener(conn net.Conn, nc *nats.Conn, u_ref string) (subs []*nats.Subscription) {
    subject := "ones.user"

    on_subscribe := func(msg *nats.Msg) {
        // packet := make(map[string]string)

        // packet["subject"] = msg.Subject
        // packet["data"] = string(msg.Data)


        // fmt.Println("GOT DATA: ", packet)

        // data, _ := json.Marshal(packet)
        // conn.Write(data)
        nc.Publish(msg.Reply, []byte("found"))
        // nc.Publish(msg.Subject + ".found", []byte("found"))
        conn.Write(msg.Data)
        conn.Write([]byte("\r\n"))
    }

    subjects := []string {
        subject + "." + u_ref,
        // subject + "." + u_ref + ".*",
    }

    for _, subj := range subjects {
        sub, _ := nc.Subscribe(subj, on_subscribe)
        fmt.Println("User subscribed:", subj)
        subs = append(subs, sub)
    }

    return
}

func handleRequest(conn net.Conn, nc *nats.Conn) {
    defer conn.Close()

    buf := make([]byte, 1024)
    u_ip := conn.RemoteAddr().String()

    for true {
        n, err := conn.Read(buf)
        if err != nil {
            fmt.Println("Error reading:", err.Error())
            break
        }

        s_buf := string(buf[:n])
        s_buf = strings.TrimSpace(s_buf)
        pack := strings.SplitN(s_buf, ":", 2)

        fmt.Printf("Message: %s %q\n", u_ip, s_buf)

        if len(pack) > 1 {
            if pack[0] == "u_ref" {
                subs := add_listener(conn, nc, pack[1])
                for _, sub := range subs {
                    defer sub.Unsubscribe()
                }
            } else {
                err = nc.Publish(pack[0], []byte(pack[1]))
            }
        } else {
            if pack[0] == "logout" {
                break
            }
        }
    }
    fmt.Println("Disconnected", u_ip)
}
