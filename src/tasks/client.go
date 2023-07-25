package tasks

import (
	"text2phenotype.com/fdl/redis"
	"fmt"
)

type Client struct {
	Documents DocumentTasks
	Chunks    ChunkTasks
	Jobs      JobTasks
}

// NewClient is a preferred way for working with TaskInfos
func NewClient() (Client, error) {
	docRedisClient, err := redis.NewClient(DocumentsDB)
	if err != nil {
		return Client{}, err
	}
	jobsRedisClient, err := redis.NewClient(JobsDB)
	if err != nil {
		return Client{}, err
	}
	chunksRedisClient, err := redis.NewClient(ChunksDB)
	if err != nil {
		return Client{}, err
	}
	return Client{
		Documents: DocumentTasks{client: docRedisClient},
		Jobs:      JobTasks{client: jobsRedisClient},
		Chunks:    ChunkTasks{client: chunksRedisClient},
	}, nil
}

func (client *Client) Close() {
	_ = client.Chunks.client.Close()
	_ = client.Documents.client.Close()
	_ = client.Jobs.client.Close()
}
func cachedPropertiesKey(redisKey string) string {
	return fmt.Sprintf("%s-cached-properties", redisKey)
}
