package bing

import (
	"encoding/json"
	"errors"
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
	"time"
)

const conversationUrl = "https://www.bing.com/turing/conversation/create"
const maxRetryTimes = 3

var (
	jar     = tlsclient.NewCookieJar()
	options = []tlsclient.HttpClientOption{
		tlsclient.WithTimeoutSeconds(360),
		tlsclient.WithClientProfile(tlsclient.Safari_IOS_16_0),
		tlsclient.WithNotFollowRedirects(),
		tlsclient.WithCookieJar(jar), // create cookieJar instance and pass it as argument
	}
	client, _ = tlsclient.NewHttpClient(tlsclient.NewNoopLogger(), options...)
	httpProxy = os.Getenv("http_proxy")
)

func init() {
	if httpProxy != "" {
		_ = client.SetProxy(httpProxy)
		logrus.Infof("Proxy set: %s", httpProxy)
	}
}

type Conversation struct {
	ClientId               string `json:"clientId"`
	ConversationId         string `json:"conversationId"`
	ConversationSignature  string `json:"conversationSignature"`
	ConversationExpiryTime string `json:"conversationExpiryTime"`
}

func CreateConversationWithRetry(cookie string, retryTimes int) (*Conversation, error) {
	var err error
	var convs *Conversation
	for i := 0; i < retryTimes+1; i++ {
		convs, err = CreateConversation(cookie)
		if err == nil {
			logrus.Debugf("new conversation created, conversation: %+v", convs)
			return convs, err
		}
		if i != retryTimes {
			logrus.Warnf("create conversation failed: %s, waiting 3s for next retry", err.Error())
			time.Sleep(3 * time.Second)
		}
	}
	if err != nil {
		logrus.Errorf("failed to create conversation: %s", err.Error())
	}
	return convs, err
}

func CreateConversation(cookie string) (*Conversation, error) {
	requestId := uuid.New().String()
	if cookie == "" {
		return nil, errors.New("empty cookie")
	}
	req, err := http.NewRequest(http.MethodGet, conversationUrl, nil)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(cookie, "_U=") {
		cookie = "_U=" + cookie
	}
	logrus.Debugf("cookie: %s", cookie)
	req.Header.Set("accept", "application/json")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("sec-ch-ua", `"Not_A Brand";v="99", "Microsoft Edge";v="109", "Chromium";v="109"`)
	req.Header.Set("sec-ch-ua-arch", `"x86"`)
	req.Header.Set("sec-ch-ua-bitness", `"64"`)
	req.Header.Set("sec-ch-ua-full-version", `"109.0.1518.78"`)
	req.Header.Set("sec-ch-ua-full-version-list", `"Not_A Brand";v="99.0.0.0", "Microsoft Edge";v="109.0.1518.78", "Chromium";v="109.0.5414.120"`)
	req.Header.Set("sec-ch-ua-mobile", `?0`)
	req.Header.Set("sec-ch-ua-model", ``)
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("sec-ch-ua-platform-version", `"12.6.0"`)
	req.Header.Set("sec-fetch-dest", `empty`)
	req.Header.Set("sec-fetch-mode", `cors`)
	req.Header.Set("sec-fetch-site", `same-origin`)
	req.Header.Set("x-ms-client-request-id", requestId)
	req.Header.Set("x-ms-useragent", `azsdk-js-api-client-factory/1.0.0-beta.1 core-rest-pipeline/1.10.0 OS/Win32`)
	req.Header.Set("user-agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36 Edg/113.0.1774.50`)
	req.Header.Set("cookie", cookie)
	req.Header.Set("Referrer", `https://www.bing.com/search?q=Bing+AI&showconv=1`)
	req.Header.Set("Referrer-policy", `origin-when-cross-origin`)

	var res *http.Response
	for i := 0; i < maxRetryTimes; i++ {
		res, err = client.Do(req)
		if err == nil {
			break
		} else {
			if i != maxRetryTimes-1 {
				time.Sleep(3 * time.Second)
				logrus.Warnf(fmt.Sprintf("request failed, retry to request... retry times: %d", i+1))
			}
		}
	}

	if err != nil {
		logrus.Errorf("request failed: %s", err.Error())
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		var result ConversationResult
		data, _ := io.ReadAll(res.Body)
		err = json.Unmarshal(data, &result)
		if err != nil {
			logrus.Errorf("/turing/conversation/create: failed to parse response body. %s, error: %s", data, err.Error())
			return nil, errors.New(fmt.Sprintf("/turing/conversation/create: failed to parse response body. %s", data))
		}
		if result.ClientId == "" || result.ConversationId == "" || result.ConversationSignature == "" {
			logrus.Errorf("unexpected response: %s: %s", result.Result.Value, result.Result.Message)
			return nil, errors.New(fmt.Sprintf("unexpected response: %s: %s", result.Result.Value, result.Result.Message))
		}
		return &Conversation{
			ClientId:              result.ClientId,
			ConversationId:        result.ConversationId,
			ConversationSignature: result.ConversationSignature,
		}, nil
	} else {
		logrus.Errorf("unexpected HTTP error createConversation %d: %s", res.StatusCode, res.Status)
		return nil, errors.New(fmt.Sprintf("unexpected HTTP error createConversation %d: %s", res.StatusCode, res.Status))
	}
}
