package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRequireAppScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		subject    *AppAuthSubject
		wantStatus int
		wantNext   bool
	}{
		{name: "missing subject", wantStatus: http.StatusUnauthorized},
		{name: "missing scope", subject: &AppAuthSubject{Scopes: map[string]struct{}{}}, wantStatus: http.StatusForbidden},
		{name: "allowed", subject: &AppAuthSubject{Scopes: map[string]struct{}{"keys:read": {}}}, wantStatus: http.StatusNoContent, wantNext: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/", func(c *gin.Context) {
				if tt.subject != nil {
					c.Set(contextKeyAppAuth, *tt.subject)
				}
				c.Next()
			}, RequireAppScope("keys:read"), func(c *gin.Context) { c.Status(http.StatusNoContent) })
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
			require.Equal(t, tt.wantStatus, recorder.Code)
			if !tt.wantNext {
				var body map[string]any
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
				require.Equal(t, false, body["success"])
			}
		})
	}
}
