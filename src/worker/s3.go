package worker

import (
	"text2phenotype.com/fdl/s3client"
)

type s3Transactions interface {
	saveResultsFile(task *Task, result string) error
	getProcessedData(task *Task) ([]byte, error)
	close()
}

type s3ClientWrapper struct {
	s3Client *s3client.Client
}

func (wrapper *s3ClientWrapper) close() {
	wrapper.s3Client.Close()
}

func (wrapper *s3ClientWrapper) saveResultsFile(task *Task, result string) error {
	resultsFileKey := getResultsFileKey(task)
	_, err := wrapper.s3Client.Upload(result, resultsFileKey)
	return err
}

func (wrapper *s3ClientWrapper) getProcessedData(task *Task) ([]byte, error) {
	return wrapper.s3Client.Download(task.chunkTask.TextFileKey)
}
