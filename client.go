package foxesscloud

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL = "https://www.foxesscloud.com"
)

type Config struct {
	Client    *http.Client
	Token     string
	UserAgent string
}

func NewClient(cfg Config) (*Client, error) {
	client := http.DefaultClient
	if cfg.Client != nil {
		client = cfg.Client
	}
	c := &Client{
		client:    client,
		token:     cfg.Token,
		userAgent: cfg.UserAgent,
	}
	c.PowerStations = &PowerStationService{client: c}
	c.Inverters = &InverterService{client: c}
	return c, nil
}

type Client struct {
	client    *http.Client
	token     string
	userAgent string

	PowerStations *PowerStationService
	Inverters     *InverterService
}

const (
	errNoNoError           = 0
	errNoTokenExpired      = 41808
	errNoTokenInvalid      = 41809
	errHeadersMissing      = 40256
	errBodyInvalid         = 40257
	errRequestsTooFrequent = 40400
	errRateLimitExceeded   = 40402
)

type RateLimitExceededError struct {
	msg string
}

func (e *RateLimitExceededError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %s", e.msg)
}

type errorResponse struct {
	ErrNo int    `json:"errno"`
	Msg   string `json:"msg"`
}

func (c *Client) do(req *http.Request, res any) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	var errResp errorResponse
	if err := json.Unmarshal(data, &errResp); err != nil {
		return nil, fmt.Errorf("could not unmarshal json response: %w", err)
	}

	switch errResp.ErrNo {
	case errNoNoError:
		break
	case errNoTokenInvalid, errNoTokenExpired:
		return nil, fmt.Errorf("invalid token: %v", errResp.Msg)
	case errHeadersMissing:
		return nil, fmt.Errorf("missing headers: %v", errResp.Msg)
	case errBodyInvalid:
		return nil, fmt.Errorf("invalid body: %v", errResp.Msg)
	case errRequestsTooFrequent:
		return nil, fmt.Errorf("rate limit exceeded: %v", errResp.Msg)
	case errRateLimitExceeded:
		return nil, &RateLimitExceededError{msg: errResp.Msg}
	default:
		return nil, fmt.Errorf("invalid response, got error code: %v, message: %v", errResp.ErrNo, errResp.Msg)
	}

	if err := json.Unmarshal(data, res); err != nil {
		return nil, fmt.Errorf("could not unmarshal json response: %w", err)
	}
	return resp, nil
}

func (c *Client) newGetRequest(ctx context.Context, path string, qry url.Values) (*http.Request, error) {
	return c.newRequest(ctx, http.MethodGet, path, qry, nil)
}

func (c *Client) newPostRequest(ctx context.Context, path string, pld any) (*http.Request, error) {
	return c.newRequest(ctx, http.MethodPost, path, nil, pld)
}

func (c *Client) newRequest(ctx context.Context, method, path string, qry url.Values, pld any) (*http.Request, error) {
	var body io.Reader
	if pld != nil {
		data, err := json.Marshal(pld)
		if err != nil {
			return nil, fmt.Errorf("could not marshal request body: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, buildPath(path, qry), body)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for k, v := range c.buildHeaders(path, now) {
		req.Header.Set(k, v)
	}
	return req, nil
}

func (c *Client) buildHeaders(path string, date time.Time) map[string]string {
	timestamp := date.UnixNano() / int64(time.Millisecond)
	headers := map[string]string{
		"content-type": "application/json",
		"lang":         "en",
		"timestamp":    strconv.FormatInt(timestamp, 10),
		"user-agent":   c.userAgent,
		"token":        c.token,
		"signature":    buildSignature(c.token, path, timestamp),
	}
	return headers
}

func buildPath(path string, qry url.Values) string {
	if len(qry) > 0 {
		return baseURL + path + "?" + qry.Encode()
	}
	return baseURL + path
}

func buildSignature(token string, path string, timestamp int64) string {
	var b strings.Builder
	b.WriteString(path)
	b.WriteString("\\r\\n")
	b.WriteString(token)
	b.WriteString("\\r\\n")
	b.WriteString(strconv.FormatInt(timestamp, 10))
	return hashMD5(b.String())
}

func hashMD5(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	return hex.EncodeToString(hash.Sum(nil))
}

type Pagination struct {
	CurrentPage int `json:"currentPage"`
	Pagesize    int `json:"pageSize"`
}

type QueryTimestamp struct {
	time.Time
}

func (t *QueryTimestamp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.UnixMilli(), 10)), nil
}

type DataTimestamp struct {
	time.Time
}

func (t *DataTimestamp) UnmarshalJSON(b []byte) error {
	d, err := time.Parse("2006-01-02 15:04:05 MST-0700", strings.Trim(string(b), "\""))
	if err != nil {
		return err
	}
	t.Time = d
	return nil
}

type DataFloat struct {
	Value float64
}

func (f *DataFloat) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	f.Value = v
	return nil
}

type DataListResponse[T any] struct {
	Items []T
}

type wrappedDataListResponse[T any] struct {
	Result []T `json:"result"`
}

func (l *wrappedDataListResponse[T]) unwrap() *DataListResponse[T] {
	return &DataListResponse[T]{
		Items: l.Result,
	}
}

type ListResponse[T any] struct {
	Items       []T
	CurrentPage int
	Pagesize    int
	Total       int
}

type wrappedListResponse[T any] struct {
	Result struct {
		Data        []T `json:"data"`
		CurrentPage int `json:"currentPage"`
		Pagesize    int `json:"pageSize"`
		Total       int `json:"total"`
	} `json:"result"`
}

func (l *wrappedListResponse[T]) unwrap() *ListResponse[T] {
	return &ListResponse[T]{
		Items:       l.Result.Data,
		CurrentPage: l.Result.CurrentPage,
		Pagesize:    l.Result.Pagesize,
		Total:       l.Result.Total,
	}
}

type wrappedDetailResponse[T any] struct {
	Result *T `json:"result"`
}

func (l *wrappedDetailResponse[T]) unwrap() *T {
	return l.Result
}
