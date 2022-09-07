package main

import (
	"flag"
	"fmt"
	"github.com/armon/go-socks5"
	"github.com/elazarl/goproxy"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	host := "0.0.0.0"
	sport := flag.Int("s", 1234, "SOCKS5 proxy port")
	pport := flag.Int("p", 1235, "HTTP proxy port")
	dport := flag.Int("d", 80, "HTTP port for auto proxy configuration discovery")
	loc := flag.Bool("l", false, "Whether to pool location details")
	flag.Parse()

	go httpAutoDiscover(host, *dport)
	go httpProxy(host, *pport)
	go socksProxy(host, *sport)

	if *loc {
		go fetchLocation()
	}

	loop()
}

func fetchLocation() {
	fmt.Println("Starting location streaming")
	reader, err := os.Open("/dev/location")
	if err != nil {
		fmt.Println("Failed reading location data")
	}
	p := make([]byte, 256)
	for {
		n, err := reader.Read(p)
		if err == io.EOF {
			fmt.Println("Reached end of location data")
		}
		fmt.Printf("%q\n", p[:n])
	}
}

func httpAutoDiscover(host string, port int) {
	addr := fmt.Sprint(host, ":", port)
	fmt.Println("Starting http discovery at", addr)
	handler := func(w http.ResponseWriter, _ *http.Request) {
		fmt.Println("Serving proxy discovery request")
		_, err := io.WriteString(w, "function FindProxyForURL (url, host) {\n  return 'SOCKS5 172.20.10.1:1234; HTTP 172.20.10.1:1235; DIRECT';\n}")
		if err != nil {
			return
		}
	}
	http.HandleFunc("/proxy", handler)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func loop() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()

	for {
		time.Sleep(1 * time.Second)
	}
}

func httpProxy(host string, port int) {
	addr := fmt.Sprint(host, ":", port)
	fmt.Println("Starting http proxy at", addr)
	proxy := goproxy.NewProxyHttpServer()
	//proxy.Verbose = true
	log.Fatal(http.ListenAndServe(addr, proxy))
}

func socksProxy(host string, port int) {
	addr := fmt.Sprint(host, ":", port)
	fmt.Println("Starting socks5 proxy at", addr)
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	// Create SOCKS5 proxy on localhost port 8000
	if err := server.ListenAndServe("tcp", addr); err != nil {
		panic(err)
	}
}
