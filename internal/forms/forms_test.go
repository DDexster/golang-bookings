package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func getRequest() *http.Request {
	return httptest.NewRequest("POST", "/whatever", nil)
}

func getRequestWithParams(params url.Values) *http.Request {
	r := httptest.NewRequest("POST", "/whatever", nil)
	r.PostForm = params
	return r
}

func TestNew(t *testing.T) {
	var values url.Values
	form := New(values)
	switch tp := interface{}(form).(type) {
	case *Form:
	//	All Good
	default:
		t.Error("Form values is not of url.Values", tp)
	}
}

func TestForm_Required(t *testing.T) {
	r := getRequest()
	form := New(r.PostForm)
	form.Required("fname", "lname")
	if form.Valid() {
		t.Error("Empty Form Should not be valid")
	}

	data := url.Values{}
	data.Add("fname", "Dex")
	r = getRequestWithParams(data)
	form = New(r.PostForm)
	form.Required("fname", "lname")

	if form.Valid() {
		t.Error("Not Full Form Should not be valid")
	}

	data = url.Values{}
	data.Add("fname", "Dex")
	data.Add("lname", "Dex")
	r = getRequestWithParams(data)
	form = New(r.PostForm)
	form.Required("fname", "lname")

	if !form.Valid() {
		t.Error("Form should be valid")
	}
}

func TestForm_Has(t *testing.T) {
	data := url.Values{}
	data.Add("fname", "Dex")
	r := getRequestWithParams(data)
	form := New(r.PostForm)
	if !form.Has("fname") {
		t.Error("Form should be a 'fname' key")
	}
}

func TestForm_MinLength(t *testing.T) {
	data := url.Values{}
	data.Add("fname", "Dex")
	r := getRequestWithParams(data)
	form := New(r.PostForm)
	if !form.MinLength("fname", 3) {
		t.Error("Form min length should be valid")
	}
	if form.MinLength("fname", 5) {
		t.Error("Form min length should not be valid")
	}
}

func TestForm_IsEmail(t *testing.T) {
	data := url.Values{}
	data.Add("email", "Dex")
	r := getRequestWithParams(data)
	form := New(r.PostForm)
	form.IsEmail("email")
	if form.Valid() {
		t.Error("Form should not be valid")
	}

	data = url.Values{}
	data.Add("email", "dex@cabdo.de")
	r = getRequestWithParams(data)
	form = New(r.PostForm)
	form.IsEmail("email")
	if !form.Valid() {
		t.Error("Form should be valid")
	}
}

func TestErrors_Get(t *testing.T) {
	data := url.Values{}
	data.Add("email", "Dex")
	r := getRequestWithParams(data)
	form := New(r.PostForm)
	form.IsEmail("email")
	if len(form.Errors.Get("email")) == 0 {
		t.Error("Form should have errors")
	}

	data = url.Values{}
	data.Add("email", "dex@cabdo.de")
	r = getRequestWithParams(data)
	form = New(r.PostForm)
	form.IsEmail("email")
	if len(form.Errors.Get("email")) != 0 {
		t.Error("Form should have not errors")
	}
}
