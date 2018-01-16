// This is a wrapper of standard http package, exposing out-of-the-box flags like proxy, timeout, json handling.
package goutils

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/ernesto-jimenez/httplogger"
	"github.com/hoveychen/go-utils/flags"
	"github.com/pkg/errors"

	"golang.org/x/net/proxy"
)

var (
	proxyAddr      = flags.String("proxy", "", "Specify proxy address to fetch data")
	proxyType      = flags.String("proxyType", "sock5", "Either sock5 or http for proxy.")
	requestTimeout = flags.Int("requestTimeout", 10, "Timeout in sec when fetching a remote page.")
	logAccess      = flags.Bool("logAccess", false, "True to log every requests.")
	downloadClient *http.Client
	requestOnce    sync.Once
)

func modifiedCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 15 {
		return errors.New("stopped after 15 redirects")
	}
	return nil
}

type httpLogger struct{}

func (l *httpLogger) LogRequest(req *http.Request) {
	LogInfo("[Request]", req.Method, req.URL.String())
}

func (l *httpLogger) LogResponse(req *http.Request, res *http.Response, err error, duration time.Duration) {
	if err != nil {
		LogError("[Response]", err, req.URL.String())
		return
	}
	LogInfo("[Response]", req.Method, res.StatusCode, duration.Seconds(), req.URL.String())
}

func GetDownloadClient() *http.Client {
	requestOnce.Do(func() {
		httpTransport := &http.Transport{}
		if *proxyAddr != "" {
			switch *proxyType {
			case "http":
				proxyUrl, err := url.Parse(*proxyAddr)
				if err != nil {
					LogFatal("Failed to parse --proxyAddr", err)
				}
				httpTransport.Proxy = http.ProxyURL(proxyUrl)
				httpTransport.TLSClientConfig = &tls.Config{
					InsecureSkipVerify: true,
				}
			case "sock5":
				dialer, err := proxy.SOCKS5("tcp", *proxyAddr, nil, proxy.Direct)
				if err != nil {
					LogFatal("Failed to dial sock5", err)
				}
				httpTransport.Dial = dialer.Dial
			default:
				LogFatal("Unknown proxy type:", *proxyType)
			}
		}

		var roundTripper http.RoundTripper = httpTransport

		if *logAccess {
			roundTripper = httplogger.NewLoggedTransport(roundTripper, &httpLogger{})
		}

		downloadClient = &http.Client{Transport: roundTripper}
		downloadClient.Timeout = time.Duration(*requestTimeout) * time.Second
		downloadClient.CheckRedirect = modifiedCheckRedirect
	})

	return downloadClient
}

func GetWithContext(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "New get request")
	}
	req = req.WithContext(ctx)
	resp, err := GetDownloadClient().Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Do get request")
	}
	return resp, nil
}

func Get(url string) (*http.Response, error) {
	return GetWithContext(context.Background(), url)
}

func PostFormWithContext(ctx context.Context, uri string, data map[string]string) (*http.Response, error) {
	values := url.Values{}
	for k, v := range data {
		values.Set(k, v)
	}
	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(values.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "New post request")
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := GetDownloadClient().Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Do post request")
	}
	return resp, nil
}

func PostForm(uri string, data map[string]string) (*http.Response, error) {
	return PostFormWithContext(context.Background(), uri, data)
}

func PostJsonWithContext(ctx context.Context, url string, data interface{}) (*http.Response, error) {
	encodedData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "Encode json")
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(encodedData))
	if err != nil {
		return nil, errors.Wrap(err, "New post request")
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	resp, err := GetDownloadClient().Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Do post request")
	}
	return resp, nil
}

func PostJson(url string, data interface{}) (*http.Response, error) {
	return PostJsonWithContext(context.Background(), url, data)
}

func FetchDataWithContext(ctx context.Context, path string) ([]byte, error) {
	url, err := url.Parse(path)
	if err != nil {
		return nil, errors.Wrap(err, "Decode url")
	}

	switch url.Scheme {
	case "http", "https":
		resp, err := GetWithContext(ctx, path)
		if err != nil {
			return nil, errors.Wrap(err, "Http get")
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "Read response")
		}
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(resp.Status + ":" + string(data))
		}

		return data, nil
	case "":
		data, err := ioutil.ReadFile(url.Path)
		if err != nil {
			return nil, errors.Wrap(err, "Read file")
		}
		return data, nil
	default:
		return nil, errors.New("Unknown scheme:" + url.Scheme)
	}
}

// FetchData is a helper function to load local/remote data in the same function.
// Local: goutils.FetchData("/absolute/path/to/file")
// Remote: goutils.FetchData("https://www.google.com")
// Also, it's integrated with proxy in flags.
// TODO(yuheng): Allow more options, while keeping easy use.
func FetchData(path string) ([]byte, error) {
	return FetchDataWithContext(context.Background(), path)
}

func FetchJsonWithContext(ctx context.Context, path string, resp interface{}) error {
	d, err := FetchDataWithContext(ctx, path)
	if err != nil {
		return errors.Wrap(err, "Fetch data")
	}
	if err := json.Unmarshal(d, resp); err != nil {
		return errors.Wrap(err, "Decode json")
	}
	return nil
}

// FetchJson is a wrapper to call FetchData() and parse results from json.
func FetchJson(path string, resp interface{}) error {
	return FetchJsonWithContext(context.Background(), path, resp)
}

func FetchXmlWithContext(ctx context.Context, path string, resp interface{}) error {
	d, err := FetchDataWithContext(ctx, path)
	if err != nil {
		return errors.Wrap(err, "Fetch data")
	}
	if err := xml.Unmarshal(d, resp); err != nil {
		return errors.Wrap(err, "Decode xml")
	}
	return nil
}

// FetchXml is a wrapper to call FetchData() and parse results from xml.
func FetchXml(path string, resp interface{}) error {
	return FetchXmlWithContext(context.Background(), path, resp)
}
