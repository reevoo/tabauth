package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/url"
)

func main() {
        client := &Client{"http://localhost", &http.Client{}}
	r := TabAuth(client)
	r.RunTLS("0.0.0.0:443", "cert.pem", "key.pem")
}

func TabAuth(client *Client) *gin.Engine {
	r := gin.Default()
	r.Use(gin.BasicAuth(accounts()))

	r.GET("/user/:username/token", func(c *gin.Context) {
		token, err := client.getToken(c)
		if err != nil {
			c.String(500, err.Error())
		} else {
			if token == "-1\n" {
				c.String(404, token)
			} else {
				c.String(200, token)
			}
		}
	})

	return r
}

type Client struct {
  BaseUrl string
  HTTPClient  *http.Client
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
  resp, err := c.HTTPClient.PostForm(c.BaseUrl + "/trusted", form)
	if err != nil {
		return "", err
	}
	token, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}
	return string(token), nil
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
