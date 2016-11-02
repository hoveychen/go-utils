package goutils

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

var (
	proxyAddr      = flag.String("proxy", "", "Specify proxy address to fetch data")
	proxyType      = flag.String("proxyType", "sock5", "Either sock5 or http for proxy.")
	requestTimeout = flag.Int("requestTimeout", 10, "Timeout in sec when fetching a remote page.")
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
			return nil, err
		}
		httpTransport.Proxy = http.ProxyURL(proxyUrl)
	case "sock5":
		dialer, err := proxy.SOCKS5("tcp", *proxyAddr, nil, proxy.Direct)
		if err != nil {
			return nil, err
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
		return nil, err
	}

	switch url.Scheme {
	case "http", "https":
		client, err := getDownloadClient()
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return data, nil
	case "":
		data, err := ioutil.ReadFile(url.Path)
		if err != nil {
			return nil, err
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
		return err
	}
	if err := json.Unmarshal(d, resp); err != nil {
		return err
	}
	return nil
}

// FetchXml is a wrapper to call FetchData() and parse results from xml.
func FetchXml(path string, resp interface{}) error {
	d, err := FetchData(path)
	if err != nil {
		return err
	}
	if err := xml.Unmarshal(d, resp); err != nil {
		return err
	}
	return nil
}