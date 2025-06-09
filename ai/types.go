package ai

import (
	"sync"

	"github.com/Wlczak/blogfinity/database/models"
)

const (
	MaxAiQueueSize   = 15
	MaxSearchResults = 25
	SearchDistance   = 55
)

func GetModels() []string {
	return []string{"qwen:0.5b", "deepseek-r1:1.5b-qwen-distill-q4_K_M", "gemma3:1b", "gemma3:4b", "deepseek-coder:latest", "llama3.2:3b", "llama3.1:8b", "smollm:135m"}
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
