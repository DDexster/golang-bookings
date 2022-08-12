package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name           string
	url            string
	method         string
	params         []postData
	expectedStatus int
}{
	{"Home Page", "/", "GET", []postData{}, http.StatusOK},
	{"About Page", "/about", "GET", []postData{}, http.StatusOK},
	{"General's Quarters Page", "/generals-quarters", "GET", []postData{}, http.StatusOK},
	{"Major's Suite Page", "/majors-suite", "GET", []postData{}, http.StatusOK},
	{"Reservation Form Page", "/reservation", "GET", []postData{}, http.StatusOK},
	{"Search Page", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"Contact Page", "/contact", "GET", []postData{}, http.StatusOK},
	{"Reservation Page POST", "/reservation", "POST", []postData{
		{key: "first_name", value: "John"},
		{key: "last_name", value: "Smith"},
		{key: "email", value: "John@go.com"},
		{key: "phone", value: "555-55-555"},
	}, http.StatusOK},
	{"Search Availability POST", "/search-availability", "POST", []postData{
		{key: "start", value: "2022-02-01"},
		{key: "end", value: "2022-04-01"},
	}, http.StatusOK},
	{"Search Availability POST", "/search-availability-json", "POST", []postData{
		{key: "start", value: "2022-02-01"},
		{key: "end", value: "2022-04-01"},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		if e.method == "GET" {
			resp, err := ts.Client().Get(ts.URL + e.url)
			if err != nil {
				t.Log(err)
				t.Fail()
			}

			if resp.StatusCode != e.expectedStatus {
				t.Errorf("For %s expected %d, but got, %d", e.name, e.expectedStatus, resp.StatusCode)
			}

		} else if e.method == "POST" {
			values := url.Values{}
			for _, x := range e.params {
				values.Add(x.key, x.value)
			}
			resp, err := ts.Client().PostForm(ts.URL+e.url, values)
			if err != nil {
				t.Log(err)
				t.Fail()
			}

			if resp.StatusCode != e.expectedStatus {
				t.Errorf("For %s expected %d, but got, %d", e.name, e.expectedStatus, resp.StatusCode)
			}
		}
	}
}
