package ai

import (
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Wlczak/blogfinity/database"
	"github.com/Wlczak/blogfinity/database/models"
	"github.com/Wlczak/blogfinity/logger"
	"gorm.io/gorm"
)

func GetOllamaServer() (url string, success bool) {
	zap := logger.GetLogger()

	db, err := database.GetDB()
	if err != nil {
		zap.Error(err.Error())
	}

	if db == nil {
		return "", false
	}

	servers := os.Getenv("OLLAMA_SERVER_IPS")

	serverSlice := strings.Split(servers, ";")

	for _, server := range serverSlice {
		serverCache := models.GetServerCache(db, server, "11434")
		if serverCache.Online {
			return "http://" + serverCache.Host + ":" + serverCache.Port, true
		}
		if serverCache.LastChecked.Add(5 * time.Minute).Before(time.Now()) {
			// fmCt.Println("Updating server status")
			go UpdateServerStatus(db, &serverCache)
		}
	}

	return "", false
}

func UpdateServerStatus(db *gorm.DB, server *models.Server) {
	zap := logger.GetLogger()

	data, err := http.Get("http://" + server.Host + ":" + server.Port + "/")
	if err != nil {
		zap.Error(err.Error())
		server.Online = false
		server.LastChecked = time.Now()
		server.Update(db)
		return
	}
	resp, err := io.ReadAll(data.Body)
	if err != nil {
		zap.Error(err.Error())
		server.Online = false
		server.LastChecked = time.Now()
		server.Update(db)
		return
	}
	str := string(resp)
	server.Online = (str == "Ollama is running")
	server.LastChecked = time.Now()
	server.Update(db)
}
