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

// NewGinJWTMiddlewareMust creates a new GinJWTMiddleware instance with the provided authentication key.
// It is a convenience function that panics if the creation of the middleware fails, making it suitable
// for use in situations where failure to create the middleware should result in an immediate halt of the application.
// This function is useful for initializing middleware in global scope or during application startup.
//
// Parameters:
// - authKey []byte: The secret key used for token generation and validation. It must be at least 128 bits long.
//
// Returns:
// - *jwt.GinJWTMiddleware: A pointer to the newly created GinJWTMiddleware instance.
func NewGinJWTMiddlewareMust(authKey []byte) *jwt.GinJWTMiddleware {
	middleware, err := NewGinJWTMiddleware(authKey)
	if err != nil {
		panic(err)
	}
	return middleware
}

// NewGinJWTMiddleware creates a new instance of GinJWTMiddleware configured with the provided authentication key.
// This middleware is used to handle JWT authentication for Gin-based web applications. It sets up the necessary
// configuration for JWT token validation, including the realm, key, token expiration, payload function, and response
// handlers for various authentication events.
//
// Parameters:
// - authKey []byte: The secret key used for signing and verifying JWT tokens. It must be at least 128 bits (16 bytes) long.
//
// Returns:
// - *jwt.GinJWTMiddleware: A pointer to the newly created GinJWTMiddleware instance, configured with the provided settings.
// - error: An error that indicates why the middleware could not be created. This is typically due to an insufficiently long authKey.
//
// Note: The function checks if the authKey is at least 128 bits long. If it is not, it returns an error indicating
// that the key is invalid. This is a critical security requirement to ensure the strength of the JWT tokens.
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
		CookieSameSite: http.SameSiteNoneMode,
	})
}

type Response struct {
	Code   int    `json:"code"`
	Token  string `json:"token"`
	Expire string `json:"expire"`
}

// PayloadFunc generates JWT claims based on the provided data.
// This function is intended to be used as a callback for the jwt.GinJWTMiddleware,
// allowing customization of the JWT payload. It checks if the input data is of type *JWTBody,
// and if so, extracts the 'Id' and 'Email' fields to be included in the JWT claims.
//
// Parameters:
// - data interface{}: The data from which to generate the JWT claims. Expected to be of type *JWTBody.
//
// Returns:
// - jwt.MapClaims: A map representing the JWT claims. Includes 'Id' and 'Email' if the input is of type *JWTBody;
// otherwise, returns an empty map.
func PayloadFunc(data interface{}) jwt.MapClaims {
	if v, ok := data.(*JWTBody); ok {
		return jwt.MapClaims{
			"Id":    v.Id,
			"Email": v.Email,
			"Info":  v.Info,
		}
	}
	return jwt.MapClaims{}
}

// Unauthorized handles unauthorized access attempts by either returning a JSON response with an error code and message
// if the request accepts JSON, or by redirecting the user to a predefined unauthorized access URL.
// It checks the request headers to determine the preferred response format.
//
// Parameters:
// - c *gin.Context: The context of the current HTTP request. It contains request information, response writers, and other middleware-related data.
// - code int: The HTTP status code to be returned in the case of a JSON response.
// - message string: The error message to be included in the JSON response or to be used as a basis for redirection in case of non-JSON requests.
//
// This function does not return any value. It directly writes to the HTTP response or redirects the user.
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

// LoginResponse handles the response to a successful login attempt. It determines the response format based on the client's request headers.
// If the client accepts JSON, it responds with a JSON object containing the login status code, token, and token expiration time.
// Otherwise, it redirects the client to a URL specified by query parameters or environment variables, appending the request host to the URL.
//
// Parameters:
// - c *gin.Context: The context of the current HTTP request. It contains request information, response writers, and other middleware-related data.
// - code int: The HTTP status code to be returned in the case of a JSON response.
// - token string: The JWT token generated upon successful authentication.
// - expire time.Time: The expiration time of the generated JWT token.
//
// This function does not return any value. It directly writes to the HTTP response or redirects the user based on the request headers.
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

// LogoutResponse handles the response to a logout request. It determines the response format based on the client's request headers.
// If the client accepts JSON, it responds with a JSON object containing the HTTP status code indicating a successful operation.
// Otherwise, it redirects the client to a URL specified by query parameters, environment variables, or default settings.
//
// Parameters:
// - c *gin.Context: The context of the current HTTP request. It contains request information, response writers, and other middleware-related data.
// - code int: The HTTP status code to be returned in the case of a JSON response. This is typically http.StatusOK to indicate a successful logout.
//
// This function does not return any value. It directly writes to the HTTP response or redirects the user based on the request headers.
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

// AuthMiddleware description of the Go function.
//
// No parameters.
// Returns a gin.HandlerFunc.
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

// NewMustJwtAuthService creates a new instance of Auth using the provided API key for JWT authentication.
// This function is a convenience wrapper around NewGinJWTMiddleware, ensuring that the JWT middleware is
// successfully created or panics otherwise. It is useful for initializing the Auth service in scenarios
// where failing to create the service should halt the application startup.
//
// Parameters:
//   - apiKey []byte: The secret key used for signing and verifying JWT tokens. It must meet the security
//     requirements enforced by NewGinJWTMiddleware, including being at least 128 bits long.
//
// Returns:
// - *Auth: A pointer to the newly created Auth instance, which includes the configured JWT middleware.
//
// Note: This function will panic if the provided apiKey is invalid or if any other error occurs during
// the creation of the JWT middleware. This is intended to catch configuration errors during application
// initialization rather than at runtime.
func NewMustJwtAuthService(apiKey []byte) *Auth {
	j, err := NewGinJWTMiddleware(apiKey)
	if err != nil {
		panic(err)
	}
	return &Auth{Jwt: j}
}
