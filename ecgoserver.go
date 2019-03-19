// implement echoserver in go

package main

import (
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "math/rand"
    "net/http"
    "os"
    "runtime"
    "strconv"
    "strings"
    "time"
    "github.com/gorilla/handlers"
    "github.com/mediocregopher/radix"
)

var bind = ""
var port int = 8000

var redisUrl string
var redisPool *radix.Pool

var version = false
var versionNumber string
var gitCommit string

var readTimeout = 10
var writeTimeout = 300

const contentTypeHeader = "Content-Type"

const atoz = "abcdefghijklmnopqrstuvwxyz"
const welcome = "Welcome to Ecgo server\n"
var message1 = "Ecgo gives you %d"


func printVersion() {
    fmt.Printf("ecgoserver %s\n", versionNumber)
    fmt.Printf("  git commit: %s\n", gitCommit)
    fmt.Printf("  go version: %s\n", runtime.Version())
    os.Exit(0)
}

func validate() {
    // check if required arguments are available
    if version == true {
        printVersion()
    }
}

// generate random string
func generateRandomString(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = atoz[rand.Intn(len(atoz))]
    }
    return string(b)
}

// handle home page
func indexHandler(w http.ResponseWriter, req *http.Request) {
    io.WriteString(w, welcome)
}

// handle GET calls
func getHandler(w http.ResponseWriter, req *http.Request) {
    message := welcome

    // custom response code
    statusCode, err := strconv.Atoi(req.URL.Query().Get("status"))
    if err != nil || statusCode < 200 || statusCode > 599 {
        statusCode = http.StatusOK
    }

    // sleep for some time, to simulate delays
    sleepSeconds, err := strconv.Atoi(req.URL.Query().Get("sleep"))
    if err == nil {
        time.Sleep(time.Duration(sleepSeconds) * time.Second)
    }

    // generate random output back
    bytes, err := strconv.Atoi(req.URL.Query().Get("bytes"))
    if err == nil {
        message = generateRandomString(bytes)
    } else {
        message = fmt.Sprintf(message1, statusCode)
    }

    // get response string from redis
    redisKey := req.URL.Query().Get("key")
    if redisKey != "" && redisPool != nil {
        err = redisPool.Do(radix.Cmd(&message, "GET", redisKey))
        if err != nil {
            statusCode = http.StatusInternalServerError
        } else if message == "" {
            statusCode = http.StatusNotFound
        } else {
            // assume redis store has json content
            w.Header().Set(contentTypeHeader, "application/json")
        }
    }

    // echo all custom headers back, along with "Set-Cookie"
    for key, values := range req.Header {
        if key == "Set-Cookie" || strings.HasPrefix(key, "X-") {
            w.Header().Set(key, values[0])
        }
    }

    w.WriteHeader(statusCode)
    io.WriteString(w, message)
}

// handle POST calls
func postHandler(w http.ResponseWriter, req *http.Request) {
    if req.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
    } else {
        body, err := ioutil.ReadAll(req.Body)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
        } else {
            // set the same "Content-Type" back
            w.Header().Set(contentTypeHeader, req.Header.Get(contentTypeHeader))
            io.WriteString(w, string(body))
        }
    }
}

func init() {
    flag.BoolVar(&version, "version", version, "print version")
    flag.IntVar(&port, "port", port, "Port on which ecgo server runs")
    flag.StringVar(&bind, "bind", bind, "To bind to a specific address")
    flag.StringVar(&redisUrl, "redis-url", redisUrl, "Redis url of response store")
    flag.IntVar(&readTimeout, "read-timeout", readTimeout, "Read timeout for ecgo server")
    flag.IntVar(&writeTimeout, "write-timeout", writeTimeout, "Write timeout for ecgo server")

    // parse and validate input
    flag.Parse()
    validate()
}

func main() {
    addr := fmt.Sprintf("%s:%d", bind, port)
    fmt.Printf("Starting ecgo server address=%s version=%s gitcommit=%s\n", addr, versionNumber, gitCommit)

    if redisUrl != "" {
        var err error
        redisPool, err = radix.NewPool("tcp", redisUrl, 10)
        if err != nil {
            fmt.Printf("Could not connect to %s - %v\n", redisUrl, err)
            os.Exit(1)
        }
        fmt.Printf("Connected to %s\n", redisUrl)
    }
    defer redisPool.Close()

    handler := http.NewServeMux()
    handler.HandleFunc("/", indexHandler)
    handler.HandleFunc("/get/", getHandler)
    handler.HandleFunc("/post/", postHandler)

    server := http.Server{
        Addr: addr,
        Handler: handlers.CombinedLoggingHandler(os.Stdout, handler),
        ReadTimeout: time.Duration(readTimeout) * time.Second,
        WriteTimeout: time.Duration(writeTimeout) * time.Second,
    }

    server.ListenAndServe()
}
