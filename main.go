package main

import (
    "strings"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

type requestResponse struct {
    statusCode int
    data string
}

func main() {
    var port, server string
    cache := make(map[string]requestResponse)

    fmt.Print("Enter port to connect to: ")
    fmt.Scanln(&port)
    fmt.Print("Enter server to connect to: ")
    fmt.Scanln(&server)

    port = fmt.Sprintf(":%v", port)

    fmt.Println("Running on port ", port)

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        seek := r.URL.Path[1:]

        if value, exists := cache[seek]; exists {
            fmt.Println("Cache HIT")
            w.WriteHeader(value.statusCode)
            io.Copy(w, strings.NewReader(value.data))
        } else {
            fmt.Println("Cache MISS")
            resp := forward(r, server, seek)
            defer resp.Body.Close()

            for k, vals := range resp.Header {
                for _, v := range vals {
                    w.Header().Add(k, v)
                }
            }

            s, _ := io.ReadAll(resp.Body)
            cache[seek] = requestResponse{
                statusCode: resp.StatusCode,
                data: string(s),
            }

            w.WriteHeader(resp.StatusCode)
            io.Copy(w, resp.Body)
        }
    })

    log.Fatal(http.ListenAndServe(port, nil))
}

func forward(r *http.Request, server string, seek string) *http.Response {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        log.Fatal(err)
    }

    defer r.Body.Close()

    url := fmt.Sprintf("%v/%v", server, seek)

    proxyRequest, err := http.NewRequest(r.Method, url, bytes.NewReader(body))
    if err != nil {
        log.Fatal(err)
    }

    for h, vals := range r.Header {
        for _, v := range vals {
            proxyRequest.Header.Add(h, v)
        }
    }

    client := &http.Client{}
    resp, err := client.Do(proxyRequest)
    if err != nil {
        log.Fatal(err)
    }

    return resp
}

