package gosuv2

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// SuvClient is the client to interact with the SUV service.
type SuvClient struct {
	SuvURL        url.URL
	HttpClient    http.Client
	Config        SuvConfig
	PrintRequest  func(req *http.Request, body io.Reader)
	PrintResponse func(res *http.Response)
}

// SuvConfig holds the configuration for SuvClient.
type SuvConfig struct {
	Host       string
	PhpSession string
	UserCode   string
	Password   string
	Detailed   bool
}

type SuvCookieJar struct {
	cookies []*http.Cookie
}

// NewSuvClient creates a new SuvClient with the provided configuration.
func NewSuvClient(cfg SuvConfig) *SuvClient {
	client := &SuvClient{
		SuvURL: url.URL{
			Scheme: "http",
			Host:   cfg.Host,
			Path:   "/portal",
		},
		HttpClient: http.Client{
			Jar: http.CookieJar(&SuvCookieJar{}),
		},
		Config: cfg,
	}

	// Default print functions
	client.PrintRequest = client.defaultPrintRequest
	client.PrintResponse = client.defaultPrintResponse

	return client
}

func (cj *SuvCookieJar) Cookies(u *url.URL) []*http.Cookie {
	return cj.cookies
}

func (cj *SuvCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	cj.cookies = cookies
}

func (c *SuvClient) defaultPrintRequest(req *http.Request, body io.Reader) {
	fmt.Println()
	fmt.Println("[Request]")
	fmt.Println(req.Method, req.URL.String())
	for k, v := range req.Header {
		fmt.Println(k+":", v)
	}
	fmt.Println()
	if body != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(body)
		fmt.Println(buf.String())
	}
	fmt.Println()
	fmt.Println("Performing request...")
	fmt.Println()
}

func (c *SuvClient) defaultPrintResponse(res *http.Response) {
	fmt.Println()
	fmt.Println("[Response]")
	fmt.Println(res.Proto, res.Status)
	for k, v := range res.Header {
		fmt.Println(k+":", v)
	}
	fmt.Println()
	bodyCopy := new(bytes.Buffer)
	bodyCopy.ReadFrom(res.Body)
	res.Body = io.NopCloser(bodyCopy)
	fmt.Println(bodyCopy.String())
	fmt.Println()
}

// setCookiesFromResponse sets cookies from the response to the client's cookie jar.
func (c *SuvClient) setCookiesFromResponse(res *http.Response) error {
	if len(res.Cookies()) == 0 {
		return fmt.Errorf("no cookies in response")
	}

	for _, cookie := range res.Cookies() {
		c.HttpClient.Jar.SetCookies(&c.SuvURL, []*http.Cookie{cookie})
	}

	return nil
}

func (c *SuvClient) postRequest(path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest("POST", c.SuvURL.String()+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "suvctl/0.1")

	return req, nil
}

func (c *SuvClient) getRequest(path string) (*http.Request, error) {
	req, err := http.NewRequest("GET", c.SuvURL.String()+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "suvctl/0.1")

	return req, nil
}

func (c *SuvClient) urlEncodedPostRequest(data url.Values, path string) (*http.Response, error) {
	payload := data.Encode()

	req, err := c.postRequest(path, bytes.NewReader([]byte(payload)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(payload)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if c.Config.Detailed {
		c.PrintRequest(req, bytes.NewReader([]byte(payload)))
	}

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if c.Config.Detailed {
		c.PrintResponse(res)
	}

	err = checkResponseBody(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Check if the response body contains an error message without consuming it
func checkResponseBody(res *http.Response) error {
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", res.Status)
	}
	bodyCopy := new(bytes.Buffer)
	bodyCopy.ReadFrom(res.Body)
	res.Body = io.NopCloser(bodyCopy)
	bodyString := bodyCopy.String()
	if bytes.Contains([]byte(bodyString), []byte("C:\\wamp64\\www\\SistemaSUV2")) {
		return errors.New("suv2 replied with an error, check your session id or try logging in again")
	}
	return nil
}
