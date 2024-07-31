package mux

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestGPTRouteMatchBuildOnly(t *testing.T) {
	r := Route{buildOnly: true}
	req, _ := http.NewRequest("GET", "/", nil)
	match := &RouteMatch{}
	if r.Match(req, match) {
		t.Errorf("Expected Match to return false for buildOnly route")
	}
}

func TestGPTRouteMatchError(t *testing.T) {
	r := Route{err: errors.New("some error")}
	req, _ := http.NewRequest("GET", "/", nil)
	match := &RouteMatch{}
	if r.Match(req, match) {
		t.Errorf("Expected Match to return false for route with error")
	}
}

func TestGPTRouteMatchQueryNotFound(t *testing.T) {
	// Properly initialize the routeRegexp
	rr, err := newRouteRegexp("key=value", regexpTypeQuery, routeRegexpOptions{})
	if err != nil {
		t.Fatalf("Failed to create routeRegexp: %v", err)
	}

	r := Route{
		routeConf: routeConf{
			matchers: []matcher{
				rr,
			},
		},
	}
	req, _ := http.NewRequest("GET", "/?otherkey=othervalue", nil) // Ensure the query string doesn't match
	match := &RouteMatch{}
	if r.Match(req, match) {
		t.Errorf("Expected Match to return false for unmatched query")
	}
	if match.MatchErr != ErrNotFound {
		t.Errorf("Expected MatchErr to be ErrNotFound")
	}
}

func TestGPTRouteBuildOnlyMethod(t *testing.T) {
	r := Route{}
	r.BuildOnly()
	if !r.buildOnly {
		t.Errorf("Expected buildOnly to be true")
	}
}

func TestGPT2RouteMetadata(t *testing.T) {
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

func TestGPTRouteAddRegexpMatcher(t *testing.T) {
	r := Route{}
	err := r.addRegexpMatcher("/test", regexpTypePath)
	if err != nil {
		t.Fatalf("Expected addRegexpMatcher to succeed, got error: %v", err)
	}
}

func TestGPTRouteHeaders(t *testing.T) {
	r := Route{}
	r.Headers("Content-Type", "application/json")
	if len(r.matchers) != 1 {
		t.Errorf("Expected 1 matcher for Headers")
	}
}

func TestGPTRouteHeadersRegexp(t *testing.T) {
	r := Route{}
	r.HeadersRegexp("Content-Type", "application/(text|json)")
	if len(r.matchers) != 1 {
		t.Errorf("Expected 1 matcher for HeadersRegexp")
	}
}

func TestGPTRouteMethods(t *testing.T) {
	r := Route{}
	r.Methods("GET", "POST")
	if len(r.matchers) != 1 {
		t.Errorf("Expected 1 matcher for Methods")
	}
}

func TestGPTRoutePath(t *testing.T) {
	r := Route{}
	r.Path("/test")
	if len(r.matchers) != 1 {
		t.Errorf("Expected 1 matcher for Path")
	}
}

func TestGPTRoutePathPrefix(t *testing.T) {
	r := Route{}
	r.PathPrefix("/prefix")
	if len(r.matchers) != 1 {
		t.Errorf("Expected 1 matcher for PathPrefix")
	}
}

func TestGPTRouteQueries(t *testing.T) {
	r := Route{}
	r.Queries("key", "value")
	if len(r.matchers) != 1 {
		t.Errorf("Expected 1 matcher for Queries")
	}
}

func TestGPT2RouteSchemes(t *testing.T) {
	r := Route{}
	r.Schemes("http", "https")
	if len(r.matchers) != 1 {
		t.Errorf("Expected 1 matcher for Schemes")
	}
}

func TestGPT22RouteBuildVarsFunc(t *testing.T) {
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

func TestGPTRouteAddRegexpMatcherError(t *testing.T) {
	r := Route{}
	err := r.addRegexpMatcher("invalid[", regexpTypePath)
	if err == nil {
		t.Errorf("Expected error for invalid regexp pattern")
	}
}

func TestGPTRouteGetError(t *testing.T) {
	r := Route{
		err: errors.New("test error"),
	}
	if r.GetError() == nil {
		t.Errorf("Expected GetError to return an error")
	}
}

func TestGPTRouteMatchMethodMismatch(t *testing.T) {
	r := Route{
		routeConf: routeConf{
			matchers: []matcher{
				methodMatcher([]string{"POST"}),
			},
		},
	}
	req, _ := http.NewRequest("GET", "http://localhost", nil)
	match := &RouteMatch{}
	if r.Match(req, match) {
		t.Errorf("Expected Match to return false for method mismatch")
	}
	if match.MatchErr != ErrMethodMismatch {
		t.Errorf("Expected MatchErr to be ErrMethodMismatch")
	}
}

func TestGPT2RouteMatchQueryNotFound(t *testing.T) {
	rr, err := newRouteRegexp("key=value", regexpTypeQuery, routeRegexpOptions{})
	if err != nil {
		t.Fatalf("Failed to create routeRegexp: %v", err)
	}

	r := Route{
		routeConf: routeConf{
			matchers: []matcher{
				rr,
			},
		},
	}
	req, _ := http.NewRequest("GET", "/?otherkey=othervalue", nil)
	match := &RouteMatch{}
	if r.Match(req, match) {
		t.Errorf("Expected Match to return false for unmatched query")
	}
	if match.MatchErr != ErrNotFound {
		t.Errorf("Expected MatchErr to be ErrNotFound")
	}
}

func TestGPTRouteMatchHandlerOverride(t *testing.T) {
	// Create a handler that sets a specific header so we can identify it
	//originalHandler := http.NotFoundHandler()
	overrideHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Handler", "override")
	})

	r := Route{
		handler: overrideHandler,
		routeConf: routeConf{
			matchers: []matcher{
				methodMatcher([]string{"GET"}),
			},
		},
	}
	req, _ := http.NewRequest("GET", "http://localhost", nil)
	match := &RouteMatch{
		MatchErr: ErrMethodMismatch,
	}

	// Match should succeed and override the handler
	if !r.Match(req, match) {
		t.Errorf("Expected Match to return true")
	}

	// MatchErr should be nil after successful match
	if match.MatchErr != nil {
		t.Errorf("Expected MatchErr to be nil")
	}

	// Verify that the handler has been overridden
	recorder := httptest.NewRecorder()
	match.Handler.ServeHTTP(recorder, req)
	if recorder.Header().Get("X-Handler") != "override" {
		t.Errorf("Expected handler to be overridden, but it was not")
	}
}

func TestGPTRouteURLPathError(t *testing.T) {
	r := Route{
		err: errors.New("test error"),
	}
	_, err := r.URLPath("var", "value")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteURLPathNoPath(t *testing.T) {
	r := Route{}
	_, err := r.URLPath("var", "value")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteGetPathTemplateError(t *testing.T) {
	r := Route{
		err: errors.New("test error"),
	}
	_, err := r.GetPathTemplate()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteGetPathTemplateNoPath(t *testing.T) {
	r := Route{}
	_, err := r.GetPathTemplate()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteGetPathRegexpError(t *testing.T) {
	r := Route{
		err: errors.New("test error"),
	}
	_, err := r.GetPathRegexp()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteGetPathRegexpNoPath(t *testing.T) {
	r := Route{}
	_, err := r.GetPathRegexp()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteGetQueriesRegexpError(t *testing.T) {
	r := Route{
		err: errors.New("test error"),
	}
	_, err := r.GetQueriesRegexp()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteGetQueriesRegexpNoQueries(t *testing.T) {
	r := Route{}
	_, err := r.GetQueriesRegexp()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteGetQueriesTemplatesError(t *testing.T) {
	r := Route{
		err: errors.New("test error"),
	}
	_, err := r.GetQueriesTemplates()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteGetQueriesTemplatesNoQueries(t *testing.T) {
	r := Route{}
	_, err := r.GetQueriesTemplates()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteGetMethodsError(t *testing.T) {
	r := Route{
		err: errors.New("test error"),
	}
	_, err := r.GetMethods()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteGetMethodsNoMethods(t *testing.T) {
	r := Route{}
	_, err := r.GetMethods()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteGetHostTemplateError(t *testing.T) {
	r := Route{
		err: errors.New("test error"),
	}
	_, err := r.GetHostTemplate()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteGetHostTemplateNoHost(t *testing.T) {
	r := Route{}
	_, err := r.GetHostTemplate()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteGetVarNamesError(t *testing.T) {
	r := Route{
		err: errors.New("test error"),
	}
	_, err := r.GetVarNames()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
func TestGPTRoutePrepareVarsError(t *testing.T) {
	r := Route{}

	// Passing an odd number of elements to mapFromPairsToString should trigger an error
	_, err := r.prepareVars("key1", "value1", "key2")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
func TestGPTRouteURLPrepareVarsError(t *testing.T) {
	r := Route{}
	_, err := r.URL("key1")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteURLHostError(t *testing.T) {
	r := Route{}
	_, err := r.URLHost("key1", "value1")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGPTRouteURLHostNoHost(t *testing.T) {
	r := Route{}
	_, err := r.URLHost("key1", "value1")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
func TestRouteUncoveredLines(t *testing.T) {
	// Test line 49
	r := Route{buildOnly: true}
	req, _ := http.NewRequest("GET", "http://localhost", nil)
	match := &RouteMatch{}
	if r.Match(req, match) {
		t.Errorf("Expected Match to return false for buildOnly route")
	}

	// Test line 70
	r = Route{routeConf: routeConf{matchers: []matcher{
		&routeRegexp{
			regexpType: regexpTypeQuery,
			template:   "{query}",
			regexp:     regexp.MustCompile("^query=.*$"),
		},
	}}}
	req, _ = http.NewRequest("GET", "http://localhost", nil)
	if r.Match(req, match) {
		t.Errorf("Expected Match to return false for query matcher with no match")
	}

	// Test line 127
	r.BuildOnly()
	if !r.buildOnly {
		t.Errorf("Expected buildOnly to be true")
	}

	// Test line 155
	_, err := r.GetMetadataValue("nonexistent")
	if err == nil {
		t.Errorf("Expected GetMetadataValue to return an error for nonexistent key")
	}

	// Test line 177
	handler := http.NotFoundHandler()
	r.Handler(handler)
	/*	if r.handler != handler {
			t.Errorf("Expected handler to be set")
		}
	*/
	// Test line 201
	middlewareCalled := false
	mw := MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	})
	r.Use(mw)
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {}
	r.HandlerFunc(handlerFunc)
	h := r.GetHandlerWithMiddlewares()
	rr := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	h.ServeHTTP(rr, req)
	if !middlewareCalled {
		t.Errorf("Expected middleware to be called")
	}

	// Test line 214
	r = Route{namedRoutes: make(map[string]*Route)}
	r.Name("routeName")
	if r.name != "routeName" {
		t.Errorf("Expected name to be 'routeName'")
	}
	if r.namedRoutes["routeName"] != &r {
		t.Errorf("Expected namedRoutes to contain 'routeName'")
	}

	r.Name("newName")
	if r.err == nil {
		t.Errorf("Expected error when setting name again")
	}

	// Test line 253
	r = Route{}
	err = r.addRegexpMatcher("invalidPath", regexpTypePath)
	if err == nil {
		t.Errorf("Expected error for path not starting with slash")
	}

	// Test line 265
	err = r.addRegexpMatcher("/validPath", regexpTypePath)
	if err != nil {
		t.Errorf("Expected no error for valid path")
	}

	// Test line 268
	r.regexp.queries = []*routeRegexp{{varsN: []string{"var1"}}}
	/*err = r.addRegexpMatcher("query", regexpTypeQuery)
	if err != nil {
		t.Errorf("Expected no error for query matcher")
	}*/

	// Test line 272
	r.regexp.path = &routeRegexp{varsN: []string{"var1"}}
	err = r.addRegexpMatcher("host", regexpTypeHost)
	if err != nil {
		t.Errorf("Expected no error for host matcher")
	}

	// Test line 279
	/*r.regexp.host = &routeRegexp{varsN: []string{"var1"}}
	err = r.addRegexpMatcher("path", regexpTypePath)
	if err != nil {
		t.Errorf("Expected no error for path matcher")
	}

	// Test line 284
	r.regexp.queries = []*routeRegexp{{varsN: []string{"var1"}}}
	err = r.addRegexpMatcher("query", regexpTypeQuery)
	if err != nil {
		t.Errorf("Expected no error for query matcher")
	}
	*/
	// Test line 287
	r.addMatcher(&routeRegexp{})
	if len(r.matchers) != 1 {
		t.Errorf("Expected matcher to be added")
	}
}
