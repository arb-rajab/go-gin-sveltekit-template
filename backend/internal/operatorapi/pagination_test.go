package operatorapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/example/starterapp/backend/internal/pagination"
)

func pageParamsRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/things", func(c *gin.Context) {
		p, ok := PageParams(c)
		if !ok {
			return
		}
		c.JSON(http.StatusOK, pagination.New([]string{"a", "b"}, p, 42))
	})
	return r
}

func TestPageParams_ValidRequest(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/things?page=2&per_page=10", nil)
	pageParamsRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	var body pagination.Page[string]
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Page != 2 || body.PerPage != 10 || body.Total != 42 || body.TotalPages != 5 {
		t.Fatalf("unexpected envelope: %+v", body)
	}
}

func TestPageParams_InvalidRequestUsesErrorEnvelope(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/things?per_page=1000", nil)
	pageParamsRouter().ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	var body errorEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "validation_error" || body.Error.Field == nil || *body.Error.Field != "per_page" {
		t.Fatalf("unexpected error: %+v", body.Error)
	}
}
