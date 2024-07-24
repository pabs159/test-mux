package mux

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestGPTNewRouter tests the NewRouter function.
func TestGPTNewRouter(t *testing.T) {
	router := NewRouter()
	if router == nil {
		t.Error("Expected a new Router, got nil")
	}
	if len(router.namedRoutes) != 0 {
		t.Errorf("Expected empty namedRoutes, got %v", router.namedRoutes)
	}
}

// TestGPTStrictSlash tests the StrictSlash method.
func TestGPTStrictSlash(t *testing.T) {
	router := NewRouter().StrictSlash(true)
	if !router.strictSlash {
		t.Error("Expected StrictSlash to be true")
	}

	router = NewRouter().StrictSlash(false)
	if router.strictSlash {
		t.Error("Expected StrictSlash to be false")
	}
}

// TestGPTSkipClean tests the SkipClean method.
func TestGPTSkipClean(t *testing.T) {
	router := NewRouter().SkipClean(true)
	if !router.skipClean {
		t.Error("Expected SkipClean to be true")
	}

	router = NewRouter().SkipClean(false)
	if router.skipClean {
		t.Error("Expected SkipClean to be false")
	}
}

// TestGPTOmitRouteFromContext tests the OmitRouteFromContext method.
func TestGPTOmitRouteFromContext(t *testing.T) {
	router := NewRouter().OmitRouteFromContext(true)
	if !router.omitRouteFromContext {
		t.Error("Expected OmitRouteFromContext to be true")
	}

	router = NewRouter().OmitRouteFromContext(false)
	if router.omitRouteFromContext {
		t.Error("Expected OmitRouteFromContext to be false")
	}
}

// TestGPTOmitRouterFromContext tests the OmitRouterFromContext method.
func TestGPTOmitRouterFromContext(t *testing.T) {
	router := NewRouter().OmitRouterFromContext(true)
	if !router.omitRouterFromContext {
		t.Error("Expected OmitRouterFromContext to be true")
	}

	router = NewRouter().OmitRouterFromContext(false)
	if router.omitRouterFromContext {
		t.Error("Expected OmitRouterFromContext to be false")
	}
}

// TestGPTUseEncodedPath tests the UseEncodedPath method.
func TestGPTUseEncodedPath(t *testing.T) {
	router := NewRouter().UseEncodedPath()
	if !router.useEncodedPath {
		t.Error("Expected UseEncodedPath to be true")
	}
}

// TestGPTRouterMatch tests the Match method of the Router.
func TestGPTRouterMatch(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("/path", stringHandler("hello")).Methods("GET")

	req := newRequest("GET", "/path")
	var match RouteMatch
	if !router.Match(req, &match) {
		t.Errorf("Expected Match to be true")
	}
}

// TestGPTServeHTTP tests the ServeHTTP method of the Router.
func TestGPTServeHTTP(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req, _ := http.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, status)
	}

	req, _ = http.NewRequest("GET", "/notfound", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected status code %v, got %v", http.StatusNotFound, status)
	}
}

// TestGPTGet tests the Get method of the Router.
func TestGPTGet(t *testing.T) {
	router := NewRouter()
	router.Name("test").Path("/test")
	route := router.Get("test")
	if route == nil {
		t.Error("Expected to get the named route, got nil")
	}

	route = router.Get("nonexistent")
	if route != nil {
		t.Error("Expected to get nil for nonexistent route, got a route")
	}
}

// TestGPTBuildVarsFunc tests the BuildVarsFunc method of the Router.
/*func TestGPTBuildVarsFunc(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("/path/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := Vars(r)
		if vars["key"] != "value" {
			t.Errorf("Expected Vars[key] to be 'value', got %s", vars["key"])
		}
		w.WriteHeader(http.StatusOK)
	}).BuildVarsFunc(func(vars map[string]string) map[string]string {
		vars["key"] = "value"
		return vars
	})

	req := newRequest("GET", "/path/test")
	rec := NewRecorder()
	router.ServeHTTP(rec, req)
}
*/
// ExampleSetURLVars demonstrates how to set URL variables for a request.
func ExampleSetURLVarsGPT() {
	req, _ := http.NewRequest("GET", "/foo", nil)
	req = SetURLVars(req, map[string]string{"foo": "bar"})

	fmt.Println(Vars(req)["foo"])

	// Output: bar
}

// testMethodsSubrouter runs an individual methodsSubrouterTest.
func testMethodsSubrouterGPT(t *testing.T, test methodsSubrouterTest) {
	// Execute request
	req, _ := http.NewRequest(test.method, test.path, nil)
	resp := NewRecorder()
	test.router.ServeHTTP(resp, req)

	switch test.wantCode {
	case http.StatusMethodNotAllowed:
		if resp.Code != http.StatusMethodNotAllowed {
			t.Errorf(`(%s) Expected "405 Method Not Allowed", but got %d code`, test.title, resp.Code)
		} else if matchedMethod := resp.Body.String(); matchedMethod != "" {
			t.Errorf(`(%s) Expected "405 Method Not Allowed", but %q handler was called`, test.title, matchedMethod)
		}

	case http.StatusMovedPermanently:
		if gotLocation := resp.HeaderMap.Get("Location"); gotLocation != test.redirectTo {
			t.Errorf("(%s) Expected %q route-match to redirect to %q, but got %q", test.title, test.method, test.redirectTo, gotLocation)
		}

	case http.StatusOK:
		if matchedMethod := resp.Body.String(); matchedMethod != test.method {
			t.Errorf("(%s) Expected %q handler to be called, but %q handler was called", test.title, test.method, matchedMethod)
		}

	default:
		expectedCodes := []int{http.StatusMethodNotAllowed, http.StatusMovedPermanently, http.StatusOK}
		t.Errorf("(%s) Expected wantCode to be one of: %v, but got %d", test.title, expectedCodes, test.wantCode)
	}
}

func TestGPTSubrouterMatching(t *testing.T) {
	const (
		none, stdOnly, subOnly uint8 = 0, 1 << 0, 1 << 1
		both                         = subOnly | stdOnly
	)

	type request struct {
		Name    string
		Request *http.Request
		Flags   uint8
	}

	cases := []struct {
		Name                string
		Standard, Subrouter func(*Router)
		Requests            []request
	}{
		{
			"pathPrefix",
			func(r *Router) {
				r.PathPrefix("/before").PathPrefix("/after")
			},
			func(r *Router) {
				r.PathPrefix("/before").Subrouter().PathPrefix("/after")
			},
			[]request{
				{"no match final path prefix", newRequest("GET", "/after"), none},
				{"no match parent path prefix", newRequest("GET", "/before"), none},
				{"matches append", newRequest("GET", "/before/after"), both},
				{"matches as prefix", newRequest("GET", "/before/after/1234"), both},
			},
		},
		{
			"path",
			func(r *Router) {
				r.Path("/before").Path("/after")
			},
			func(r *Router) {
				r.Path("/before").Subrouter().Path("/after")
			},
			[]request{
				{"no match subroute path", newRequest("GET", "/after"), none},
				{"no match parent path", newRequest("GET", "/before"), none},
				{"no match as prefix", newRequest("GET", "/before/after/1234"), none},
				{"no match append", newRequest("GET", "/before/after"), none},
			},
		},
		{
			"host",
			func(r *Router) {
				r.Host("before.com").Host("after.com")
			},
			func(r *Router) {
				r.Host("before.com").Subrouter().Host("after.com")
			},
			[]request{
				{"no match before", newReqHostGPT("GET", "/", "before.com"), none},
				{"no match other", newReqHostGPT("GET", "/", "other.com"), none},
				{"matches after", newReqHostGPT("GET", "/", "after.com"), none},
			},
		},
		{
			"queries variant keys",
			func(r *Router) {
				r.Queries("foo", "bar").Queries("cricket", "baseball")
			},
			func(r *Router) {
				r.Queries("foo", "bar").Subrouter().Queries("cricket", "baseball")
			},
			[]request{
				{"matches with all", newRequest("GET", "/?foo=bar&cricket=baseball"), both},
				{"matches with more", newRequest("GET", "/?foo=bar&cricket=baseball&something=else"), both},
				{"no match with none", newRequest("GET", "/"), none},
				{"no match with some", newRequest("GET", "/?cricket=baseball"), none},
			},
		},
		{
			"queries overlapping keys",
			func(r *Router) {
				r.Queries("foo", "bar").Queries("foo", "baz")
			},
			func(r *Router) {
				r.Queries("foo", "bar").Subrouter().Queries("foo", "baz")
			},
			[]request{
				{"no match old value", newRequest("GET", "/?foo=bar"), none},
				{"no match diff value", newRequest("GET", "/?foo=bak"), none},
				{"no match with none", newRequest("GET", "/"), none},
				{"matches override", newRequest("GET", "/?foo=baz"), none},
			},
		},
		{
			"header variant keys",
			func(r *Router) {
				r.Headers("foo", "bar").Headers("cricket", "baseball")
			},
			func(r *Router) {
				r.Headers("foo", "bar").Subrouter().Headers("cricket", "baseball")
			},
			[]request{
				{
					"matches with all",
					newReqWithHeaders("GET", "/", "foo", "bar", "cricket", "baseball"),
					both,
				},
				{
					"matches with more",
					newReqWithHeaders("GET", "/", "foo", "bar", "cricket", "baseball", "something", "else"),
					both,
				},
				{"no match with none", newRequest("GET", "/"), none},
				{"no match with some", newReqWithHeaders("GET", "/", "cricket", "baseball"), none},
			},
		},
		{
			"header overlapping keys",
			func(r *Router) {
				r.Headers("foo", "bar").Headers("foo", "baz")
			},
			func(r *Router) {
				r.Headers("foo", "bar").Subrouter().Headers("foo", "baz")
			},
			[]request{
				{"no match old value", newReqWithHeaders("GET", "/", "foo", "bar"), none},
				{"no match diff value", newReqWithHeaders("GET", "/", "foo", "bak"), none},
				{"no match with none", newRequest("GET", "/"), none},
				{"matches override", newReqWithHeaders("GET", "/", "foo", "baz"), none},
			},
		},
		{
			"method",
			func(r *Router) {
				r.Methods("POST").Methods("GET")
			},
			func(r *Router) {
				r.Methods("POST").Subrouter().Methods("GET")
			},
			[]request{
				{"matches before", newRequest("POST", "/"), none},
				{"no match other", newRequest("HEAD", "/"), none},
				{"matches override", newRequest("GET", "/"), none},
			},
		},
		{
			"schemes",
			func(r *Router) {
				r.Schemes("http").Schemes("https")
			},
			func(r *Router) {
				r.Schemes("http").Subrouter().Schemes("https")
			},
			[]request{
				{"matches overrides", newRequest("GET", "https://www.example.com/"), none},
				{"matches original", newRequest("GET", "http://www.example.com/"), none},
				{"no match other", newRequest("GET", "ftp://www.example.com/"), none},
			},
		},
	}

	// case -> request -> router
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			for _, req := range c.Requests {
				t.Run(req.Name, func(t *testing.T) {
					for _, v := range []struct {
						Name     string
						Config   func(*Router)
						Expected bool
					}{
						{"subrouter", c.Subrouter, (req.Flags & subOnly) != 0},
						{"standard", c.Standard, (req.Flags & stdOnly) != 0},
					} {
						r := NewRouter()
						v.Config(r)
						if r.Match(req.Request, &RouteMatch{}) != v.Expected {
							if v.Expected {
								t.Errorf("expected %v match", v.Name)
							} else {
								t.Errorf("expected %v no match", v.Name)
							}
						}
					}
				})
			}
		})
	}
}

// verify that copyRouteConf copies fields as expected.
func TestGPT_copyRouteConf(t *testing.T) {
	var (
		m MatcherFunc = func(*http.Request, *RouteMatch) bool {
			return true
		}
		b BuildVarsFunc = func(i map[string]string) map[string]string {
			return i
		}
		r, _ = newRouteRegexp("hi", regexpTypeHost, routeRegexpOptions{})
	)

	tests := []struct {
		name string
		args routeConf
		want routeConf
	}{
		{
			"empty",
			routeConf{},
			routeConf{},
		},
		{
			"full",
			routeConf{
				useEncodedPath: true,
				strictSlash:    true,
				skipClean:      true,
				regexp:         routeRegexpGroup{host: r, path: r, queries: []*routeRegexp{r}},
				matchers:       []matcher{m},
				buildScheme:    "https",
				buildVarsFunc:  b,
			},
			routeConf{
				useEncodedPath: true,
				strictSlash:    true,
				skipClean:      true,
				regexp:         routeRegexpGroup{host: r, path: r, queries: []*routeRegexp{r}},
				matchers:       []matcher{m},
				buildScheme:    "https",
				buildVarsFunc:  b,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// special case some incomparable fields of routeConf before delegating to reflect.DeepEqual
			got := copyRouteConf(tt.args)

			// funcs not comparable, just compare length of slices
			if len(got.matchers) != len(tt.want.matchers) {
				t.Errorf("matchers different lengths: %v %v", len(got.matchers), len(tt.want.matchers))
			}
			got.matchers, tt.want.matchers = nil, nil

			// deep equal treats nil slice differently to empty slice so check for zero len first
			{
				bothZero := len(got.regexp.queries) == 0 && len(tt.want.regexp.queries) == 0
				if !bothZero && !reflect.DeepEqual(got.regexp.queries, tt.want.regexp.queries) {
					t.Errorf("queries unequal: %v %v", got.regexp.queries, tt.want.regexp.queries)
				}
				got.regexp.queries, tt.want.regexp.queries = nil, nil
			}

			// funcs not comparable, just compare nullity
			if (got.buildVarsFunc == nil) != (tt.want.buildVarsFunc == nil) {
				t.Errorf("build vars funcs unequal: %v %v", got.buildVarsFunc == nil, tt.want.buildVarsFunc == nil)
			}
			got.buildVarsFunc, tt.want.buildVarsFunc = nil, nil

			// finish the deal
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("route confs unequal: %v %v", got, tt.want)
			}
		})
	}
}

func TestGPTMethodNotAllowed(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }
	router := NewRouter()
	router.HandleFunc("/thing", handler).Methods(http.MethodGet)
	router.HandleFunc("/something", handler).Methods(http.MethodGet)

	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	w := NewRecorder()
	req := newRequest("PUT", "/thing")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected status code 405 (got %d)", w.Code)
	}
}

func TestGPTMethodNotAllowedSubrouterWithSeveralRoutes(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }

	router := NewRouter()
	subrouter := router.PathPrefix("/v1").Subrouter()
	subrouter.HandleFunc("/api", handler).Methods(http.MethodGet)
	subrouter.HandleFunc("/api/{id}", handler).Methods(http.MethodGet)

	subrouter.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	w := NewRecorder()
	req := newRequest("PUT", "/v1/api")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code 405 (got %d)", w.Code)
	}
}

type customMethodNotAllowedHandlerGPT struct {
	msg string
}

func (h customMethodNotAllowedHandlerGPT) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprint(w, h.msg)
}

func TestGPTSubrouterCustomMethodNotAllowed(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }

	router := NewRouter()
	router.HandleFunc("/test", handler).Methods(http.MethodGet)
	router.MethodNotAllowedHandler = customMethodNotAllowedHandlerGPT{msg: "custom router handler"}

	subrouter := router.PathPrefix("/sub").Subrouter()
	subrouter.HandleFunc("/test", handler).Methods(http.MethodGet)
	subrouter.MethodNotAllowedHandler = customMethodNotAllowedHandlerGPT{msg: "custom sub router handler"}

	testCases := map[string]struct {
		path   string
		expMsg string
	}{
		"router method not allowed": {
			path:   "/test",
			expMsg: "custom router handler",
		},
		"subrouter method not allowed": {
			path:   "/sub/test",
			expMsg: "custom sub router handler",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			w := NewRecorder()
			req := newRequest("PUT", tc.path)

			router.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				tt.Errorf("Expected status code 405 (got %d)", w.Code)
			}

			b, err := io.ReadAll(w.Body)
			if err != nil {
				tt.Errorf("failed to read body: %v", err)
			}

			if string(b) != tc.expMsg {
				tt.Errorf("expected msg %q, got %q", tc.expMsg, string(b))
			}
		})
	}
}

func TestGPTSubrouterNotFound(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }
	router := NewRouter()
	router.Path("/a").Subrouter().HandleFunc("/thing", handler).Methods(http.MethodGet)
	router.Path("/b").Subrouter().HandleFunc("/something", handler).Methods(http.MethodGet)

	w := NewRecorder()
	req := newRequest("PUT", "/not-present")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status code 404 (got %d)", w.Code)
	}
}

func TestGPTContextMiddleware(t *testing.T) {
	withTimeout := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
			defer cancel()
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	r := NewRouter()
	r.Handle("/path/{foo}", withTimeout(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := Vars(r)
		if vars["foo"] != "bar" {
			t.Fatal("Expected foo var to be set")
		}
	})))

	rec := NewRecorder()
	req := newRequest("GET", "/path/bar")
	r.ServeHTTP(rec, req)
}

func TestGPTGetVarNames(t *testing.T) {
	r := NewRouter()

	route := r.Host("{domain}").
		Path("/{group}/{item_id}").
		Queries("some_data1", "{some_data1}").
		Queries("some_data2_and_3", "{some_data2}.{some_data3}")

	// Order of vars in the slice is not guaranteed, so just check for existence
	expected := map[string]bool{
		"domain":     true,
		"group":      true,
		"item_id":    true,
		"some_data1": true,
		"some_data2": true,
		"some_data3": true,
	}

	varNames, err := route.GetVarNames()
	if err != nil {
		t.Fatal(err)
	}

	if len(varNames) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(varNames))
	}

	for _, varName := range varNames {
		if !expected[varName] {
			t.Fatalf("got unexpected %s", varName)
		}
	}
}

func getPopulateContextTestGPTCases() []struct {
	name                 string
	path                 string
	omitRouteFromContext bool
	wantVar              string
	wantStaticRoute      bool
	wantDynamicRoute     bool
} {
	return []struct {
		name                 string
		path                 string
		omitRouteFromContext bool
		wantVar              string
		wantStaticRoute      bool
		wantDynamicRoute     bool
	}{
		{
			name:            "no populated vars",
			path:            "/static",
			wantVar:         "",
			wantStaticRoute: true,
		},
		{
			name:             "empty var",
			path:             "/dynamic/",
			wantVar:          "",
			wantDynamicRoute: true,
		},
		{
			name:             "populated vars",
			path:             "/dynamic/foo",
			wantVar:          "foo",
			wantDynamicRoute: true,
		},
		{
			name:                 "omit route /static",
			path:                 "/static",
			omitRouteFromContext: true,
			wantVar:              "",
			wantStaticRoute:      false,
		},
		{
			name:                 "omit route /dynamic",
			path:                 "/dynamic/",
			omitRouteFromContext: true,
			wantVar:              "",
			wantDynamicRoute:     false,
		},
	}
}

func TestGPTPopulateContext(t *testing.T) {
	testCases := getPopulateContextTestGPTCases()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matched := false
			r := NewRouter()
			r.OmitRouteFromContext(tc.omitRouteFromContext)
			var static *Route
			var dynamic *Route
			fn := func(w http.ResponseWriter, r *http.Request) {
				matched = true
				if got := Vars(r)["x"]; got != tc.wantVar {
					t.Fatalf("wantVar=%q, got=%q", tc.wantVar, got)
				}
				switch {
				case tc.wantDynamicRoute:
					r2 := CurrentRoute(r)
					if r2 != dynamic || r2.GetName() != "dynamic" {
						t.Fatalf("expected dynamic route in ctx, got %v", r2)
					}
				case tc.wantStaticRoute:
					r2 := CurrentRoute(r)
					if r2 != static || r2.GetName() != "static" {
						t.Fatalf("expected static route in ctx, got %v", r2)
					}
				default:
					if r2 := CurrentRoute(r); r2 != nil {
						t.Fatalf("expected no route in ctx, got %v", r2)
					}
				}
				w.WriteHeader(http.StatusNoContent)
			}
			static = r.Name("static").Path("/static").HandlerFunc(fn)
			dynamic = r.Name("dynamic").Path("/dynamic/{x:.*}").HandlerFunc(fn)
			req := newRequest("GET", "http://localhost"+tc.path)
			rec := NewRecorder()
			r.ServeHTTP(rec, req)
			if !matched {
				t.Fatal("Expected route to match")
			}
		})
	}
}

func BenchmarkPopulateContextGPT(b *testing.B) {
	testCases := getPopulateContextTestGPTCases()
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			matched := false
			r := NewRouter()
			r.OmitRouteFromContext(tc.omitRouteFromContext)
			fn := func(w http.ResponseWriter, r *http.Request) {
				matched = true
				w.WriteHeader(http.StatusNoContent)
			}
			r.Name("static").Path("/static").HandlerFunc(fn)
			r.Name("dynamic").Path("/dynamic/{x:.*}").HandlerFunc(fn)
			req := newRequest("GET", "http://localhost"+tc.path)
			rec := NewRecorder()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.ServeHTTP(rec, req)
			}
			if !matched {
				b.Fatal("Expected route to match")
			}
		})
	}
}

// mapToPairsGpt converts a string map to a slice of string pairs
func mapToPairsGPT(m map[string]string) []string {
	var i int
	p := make([]string, len(m)*2)
	for k, v := range m {
		p[i] = k
		p[i+1] = v
		i += 2
	}
	return p
}

// stringMapEqualGpt checks the equality of two string maps
func stringMapEqualGPT(m1, m2 map[string]string) bool {
	nil1 := m1 == nil
	nil2 := m2 == nil
	if nil1 != nil2 || len(m1) != len(m2) {
		return false
	}
	for k, v := range m1 {
		if v != m2[k] {
			return false
		}
	}
	return true
}

// stringHandler returns a handler func that writes a message 's' to the
// http.ResponseWriter.
func stringHandlerGPT(s string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(s))
		if err != nil {
			log.Printf("Failed writing HTTP response: %v", err)
		}
	}
}

// newRequestGPT is a helper function to create a new request with a method and url.
// The request returned is a 'server' request as opposed to a 'client' one through
// simulated write onto the wire and read off of the wire.
// The differences between requests are detailed in the net/http package.
func newRequestGPT(method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	// extract the escaped original host+path from url
	// http://localhost/path/here?v=1#frag -> //localhost/path/here
	opaque := ""
	if i := len(req.URL.Scheme); i > 0 {
		opaque = url[i+1:]
	}

	if i := strings.LastIndex(opaque, "?"); i > -1 {
		opaque = opaque[:i]
	}
	if i := strings.LastIndex(opaque, "#"); i > -1 {
		opaque = opaque[:i]
	}

	// Escaped host+path workaround as detailed in https://golang.org/pkg/net/url/#URL
	// for < 1.5 client side workaround
	req.URL.Opaque = opaque

	// Simulate writing to wire
	var buff bytes.Buffer
	err = req.Write(&buff)
	if err != nil {
		log.Printf("Failed writing HTTP request: %v", err)
	}
	ioreader := bufio.NewReader(&buff)

	// Parse request off of 'wire'
	req, err = http.ReadRequest(ioreader)
	if err != nil {
		panic(err)
	}
	return req
}

// create a new request with the provided headers
func newReqWithHeaders(method, url string, headers ...string) *http.Request {
	req := newRequestGPT(method, url)

	if len(headers)%2 != 0 {
		panic(fmt.Sprintf("Expected headers length divisible by 2 but got %v", len(headers)))
	}

	for i := 0; i < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}

	return req
}

// newReqHostGPT a new request with a method, url, and host header
func newReqHostGPT(method, url, host string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	req.Host = host
	return req
}

