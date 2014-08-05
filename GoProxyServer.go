package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)
/**
*The proxy server port
*/
var port string
/**
*The real proxy addresses
*/
var addrs string

func proxyHandler(respw http.ResponseWriter, req *http.Request) {

	addrsArray := strings.Split(addrs, ";")
	addr := addrsArray[rand.Intn(len(addrsArray))]
	c, err := net.Dial("tcp", addr)
	if err != nil {
		http.Error(respw, err.Error(), http.StatusGatewayTimeout)
		loghit(req, addr, http.StatusGatewayTimeout)
		return
	}
	c.SetReadDeadline(time.Now().Add(30 * time.Second))
	cc := httputil.NewClientConn(c, nil)
	err = cc.Write(req)
	if err != nil {
		http.Error(respw, err.Error(), http.StatusGatewayTimeout)
		loghit(req, addr, http.StatusGatewayTimeout)
		return
	}

	resp, err := cc.Read(req)
	defer resp.Body.Close()
	if err != nil && err != httputil.ErrPersistEOF {
		http.Error(respw, err.Error(), http.StatusGatewayTimeout)
		loghit(req, addr, http.StatusGatewayTimeout)
		return
	}
	
	for k, v := range resp.Header {
		for _, vv := range v {
			respw.Header().Add(k, vv)
		}
	}
	respw.Header().Add("X-Forwarded-For","GoProxyServer")
	respw.WriteHeader(resp.StatusCode)
	io.Copy(respw, resp.Body)
	loghit(req, addr, resp.StatusCode)
}

func loghit(r *http.Request, addr string, code int) {
	log.Printf("%v %v %v %v", addr, r.Method, r.RequestURI, code)
}
func loadConfig() {
	configMap := map[string]string{}
	file, err := os.Open("config.properties")
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		kv := strings.Split(line, "=")
		if len(kv) == 2 {
			configMap[kv[0]] = kv[1]
		}
	}
	if value, ok := configMap["port"]; ok {
		port = value
		fmt.Println("The server port:"+value)
	}
	if value, ok := configMap["addrs"]; ok {
		addrs = value
		fmt.Println("The real addrs:"+value)
	}

}
func startGoProxyServer() {
	log.Println("Start GoProxyServer on port "+port)
	http.HandleFunc("/", proxyHandler)
	err := http.ListenAndServe("[::]:"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	loadConfig()
	startGoProxyServer()
}
