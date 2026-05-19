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

	for _, serverUrl := range serverSlice {
		serverUrl = strings.TrimSpace(serverUrl)
		if serverUrl == "" {
			continue
		}
		serverCache := models.GetServerCache(db, serverUrl, "11434")
		if serverCache.Online {
			return serverCache.Host, true
		}
		if UpdateServerStatus(db, &serverCache) {
			return serverCache.Host, true
		}
	}

	return "", false
}

func UpdateServerStatus(db *gorm.DB, server *models.Server) bool {
	zap := logger.GetLogger()

	client := &http.Client{Timeout: 5 * time.Second}
	data, err := client.Get("http://" + server.Host + ":" + server.Port + "/api/tags")
	if err != nil {
		zap.Warn(err.Error())
		server.Online = false
		server.LastChecked = time.Now()
		server.Update(db)
		return false
	}
	defer func() {
		err := data.Body.Close()
		if err != nil {
			zap.Warn(err.Error())
		}
	}()

	resp, err := io.ReadAll(data.Body)
	if err != nil {
		zap.Error(err.Error())
		server.Online = false
		server.LastChecked = time.Now()
		server.Update(db)
		return false
	}

	server.Online = data.StatusCode == http.StatusOK && len(resp) > 0
	server.LastChecked = time.Now()
	server.Update(db)
	return server.Online
}
