package ai

import (
	"bytes"
	"fmt"
	"net/http"
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

func PromptAi(query string) {
	zap := logger.GetLogger()
	body := []byte(`{
		{"model":"deepseek-r1:8b",
		"prompt":"i need you to generate an article title based on this search prompt: “` + query +
		`“ and format it into a json format like so: {“title“:”<insert title here>”}","stream":false}
	}`)

	request, err := http.NewRequest("POST", "http://nix:11434/api/generate", bytes.NewBuffer(body))

	if err != nil {
		zap.Error(err.Error())
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		zap.Error(err.Error())
	}
	fmt.Println(response.Body)

	err = response.Body.Close()

	if err != nil {
		zap.Error(err.Error())
	}
}
