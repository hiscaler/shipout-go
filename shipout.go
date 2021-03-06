package shipout

import (
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hiscaler/gox/bytex"
	"github.com/hiscaler/gox/cryptox"
	"github.com/hiscaler/shipout-go/config"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 返回代码

const (
	OK = 200 // 无错误
)

func init() {
	extra.RegisterFuzzyDecoders()
}

var ErrNotFound = errors.New("shipout: not found")

type ShipOut struct {
	Debug       bool        // 是否调试模式
	EnableCache bool        // 是否激活缓存
	OMS         omsServices // OMS API Services
}

func NewShipOut(config config.Config) *ShipOut {
	logger := log.New(os.Stdout, "[ ShipOut ] ", log.LstdFlags|log.Llongfile)
	shipOutClient := &ShipOut{
		Debug: config.Debug,
	}
	httpClient := resty.New().
		SetDebug(config.Debug).
		SetBaseURL("https://open.shipout.com/api/").
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
			"appKey":       config.AppKey,
		}).
		SetAuthToken(config.Authorization).
		SetAllowGetMethodPayload(true).
		SetTimeout(10 * time.Second).
		OnBeforeRequest(func(client *resty.Client, request *resty.Request) error {
			path := request.URL
			qp := request.QueryParam.Encode()
			if qp != "" {
				path += "?" + qp
			}
			headers := map[string]string{
				"timestamp": strconv.Itoa(int(time.Now().UnixMicro())),
				"version":   "1.0.0",
				"path":      path,
			}
			keys := make([]string, len(headers))
			i := 0
			for k := range headers {
				keys[i] = k
				i++
			}
			sort.Strings(keys)
			sb := strings.Builder{}
			for _, key := range keys {
				sb.WriteString(key)
				sb.WriteString(headers[key])
			}
			sb.WriteString(config.SecretKey)
			headers["sign"] = strings.ToUpper(cryptox.Md5(sb.String()))
			request.SetHeaders(headers)
			return nil
		}).
		OnAfterResponse(func(client *resty.Client, response *resty.Response) (err error) {
			if response.IsError() {
				return fmt.Errorf("%s: %s", response.Status(), bytex.ToString(response.Body()))
			}

			r := struct {
				Result         string `json:"result"`
				ErrorCode      string `json:"ErrorCode"`
				Message        string `json:"message"`
				ChineseMessage string `json:"zhMessage"`
			}{}
			if err = jsoniter.Unmarshal(response.Body(), &r); err == nil {
				if r.Result != "OK" {
					err = ErrorWrap(r.ErrorCode, r.ChineseMessage, r.Message)
				}
			}

			if err != nil {
				logger.Printf("OnAfterResponse error: %s", err.Error())
			}
			return
		})
	if config.Debug {
		httpClient.SetBaseURL("https://opendev.shipout.com/api/")
		httpClient.EnableTrace()
	}
	httpClient.JSONMarshal = jsoniter.Marshal
	httpClient.JSONUnmarshal = jsoniter.Unmarshal
	xService := service{
		debug:      config.Debug,
		logger:     logger,
		httpClient: httpClient,
	}
	shipOutClient.OMS = omsServices{
		BaseInfo:          (baseInfoService)(xService),
		Product:           (productService)(xService),
		Order:             (orderService)(xService),
		ValueAddedService: (valueAddedService)(xService),
	}
	return shipOutClient
}

// NormalResponse Normal API response
type NormalResponse struct {
	Result    string        `json:"result"`
	ErrorCode string        `json:"errorCode"`
	Message   string        `json:"message"`
	ZhMessage string        `json:"zhMessage"`
	ErrorType string        `json:"errorType"`
	Data      []interface{} `json:"data"`
}

// ErrorWrap 错误包装
func ErrorWrap(code string, messages ...string) error {
	msg := ""
	for _, message := range messages {
		message = strings.TrimSpace(message)
		if message != "" {
			msg = message
			break
		}
	}
	if code == "" {
		return errors.New(msg)
	}
	return fmt.Errorf("%s: %s", code, msg)
}
