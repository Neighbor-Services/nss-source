package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestResponseHelpers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		Success(c, http.StatusOK, "Operation successful", map[string]string{"foo": "bar"})

		assert.Equal(t, http.StatusOK, w.Code)
		var res APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.True(t, res.Success)
		assert.Equal(t, "Operation successful", res.Message)
	})

	t.Run("Created", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		Created(c, "Resource created", 123)

		assert.Equal(t, http.StatusCreated, w.Code)
		var res APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.True(t, res.Success)
	})

	t.Run("BadRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		BadRequest(c, "Invalid input", "field 'email' is required")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var res APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.False(t, res.Success)
		assert.Equal(t, "Invalid input", res.Message)
		assert.Equal(t, "field 'email' is required", res.Errors)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		Unauthorized(c, "Token missing")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Forbidden", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		Forbidden(c, "Access denied")

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		NotFound(c, "User not found")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InternalError", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		InternalError(c, "Database failure")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
