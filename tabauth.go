package main

import (
	"encoding/json"
	"github.com/reevoo/tabauth/Godeps/_workspace/src/github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
        "log"
)

func main() {
	if os.Getenv("TABLEAU_ENDPOINT") == "" {
		log.Fatalln("TABLEAU_ENDPOINT not set in environment")
	}

	if os.Getenv("BIND_ADDR") == "" {
		log.Fatalln("BIND_ADDR not set in environment")
	}
	client := &Client{os.Getenv("TABLEAU_ENDPOINT"), &http.Client{}}
	r := TabAuth(client)
	r.RunTLS(os.Getenv("BIND_ADDR"), "cert.pem", "key.pem")
}

func TabAuth(client *Client) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(gin.BasicAuth(accounts()))

	r.GET("/user/:username/ticket", func(c *gin.Context) {
		ticket, err := client.getToken(c)
		if err != nil {
			c.String(500, err.Error())
		} else {
			if ticket == "-1\n" {
				c.String(404, ticket)
			} else {
				c.String(200, ticket)
			}
		}
	})

	return r
}

type Client struct {
	BaseUrl    string
	HTTPClient *http.Client
}

func (c *Client) getToken(g *gin.Context) (string, error) {
	form := url.Values{}
	form.Add("username", g.Param("username"))
	site_id := g.Query("site_id")
	if site_id != "" {
		form.Add("target_site", site_id)
	}
	client_ip := g.Query("client_ip")
	if client_ip != "" {
		form.Add("client_ip", client_ip)
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
