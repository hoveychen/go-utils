package goutils

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/hoveychen/go-utils/flags"
	"github.com/pkg/errors"

	"golang.org/x/net/proxy"
)

var (
	proxyAddr      = flags.String("proxy", "", "Specify proxy address to fetch data")
	proxyType      = flags.String("proxyType", "sock5", "Either sock5 or http for proxy.")
	requestTimeout = flags.Int("requestTimeout", 10, "Timeout in sec when fetching a remote page.")
	downloadClient *http.Client
)

func init() {
	PkgInit(func() {
		httpTransport := &http.Transport{}
		if *proxyAddr != "" {
			switch *proxyType {
			case "http":
				proxyUrl, err := url.Parse(*proxyAddr)
				if err != nil {
					LogFatal("Failed to parse --proxyAddr", err)
				}
				httpTransport.Proxy = http.ProxyURL(proxyUrl)
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

		downloadClient = &http.Client{Transport: httpTransport}
		downloadClient.Timeout = time.Duration(*requestTimeout) * time.Second
		downloadClient.CheckRedirect = modifiedCheckRedirect
	})
}

func modifiedCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 15 {
		return errors.New("stopped after 15 redirects")
	}
	return nil
}

func Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "New get request")
	}
	resp, err := downloadClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Do get request")
	}
	return resp, nil
}

func PostJson(url string, data interface{}) (*http.Response, error) {
	encodedData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "Encode json")
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(encodedData))
	if err != nil {
		return nil, errors.Wrap(err, "New post request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := downloadClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Do post request")
	}
	return resp, nil
}

// FetchData is a helper function to load local/remote data in the same function.
// Local: goutils.FetchData("/absolute/path/to/file")
// Remote: goutils.FetchData("https://www.google.com")
// Also, it's integrated with proxy in flags.
// TODO(yuheng): Allow more options, while keeping easy use.
func FetchData(path string) ([]byte, error) {
	url, err := url.Parse(path)
	if err != nil {
		return nil, errors.Wrap(err, "decode url")
	}

	switch url.Scheme {
	case "http", "https":
		resp, err := Get(path)
		if err != nil {
			return nil, errors.Wrap(err, "Http get")
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "read response")
		}

		return data, nil
	case "":
		data, err := ioutil.ReadFile(url.Path)
		if err != nil {
			return nil, errors.Wrap(err, "read file")
		}
		return data, nil
	default:
		return nil, errors.New("Unknown scheme:" + url.Scheme)
	}
}

// FetchJson is a wrapper to call FetchData() and parse results from json.
func FetchJson(path string, resp interface{}) error {
	d, err := FetchData(path)
	if err != nil {
		return errors.Wrap(err, "fetch data")
	}
	if err := json.Unmarshal(d, resp); err != nil {
		return errors.Wrap(err, "decode json")
	}
	return nil
}

// FetchXml is a wrapper to call FetchData() and parse results from xml.
func FetchXml(path string, resp interface{}) error {
	d, err := FetchData(path)
	if err != nil {
		return errors.Wrap(err, "fetch data")
	}
	if err := xml.Unmarshal(d, resp); err != nil {
		return errors.Wrap(err, "decode xml")
	}
	return nil
}
