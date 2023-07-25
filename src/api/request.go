package api

import (
	"text2phenotype.com/fdl/pipeline"
	"io/ioutil"
	"net/http"
)

type Request struct {
	Pipeline pipeline.Pipeline
}

func (req *Request) ProcessData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	logger := makeRequestLogger(r)

	if r.Method != "POST" {
		logger.Err(nil).Int("status", http.StatusMethodNotAllowed).Msg("Only 'POST' method is allowed here")
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Err(err).Int("status", http.StatusBadRequest).Msg("Could not read request body")
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	request := pipeline.Request{
		Tid:  "test_api",
		Text: string(msg),
	}
	logger.Info().Str("tid", request.Tid).Msg("Starting pipeline for request from API")
	resp := <-req.Pipeline(request)
	_, _ = w.Write([]byte(resp))
	logger.Info().Int("status", http.StatusOK).Msg("Finished processing request")
}
