package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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

	New(os.Getenv("BIND_ADDR"), tableauEndpoint).ListenAndServeTLS("cert.pem", "key.pem")
}

//New returns a new *http.Server with the provided configuration
func New(bindAddress string, tableauEndpoint string) *http.Server {
	return &http.Server{
		Addr: bindAddress,
		Handler: TabAuth{
			&Client{tableauEndpoint, &http.Client{}},
			accounts(),
		},
	}
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

//TabAuth is a http.Handler for the Tabauth Application
type TabAuth struct {
	*Client
	accounts map[string]string
}

//ServeHTTP impliments the http.Handler interface on TabAuth
//We ensure that the request is authenticated before passing it to the TabauthHandler
//We also handle any errors here.
func (t TabAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !t.isAuthenticated(r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	status, err := t.TabauthHandler(w, r)
	if err != nil {
		log.Printf("HTTP %d: %s", status, err.Error())
		http.Error(w, http.StatusText(status), status)
	}
}

func (t TabAuth) isAuthenticated(r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	if ok && password == t.accounts[username] {
		return true
	}
	return false
}

func (t TabAuth) TabauthHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	user, err := username(r.URL.Path)
	if err != nil {
		return http.StatusBadRequest, err
	}
	token, err := t.Client.GetToken(user, r.FormValue("site_id"), r.FormValue("client_ip"))
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if token == "-1\n" {
		return http.StatusNotFound, errors.New("tabauth: token not found")
	}
	fmt.Fprintf(w, token)
	return http.StatusOK, nil
}

func username(path string) (string, error) {
	match := pathRegexp.FindStringSubmatch(path)
	if len(match) > 1 {
		return match[1], nil
	}
	return "", errors.New("tabauth: bad request")
}

//Client is a client for Tableau Server's Trusted Authentication API
// http://onlinehelp.tableau.com/current/server/en-us/trusted_auth.htm
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

//GetToken returns the authentication token for the given params
func (c *Client) GetToken(username, siteID, clientIP string) (string, error) {
	form := url.Values{}
	form.Add("username", username)
	if siteID != "" {
		form.Add("target_site", siteID)
	}
	if clientIP != "" {
		form.Add("client_ip", clientIP)
	}
	resp, err := c.HTTPClient.PostForm(c.BaseURL+"/trusted", form)
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
