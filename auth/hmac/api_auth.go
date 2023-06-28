package hmac

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/sha3"
)

// Simple Header Authorization = APIAuth 'client access id':'signature'
// canonical_string = "#{http method},#{content-type},#{X-Authorization-Content-SHA256},#{request URI},#{timestamp}"

var AuthHeaderPattern = `APIAuth(?P<digest>-HMAC-(MD5|SHA(?:1|224|256|384|512)?))? (?P<access_id>[^:]+):(?P<sign>.+)$`

type APIAuth struct {
	SecretKey []byte
	AccessID  string
}

type signParams struct {
	accessID            string
	contentType         string
	xAuthHeader         string
	requestURI          string
	timestamp           string
	authorizationHeader string
	digest              string
	hmacStr             string
	httpMethod          string
}

func (a *APIAuth) CheckRequest(r *http.Request) error {
	h, err := a.parseHeaders(r)
	if err != nil {
		return err
	}

	canonicalString, err := a.CanonicalString(h)
	if err != nil {
		return err
	}
	hh, err := a.getDigest(h.digest)
	if err != nil {
		return err
	}

	mac := hmac.New(hh, a.SecretKey)
	mac.Write([]byte(canonicalString))
	calculationHmac := mac.Sum(nil)
	if hex.EncodeToString(calculationHmac) == h.hmacStr {
		return nil
	}

	return ErrRequestUnAuthorized
}

func (a *APIAuth) CanonicalString(h *signParams) (string, error) {
	return fmt.Sprintf("%s,%s,%s,%s,%s",
		h.httpMethod,
		h.contentType,
		h.xAuthHeader,
		h.requestURI,
		h.timestamp,
	), nil
}

func (a *APIAuth) SignRequest(r *http.Request, digest string) error {
	if a.AccessID == "" {
		return ErrAccessIDNotSet
	}

	h, err := a.setHeaders(r)
	if err != nil {
		return err
	}

	canonicalString, err := a.CanonicalString(h)
	if err != nil {
		return err
	}

	return a.setAuthHeader(r, a.AccessID, canonicalString, digest)
}

func (a *APIAuth) parseHeaders(r *http.Request) (*signParams, error) {
	headers := make(map[string][]string)
	for name, header := range r.Header {
		headers[strings.ToUpper(name)] = header
	}
	h := &signParams{}
	h.httpMethod = r.Method
	h.xAuthHeader = a.findHeader(headers, XAuthorizationContentSha256Variants...)
	h.authorizationHeader = a.findHeader(headers, AuthorizeVariants...)
	h.timestamp = a.findHeader(headers, DateVariants...)
	h.contentType = a.findHeader(headers, ContentTypeVariants...)
	h.requestURI = a.findHeader(headers, OriginURIVariants...)

	if h.requestURI == "" {
		h.requestURI = r.RequestURI
	}
	if h.authorizationHeader == "" {
		return nil, ErrAuthHeaderNotFound
	}
	rg := regexp.MustCompile(AuthHeaderPattern)
	res := rg.FindStringSubmatch(h.authorizationHeader)
	if len(res) < 5 || res[3] == "" || res[4] == "" {
		return nil, ErrAuthHeaderInvalid
	}

	if res[2] == "" {
		h.digest = "SHA1"
	} else {
		h.digest = res[2]
	}

	h.accessID = res[3]
	h.hmacStr = res[4]
	return h, nil
}

func (a *APIAuth) setHeaders(r *http.Request) (*signParams, error) {
	h := &signParams{}
	h.httpMethod = r.Method
	h.contentType = setContentType(r)
	h.timestamp = a.setDate(r)
	h.requestURI = a.setRequestURI(r)
	h.xAuthHeader = a.setAuthContentSha(r)
	return h, nil
}

func (a *APIAuth) findHeader(headers map[string][]string, variant ...string) string {
	for _, s := range variant {
		h, ok := headers[s]
		if ok {
			return h[0]
		}
	}
	return ""
}

func (a *APIAuth) setAuthHeader(r *http.Request, accessID string, canonicalString string, digest string) error {
	if digest == "" {
		digest = "sha1"
	}

	hh, err := a.getDigest(digest)
	if err != nil {
		return err
	}
	mac := hmac.New(hh, a.SecretKey)
	mac.Write([]byte(canonicalString))
	hmacString := hex.EncodeToString(mac.Sum(nil))
	r.Header.Set(AuthorizeVariants[0], fmt.Sprintf("APIAuth-HMAC-%s %s:%s", strings.ToUpper(digest), accessID, hmacString))
	return nil
}

func (a *APIAuth) getDigest(digest string) (func() hash.Hash, error) {
	switch strings.ToUpper(digest) {
	case "SHA1":
		return sha1.New, nil
	case "SHA256":
		return sha256.New, nil
	case "SHA224":
		return sha3.New224, nil
	case "SHA386":
		return sha3.New384, nil
	case "SHA512":
		return sha512.New, nil
	default:
		return nil, ErrMethodNotSupported
	}
}

func (a *APIAuth) setDate(r *http.Request) string {
	for _, variant := range DateVariants {
		if r.Header.Get(variant) != "" {
			return r.Header.Get(variant)
		}
	}
	date := time.Now().Format(time.RFC1123)
	r.Header.Set(DateVariants[0], date)
	return date
}

func setContentType(r *http.Request) string {
	for _, variant := range ContentTypeVariants {
		if r.Header.Get(variant) != "" {
			return r.Header.Get(variant)
		}
	}
	r.Header.Set("Content-Type", "application/json")
	return "application/json"
}

func (a *APIAuth) setRequestURI(r *http.Request) string {
	for _, variant := range OriginURIVariants {
		if r.Header.Get(variant) != "" {
			return r.Header.Get(variant)
		}
	}
	if r.RequestURI != "" {
		r.Header.Set(OriginURIVariants[0], r.RequestURI)
		return r.RequestURI
	}
	r.Header.Set(OriginURIVariants[0], "/")
	return "/"
}

func (a *APIAuth) setAuthContentSha(r *http.Request) string {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		r.Header.Set(XAuthorizationContentSha256Variants[0], "")
		return ""
	}

	reader, err := r.GetBody()
	if err != nil {
		r.Header.Set(XAuthorizationContentSha256Variants[0], "")
		return ""
	}
	s, err := io.ReadAll(reader)
	if err != nil {
		r.Header.Set(XAuthorizationContentSha256Variants[0], "")
		return ""
	}
	sh := sha256.New()
	sh.Write(s)
	xAuth := base64.StdEncoding.EncodeToString(sh.Sum(nil))
	r.Header.Set(XAuthorizationContentSha256Variants[0], xAuth)
	return xAuth
}
