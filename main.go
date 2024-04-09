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

func (conn *Connection) AddHeaders(headers map[string]string) {
	for k, v := range headers {
		conn.AddHeader(k, v)
	}
}

type RequestTemplate struct {
	Probability float64         `json:"probability"`
	Template    json.RawMessage `json:"template"`
}

type ConfigFile struct {
	Rps              int                 `json:"rps"`
	NumConnections   int                 `json:"numConnections"`
	Url              string              `json:"url"`
	Headers          map[string]string   `json:"headers,omitempty"`
	RequestTemplates []RequestTemplate   `json:"requestTemplates"`
	IdLists          map[string][]string `json:"idLists,omitempty"`
}

var config ConfigFile

// GetRandomMessageTemplate Pick a random template from template list
func GetRandomMessageTemplate() string {
	randomValue := rand.Float64()

	totalProbability := 0.0
	for i := range config.RequestTemplates {
		totalProbability += config.RequestTemplates[i].Probability
		if randomValue < totalProbability {
			return string(config.RequestTemplates[i].Template)
		}
	}

	return `{}`
}

// GetRandomMessage pick a random message
func GetRandomMessage() json.RawMessage {

	msg := GetRandomMessageTemplate()
	id := rand.Uint64()

	// replace patterns in the message with required elements
	msg = strings.ReplaceAll(msg, "#ID#", fmt.Sprintf("CALLER-%d", id))

	for key, list := range config.IdLists {
		if len(list) > 0 {
			msg = strings.ReplaceAll(msg, "\"##"+key+"##\"", list[rand.Intn(len(list))])
			msg = strings.ReplaceAll(msg, "#"+key+"#", list[rand.Intn(len(list))])
		}
	}

	return []byte(msg)
}

func LoadConfig(fileName string) error {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &config)
	if err != nil {
		return err
	}

	totalProbability := 0.0
	for _, v := range config.RequestTemplates {
		totalProbability = totalProbability + v.Probability
	}

	for p := range config.RequestTemplates {
		config.RequestTemplates[p].Probability = config.RequestTemplates[p].Probability / totalProbability
	}

	return nil
}

func main() {

	args := os.Args[1:]

	configFileName := "config.json"
	if len(args) > 0 {
		configFileName = args[1]
	}

	err := LoadConfig(configFileName)
	if err != nil {
		fmt.Println("config loading error:", err)
		return
	}

	// Delay between requests
	delay := time.Duration(1000000/config.Rps) * time.Microsecond

	// Create the connections
	connections := make([]*Connection, config.NumConnections)
	for i := 0; i < config.NumConnections; i++ {
		connections[i] = NewConnection(config.Url)
		connections[i].AddHeaders(config.Headers)
	}

	// Perform the requests
	connCycle := 0
	for {
		time.Sleep(delay)
		// Each request is done in its own goroutine
		go func() { connections[connCycle].Call(GetRandomMessage()) }()
		// The connections are cycled
		connCycle = (connCycle + 1) % config.NumConnections
	}
}
