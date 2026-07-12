package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCreateAppKeyRejectsFieldsOtherThanName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		body string
	}{
		{name: "group id override", body: `{"name":"ZeroBox","group_id":99}`},
		{name: "custom key", body: `{"name":"ZeroBox","custom_key":"attacker"}`},
		{name: "second json value", body: `{"name":"ZeroBox"}{"name":"second"}`},
		{name: "blank name", body: `{"name":"   "}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &AppResourceHandler{}
			router := gin.New()
			router.POST("/groups/:groupId/keys", func(c *gin.Context) {
				c.Set("app_auth_subject", servermiddleware.AppAuthSubject{UserID: 7})
				c.Next()
			}, handler.CreateKey)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/groups/3/keys", strings.NewReader(tt.body))
			request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(recorder, request)
			require.Equal(t, http.StatusBadRequest, recorder.Code)
			var body map[string]any
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
			require.Equal(t, false, body["success"])
		})
	}
}

func TestResolveAppGroupRatesUsesOverrideSemantics(t *testing.T) {
	userRate, effective := resolveAppGroupRates(1.2, 1.7, true)
	require.Equal(t, 1.7, userRate)
	require.Equal(t, 1.7, effective)

	userRate, effective = resolveAppGroupRates(1.2, 0, false)
	require.Equal(t, 1.0, userRate)
	require.Equal(t, 1.2, effective)
}
