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

type Props struct {
	addr string
	bind string

	socks     int
	http      int
	discovery int

	verbose  bool
	location bool
}

func main() {
	props := Props{
		addr: *flag.String("a", "172.20.10.1", "Proxy address to expose to clients"),
		bind: *flag.String("b", "0.0.0.0", "Address to bind to"),

		socks:     *flag.Int("s", 0, "SOCKS5 proxy port"),
		http:      *flag.Int("p", 0, "HTTP proxy port"),
		discovery: *flag.Int("d", 0, "HTTP port for auto proxy configuration discovery"),

		location: *flag.Bool("l", false, "Whether to pool location details"),
		verbose:  *flag.Bool("v", false, "Enable verbose output"),
	}
	help := *flag.Bool("h", false, "Show help")

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if props.discovery != 0 {
		go httpAutoDiscover(props)
	}
	if props.http != 0 {
		go httpProxy(props)
	}
	if props.socks != 0 {
		go socksProxy(props)
	}
	if props.location {
		go fetchLocation(props.verbose)
	}

	loop()
}

func fetchLocation(verbose bool) {
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
		if verbose {
			fmt.Printf("%q", p[:n])
		}
	}
}

func httpAutoDiscover(props Props) {
	addr := fmt.Sprint(props.bind, ":", props.discovery)
	fmt.Println("Starting http discovery at", addr)
	handler := func(w http.ResponseWriter, _ *http.Request) {
		if props.verbose {
			fmt.Println("Serving proxy discovery request")
		}
		funct := "function FindProxyForURL (url, host) {\n  return '"
		if props.socks != 0 {
			funct += fmt.Sprint("SOCKS5 ", props.addr, ":", props.socks, "; ")
		}
		if props.http != 0 {
			funct += fmt.Sprint("HTTP ", props.addr, ":", props.http, "; ")
		}

		funct += "DIRECT';}"

		_, err := io.WriteString(w, funct)
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

func httpProxy(props Props) {
	addr := fmt.Sprint(props.addr, ":", props.http)
	fmt.Println("Starting http proxy at", addr)
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = props.verbose
	log.Fatal(http.ListenAndServe(addr, proxy))
}

func socksProxy(props Props) {
	addr := fmt.Sprint(props.addr, ":", props.socks)
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
