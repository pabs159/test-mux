package mux

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGPTRouteSkipClean(t *testing.T) {
	r := Route{routeConf: routeConf{skipClean: true}}
	if !r.SkipClean() {
		t.Errorf("Expected SkipClean to return true")
	}
}

func TestGPTRouteMatch(t *testing.T) {
	r := Route{
		handler: http.NotFoundHandler(),
		routeConf: routeConf{
			matchers: []matcher{
				methodMatcher([]string{"GET"}),
			},
		},
	}
	req, _ := http.NewRequest("GET", "http://localhost", nil)
	match := &RouteMatch{}
	if !r.Match(req, match) {
		t.Errorf("Expected Match to return true")
	}

	req, _ = http.NewRequest("POST", "http://localhost", nil)
	if r.Match(req, match) {
		t.Errorf("Expected Match to return false for POST method")
	}
}

func TestGPTRouteBuildOnly(t *testing.T) {
	r := Route{}
	r.BuildOnly()
	if !r.buildOnly {
		t.Errorf("Expected buildOnly to be true")
	}
}

func TestGPTRouteMetadata(t *testing.T) {
	r := Route{}
	r.Metadata("key", "value")
	if r.metadata["key"] != "value" {
		t.Errorf("Expected metadata value to be 'value'")
	}

	if !r.MetadataContains("key") {
		t.Errorf("Expected metadata to contain 'key'")
	}

	value, err := r.GetMetadataValue("key")
	if err != nil || value != "value" {
		t.Errorf("Expected GetMetadataValue to return 'value'")
	}

	value = r.GetMetadataValueOr("nonexistent", "default")
	if value != "default" {
		t.Errorf("Expected GetMetadataValueOr to return 'default'")
	}
}

func TestGPTRouteHandler(t *testing.T) {
	r := Route{}
	handler := http.NotFoundHandler()
	r.Handler(handler)

	// Test if the handler is set correctly by invoking it and checking the response
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	r.GetHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected handler to return status 404, got %v", rr.Code)
	}

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handlerFunc called"))
	}
	r.HandlerFunc(handlerFunc)
	//expectedHandler := http.HandlerFunc(handlerFunc)
	actualHandler := r.GetHandler()

	// Test if the handler function is set correctly by invoking it and checking the response
	rr = httptest.NewRecorder()
	actualHandler.ServeHTTP(rr, req)
	expectedResponse := "handlerFunc called"

	if rr.Body.String() != expectedResponse {
		t.Errorf("Expected handler function to be called correctly, got %v", rr.Body.String())
	}
}
func TestGPTRouteName(t *testing.T) {
	r := Route{namedRoutes: make(map[string]*Route)}
	r.Name("routeName")
	if r.name != "routeName" {
		t.Errorf("Expected name to be 'routeName'")
	}
	if r.namedRoutes["routeName"] != &r {
		t.Errorf("Expected namedRoutes to contain 'routeName'")
	}
}

func TestGPTRouteMatchers(t *testing.T) {
	r := Route{}
	r.Methods("GET", "POST")
	if len(r.matchers) != 1 {
		t.Errorf("Expected 1 matcher")
	}

	r.Headers("Content-Type", "application/json")
	if len(r.matchers) != 2 {
		t.Errorf("Expected 2 matchers")
	}

	r.Host("www.example.com")
	if len(r.matchers) != 3 {
		t.Errorf("Expected 3 matchers")
	}

	r.Queries("key", "value")
	if len(r.matchers) != 4 {
		t.Errorf("Expected 4 matchers")
	}
}

func TestGPTRouteSchemes(t *testing.T) {
	r := Route{}
	r.Schemes("http", "https")
	if len(r.matchers) != 1 {
		t.Errorf("Expected 1 matcher")
	}
}

func TestGPTRouteBuildVarsFunc(t *testing.T) {
	r := Route{}
	r.BuildVarsFunc(func(vars map[string]string) map[string]string {
		vars["key"] = "value"
		return vars
	})

	vars := map[string]string{}
	r.buildVars(vars)
	if vars["key"] != "value" {
		t.Errorf("Expected buildVarsFunc to modify vars")
	}
}

func TestRouteURLBuilding(t *testing.T) {
	// Initialize the regexp fields to avoid nil pointer dereference
	hostRegexp, _ := newRouteRegexp("{subdomain}.example.com", regexpTypeHost, routeRegexpOptions{})
	pathRegexp, _ := newRouteRegexp("/path/{var}", regexpTypePath, routeRegexpOptions{})

	r := Route{
		routeConf: routeConf{
			regexp: routeRegexpGroup{
				host: hostRegexp,
				path: pathRegexp,
			},
		},
	}

	builtURL, err := r.URL("subdomain", "test", "var", "value")
	if err != nil {
		t.Fatalf("Expected URL to build successfully, got error: %v", err)
	}
	expected := &url.URL{
		Scheme: "http",
		Host:   "test.example.com",
		Path:   "/path/value",
	}

	if builtURL.Scheme != expected.Scheme || builtURL.Host != expected.Host || builtURL.Path != expected.Path {
		t.Errorf("Expected URL to be %v, got %v", expected, builtURL)
	}

	_, err = r.URLHost("subdomain", "test")
	if err != nil {
		t.Fatalf("Expected URLHost to build successfully, got error: %v", err)
	}

	_, err = r.URLPath("var", "value")
	if err != nil {
		t.Fatalf("Expected URLPath to build successfully, got error: %v", err)
	}
}
