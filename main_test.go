package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}
func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	for town, list := range cafeList {

		requests := []struct {
			count    int
			want     int
			NameTown string
		}{
			{0, 0, town},           // не возвращается ни одного кафе
			{1, 1, town},           // возвращается одно кафе
			{2, 2, town},           // возвращается два кафе
			{100, len(list), town}, // возвращается минимум из общего количества кафе и 100
		}

		u := url.URL{Path: "/cafe"}
		q := u.Query()
		q.Set("city", town)

		for _, v := range requests {
			q.Set("count", strconv.Itoa(v.count))
			u.RawQuery = q.Encode()

			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", u.String(), nil)
			handler.ServeHTTP(response, req)

			body := strings.TrimSpace(response.Body.String())
			answer := strings.Split(body, ",")
			if body == "" {
				assert.Equal(t, v.want, 0)
			} else {
				assert.Equal(t, v.want, len(answer))
			}
		}

	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{"фасоль", 0}, //не возвращается ни одного кафе
		{"кофе", 2},   //возвращается два кафе
		{"вилка", 1},  //возвращается одно кафе
	}

	u := url.URL{Path: "/cafe"}
	q := u.Query()
	q.Set("city", "moscow")

	for _, v := range requests {
		q.Set("search", v.search)
		u.RawQuery = q.Encode()

		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", u.String(), nil)
		handler.ServeHTTP(response, req)

		body := strings.TrimSpace(response.Body.String())
		if body == "" {
			assert.Equal(t, v.wantCount, 0)
			continue
		}

		searchTerm := strings.ToLower(v.search)
		cafeNames := strings.Split(body, ",")
		countCofeName := 0

		for _, name := range cafeNames {
			name = strings.TrimSpace(strings.ToLower(name))
			if strings.Contains(name, searchTerm) {
				countCofeName++
			}

		}

		assert.Equal(t, v.wantCount, countCofeName)
	}
}
