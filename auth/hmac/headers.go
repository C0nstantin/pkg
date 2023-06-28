package hmac

var ContentTypeVariants = []string{
	"Content-Type",
	"CONTENT-TYPE",
	"CONTENT_TYPE",
	"HTTP_CONTENT_TYPE",
	"HTTP_X_HMAC_CONTENT_TYPE",
	"HTTP_X_CONTENT_TYPE",
}
var AuthorizeVariants = []string{
	"Authorization",
	"AUTHORIZATION",
	"HTTP_AUTHORIZATION",
	"HTTP_X_HMAC_AUTHORIZATION",
	"HTTP_X_AUTHORIZATION",
}
var XAuthorizationContentSha256Variants = []string{
	"X-Authorization-Content-Sha256",
	"X-AUTHORIZATION-CONTENT-SHA256",
	"X_AUTHORIZATION_CONTENT_SHA256",
	"HTTP_X_AUTHORIZATION_CONTENT_SHA256",
}

var OriginURIVariants = []string{
	"X-Original-Uri",
	"X-ORIGINAL-URI",
	"HTTP_X_HMAC_ORIGINAL_URI",
	"HTTP_X_ORIGINAL_URI",
	"X_ORIGINAL_URI",
}
var DateVariants = []string{
	"Date",
	"DATE",
	"HTTP_DATE",
	"HTTP_X_HMAC_DATE",
	"HTTP_X_DATE",
}
