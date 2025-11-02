package ai

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Wlczak/blogfinity/database/models"
	"github.com/Wlczak/blogfinity/logger"
	"github.com/gorilla/websocket"
)

const (
	MaxAiQueueSize   = 15
	MaxSearchResults = 25
	SearchDistance   = 55
)

func GetModels() []string {
	zap := logger.GetLogger()
	if IsServerOnline() {
		resp, err := http.Get("http://ollama-server:11434/api/tags")
		if err != nil {
			zap.Error(err.Error())
		}
		var body []byte
		body, err = io.ReadAll(resp.Body)

		if err != nil {
			zap.Error(err.Error())
		}
		modelResp := ModelResponse{}
		err = json.Unmarshal(body, &modelResp)
		if err != nil {
			zap.Error(err.Error())
		}
		// fmt.Println(modelResp)
		var modelList []string
		for _, v := range modelResp.Models {
			modelList = append(modelList, v.Model)
		}
		err = resp.Body.Close()
		if err != nil {
			zap.Error(err.Error())
		}
		return modelList

	}
	return []string{}
}

type ModelResponse struct {
	Models []ModelItem `json:"models"`
}

type ModelItem struct {
	Model string `json:"model"`
}

type Queue struct {
	mutex   *sync.Mutex
	queries []*AiQuery
}

func (q *Queue) Push(query *AiQuery) {
	q.mutex.Lock()

	defer q.mutex.Unlock()
	if len(q.queries) < MaxAiQueueSize {
		// fmt.Println("Added query: " + query.Query)
		q.queries = append(q.queries, query)
	}
}

func (q *Queue) Pop() (*AiQuery, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.queries) == 0 {
		return nil, false
	} else {
		query := q.queries[0]
		q.queries = q.queries[1:]
		return query, true
	}
}

func (q *Queue) Copy() []*AiQuery {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	dst := make([]*AiQuery, len(q.queries))
	copy(dst, q.queries)
	return dst
}

func (q *Queue) AddConn(conn *websocket.Conn, articleId int) {
	zap := logger.GetLogger()
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for _, query := range q.queries {
		if query.Article.ID == articleId {
			query.EventConns = append(query.EventConns, conn)
			return
		}
	}
	go func(conn *websocket.Conn) {
		time.Sleep(1 * time.Second)
		err := conn.Close()
		zap.Debug("closed connection in types")
		if err != nil {
			zap.Error(err.Error())
		}
	}(conn)
}

type AiQuery struct {
	Query      string
	Article    models.Article
	Type       string
	Model      string
	RequestId  string
	EventConns []*websocket.Conn
}
