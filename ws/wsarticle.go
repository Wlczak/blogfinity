package ws

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Wlczak/blogfinity/ai"
	"github.com/Wlczak/blogfinity/logger"
	"github.com/gorilla/websocket"
)

func HandleWsArticle(w http.ResponseWriter, r *http.Request, queue *ai.Queue) {
	zap := logger.GetLogger()
	articleId := r.PathValue("articleId")
	articleIdInt, err := strconv.Atoi(articleId)
	if err != nil {
		zap.Error(err.Error())
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	time.Sleep(500 * time.Millisecond)
	queue.AddConn(conn, articleIdInt)
}
