package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

var pathRegexp = regexp.MustCompile(`\A\/user\/([^/]+)\/ticket\z`)

func main() {
	tableauEndpoint := os.Getenv("TABLEAU_ENDPOINT")
	if tableauEndpoint == "" {
		tableauEndpoint = "http://localhost"
	}

	Tabauth(os.Getenv("BIND_ADDR"), tableauEndpoint).ListenAndServeTLS("cert.pem", "key.pem")
}


func Tabauth(bindAddress string, tableauEndpoint string) *http.Server {
	return &http.Server{
		Addr: bindAddress,
		Handler: tabauth{
			&Client{tableauEndpoint, &http.Client{}},
			accounts(),
		},
	}
}

type tabauth struct {
	*Client
	accounts map[string]string
	//handler  func(tabauth, http.ResponseWriter, *http.Request) (int, error)
}

func (t tabauth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, err := t.TabauthHandler(w, r)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
	}
}

func (t tabauth) TabauthHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	if !isAuthenticated(r, t.accounts) {
		return http.StatusUnauthorized, errors.New("tabauth: unauthenticated")
	}
	user, err := username(r.URL.Path)
	if err != nil {
		return http.StatusBadRequest, err
	}
	token, err := t.Client.getToken(user, r.FormValue("site_id"), r.FormValue("client_ip"))
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if token == "-1\n" {
		return http.StatusNotFound, errors.New("tabauth: token not found")
	}
	fmt.Fprintf(w, token)
	return http.StatusOK, nil
}

func isAuthenticated(r *http.Request, accounts map[string]string) bool {
	username, password, ok := r.BasicAuth()
	if ok && password == accounts[username] {
		return true
	}
	return false
}

func accounts() map[string]string {
	file, err := ioutil.ReadFile("accounts.json")
	if err != nil {
		panic(err)
	}
	var accs map[string]string
	err = json.Unmarshal([]byte(file), &accs)
	if err != nil {
		panic(err)
	}
	return accs
}

func username(path string) (string, error) {
	match := pathRegexp.FindStringSubmatch(path)
	if len(match) > 1 {
		return match[1], nil
	}
	return "", errors.New("tabauth: bad request")
}

type Client struct {
	BaseUrl    string
	HTTPClient *http.Client
}

func (c *Client) getToken(username, siteId, clientIp string) (string, error) {
	form := url.Values{}
	form.Add("username", username)
	if siteId != "" {
		form.Add("target_site", siteId)
	}
	if clientIp != "" {
		form.Add("client_ip", clientIp)
	}
	resp, err := c.HTTPClient.PostForm(c.BaseUrl+"/trusted", form)
	if err != nil {
		return "", err
	}
	ticket, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}
	return string(ticket), nil
}
