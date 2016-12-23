package goutils

import (
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
)

func getDownloadClient() (*http.Client, error) {
	if *proxyAddr == "" {
		return http.DefaultClient, nil
	}

	httpTransport := &http.Transport{}
	switch *proxyType {
	case "http":
		proxyUrl, err := url.Parse(*proxyAddr)
		if err != nil {
			return nil, errors.Wrap(err, "parse --proxyAddr")
		}
		httpTransport.Proxy = http.ProxyURL(proxyUrl)
	case "sock5":
		dialer, err := proxy.SOCKS5("tcp", *proxyAddr, nil, proxy.Direct)
		if err != nil {
			return nil, errors.Wrap(err, "dial sock5")
		}
		httpTransport.Dial = dialer.Dial
	default:
		return nil, errors.New("Unknown proxy type:" + *proxyType)
	}

	httpClient := &http.Client{Transport: httpTransport}
	httpClient.Timeout = time.Duration(*requestTimeout) * time.Second
	return httpClient, nil
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
		client, err := getDownloadClient()
		if err != nil {
			return nil, errors.Wrap(err, "get download client")
		}
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			return nil, errors.Wrap(err, "new request")
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "do request")
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
