package ai

import (
	"sync"
	"time"
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
			PromptAi(query)
		} else {
			time.Sleep(1 * time.Second)
		}

	}
}

func PromptAi(query string) {
	println("Prompting AI with query: " + query)
}
