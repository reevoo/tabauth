package main

import (
	"encoding/base64"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestTabAuth(t *testing.T) {
	Convey("Given an instance of the engine", t, func() {
		Convey("And the user is authorized", func() {
			Convey("They should see a 404 when visting the wrong path", func() {
				w := makeRequest("http://example.com/foo", "foo:bar", 200)
				So(w.Code, ShouldEqual, 404)
			})

			Convey("They should see a 200 when visting the correct path", func() {
				w := makeRequest("http://example.com/user/foofah/token", "foo:bar", 200)
				So(w.Code, ShouldEqual, 200)
			})

			Convey("They should see a 500 when the backend is down", func() {
				w := makeRequest("http://example.com/user/foofah/token", "foo:bar", 500)
				So(w.Code, ShouldEqual, 500)
			})

			Convey("They should see a 404 when the backend cannot find the token", func() {
				w := makeRequest("http://example.com/user/foofah/token", "foo:bar", 404)
				So(w.Code, ShouldEqual, 404)
				So(w.Body.String(), ShouldEqual, "-1\n")
			})

			Convey("the backend is called with the correct body, the token is returned unchanged", func() {
				w := makeRequest("http://example.com/user/iamtheuser/token", "foo:bar", 200)
				So(w.Body.String(), ShouldEqual, "username=iamtheuser\n")
			})

			Convey("When the client IP is given, it is passed to the backend, the token is returned unchanged", func() {
				w := makeRequest("http://example.com/user/iamtheuser/token?client_ip=10.10.10.15", "foo:bar", 200)
				So(w.Body.String(), ShouldEqual, "client_ip=10.10.10.15&username=iamtheuser\n")
			})

			Convey("When the site ID is given, it is passed to the backend, the token is returned unchanged", func() {
				w := makeRequest("http://example.com/user/iamtheuser/token?site_id=sdf86438brf34", "foo:bar", 200)
				So(w.Body.String(), ShouldEqual, "target_site=sdf86438brf34&username=iamtheuser\n")
			})

			Convey("When the site ID and client IP are given, they are passed to the backend, the token is returned unchanged", func() {
				w := makeRequest("http://example.com/user/iamtheuser/token?site_id=sdf86438brf34&client_ip=127.0.0.1", "foo:bar", 200)
				So(w.Body.String(), ShouldEqual, "client_ip=127.0.0.1&target_site=sdf86438brf34&username=iamtheuser\n")
			})
		})

		Convey("And the user is unauthorized", func() {
			Convey("They should see a 401 error when visiting the wrong path", func() {
				w := makeRequest("http://example.com/foo", "foo:inccorrect", 200)
				So(w.Code, ShouldEqual, 401)
			})

			Convey("They should see a 401 error when visiting a valid path", func() {
				w := makeRequest("http://example.com/user/thenameofauser/token", "foo:inccorrect", 200)
				So(w.Code, ShouldEqual, 401)
			})
		})
	})
}

func makeRequest(url string, auth string, code int) httptest.ResponseRecorder {
	server, client := mockHTTP(code)
	engine := TabAuth(client)
	req, _ := http.NewRequest("GET", url, nil)
	authString := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+authString)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	server.Close()
	return *recorder
}

func mockHTTP(code int) (*httptest.Server, *Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trusted" {
			panic("hit backend with the wrong path")
		}
		if code == 500 {
			panic("Mock backend internal server error")
		}
		if code == 404 {
			w.WriteHeader(200)
			fmt.Fprintln(w, "-1")
		} else {
			body, _ := ioutil.ReadAll(r.Body)
			w.WriteHeader(code)
			fmt.Fprintln(w, string(body))
		}
	}))

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	httpClient := &http.Client{Transport: transport}
	client := &Client{server.URL, httpClient}

	return server, client
}
