package jwt

import (
	"net/http/httptest"
	"testing"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoginResponse(t *testing.T) {

}

func TestLogoutResponse(t *testing.T) {

}

func TestNewGinJWTMiddleware(t *testing.T) {
	j, err := NewGinJWTMiddleware([]byte("asfasdf"))
	assert.Nil(t, j)
	assert.Error(t, err)
	j2, err2 := NewGinJWTMiddleware([]byte("asdfasdfasdfasdfasdwerwqresdfasdfasdfasdfasqwrqwerfasdadsfasdfasdfadsfasdfasdfasdfadsfadsfasdfasdfadsfasdfasdfasdfasdfasdfadsfadsfasdfasdfasdf"))
	assert.NoError(t, err2)
	assert.IsType(t, &jwt.GinJWTMiddleware{}, j2)
}

func TestPayloadFunc(t *testing.T) {

}

func TestUnauthorized(t *testing.T) {

}

func Test_getRedirectURL(t *testing.T) {

}

func Test_isJson(t *testing.T) {
	tr := []string{
		"application/json",
		"text/json",
		"text/x-json",
	}
	fs := []string{
		"application/x-javascript",
		"text/x-javascript",
		"application/ogg",
		"application/zip",
	}
	for _, trr := range tr {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Add("Content-Type", trr)
		assert.True(t, isJSON(c))
	}
	for _, trr := range fs {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Add("Content-Type", trr)
		assert.False(t, isJSON(c))
	}

}
