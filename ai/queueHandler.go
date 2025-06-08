package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Wlczak/blogfinity/database"
	"github.com/Wlczak/blogfinity/database/models"
	"github.com/Wlczak/blogfinity/logger"
)

type Queue struct {
	mutex   *sync.Mutex
	queries []AiQuery
}

type AiQuery struct {
	Query   string
	Article models.Article
	Type    string
	Model   string
}

func NewQueue() *Queue {
	return &Queue{
		mutex:   new(sync.Mutex),
		queries: make([]AiQuery, 0),
	}
}

func (q *Queue) Push(query AiQuery) {
	q.mutex.Lock()
	maxQueries := 5
	defer q.mutex.Unlock()
	if len(q.queries) <= maxQueries {
		fmt.Println("Added query: " + query.Query)
		q.queries = append(q.queries, query)
	}
}

func (q *Queue) Pop() (AiQuery, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.queries) == 0 {
		return AiQuery{}, false
	} else {
		query := q.queries[0]
		q.queries = q.queries[1:]
		return query, true
	}
}

func HandleQueue(queryCh chan AiQuery) {
	var queue = NewQueue()
	go CheckQueue(queue)
	for {
		query := <-queryCh
		queue.Push(query)
		fmt.Println("Received query: " + query.Query)
	}
}

func FilterModel(art AiQuery) string {
	var model string
	switch art.Model {
	case "deepseek-r1:8b":
		model = "deepseek-r1:8b"
		break
	default:
		model = "qwen:0.5b"
		break
	}
	return model
}

func CheckQueue(queue *Queue) {
	zap := logger.GetLogger()
	for {
		query, ok := queue.Pop()
		if ok {
			model := FilterModel(query)
			if query.Type == "title" {
				prompt1 := " i need you to generate an article title based on this search prompt: “"
				prompt2 := "“, the answer must be in the form of a non formatted string and must be completely plain text. It must also be searchable with fuzzy search. Meaning it has to be similar to the search prompt, thogh it doesnt have to have the same exact words every time. Please output only the title and nothing else since the output is not filtered and will end up directly on the website. Also be creative and make sure the title is around 5-15 words long. Do not put the title into quotes."

				response := PromptAi(prompt1+strings.Trim(query.Query, `"`)+prompt2, model)

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
					response := PromptAi(prompt1+strings.Trim(query.Query, `"`)+prompt2, model)

					article.Body = response.Text

					article.Update(db)
				}
			}
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

func PromptAi(query string, model string) PrompResult {
	zap := logger.GetLogger()
	// deepseek-r1:8b
	// deepseek-r1:1.5b-qwen-distill-q4_K_M
	// gemma3:1b
	// gemma3:4b
	// deepseek-coder:latest
	// qwen:0.5b
	// llama3.2:3b
	// llama3.1:8b
	// smollm:135m

	fmt.Println("Prompting AI with query: " + query)
	requestJson := []byte(`{"model":"` + model + `", "options": {"temperature": 0.6},
		"prompt":"` + query + `","stream":false}`)

	request, err := http.NewRequest("POST", "http://nix:11434/api/generate", bytes.NewBuffer(requestJson))
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

	body, err := io.ReadAll(response.Body)
	if err != nil {
		zap.Error(err.Error())
	}

	raw := string(body)

	//fmt.Println(raw)

	var out OllamaResp
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		zap.Error(err.Error())
	}

	resp := strings.Trim(out.Response, "`")                    // remove backticks
	resp = strings.TrimSpace(strings.TrimPrefix(resp, "json")) // strip leading “json” label
	resp = strings.TrimSpace(resp)

	regex := regexp.MustCompile(`(?s)<think>.*?</think>`)
	cleanResp := regex.ReplaceAll([]byte(resp), []byte{})

	bom := []byte{0xEF, 0xBB, 0xBF}
	cleanResp = bytes.TrimPrefix(cleanResp, bom)

	reTags := regexp.MustCompile(`(?s)<[^>]+>`)
	cleanResp = reTags.ReplaceAll(cleanResp, []byte{})

	// drop or replace any invalid UTF-8 sequences
	cleanResp = bytes.ToValidUTF8(cleanResp, nil)

	output := strings.Trim(string(cleanResp), `"`)

	output = strings.TrimSpace(output)

	fmt.Println("Cleaned response:", output)

	return PrompResult{Text: output, Model: out.Model}

}
