package ai

import (
	"io"
	"net/http"
)

func IsServerOnline() bool {
	data, err := http.Get("http://nix:11434/")
	if err != nil {
		return false
	}
	resp, err := io.ReadAll(data.Body)
	if err != nil {
		return false
	}
	str := string(resp)

	return str == "Ollama is running"
}
