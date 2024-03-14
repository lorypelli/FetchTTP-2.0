package main

import (
	"bytes"
	"context"
	j "encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type Header struct {
	Enabled bool
	Name    string
	Value   string
}

type Query struct {
	Enabled bool
	Name    string
	Value   string
}

type HTTPResponse struct {
	URL    string
	Status string
	Header http.Header
	Body   string
}

type WSResponse struct {
	Ws      *websocket.Conn
	Status  string
	Header  http.Header
	Message string
}

func (a *App) HTTP(method string, url string, headers []Header, query []Query, body string) HTTPResponse {
	data := []byte(body)
	for i := 0; i < len(query); i++ {
		if query[i].Enabled && strings.TrimSpace(query[i].Name) != "" && strings.TrimSpace(query[i].Value) != "" {
			if strings.Contains(url, "?") {
				url += fmt.Sprintf("&%s=%s", query[i].Name, query[i].Value)
			} else {
				url += fmt.Sprintf("?%s=%s", query[i].Name, query[i].Value)
			}
		}
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return HTTPResponse{
			url, "", http.Header{}, "",
		}
	}
	for i := 0; i < len(headers); i++ {
		regexp, _ := regexp.Compile(`^[A-Za-z\d[\]{}()<>\/@?=:";,-]*$`)
		if headers[i].Enabled && strings.TrimSpace(headers[i].Name) != "" && regexp.MatchString(headers[i].Name) && strings.TrimSpace(headers[i].Value) != "" {
			req.Header.Add(headers[i].Name, headers[i].Value)
		}
	}
	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return HTTPResponse{
			url, "", http.Header{}, "",
		}
	}
	var resBody []byte
	if strings.Contains(res.Header.Get("Content-Type"), "application/json") {
		var jsonBody interface{}
		bytes, _ := io.ReadAll(res.Body)
		j.Unmarshal(bytes, &jsonBody)
		resBody, _ = j.MarshalIndent(jsonBody, "", "\t")
	} else {
		bytes, _ := io.ReadAll(res.Body)
		resBody = bytes
	}
	return HTTPResponse{
		url, res.Status, res.Header, string(resBody),
	}
}

var currentConnection *websocket.Conn
var currentResponse *http.Response
var currentError error

func (a *App) WS(url string, headers []Header, query []Query, connected bool) {
	for i := 0; i < len(query); i++ {
		if query[i].Enabled && strings.TrimSpace(query[i].Name) != "" && strings.TrimSpace(query[i].Value) != "" {
			if strings.Contains(url, "?") {
				url += fmt.Sprintf("&%s=%s", query[i].Name, query[i].Value)
			} else {
				url += fmt.Sprintf("?%s=%s", query[i].Name, query[i].Value)
			}
		}
	}
	header := http.Header{}
	for i := 0; i < len(headers); i++ {
		regexp, _ := regexp.Compile(`^[A-Za-z\d[\]{}()<>\/@?=:";,-]*$`)
		if headers[i].Enabled && strings.TrimSpace(headers[i].Name) != "" && regexp.MatchString(headers[i].Name) && strings.TrimSpace(headers[i].Value) != "" {
			header.Add(headers[i].Name, headers[i].Value)
		}
	}
	if connected {
		ws, res, err := websocket.DefaultDialer.Dial(url, header)
		currentConnection = ws
		currentResponse = res
		currentError = err
	}
	if currentError == nil {
		go Connect(currentResponse, currentConnection, connected, a)
	}
}

func Connect(res *http.Response, ws *websocket.Conn, connected bool, a *App) {
	runtime.EventsOn(a.ctx, "connected", func(data ...interface{}) {
		connected, _ = strconv.ParseBool(fmt.Sprint(data[0]))
	})
	for {
		if connected {
			_, msg, err := ws.ReadMessage()
			if err == nil {
				runtime.EventsEmit(a.ctx, "websocket", WSResponse{
					ws, res.Status, res.Header, string(msg),
				})
			}
		} else {
			ws.Close()
			return
		}
		time.Sleep(1 * time.Second)
	}
}
