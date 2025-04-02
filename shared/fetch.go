package shared

import (
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"sync"
)

var RequestClient *http.Client

var TransportPool = sync.Pool{}

// TODO init Fetch without init()
func init() {
	PutTransport(InitTransport())
	InitClient()
	log.Println("client: init!")
}

func InitClient() {
	jar, _ := cookiejar.New(nil)
	tr := TransportPool.Get()
	if tr == nil {
		tr = InitTransport()
	}
	RequestClient = &http.Client{
		Timeout:   RequestTimeout,
		Jar:       jar,
		Transport: tr.(*http.Transport),
	}
	go PutTransport(InitTransport())
}

func PutTransport(transport *http.Transport) {
	TransportPool.Put(transport)
}

func InitTransport() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	return transport
}

var FetchBearerTokenList = make(map[string]string)

// fetch
func Fetch(_url string, _method string, __no_used_variable *int, body io.Reader, headers map[string]string, ext map[string]string) (int, *http.Response, []byte, error) {
	//s := time.Now()
	req, err := http.NewRequest(_method, _url, body)
	if err != nil {
		return 400, nil, nil, err
	}
	//req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")
	req.Header.Set("user-agent", "mockrpc-exp")
	req.Header.Set("content-type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := RequestClient.Do(req)
	if err != nil {
		return 400, nil, nil, err
	}
	defer resp.Body.Close()
	if extStatusCodeOnly, ok := ext["ext_no_body"]; ok && extStatusCodeOnly == "1" {
		io.Copy(io.Discard, resp.Body)
		return resp.StatusCode, resp, nil, nil
	}

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, resp, nil, err
	}
	//for k, v := range resp.Header {
	//	log.Println(k, v)
	//}
	// log.Println(string(response[:]))
	//log.Println(time.Since(s))
	return resp.StatusCode, resp, response, err
}
