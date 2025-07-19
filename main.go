package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

type requestResponse struct {
    statusCode int
    header map[string]map[int]string
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

        if _, exists := cache[seek]; !exists {
            fmt.Println("Cache MISS")
            resp := forward(r, server, seek)
            defer resp.Body.Close()

            header := make(map[string]map[int]string)
            for k, vals := range resp.Header {
                header[k] = make(map[int]string)
                for val, v := range vals {
                    header[k][val] = v
                }
            }

            s, _ := io.ReadAll(resp.Body)
            cache[seek] = requestResponse{
                statusCode: resp.StatusCode,
                header: header,
                data: string(s),
            }
        }
        
        for i, vals := range cache[seek].header {
            for _, v := range vals {
                w.Header().Add(i, v)
            }
        }
        w.WriteHeader(cache[seek].statusCode)
        fmt.Fprintf(w, "%v", cache[seek].data)
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

