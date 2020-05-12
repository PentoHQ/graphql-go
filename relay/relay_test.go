package relay_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PentoHQ/graphql-go"
	"github.com/PentoHQ/graphql-go/example/starwars"
	"github.com/PentoHQ/graphql-go/relay"
)

var starwarsSchema = graphql.MustParseSchema(starwars.Schema, &starwars.Resolver{})

func TestServeHTTP(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/some/path/here", strings.NewReader(`{"query":"{ hero { name } }", "operationName":"", "variables": null}`))
	h := relay.Handler{Schema: starwarsSchema}

	h.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Fatalf("Expected status code 200, got %d.", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("Invalid content-type. Expected [application/json], but instead got [%s]", contentType)
	}

	expectedResponse := `{"data":{"hero":{"name":"R2-D2"}}}`
	actualResponse := w.Body.String()
	if expectedResponse != actualResponse {
		t.Fatalf("Invalid response. Expected [%s], but instead got [%s]", expectedResponse, actualResponse)
	}
}

func TestBatchedServeHTTP(t *testing.T) {
	t.Run("no parallel queries limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/some/path/here", strings.NewReader(`[{"id":"qID", "query":"{ hero { name } }", "operationName":"", "variables": null},
	{"id":"qID", "query":"{ hero { name } }", "operationName":"", "variables": null}]`))
		h := relay.BatchedHandler{Schema: starwarsSchema}

		h.ServeHTTP(w, r)

		if w.Code != 200 {
			t.Fatalf("Expected status code 200, got %d.", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Fatalf("Invalid content-type. Expected [application/json], but instead got [%s]", contentType)
		}

		expectedResponse := `[{"id":"qID","data":{"hero":{"name":"R2-D2"}}},{"id":"qID","data":{"hero":{"name":"R2-D2"}}}]`
		actualResponse := w.Body.String()
		if expectedResponse != actualResponse {
			t.Fatalf("Invalid response. Expected [%s], but instead got [%s]", expectedResponse, actualResponse)
		}
	})
	t.Run("with parallel queries limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/some/path/here", strings.NewReader(`[{"id":"qID", "query":"{ hero { name } }", "operationName":"", "variables": null},
	{"id":"qID", "query":"{ hero { name } }", "operationName":"", "variables": null}]`))
		h := relay.BatchedHandler{Schema: starwarsSchema, MaxParallelQueries: 1}

		h.ServeHTTP(w, r)

		if w.Code != 200 {
			t.Fatalf("Expected status code 200, got %d.", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Fatalf("Invalid content-type. Expected [application/json], but instead got [%s]", contentType)
		}

		expectedResponse := `[{"id":"qID","data":{"hero":{"name":"R2-D2"}}},{"id":"qID","data":{"hero":{"name":"R2-D2"}}}]`
		actualResponse := w.Body.String()
		if expectedResponse != actualResponse {
			t.Fatalf("Invalid response. Expected [%s], but instead got [%s]", expectedResponse, actualResponse)
		}
	})
}
