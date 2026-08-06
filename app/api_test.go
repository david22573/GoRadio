package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSearchStations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock upstream server
	var requestedPath string
	var returnStatus int = http.StatusOK
	var returnBody []byte = []byte(`[{"name": "test station", "stationuuid": "123"}]`)
	
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.String()
		w.WriteHeader(returnStatus)
		w.Write(returnBody)
	}))
	defer ts.Close()

	RadioBrowserBaseURL = ts.URL
	defer func() { RadioBrowserBaseURL = "" }()

	handler := &APIHandler{
		cache: newSearchCache(5 * time.Minute),
	}

	tests := []struct {
		name           string
		query          string
		page           string
		limit          string
		setupMock      func()
		expectedStatus int
		expectedPage   int
		expectedLimit  int
		expectedMore   bool
		expectedURL    string
	}{
		{
			name:           "default pagination",
			query:          "jazz",
			setupMock:      func() { returnStatus = 200; returnBody = []byte(`[{},{},{},{},{},{},{},{},{},{},{},{}]`) },
			expectedStatus: 200,
			expectedPage:   1,
			expectedLimit:  12,
			expectedMore:   true,
			expectedURL:    "/json/stations/search?name=jazz&limit=12&offset=0&hidebroken=true",
		},
		{
			name:           "custom pagination",
			query:          "jazz",
			page:           "3",
			limit:          "15",
			setupMock:      func() { returnStatus = 200; returnBody = []byte(`[{}]`) },
			expectedStatus: 200,
			expectedPage:   3,
			expectedLimit:  15,
			expectedMore:   false,
			expectedURL:    "/json/stations/search?name=jazz&limit=15&offset=30&hidebroken=true",
		},
		{
			name:           "invalid pagination uses defaults",
			query:          "rock",
			page:           "invalid",
			limit:          "-5",
			setupMock:      func() { returnStatus = 200; returnBody = []byte(`[]`) },
			expectedStatus: 200,
			expectedPage:   1,
			expectedLimit:  12,
			expectedMore:   false,
			expectedURL:    "/json/stations/search?name=rock&limit=12&offset=0&hidebroken=true",
		},
		{
			name:           "max limit is 48",
			query:          "pop",
			limit:          "100",
			setupMock:      func() { returnStatus = 200; returnBody = []byte(`[]`) },
			expectedStatus: 200,
			expectedPage:   1,
			expectedLimit:  48,
			expectedMore:   false,
			expectedURL:    "/json/stations/search?name=pop&limit=48&offset=0&hidebroken=true",
		},
		{
			name:           "missing query",
			query:          "",
			setupMock:      func() {},
			expectedStatus: 400,
		},
		{
			name:           "upstream error",
			query:          "error",
			setupMock:      func() { returnStatus = 500; returnBody = []byte(`error`) },
			expectedStatus: 500,
		},
		{
			name:           "invalid upstream json",
			query:          "badjson",
			setupMock:      func() { returnStatus = 200; returnBody = []byte(`{invalid`) },
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			requestedPath = "" // Reset

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			
			q := url.Values{}
			if tt.query != "" {
				q.Add("q", tt.query)
			}
			if tt.page != "" {
				q.Add("page", tt.page)
			}
			if tt.limit != "" {
				q.Add("limit", tt.limit)
			}
			
			c.Request, _ = http.NewRequest("GET", "/api/search?"+q.Encode(), nil)
			handler.SearchStations(c)

			if w.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == 200 {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}

				if int(response["page"].(float64)) != tt.expectedPage {
					t.Errorf("expected page %d, got %v", tt.expectedPage, response["page"])
				}
				if int(response["limit"].(float64)) != tt.expectedLimit {
					t.Errorf("expected limit %d, got %v", tt.expectedLimit, response["limit"])
				}
				if response["has_more"].(bool) != tt.expectedMore {
					t.Errorf("expected has_more %v, got %v", tt.expectedMore, response["has_more"])
				}
				if requestedPath != tt.expectedURL {
					t.Errorf("expected url %s, got %s", tt.expectedURL, requestedPath)
				}
			}
		})
	}
}

func TestSearchStationsCacheIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var returnBody []byte
	requestCount := 0
	
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write(returnBody)
	}))
	defer ts.Close()

	RadioBrowserBaseURL = ts.URL
	defer func() { RadioBrowserBaseURL = "" }()

	handler := &APIHandler{
		cache: newSearchCache(5 * time.Minute),
	}

	performRequest := func(query, page, limit string) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		q := url.Values{}
		q.Add("q", query)
		if page != "" {
			q.Add("page", page)
		}
		if limit != "" {
			q.Add("limit", limit)
		}
		c.Request, _ = http.NewRequest("GET", "/api/search?"+q.Encode(), nil)
		handler.SearchStations(c)
	}

	// Request 1: page 1
	returnBody = []byte(`[{"id": 1}]`)
	performRequest("jazz", "1", "10")
	if requestCount != 1 {
		t.Fatalf("expected 1 request, got %d", requestCount)
	}

	// Request 2: page 1 again (should be cached)
	performRequest("jazz", "1", "10")
	if requestCount != 1 {
		t.Fatalf("expected 1 request due to cache, got %d", requestCount)
	}

	// Request 3: page 2 (should bypass cache)
	returnBody = []byte(`[{"id": 2}]`)
	performRequest("jazz", "2", "10")
	if requestCount != 2 {
		t.Fatalf("expected 2 requests, got %d", requestCount)
	}
}
