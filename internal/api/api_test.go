package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/FernandaSpineli/calculator/internal/store"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func do(t *testing.T, r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decode[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("decoding %q: %v", w.Body.String(), err)
	}
	return v
}

type errorBody struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func TestCalculationLifecycle(t *testing.T) {
	r := NewRouter(store.NewMemory())

	w := do(t, r, http.MethodPost, "/api/v1/calculations", `{"operation":"sin","operands":[30],"angle_unit":"deg"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: status %d, body %s", w.Code, w.Body)
	}
	if loc := w.Header().Get("Location"); loc != "/api/v1/calculations/1" {
		t.Errorf("Location = %q", loc)
	}
	created := decode[store.Calculation](t, w)
	if created.ID != 1 || created.AngleUnit != "deg" || created.Result < 0.4999999 || created.Result > 0.5000001 {
		t.Errorf("created = %+v", created)
	}

	w = do(t, r, http.MethodGet, "/api/v1/calculations/1", "")
	if w.Code != http.StatusOK || decode[store.Calculation](t, w).Operation != "sin" {
		t.Fatalf("get: status %d, body %s", w.Code, w.Body)
	}

	w = do(t, r, http.MethodPost, "/api/v1/calculations", `{"operation":"add","operands":[2,3]}`)
	if got := decode[store.Calculation](t, w); got.Result != 5 || got.AngleUnit != "" {
		t.Errorf("add = %+v", got)
	}

	w = do(t, r, http.MethodGet, "/api/v1/calculations", "")
	list := decode[struct{ Calculations []store.Calculation }](t, w)
	if len(list.Calculations) != 2 || list.Calculations[0].ID != 1 {
		t.Errorf("list = %+v", list)
	}

	if w = do(t, r, http.MethodDelete, "/api/v1/calculations/1", ""); w.Code != http.StatusNoContent {
		t.Errorf("delete: status %d", w.Code)
	}
	if w = do(t, r, http.MethodGet, "/api/v1/calculations/1", ""); w.Code != http.StatusNotFound {
		t.Errorf("get after delete: status %d", w.Code)
	}
	if w = do(t, r, http.MethodDelete, "/api/v1/calculations/1", ""); w.Code != http.StatusNotFound {
		t.Errorf("second delete: status %d", w.Code)
	}
}

func TestCreateCalculationErrors(t *testing.T) {
	r := NewRouter(store.NewMemory())
	tests := []struct {
		body   string
		status int
		code   string
	}{
		{`{`, http.StatusBadRequest, "invalid_body"},
		{`{"operands":[1,2]}`, http.StatusBadRequest, "invalid_body"},
		{`{"operation":"divide","operands":[1,0]}`, http.StatusUnprocessableEntity, "division_by_zero"},
		{`{"operation":"sqrt","operands":[-4]}`, http.StatusUnprocessableEntity, "domain_error"},
		{`{"operation":"add","operands":[1]}`, http.StatusUnprocessableEntity, "wrong_operand_count"},
		{`{"operation":"teleport","operands":[1]}`, http.StatusUnprocessableEntity, "unknown_operation"},
		{`{"operation":"exp","operands":[1000]}`, http.StatusUnprocessableEntity, "overflow"},
		{`{"operation":"sin","operands":[1],"angle_unit":"grad"}`, http.StatusUnprocessableEntity, "invalid_angle_unit"},
	}
	for _, tt := range tests {
		w := do(t, r, http.MethodPost, "/api/v1/calculations", tt.body)
		if w.Code != tt.status || decode[errorBody](t, w).Error.Code != tt.code {
			t.Errorf("%s: status %d body %s, want %d %s", tt.body, w.Code, w.Body, tt.status, tt.code)
		}
	}
	if list := decode[struct{ Calculations []store.Calculation }](t, do(t, r, http.MethodGet, "/api/v1/calculations", "")); len(list.Calculations) != 0 {
		t.Errorf("failed calculations were stored: %+v", list.Calculations)
	}
}

func TestOperations(t *testing.T) {
	r := NewRouter(store.NewMemory())

	w := do(t, r, http.MethodGet, "/api/v1/operations", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"name":"factorial"`) {
		t.Errorf("list: status %d, body %s", w.Code, w.Body)
	}
	if w = do(t, r, http.MethodGet, "/api/v1/operations/tan", ""); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"uses_angle_unit":true`) {
		t.Errorf("get: status %d, body %s", w.Code, w.Body)
	}
	if w = do(t, r, http.MethodGet, "/api/v1/operations/nope", ""); w.Code != http.StatusNotFound {
		t.Errorf("get unknown: status %d", w.Code)
	}
}

func TestInvalidID(t *testing.T) {
	r := NewRouter(store.NewMemory())
	for _, id := range []string{"abc", "0", "-3"} {
		if w := do(t, r, http.MethodGet, "/api/v1/calculations/"+id, ""); w.Code != http.StatusBadRequest {
			t.Errorf("id %q: status %d", id, w.Code)
		}
	}
}
