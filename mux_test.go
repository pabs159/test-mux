package mux

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
)

func TestGPTNewRouter(t *testing.T) {
	r := NewRouter()
	if r == nil {
		t.Fatal("NewRouter() returned nil")
	}
}

func TestGPTRouter_Get(t *testing.T) {
	r := NewRouter()
	route := r.NewRoute().Name("testRoute")
	if route == nil {
		t.Fatal("NewRoute().Name() returned nil")
	}

	gotRoute := r.Get("testRoute")
	if gotRoute != route {
		t.Fatalf("Get() = %v; want %v", gotRoute, route)
	}
}

func TestGPTRouter_ServeHTTP(t *testing.T) {
	r := NewRouter()
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ServeHTTP() = %v; want %v", rr.Code, http.StatusOK)
	}
}

func TestGPTRouter_StrictSlash(t *testing.T) {
	r := NewRouter().StrictSlash(true)
	if !r.strictSlash {
		t.Fatal("StrictSlash() did not set the strictSlash field to true")
	}

	r = NewRouter().StrictSlash(false)
	if r.strictSlash {
		t.Fatal("StrictSlash() did not set the strictSlash field to false")
	}
}

func TestGPTRouter_SkipClean(t *testing.T) {
	r := NewRouter().SkipClean(true)
	if !r.skipClean {
		t.Fatal("SkipClean() did not set the skipClean field to true")
	}

	r = NewRouter().SkipClean(false)
	if r.skipClean {
		t.Fatal("SkipClean() did not set the skipClean field to false")
	}
}

func TestGPTRouter_OmitRouteFromContext(t *testing.T) {
	r := NewRouter().OmitRouteFromContext(true)
	if !r.omitRouteFromContext {
		t.Fatal("OmitRouteFromContext() did not set the omitRouteFromContext field to true")
	}

	r = NewRouter().OmitRouteFromContext(false)
	if r.omitRouteFromContext {
		t.Fatal("OmitRouteFromContext() did not set the omitRouteFromContext field to false")
	}
}

func TestGPTRouter_OmitRouterFromContext(t *testing.T) {
	r := NewRouter().OmitRouterFromContext(true)
	if !r.omitRouterFromContext {
		t.Fatal("OmitRouterFromContext() did not set the omitRouterFromContext field to true")
	}

	r = NewRouter().OmitRouterFromContext(false)
	if r.omitRouterFromContext {
		t.Fatal("OmitRouterFromContext() did not set the omitRouterFromContext field to false")
	}
}

func TestGPTRouter_UseEncodedPath(t *testing.T) {
	r := NewRouter().UseEncodedPath()
	if !r.useEncodedPath {
		t.Fatal("UseEncodedPath() did not set the useEncodedPath field to true")
	}
}

func TestGPTRouter_Match(t *testing.T) {
	r := NewRouter()
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET")

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	var match RouteMatch
	if !r.Match(req, &match) {
		t.Fatal("Match() did not find a matching route")
	}
	if match.MatchErr != nil {
		t.Fatalf("Match() returned an error: %v", match.MatchErr)
	}
	if match.Handler == nil {
		t.Fatal("Match() did not return a handler")
	}
}

func TestGPTMethodNotAllowedHandler(t *testing.T) {
	handler := methodNotAllowedHandler()

	req, err := http.NewRequest("POST", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestGPTCleanPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/", "/"},
		{"//", "/"},
		{"/../", "/"},
		{"/a/b/../c", "/a/c"},
		{"/a/b/./c", "/a/b/c"},
		{"/a//b/c/", "/a/b/c/"},
	}

	for _, tt := range tests {
		got := cleanPath(tt.path)
		if got != tt.want {
			t.Errorf("cleanPath(%q) = %v; want %v", tt.path, got, tt.want)
		}
	}
}

func TestGPTReplaceURLPath(t *testing.T) {
	u, err := url.Parse("http://example.com/foo")
	if err != nil {
		t.Fatal(err)
	}

	newPath := "/bar"
	want := "http://example.com/bar"
	got := replaceURLPath(u, newPath)

	if got != want {
		t.Errorf("replaceURLPath() = %v; want %v", got, want)
	}
}

func TestGPTCheckPairs(t *testing.T) {
	_, err := checkPairs("key1", "value1", "key2")
	if err == nil {
		t.Fatal("checkPairs() did not return an error for odd number of pairs")
	}

	_, err = checkPairs("key1", "value1", "key2", "value2")
	if err != nil {
		t.Fatalf("checkPairs() returned an error for even number of pairs: %v", err)
	}
}

func TestGPTMapFromPairsToString(t *testing.T) {
	m, err := mapFromPairsToString("key1", "value1", "key2", "value2")
	if err != nil {
		t.Fatalf("mapFromPairsToString() returned an error: %v", err)
	}
	if m["key1"] != "value1" || m["key2"] != "value2" {
		t.Fatalf("mapFromPairsToString() = %v; want map[key1:value1 key2:value2]", m)
	}
}

func TestGPTMapFromPairsToRegex(t *testing.T) {
	m, err := mapFromPairsToRegex("key1", "^value1$", "key2", "^value2$")
	if err != nil {
		t.Fatalf("mapFromPairsToRegex() returned an error: %v", err)
	}
	if !m["key1"].MatchString("value1") || !m["key2"].MatchString("value2") {
		t.Fatalf("mapFromPairsToRegex() did not return expected regex matchers")
	}
}

func TestGPTMatchInArray(t *testing.T) {
	arr := []string{"one", "two", "three"}
	if !matchInArray(arr, "two") {
		t.Fatal("matchInArray() did not find the value in the array")
	}
	if matchInArray(arr, "four") {
		t.Fatal("matchInArray() found a value that should not be in the array")
	}
}

func TestGPTMatchMapWithString(t *testing.T) {
	toCheck := map[string]string{"Content-Type": "application/json"}
	toMatch := map[string][]string{"Content-Type": {"application/json", "text/plain"}}
	if !matchMapWithString(toCheck, toMatch, true) {
		t.Fatal("matchMapWithString() did not find the key/value pair")
	}
	toCheck = map[string]string{"Content-Type": "application/xml"}
	if matchMapWithString(toCheck, toMatch, true) {
		t.Fatal("matchMapWithString() found a key/value pair that should not exist")
	}
}

func TestGPTMatchMapWithRegex(t *testing.T) {
	toCheck := map[string]*regexp.Regexp{"Content-Type": regexp.MustCompile(`application/.*`)}
	toMatch := map[string][]string{"Content-Type": {"application/json", "application/xml"}}
	if !matchMapWithRegex(toCheck, toMatch, true) {
		t.Fatal("matchMapWithRegex() did not find the key/value pair with regex")
	}
	toCheck = map[string]*regexp.Regexp{"Content-Type": regexp.MustCompile(`text/.*`)}
	if matchMapWithRegex(toCheck, toMatch, true) {
		t.Fatal("matchMapWithRegex() found a key/value pair with regex that should not exist")
	}
}

func TestGPTRouter_NotFoundHandler(t *testing.T) {
	r := NewRouter()
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	})

	req, err := http.NewRequest("GET", "/not-exist", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("ServeHTTP() = %v; want %v", rr.Code, http.StatusNotFound)
	}
	if rr.Body.String() != "Not Found" {
		t.Fatalf("ServeHTTP() body = %v; want %v", rr.Body.String(), "Not Found")
	}
}

func TestGPTRouter_MethodNotAllowedHandler(t *testing.T) {
	r := NewRouter()
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed"))
	})

	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("GET")

	req, err := http.NewRequest("POST", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("ServeHTTP() = %v; want %v", rr.Code, http.StatusMethodNotAllowed)
	}
	if rr.Body.String() != "Method Not Allowed" {
		t.Fatalf("ServeHTTP() body = %v; want %v", rr.Body.String(), "Method Not Allowed")
	}
}

func TestGPTVars(t *testing.T) {
	r := NewRouter()
	r.HandleFunc("/test/{var}", func(w http.ResponseWriter, r *http.Request) {
		vars := Vars(r)
		if vars["var"] != "value" {
			t.Errorf("Vars() = %v; want %v", vars["var"], "value")
		}
	})

	req, err := http.NewRequest("GET", "/test/value", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
}

func TestGPTCurrentRoute(t *testing.T) {
	r := NewRouter()
	route := r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		currentRoute := CurrentRoute(r)
		if currentRoute == nil {
			t.Error("CurrentRoute() = nil; want non-nil")
		}
		if currentRoute.GetName() != "testRoute" {
			t.Errorf("CurrentRoute().GetName() = %v; want %v", currentRoute.GetName(), "testRoute")
		}
	}).Name("testRoute")

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if route.GetName() != "testRoute" {
		t.Fatalf("Route name = %v; want %v", route.GetName(), "testRoute")
	}
}

func TestGPTCurrentRouter(t *testing.T) {
	r := NewRouter()
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		currentRouter := CurrentRouter(r)
		if currentRouter == nil {
			t.Error("CurrentRouter() = nil; want non-nil")
		}
		if len(currentRouter.routes) != 1 {
			t.Errorf("CurrentRouter().routes = %v; want %v", len(currentRouter.routes), 1)
		}
	})

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
}

func TestGPTRoute_Headers(t *testing.T) {
	r := NewRouter()
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Headers("X-TestGPT-Header", "test-value")

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-TestGPT-Header", "test-value")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ServeHTTP() = %v; want %v", rr.Code, http.StatusOK)
	}
}

func TestGPTRoute_Host(t *testing.T) {
	r := NewRouter()
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Host("example.com")

	req, err := http.NewRequest("GET", "http://example.com/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ServeHTTP() = %v; want %v", rr.Code, http.StatusOK)
	}
}

func TestGPTRoute_Queries(t *testing.T) {
	r := NewRouter()
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Queries("foo", "bar")

	req, err := http.NewRequest("GET", "/test?foo=bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ServeHTTP() = %v; want %v", rr.Code, http.StatusOK)
	}
}

func TestGPTRoute_Schemes(t *testing.T) {
	r := NewRouter()
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Schemes("https")

	req, err := http.NewRequest("GET", "https://example.com/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ServeHTTP() = %v; want %v", rr.Code, http.StatusOK)
	}
}

func TestGPTCORSMethodMiddleware(t *testing.T) {
	r := NewRouter()
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET", "POST")

	r.Use(CORSMethodMiddleware(r))

	// Ensure OPTIONS method is handled by the router
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {}).Methods("OPTIONS")

	req, err := http.NewRequest("OPTIONS", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if allowMethods := rr.Header().Get("Access-Control-Allow-Methods"); allowMethods != "GET,POST,OPTIONS" {
		t.Fatalf("Access-Control-Allow-Methods = %v; want %v", allowMethods, "GET,POST,OPTIONS")
	}
}

// Additional tests for completeness

func TestRouter_ServeHTTP_NotFound(t *testing.T) {
	r := NewRouter()

	req, err := http.NewRequest("GET", "/notfound", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("ServeHTTP() = %v; want %v", rr.Code, http.StatusNotFound)
	}
}

