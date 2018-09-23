package relay

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	graphql "github.com/PentoHQ/graphql-go"
)

func MarshalID(kind string, spec interface{}) graphql.ID {
	d, err := json.Marshal(spec)
	if err != nil {
		panic(fmt.Errorf("relay.MarshalID: %s", err))
	}
	return graphql.ID(base64.URLEncoding.EncodeToString(append([]byte(kind+":"), d...)))
}

func UnmarshalKind(id graphql.ID) string {
	s, err := base64.URLEncoding.DecodeString(string(id))
	if err != nil {
		return ""
	}
	i := strings.IndexByte(string(s), ':')
	if i == -1 {
		return ""
	}
	return string(s[:i])
}

func UnmarshalSpec(id graphql.ID, v interface{}) error {
	s, err := base64.URLEncoding.DecodeString(string(id))
	if err != nil {
		return err
	}
	i := strings.IndexByte(string(s), ':')
	if i == -1 {
		return errors.New("invalid graphql.ID")
	}
	return json.Unmarshal([]byte(s[i+1:]), v)
}

type Handler struct {
	Schema *graphql.Schema
}

type BatchedHandler struct {
	Schema *graphql.Schema
}

type RequestData struct {
	ID            string
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

type BatchedResponse struct {
	ID string `json:"id"`
	graphql.Response
}

func (h *BatchedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	queryBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var responses []*BatchedResponse
	var queries []*RequestData
	err = json.Unmarshal(queryBytes, &queries)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	responsesChan := make(chan *BatchedResponse, len(queries))
	for i := range queries {
		qr := queries[i]
		go func(query *RequestData, responsesChan chan *BatchedResponse) {
			gqlResponse := h.Schema.Exec(r.Context(), query.Query, query.OperationName, query.Variables)
			responsesChan <- &BatchedResponse{query.ID, *gqlResponse}
		}(qr, responsesChan)
	}

	for i := 0; i < len(queries); i++ {
		responses = append(responses, <-responsesChan)
	}
	close(responsesChan)

	responseJSON, err := json.Marshal(responses)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := h.Schema.Exec(r.Context(), params.Query, params.OperationName, params.Variables)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}
