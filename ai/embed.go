package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Wlczak/blogfinity/database/models"
	"github.com/Wlczak/blogfinity/logger"
	ollama "github.com/ollama/ollama/api"
	"github.com/qdrant/go-client/qdrant"
)

const qdrantEmbeddingVectorName = "embedding"

func EmbedQueryAndUpsert(id int, collectionName string, query string) {
	zap := logger.GetLogger()
	embedings, err := PromptEmbedAi(query)

	if err != nil {
		zap.Error(err.Error())
		panic(err)
	}

	for _, embed := range embedings {
		UpsertEmbedding(id, collectionName, embed)
	}
}

func UpsertEmbedding(id int, collectionName string, embedings []float32) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host:                   "qdrant",
		Port:                   6334,
		SkipCompatibilityCheck: true,
		UseTLS:                 false,
	})
	if err != nil {
		zap := logger.GetLogger()
		zap.Error(err.Error())
		panic(err)
	}

	vectors, err := buildQdrantVectors(client, collectionName, embedings)
	if err != nil {
		zap := logger.GetLogger()
		zap.Error(err.Error())
		panic(err)
	}

	_, err = client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points: []*qdrant.PointStruct{
			{
				Id:      qdrant.NewIDNum(uint64(id)),
				Vectors: vectors,
				Payload: nil,
			},
		},
	})
	if err != nil {
		zap := logger.GetLogger()
		zap.Error(err.Error())
		panic(err)
	}
}

func buildQdrantVectors(client *qdrant.Client, collectionName string, embedding []float32) (*qdrant.Vectors, error) {
	info, err := client.GetCollectionInfo(context.Background(), collectionName)
	if err != nil {
		createErr := client.CreateCollection(context.Background(), &qdrant.CreateCollection{
			CollectionName: collectionName,
			VectorsConfig: qdrant.NewVectorsConfigMap(map[string]*qdrant.VectorParams{
				qdrantEmbeddingVectorName: {
					Size:     uint64(len(embedding)),
					Distance: qdrant.Distance_Cosine,
				},
			}),
		})
		if createErr != nil {
			return nil, createErr
		}

		return qdrant.NewVectorsMap(map[string]*qdrant.Vector{
			qdrantEmbeddingVectorName: qdrant.NewVectorDense(embedding),
		}), nil
	}

	vectorsConfig := info.GetConfig().GetParams().GetVectorsConfig()
	if vectorsConfig == nil {
		return qdrant.NewVectorsDense(embedding), nil
	}

	if params := vectorsConfig.GetParams(); params != nil {
		if params.GetSize() != 0 && params.GetSize() != uint64(len(embedding)) {
			return nil, fmt.Errorf(
				"qdrant collection %q expects vectors of size %d, got %d",
				collectionName,
				params.GetSize(),
				len(embedding),
			)
		}
		return qdrant.NewVectorsDense(embedding), nil
	}

	if paramsMap := vectorsConfig.GetParamsMap(); paramsMap != nil && len(paramsMap.GetMap()) > 0 {
		if _, ok := paramsMap.GetMap()[qdrantEmbeddingVectorName]; ok {
			return qdrant.NewVectorsMap(map[string]*qdrant.Vector{
				qdrantEmbeddingVectorName: qdrant.NewVectorDense(embedding),
			}), nil
		}

		for vectorName := range paramsMap.GetMap() {
			return qdrant.NewVectorsMap(map[string]*qdrant.Vector{
				vectorName: qdrant.NewVectorDense(embedding),
			}), nil
		}
	}

	return qdrant.NewVectorsDense(embedding), nil
}

func PromptEmbedAi(query string) ([][]float32, error) {
	zap := logger.GetLogger()

	embedRequest := &ollama.EmbedRequest{
		Model: "embeddinggemma",
		Input: query,
	}

	serverUrl, serverOnline := GetOllamaServer()
	if !serverOnline {
		zap.Error("No Ollama server available")
		return nil, errors.New("no ollama server available")
	}
	client := ollama.NewClient(&url.URL{Scheme: "http", Host: serverUrl + ":11434"}, http.DefaultClient)
	embedResult, err := client.Embed(context.Background(), embedRequest)
	if err != nil {
		zap.Error(err.Error())
		return nil, err
	}

	return embedResult.Embeddings, err
}

func EmbedArticle(article models.Article) {
	EmbedQueryAndUpsert(article.ID, "articles", ""+article.Title+". "+article.Body)
}
