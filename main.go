package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	websocketSuccessful = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_successful",
			Help: "( 0 = false , 1 = true )",
		})
	websocketResponseTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_response_time",
			Help: "( Time until we get EOSE in ms; 0 for failed )",
		})
	websocketStatusCode = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_status_code",
			Help: "( 101 is normal status code for ws )",
		})

	respCode float64
)

func probeHandler(w http.ResponseWriter, r *http.Request) {

	messages, msOk := r.URL.Query()["message"]
	if !msOk || len(messages) != 1 {
		http.Error(w, "Specify one 'message' parameter", http.StatusBadRequest)
		return
	}
	sendMessage := []byte(messages[0])

	contains, coOk := r.URL.Query()["contains"]
	if !coOk || len(contains) != 1 {
		http.Error(w, "Specify one 'contains' parameter", http.StatusBadRequest)
		return
	}
	contain := contains[0]

	targets, trOk := r.URL.Query()["target"]
	if !trOk || len(targets) != 1 {
		http.Error(w, "Specify one 'target' parameter", http.StatusBadRequest)
		return
	}
	target := targets[0]

	ur, _ := url.Parse(target)

	u := url.URL{Scheme: ur.Scheme, Host: ur.Host, Path: ur.Path, RawQuery: ur.RawQuery}

	fmt.Printf("Probing %s with message %s and checkign if response contains %s\n", target, sendMessage, contain)

	headers := http.Header{}
	headers.Add("User-Agent", "websocket-exporter/0.1.0")

	c, resp, errCon := websocket.DefaultDialer.Dial(u.String(), headers)

	start := time.Now()
	timeout := 5 * time.Second
	c.SetReadDeadline(start.Add(timeout))
	c.SetWriteDeadline(start.Add(timeout))

	if resp != nil {
		respCode = float64(resp.StatusCode)
	} else {
		respCode = 0
	}

	if (errCon != nil) || (float64(resp.StatusCode) != 101) {
		websocketSuccessful.Set(0)
		websocketStatusCode.Set(respCode)
		websocketResponseTime.Set(0)
	} else {
		// send ws message
		err := c.WriteMessage(websocket.TextMessage, sendMessage)
		if err != nil {
			websocketSuccessful.Set(0)
			websocketStatusCode.Set(respCode)
			websocketResponseTime.Set(0)
		}
	OuterLoop:
		for {
			select {
			case <-time.After(timeout):
				websocketSuccessful.Set(0)
				websocketStatusCode.Set(respCode)
				websocketResponseTime.Set(0)
				break OuterLoop
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					websocketSuccessful.Set(0)
					websocketStatusCode.Set(respCode)
					websocketResponseTime.Set(0)
					break OuterLoop
				}
				if strings.Contains(string(message), contain) {
					websocketSuccessful.Set(1)
					websocketStatusCode.Set(respCode)
					websocketResponseTime.Set(float64(time.Since(start).Milliseconds()))
					break OuterLoop
				} else {
					websocketSuccessful.Set(0)
					websocketStatusCode.Set(respCode)
					websocketResponseTime.Set(0)
				}
			}
		}
		c.Close()
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(websocketSuccessful)
	reg.MustRegister(websocketStatusCode)
	reg.MustRegister(websocketResponseTime)

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {

	port := flag.Int("port", 9143, "Port Number to listen")
	host := flag.String("host", "127.0.0.1", "Host to listen")

	flag.Parse()

	addr := fmt.Sprint(*host, ":", *port)
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		probeHandler(w, r)
	})

	fmt.Println("Starting exporter on addr: ", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("Error staring server: ", err)
	}

}
