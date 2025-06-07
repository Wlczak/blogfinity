package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Wlczak/blogfinity/logger"
)

type Queue struct {
	mutex   *sync.Mutex
	queries []string
}

func NewQueue() *Queue {
	return &Queue{
		mutex:   new(sync.Mutex),
		queries: make([]string, 0),
	}
}

func (q *Queue) Push(query string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.queries = append(q.queries, query)
}

func (q *Queue) Pop() (string, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.queries) == 0 {
		return "", false
	} else {
		query := q.queries[0]
		q.queries = q.queries[1:]
		return query, true
	}
}

func HandleQueue(queryCh chan string) {
	var queue = NewQueue()
	go CheckQueue(queue)
	for {
		query := <-queryCh
		queue.Push(query)

		println("Received query: " + query)
	}
}

func CheckQueue(queue *Queue) {
	for {
		query, ok := queue.Pop()
		if ok {
			fmt.Println("Prompting AI with query: " + query)
			PromptAi(query)
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

type TitleResp struct {
	Title string `json:"title"`
}

func PromptAi(query string) {
	zap := logger.GetLogger()
	//deepseek-r1:8b
	requestJson := []byte(`{"model":"deepseek-r1:1.5b-qwen-distill-q4_K_M",
		"prompt":"i need you to generate an article title based on this search prompt: “` + query +
		`“, format it into a json format like so: {“title“:”<insert title here>”}, do not add any other text to the response, do not use any text formating, use only plaintext. Be very creative in your title creation. Under any circumstances do not!!! write any more than one title.","stream":false}`)

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
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	raw := string(body)

	fmt.Println(raw)

	var out OllamaResp
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		panic(err)
	}

	resp := strings.Trim(out.Response, "`")                    // remove backticks
	resp = strings.TrimSpace(strings.TrimPrefix(resp, "json")) // strip leading “json” label
	resp = strings.TrimSpace(resp)

	var titleOut TitleResp
	if err := json.Unmarshal([]byte(resp), &titleOut); err != nil {
		panic(err)
	}

	fmt.Println("Title:", titleOut.Title)
}
