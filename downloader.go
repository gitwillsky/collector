package collector

import (
	"time"
	"net/http"
	"sync"
	"math/rand"
	"net/url"
	"mime"
	"io/ioutil"
	"errors"
	"path"
)

const UA = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36"

type httpDownloader struct {
	proxyList []string
	timeout   time.Duration
}

var clientPool = sync.Pool{
	New: func() interface{} {
		return &http.Client{}
	},
}

func NewHttpDownloader(timeout time.Duration, proxy ...string) *httpDownloader {
	return &httpDownloader{
		proxyList: proxy,
		timeout:   timeout,
	}
}

func (h *httpDownloader) Download(address *Address) (*Data, error) {
	client := clientPool.Get().(*http.Client)
	defer clientPool.Put(client)

	// set timeout
	client.Timeout = h.timeout

	// set proxy
	if (len(h.proxyList) > 0) {
		rand.Seed(time.Now().Unix())
		index := rand.Intn(len(h.proxyList))
		u, err := url.Parse(h.proxyList[index])
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(u),
			}
		}
	}

	request, err := http.NewRequest("GET", address.URL, nil)
	if err != nil {
		return nil, err
	}

	// set header
	request.Header.Set("Referer", address.URL)
	request.Header.Set("User-Agent", UA)

	res, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	contentType := res.Header.Get("Content-Type")

	mt, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}

	data := &Data{}
	defer res.Body.Close()
	data.Content, _ = ioutil.ReadAll(res.Body)

	data.Name = path.Base(res.Request.URL.Path)
	resType, _ := mime.ExtensionsByType(mt)
	if (len(resType) > 0) {
		switch  resType[0] {
		case ".html":
			data.DataType = "html"
		case ".jpg":
			data.DataType = "jpg"
		case ".png":
			data.DataType = "png"
		case ".gif":
			data.DataType = "gif"
		case ".htm":
			data.DataType = "html"
		}
	}

	if (data.DataType == "") {
		return nil, errors.New("Unknown response data type")
	}

	return data, nil
}
