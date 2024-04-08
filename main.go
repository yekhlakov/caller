package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Connection struct {
	id      int
	url     *url.URL    // base url
	client  http.Client // http client to perform calls
	headers http.Header // http headers to send to the server
}

var idseq = 0

func NewConnection(baseUrl string) *Connection {
	u, _ := url.Parse(baseUrl)

	idseq = idseq + 1
	return &Connection{
		id:      idseq,
		url:     u,
		client:  http.Client{},
		headers: http.Header{},
	}
}

func (conn *Connection) Call(message json.RawMessage) {

	body, _ := json.Marshal(message)

	request := &http.Request{
		Method: http.MethodPost,
		URL:    conn.url,
		Header: conn.headers,
		Body:   ioutil.NopCloser(bytes.NewReader(body)),
	}

	response, err := conn.client.Do(request)
	if err != nil {
		fmt.Println(time.Now(), conn.id, "error:", err)
		return
	}

	defer func() { _ = response.Body.Close() }()

	responseBody, _ := ioutil.ReadAll(response.Body)

	fmt.Println(time.Now(), conn.id, "body:", string(body), "code:", response.StatusCode, "response:", string(responseBody))
}

func (conn *Connection) AddHeader(key, value string) {
	conn.headers.Add(key, value)
}

type TemplateFile map[string]json.RawMessage

type IdListFile map[string][]string

var idLists IdListFile = map[string][]string{}

var messages = map[float32]string{}

func GetRawRandomMessage() string {
	randomValue := rand.Float32()
	var probabilitySum float32 = 0.0
	for probability, message := range messages {
		if randomValue < probability {
			return message
		}
		probabilitySum = probabilitySum + probability
	}

	return `{}`
}

func GetRandomMessage() json.RawMessage {

	msg := GetRawRandomMessage()
	id := rand.Uint64()

	// replace patterns in the message with required elements
	msg = strings.ReplaceAll(msg, "#ID#", fmt.Sprintf("CALLER-%d", id))

	for key, list := range idLists {
		msg = strings.ReplaceAll(msg, "#"+key+"#", list[rand.Intn(len(list))])
	}

	return []byte(msg)
}

func main() {

	args := os.Args[1:]
	// Args order: rps numConns url headerFile requestFile idListFile

	// Required overall RPS
	rps, _ := strconv.Atoi(args[0])

	// Delay between requests
	delay := time.Duration(1000000/rps) * time.Microsecond

	// Number of connections
	numConns, _ := strconv.Atoi(args[1])

	// Read and prepare headers
	headerFile, _ := ioutil.ReadFile(args[3])
	headers := strings.Split(string(headerFile), "\n")

	// Create the connections
	connections := make([]*Connection, numConns)
	for i := 0; i < numConns; i++ {
		connections[i] = NewConnection(args[2])

		for _, v := range headers {
			nameValue := strings.Split(v, ":")
			if len(nameValue) != 2 {
				continue
			}
			connections[i].AddHeader(strings.TrimSpace(nameValue[0]), strings.TrimSpace(nameValue[1]))
		}
	}

	// Read the request templates
	var t TemplateFile
	requestFile, _ := ioutil.ReadFile(args[4])
	_ = json.Unmarshal(requestFile, &t)

	for p, v := range t {
		f, _ := strconv.ParseFloat(p, 32)
		messages[float32(f)] = string(v)
	}

	// Read the id-lists
	idListFile, _ := ioutil.ReadFile(args[5])
	_ = json.Unmarshal(idListFile, &idLists)

	// Perform the requests
	connCycle := 0
	for {
		time.Sleep(delay)
		go func() { connections[connCycle].Call(GetRandomMessage()) }()
		connCycle = (connCycle + 1) % numConns
	}
}
