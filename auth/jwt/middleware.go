package jwt

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

var (
	realmName     = "testing area"
	IdentityKey   = "ID"
	TokenLookup   = "header: Authorization, query: token, cookie: jwt"
	TokenHeadName = "Bearer"
)

const (
	DefaultRedirectLoginURL        = "/"
	DefaultRedirectLogoutURL       = "/"
	DefaultRedirectUnauthorizedUrl = "/login"
)

var (
	ErrInvalidAuthKey = errors.New("authorization key small or empty")
)

func NewGinJWTMiddlewareMust(authKey []byte) *jwt.GinJWTMiddleware {
	middleware, err := NewGinJWTMiddleware(authKey)
	if err != nil {
		panic(err)
	}
	return middleware
}

func NewGinJWTMiddleware(authKey []byte) (*jwt.GinJWTMiddleware, error) {
	if len(authKey) <= 128 {
		return nil, ErrInvalidAuthKey
	}
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:          realmName,
		Key:            authKey,
		Timeout:        time.Hour,
		MaxRefresh:     time.Hour,
		PayloadFunc:    PayloadFunc,
		Unauthorized:   Unauthorized,
		LoginResponse:  LoginResponse,
		LogoutResponse: LogoutResponse,
		IdentityKey:    IdentityKey,
		TokenLookup:    TokenLookup,
		TokenHeadName:  TokenHeadName,

		SendCookie:     true,
		SecureCookie:   true,
		CookieHTTPOnly: true,
	})
}

type Response struct {
	Code   int    `json:"code"`
	Token  string `json:"token"`
	Expire string `json:"expire"`
}

func PayloadFunc(data interface{}) jwt.MapClaims {
	if v, ok := data.(*JWTBody); ok {
		return jwt.MapClaims{
			"Id":    v.Id,
			"Email": v.Email,
		}
	}
	return jwt.MapClaims{}
}

func Unauthorized(c *gin.Context, code int, message string) {
	if isJSON(c) {
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})
	} else {
		redirectURL := getRedirectURL(c, "unauthorized")
		c.Redirect(http.StatusSeeOther, redirectURL)
	}
}

func LoginResponse(c *gin.Context, code int, token string, expire time.Time) {
	if isJSON(c) {
		c.JSON(http.StatusOK, &Response{
			Code:   code,
			Token:  token,
			Expire: expire.Format(time.RFC3339),
		})
	} else {
		redirectUrl := getRedirectURL(c, "login")
		parse, err := url.Parse(redirectUrl)
		if err != nil {
			return
		}
		parse.Host = c.Request.Host

		c.Redirect(http.StatusSeeOther, parse.String())
	}
}

func LogoutResponse(c *gin.Context, code int) {
	if isJSON(c) {
		c.JSON(code, gin.H{
			"code": http.StatusOK,
		})
	} else {
		redirectUrl := getRedirectURL(c, "logout")
		c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
	}
}

func getRedirectURL(c *gin.Context, action string) string {
	if res := c.Query("ReturnUrl"); res != "" {
		return res
	}
	if res := c.Query("redirect_url"); res != "" {
		return res
	}
	if res, _ := c.Cookie("redirect_url"); res != "" {
		return res
	}
	switch action {
	case "login":
		if res, ok := os.LookupEnv("DEFAULT_REDIRECT_LOGIN_URL"); ok { // @todo add to config
			return res
		} else {
			return DefaultRedirectLoginURL
		}
	case "logout":
		if res, ok := os.LookupEnv("DEFAULT_REDIRECT_LOGOUT_URL"); ok { // @todo add to config
			return res
		} else {
			return DefaultRedirectLogoutURL
		}
	case "unauthorized":
		if res, ok := os.LookupEnv("DEFAULT_REDIRECT_UNAUTHORIZED_URL"); ok { // @todo add to config
			return res
		} else {
			return DefaultRedirectUnauthorizedUrl
		}
	}

	return DefaultRedirectLoginURL
}

func isJSON(c *gin.Context) bool {
	contentType := []byte(c.GetHeader("Content-Type"))
	accept := []byte(c.GetHeader("Accept"))
	res, _ := regexp.Match(`json.*`, contentType)
	acResp, _ := regexp.Match(`json.*`, accept)
	return res || acResp
}

type Auth struct {
	Jwt *jwt.GinJWTMiddleware
}

func (a Auth) AuthMiddleware() gin.HandlerFunc {
	return a.Jwt.MiddlewareFunc()
}

func NewJwtAuthService(apiKey []byte) (*Auth, error) {
	j, err := NewGinJWTMiddleware(apiKey)
	if err != nil {
		return nil, err
	}
	return &Auth{Jwt: j}, nil
}
func NewMustJwtAuthService(apiKey []byte) *Auth {
	j, err := NewGinJWTMiddleware(apiKey)
	if err != nil {
		panic(err)
	}
	return &Auth{Jwt: j}
}
