package gosuv2

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// Login attempts to log in to the SUV system with the provided user code and password.
func (c *SuvClient) Login(usercode, password string) (*string, error) {
	data := url.Values{
		"user": {usercode},
		"pass": {password},
	}

	c.Config.UserCode = usercode
	c.Config.Password = password

	res, err := c.urlEncodedPostRequest(data, "/validar.php")
	if err != nil {
		return nil, err
	}

	phpSession, err := c.handleLoginResponse(res)
	if err != nil {
		return nil, err
	}

	return phpSession, nil
}

// Logout logs out from the SUV system.
func (c *SuvClient) Logout() error {
	req, err := c.getRequest("/desconectar.php")
	if err != nil {
		return err
	}

	if c.Config.Detailed {
		c.PrintRequest(req, nil)
	}

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}

	if c.Config.Detailed {
		c.PrintResponse(res)
	}

	bodyBuf := new(bytes.Buffer)
	bodyBuf.ReadFrom(res.Body)
	body := bodyBuf.String()

	if !strings.Contains(body, "Su sesion ha culminado") {
		return errors.New("logout failed: unexpected response")
	}

	c.Config.PhpSession = ""

	return nil
}

// Load existing session into the CookieJar.
func (c *SuvClient) LoadPhpSession() error {
	if c.Config.PhpSession == "" {
		return errors.New("php session is empty")
	}

	cookie := &http.Cookie{
		Name:  "PHPSESSID",
		Value: c.Config.PhpSession,
	}

	c.HttpClient.Jar.SetCookies(&c.SuvURL, []*http.Cookie{cookie})

	return nil
}

// handleLoginResponse processes the login response.
func (c *SuvClient) handleLoginResponse(res *http.Response) (*string, error) {
	var validarPhpResponse []string

	err := json.NewDecoder(res.Body).Decode(&validarPhpResponse)
	if err != nil {
		return nil, err
	}

	if len(validarPhpResponse) == 0 {
		return nil, errors.New("login failed")
	}

	if validarPhpResponse[0] == "0" {
		return nil, errors.New("login failed, check your credentials")
	}

	c.setCookiesFromResponse(res)

	for _, cookie := range c.HttpClient.Jar.Cookies(&c.SuvURL) {
		if cookie.Name == "PHPSESSID" {
			c.Config.PhpSession = cookie.Value
		}
	}

	return &c.Config.PhpSession, nil
}
