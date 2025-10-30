package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/Wlczak/blogfinity/database/models"
	"github.com/Wlczak/blogfinity/logger"
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
		defer resp.Body.Close()
		if err != nil {
			zap.Error(err.Error())
		}
		modelResp := ModelResponse{}
		json.Unmarshal(body, &modelResp)
		fmt.Println(modelResp)
		var modelList []string
		for _, v := range modelResp.Models {
			modelList = append(modelList, v.Model)
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
	queries []AiQuery
}

type AiQuery struct {
	Query   string
	Article models.Article
	Type    string
	Model   string
}
