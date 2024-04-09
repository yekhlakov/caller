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

func (conn *Connection) AddHeaders(headers []string) {
	for _, v := range headers {
		nameValue := strings.Split(v, ":")
		if len(nameValue) != 2 {
			continue
		}
		conn.AddHeader(strings.TrimSpace(nameValue[0]), strings.TrimSpace(nameValue[1]))
	}
}

// TemplateFile is the type to be extracted from a template file
type TemplateFile map[string]json.RawMessage

// IdListFile is the type to be extracted from an id list file
type IdListFile map[string][]string

// The storage for id lists
var idLists IdListFile = map[string][]string{}

type MessageTemplate struct {
	Probability float64
	Text        string
}

var messageTemplates []MessageTemplate

// GetRandomMessageTemplate Pick a random template from template list
func GetRandomMessageTemplate() string {
	randomValue := rand.Float64()

	for i := range messageTemplates {
		if randomValue < messageTemplates[i].Probability {
			return messageTemplates[i].Text
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

	for key, list := range idLists {
		if len(list) > 0 {
			msg = strings.ReplaceAll(msg, "\"##"+key+"##\"", list[rand.Intn(len(list))])
			msg = strings.ReplaceAll(msg, "#"+key+"#", list[rand.Intn(len(list))])
		}
	}

	return []byte(msg)
}

func LoadMessageTemplateFile(fileName string) {
	var t TemplateFile
	requestFile, _ := ioutil.ReadFile(fileName)
	_ = json.Unmarshal(requestFile, &t)

	totalMessageProbability := 0.0
	for p, v := range t {
		f, _ := strconv.ParseFloat(p, 64)
		totalMessageProbability = totalMessageProbability + f
		messageTemplates = append(messageTemplates, MessageTemplate{f, string(v)})
	}

	// Normalize probabilities
	for i := range messageTemplates {
		messageTemplates[i].Probability = messageTemplates[i].Probability / totalMessageProbability
	}
}

func LoadIdListFile(fileName string) {
	idListFile, _ := ioutil.ReadFile(fileName)

	_ = json.Unmarshal(idListFile, &idLists)

	for k := range idLists {
		fmt.Println("key:", k, "count:", len(idLists[k]))
	}
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
	headers := append(strings.Split(string(headerFile), "\n"), "Content-Type: application/json")

	// Create the connections
	connections := make([]*Connection, numConns)
	for i := 0; i < numConns; i++ {
		connections[i] = NewConnection(args[2])
		connections[i].AddHeaders(headers)
	}

	// Read the request templates
	LoadMessageTemplateFile(args[4])

	// Read the id-lists
	LoadIdListFile(args[5])

	// Perform the requests
	connCycle := 0
	for {
		time.Sleep(delay)
		// Each request is done in its own goroutine
		go func() { connections[connCycle].Call(GetRandomMessage()) }()
		// The connections are cycled
		connCycle = (connCycle + 1) % numConns
	}
}
