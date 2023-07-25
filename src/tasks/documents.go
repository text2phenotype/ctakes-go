package tasks

import (
	"text2phenotype.com/fdl/redis"
	"text2phenotype.com/fdl/utils/maps"
	"sync"
)

const DocumentsDB redis.DB = 0

type DocumentTask struct {
	maps.BaseDocument
	FailedTasks  []string            `json:"failed_tasks"`
	FailedChunks map[string][]string `json:"failed_chunks"`
}

type DocumentTaskCached struct {
	maps.BaseDocument
	DocInfo     map[string]interface{} `json:"document_info"`
	FailedTasks []string               `json:"failed_tasks"`
	JobID       string                 `json:"job_id"`
	WorkType    string                 `json:"work_type"`
}

type DocumentTasks struct {
	client redis.Client
}

func (tasks DocumentTasks) Get(redisKey string) (*DocumentTask, error) {
	var task DocumentTask
	err := tasks.client.GetPartialDocument(redisKey, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (tasks DocumentTasks) GetCached(redisKey string) (*DocumentTaskCached, error) {
	var task DocumentTaskCached
	err := tasks.client.GetPartialDocument(cachedPropertiesKey(redisKey), &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (tasks DocumentTasks) Update(redisKey string, updateFunc func(task *DocumentTask)) (err error) {
	releaseLock, err := tasks.client.Lock(redisKey)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = releaseLock()
			return
		}
		err = releaseLock()
	}()
	var task DocumentTask
	var cached DocumentTaskCached

	err = tasks.client.GetPartialDocument(redisKey, &task)
	if err != nil {
		return err
	}
	err = maps.ApplyUpdates(&task, updateFunc)
	if err != nil {
		return err
	}
	err = maps.CopyValues(&task, &cached)
	if err != nil {
		return err
	}
	errChan := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		errChan <- tasks.client.SaveDoc(redisKey, &task)
		wg.Done()
	}()
	go func() {
		errChan <- tasks.client.SaveDoc(cachedPropertiesKey(redisKey), &cached)
		wg.Done()
	}()
	wg.Wait()
	close(errChan)
	for err = range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
