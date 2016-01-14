package main_test

import (
	"encoding/base64"
	"fmt"
	"github.com/reevoo/tabauth"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBadRequest(t *testing.T) {
	w := makeRequest("http://example.com/foo", "foo:bar", 200)

	if expected := 400; w.Code != expected {
		t.Errorf("Expected HTTP status: %v to eq %v", w.Code, expected)
	}
}

func TestGoodRequest(t *testing.T) {
	w := makeRequest("http://example.com/user/foofah/ticket", "foo:bar", 200)

	if expected := 200; w.Code != expected {
		t.Errorf("Expected HTTP status: %v to eq %v", w.Code, expected)
	}
}

func TestBackendDown(t *testing.T) {
	w := makeRequest("http://example.com/user/foofah/ticket", "foo:bar", 500)

	if expected := 500; w.Code != expected {
		t.Errorf("Expected HTTP status: %v to eq %v", w.Code, expected)
	}
}

func TestTicketNotFound(t *testing.T) {
	w := makeRequest("http://example.com/user/foofah/ticket", "foo:bar", 404)

	if expected := 404; w.Code != expected {
		t.Errorf("Expected HTTP status: %v to eq %v", w.Code, expected)
	}

	if expected := "Not Found\n"; w.Body.String() != expected {
		t.Errorf("Expected Body: %v to eq %v", w.Body.String(), expected)
	}
}

func TestTicketBody(t *testing.T) {
	w := makeRequest("http://example.com/user/iamtheuser/ticket", "foo:bar", 200)

	if expected := "username=iamtheuser\n"; w.Body.String() != expected {
		t.Errorf("Expected Body: %v to eq %v", w.Body.String(), expected)
	}
}

func TestPassingClientIP(t *testing.T) {
	w := makeRequest("http://example.com/user/iamtheuser/ticket?client_ip=10.10.10.15", "foo:bar", 200)
	if expected := "client_ip=10.10.10.15&username=iamtheuser\n"; w.Body.String() != expected {
		t.Errorf("Expected Body: %v to eq %v", w.Body.String(), expected)
	}
}

func TestPassingSiteID(t *testing.T) {
	w := makeRequest("http://example.com/user/iamtheuser/ticket?site_id=sdf86438brf34", "foo:bar", 200)
	if expected := "target_site=sdf86438brf34&username=iamtheuser\n"; w.Body.String() != expected {
		t.Errorf("Expected Body: %v to eq %v", w.Body.String(), expected)
	}
}

func TestPassingSiteIDandClientIP(t *testing.T) {
	w := makeRequest("http://example.com/user/iamtheuser/ticket?site_id=sdf86438brf34&client_ip=127.0.0.1", "foo:bar", 200)
	if expected := "client_ip=127.0.0.1&target_site=sdf86438brf34&username=iamtheuser\n"; w.Body.String() != expected {
		t.Errorf("Expected Body: %v to eq %v", w.Body.String(), expected)
	}
}

func TestUnauthorized(t *testing.T) {
	w := makeRequest("http://example.com/foo", "foo:inccorrect", 200)
	if expected := 401; w.Code != expected {
		t.Errorf("Expected HTTP status: %v to eq %v", w.Code, expected)
	}
	w = makeRequest("http://example.com/user/thenameofauser/ticket", "foo:inccorrect", 200)
	if expected := 401; w.Code != expected {
		t.Errorf("Expected HTTP status: %v to eq %v", w.Code, expected)
	}
}

func makeRequest(url string, auth string, code int) httptest.ResponseRecorder {
	server := mockHTTP(code)
	engine := main.New("", server.URL)
	req, _ := http.NewRequest("GET", url, nil)
	authString := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+authString)
	recorder := httptest.NewRecorder()
	engine.Handler.ServeHTTP(recorder, req)
	server.Close()
	return *recorder
}

func mockHTTP(code int) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trusted" {
			panic("hit backend with the wrong path")
		}
		if code == 500 {
			panic("Mock backend internal server error")
		}
		if code == 404 {
			w.WriteHeader(200)
			fmt.Fprint(w, "-1")
		} else {
			body, _ := ioutil.ReadAll(r.Body)
			w.WriteHeader(code)
			fmt.Fprintln(w, string(body))
		}
	}))
	return server
}
