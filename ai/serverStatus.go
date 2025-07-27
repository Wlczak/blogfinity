package ai

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Wlczak/blogfinity/database"
	"github.com/Wlczak/blogfinity/database/models"
	"github.com/Wlczak/blogfinity/logger"
	"gorm.io/gorm"
)

func IsServerOnline() bool {
	zap := logger.GetLogger()

	db, err := database.GetDB()
	if err != nil {
		zap.Error(err.Error())
	}

	server := models.GetServerCache(db, "nix", "11434")
	if server.LastChecked.Add(5 * time.Second).Before(time.Now()) {
		fmt.Println("Updating server status")
		go UpdateServerStatus(db, &server)
	}
	return server.Online
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
