package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Wlczak/blogfinity/database"
	"github.com/Wlczak/blogfinity/database/models"
	"github.com/Wlczak/blogfinity/logger"
	"github.com/gorilla/websocket"
)

func NewQueue() *Queue {
	return &Queue{
		mutex:   new(sync.Mutex),
		queries: make([]*AiQuery, 0),
	}
}

func HandleQueue(queryCh chan *AiQuery, queue *Queue) {
	go CheckQueue(queue)
	for {
		query := <-queryCh
		queue.Push(query)
		// fmt.Println("Received query: " + query.Query)
	}
}

func FilterModel(art *AiQuery) (string, bool) {
	models := GetModels()
	if slices.Contains(models, art.Model) {
		return art.Model, true
	} else {
		return "", false
	}
}

func CheckQueue(queue *Queue) {
	zap := logger.GetLogger()
	for {
		query, ok := queue.Peek()
		if ok {
			model, ok := FilterModel(query)
			if !ok {
				queue.Pop()
				continue
			}
			if query.Type == "title" {
				prompt1 := " i need you to generate an article title based on this search prompt: “"
				prompt2 := "“, the answer must be in the form of a non formatted string and must be completely plain text. It must also be searchable with fuzzy search. Meaning it has to be similar to the search prompt, thogh it doesnt have to have the same exact words every time. Please output only the title and nothing else since the output is not filtered and will end up directly on the website. Also be creative and make sure the title is around 5-15 words long. Do not put the title into quotes."

				serverUrl, serverStatus := GetOllamaServer()

				if !serverStatus {
					zap.Error("No Ollama server available")
					queue.Pop()
					continue
				}

				response := PromptAi(serverUrl, prompt1+strings.Trim(query.Query, `"`)+prompt2, model, query.EventConns)

				article := models.Article{Title: response.Text, Body: "", Author: response.Model}

				db, err := database.GetDB()

				if err != nil {
					zap.Error(err.Error())
				}

				article.Create(db)

			}
			if query.Type == "body" {

				prompt1 := " i need you to generate an article body based on this article title: “"
				prompt2 := "“, the answer must be in the form of a non formatted string and must be completely plain text. Please output only the body and nothing else since the output is not filtered and will end up directly on the website. Also be creative and make sure the body is around 1-3 paragraphs long. Do not put the body into quotes."

				db, err := database.GetDB()

				if err != nil {
					zap.Error(err.Error())
				}

				article := query.Article
				if !article.HasBody(db) {
					time.Sleep(1 * time.Second)
					for _, conn := range query.EventConns {

						err = conn.WriteJSON(ArticleWebsocketMsg{
							Type: "status",
							Data: "loading",
						})

						if err != nil {
							zap.Error(err.Error())
						}
					}
					serverUrl, serverStatus := GetOllamaServer()

					if !serverStatus {
						for _, conn := range query.EventConns {
							err = conn.WriteJSON(ArticleWebsocketMsg{
								Type: "status",
								Data: "no server",
							})
							if err != nil {
								zap.Error(err.Error())
							}
						}
						queue.Pop()
						continue
					}

					response := PromptAi(serverUrl, prompt1+strings.Trim(query.Query, `"`)+prompt2, model, query.EventConns)

					article.Body = response.Text
					article.Author = response.Model

					for _, conn := range query.EventConns {
						err = conn.WriteJSON(article)
						if err != nil {
							zap.Error(err.Error())
						}
					}

					article.Update(db)
				}
			}
			for _, conn := range query.EventConns {
				go func(conn *websocket.Conn) {
					time.Sleep(1 * time.Second)
					err := conn.Close()
					zap.Debug("closed connection in queue handler")
					if err != nil {
						zap.Error(err.Error())
					}
				}(conn)
			}
			queue.Pop()
		} else {
			time.Sleep(1 * time.Second)
		}

	}
}

type OllamaResp struct {
	Model      string `json:"model"`
	CreatedAt  string `json:"created_at"`
	Response   string `json:"response"`
	Done       bool   `json:"done"`
	DoneReason string `json:"done_reason"`
	// other fields omitted
}

type PrompResult struct {
	Text  string
	Model string
}

func PromptAi(serverUrl string, query string, model string, eventConns []*websocket.Conn) PrompResult {
	zap := logger.GetLogger()
	ctx := context.Background()
	// deepseek-r1:8b
	// deepseek-r1:1.5b-qwen-distill-q4_K_M
	// gemma3:1b
	// gemma3:4b
	// deepseek-coder:latest
	// qwen:0.5b
	// llama3.2:3b
	// llama3.1:8b
	// smollm:135m

	serverUrl, serverOnline := GetOllamaServer()
	if !serverOnline {
		zap.Error("No Ollama server available")
		return PrompResult{Text: "error", Model: model}
	}
	fmt.Println("Prompting AI with query: " + query)
	requestJson := []byte(`{"model":"` + model + `", "options": {"temperature": 0.6},
		"prompt":"` + query + `","stream":true}`)

	request, err := http.NewRequestWithContext(ctx, "POST", serverUrl+"/api/generate", bytes.NewBuffer(requestJson))
	request.Header.Set("Content-Type", "application/json")

	if err != nil {
		zap.Error(err.Error())
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		zap.Error(err.Error())
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			zap.Error(err.Error())
		}
	}()

	if response.StatusCode != http.StatusOK {
		zap.Error("non-200 response from ollama server")
		return PrompResult{Text: "error", Model: model}
	}

	scanner := bufio.NewScanner(response.Body)
	// increase buffer if lines may be large
	const maxSize = 1024 * 1024 // 1MB (tune as needed)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxSize)

	var scannedString string
	for scanner.Scan() {
		line := scanner.Bytes()
		var it GenerationResponse
		if err := json.Unmarshal(line, &it); err != nil {
			// handle parse error (maybe log + continue or abort)
			zap.Error(err.Error())
		}

		scannedString += string(it.Response)
		for _, conn := range eventConns {
			err = conn.WriteJSON(ArticleWebsocketMsg{
				Type: "generation",
				Data: it.Response,
			})
			if err != nil {
				zap.Error(err.Error())
			}
		}
	}

	if err := scanner.Err(); err != nil {
		zap.Error(err.Error())
	}

	// raw := string(body)

	// fmt.Println(raw)

	// var out OllamaResp
	// if err := json.Unmarshal([]byte(raw), &out); err != nil {
	// 	zap.Error(err.Error())
	// }

	// resp := strings.Trim(out.Response, "`")                    // remove backticks
	// resp = strings.TrimSpace(strings.TrimPrefix(resp, "json")) // strip leading “json” label
	// resp = strings.TrimSpace(resp)

	// regex := regexp.MustCompile(`(?s)<think>.*?</think>`)
	// cleanResp := regex.ReplaceAll([]byte(resp), []byte{})

	// bom := []byte{0xEF, 0xBB, 0xBF}
	// cleanResp = bytes.TrimPrefix(cleanResp, bom)

	// reTags := regexp.MustCompile(`(?s)<[^>]+>`)
	// cleanResp = reTags.ReplaceAll(cleanResp, []byte{})

	// // drop or replace any invalid UTF-8 sequences
	// cleanResp = bytes.ToValidUTF8(cleanResp, nil)

	// output := strings.Trim(string(cleanResp), `"`)

	// output = strings.TrimSpace(output)

	output := scannedString
	fmt.Println("Cleaned response:", output)

	return PrompResult{Text: output, Model: model}
}
