package smartlogic

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		expectedError bool
	}{
		{
			name: "success",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path != "/token" {
					w.WriteHeader(http.StatusInternalServerError)
				}
				handleTokenRequest(t, w)
				// nothing given as status code returns 200
			}),
			expectedError: false,
		},
		{
			name: "cannot get access token",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.String() != "/token" {
					return
				}
				w.WriteHeader(http.StatusUnauthorized)
			}),
			expectedError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testServer := httptest.NewServer(test.serverHandler)
			serverURL, err := url.Parse(testServer.URL)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.TODO()
			_, err = NewClient(ctx, testServer.Client(), serverURL, "test", "test", "test")
			if err != nil && !test.expectedError {
				t.Errorf("unexpected error creating client: %v", err)
			}
			if err == nil && test.expectedError {
				t.Errorf("expected error creating client")
			}
			testServer.Close()
		})
	}
}

func TestAddConceptMetadataFieldRequestURIAndBody(t *testing.T) {
	testServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/token" {
				handleTokenRequest(t, w)
			}
			if req.URL.Path == "/sw/client/testClientID/api" &&
				req.URL.RawQuery == "path=task:testModel:testTask/%253Chttp%253A%252F%252Fwww.ft.com%252Fthing%252F7bcfe07b-0fb1-49ce-a5fa-e51d5c01c3e0%253E" {
				reqMap := make(map[string]string)
				err := json.NewDecoder(req.Body).Decode(&reqMap)
				if err != nil {
					t.Errorf("invalid body send on add concept metadata: %v", err)
				}
				if reqMap["@id"] != "http://www.ft.com/thing/7bcfe07b-0fb1-49ce-a5fa-e51d5c01c3e0" {
					t.Error("invalid body send on add concept metadata, invalid concept id")
				}
				if reqMap["http://www.ft.com/ontology/factsetIdentifier"] != "0DR49W-E" {
					t.Error("invalid body send on add concept metadata, invalid metadata field")
				}
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}))
	defer testServer.Close()

	serverURL, err := url.Parse(testServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.TODO()

	client, err := NewClient(ctx, testServer.Client(), serverURL, "testClientID", "testAPIKey", "testModel")
	if err != nil {
		t.Fatalf("failed creating Smartlogic client: %v", err)
	}

	err = client.AddConceptMetadataField(ctx, "7bcfe07b-0fb1-49ce-a5fa-e51d5c01c3e0", "factsetIdentifier", "0DR49W-E", "testTask")
	if err != nil {
		t.Errorf("failed adding concept metadata field: %v", err)
	}
}

func TestClientAddConceptMetadataField(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		expectedError bool
	}{
		{
			name: "success with 200",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/token" {
					handleTokenRequest(t, w)
				}
			}),
			expectedError: false,
		},
		{
			name: "success with 201",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/token" {
					handleTokenRequest(t, w)
				}
				w.WriteHeader(http.StatusCreated)
			}),
			expectedError: false,
		},
		{
			name: "non-200 status",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/token" {
					handleTokenRequest(t, w)
				}
				w.WriteHeader(http.StatusInternalServerError)
			}),
			expectedError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testServer := httptest.NewServer(test.serverHandler)
			serverURL, err := url.Parse(testServer.URL)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.TODO()

			client, err := NewClient(ctx, testServer.Client(), serverURL, "test", "test", "test")
			if err != nil {
				t.Fatalf("failed creating Smartlogic client: %v", err)
			}
			err = client.AddConceptMetadataField(ctx, "conceptID", "factsetIdentifier", "factsetID", "testTask")
			if err != nil && !test.expectedError {
				t.Errorf("unexpected error adding concept metadata field: %v", err)
			}
			if err == nil && test.expectedError {
				t.Errorf("expected error adding concept metadata field")
			}
			testServer.Close()
		})
	}
}

func TestClientCreateConceptRequestURIAndBody(t *testing.T) {
	testServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/token" {
				handleTokenRequest(t, w)
			}
			if req.URL.Path == "/sw/client/testClientID/api" &&
				req.URL.RawQuery == "path=task:testModel:testTask/skos:Concept/rdf:instance" {
				body, err := ioutil.ReadAll(req.Body)
				if err != nil {
					t.Errorf("invalid body send on add concept: %v", err)
				}
				if string(body) != `{"skosxl:prefLabel":[{"skosxl:literalForm":[{"@value":"Test Pref Label","@language":"en"}],"@type":["skosxl:Label"]}],"@type":["skos:Concept","Test Type"],"skos:topConceptOf":{"@id":"Test Concept Schema"}}` {
					t.Errorf("invalid body send on add concept: got %v", string(body))
				}
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}))
	defer testServer.Close()

	serverURL, err := url.Parse(testServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.TODO()

	client, err := NewClient(ctx, testServer.Client(), serverURL, "testClientID", "testAPIKey", "testModel")
	if err != nil {
		t.Fatalf("failed creating Smartlogic client: %v", err)
	}

	concept := Concept{
		PrefLabel:    "Test Pref Label",
		Type:         "Test Type",
		SchemaObject: "Test Concept Schema",
	}
	err = client.CreateConcept(ctx, concept, "testTask")
	if err != nil {
		t.Errorf("failed adding concept metadata field: %v", err)
	}
}

func TestClientCreateConcept(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		concept       Concept
		expectedError bool
	}{
		{
			name: "success with 200",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/token" {
					handleTokenRequest(t, w)
				}
			}),
			concept: Concept{
				PrefLabel:    "Test Pref Label",
				Type:         "Test Type",
				SchemaObject: "Test Concept Schema",
			},
			expectedError: false,
		},
		{
			name: "success with 201",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/token" {
					handleTokenRequest(t, w)
				}
			}),
			concept: Concept{
				PrefLabel:    "Test Pref Label",
				Type:         "Test Type",
				SchemaObject: "Test Concept Schema",
			},
			expectedError: false,
		},
		{
			name: "invalid concept - no concept pref label",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/token" {
					handleTokenRequest(t, w)
				}
			}),
			concept: Concept{
				PrefLabel:    "",
				Type:         "Test Type",
				SchemaObject: "Test Concept Schema",
			},
			expectedError: true,
		},
		{
			name: "invalid concept - no concept type",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/token" {
					handleTokenRequest(t, w)
				}
			}),
			concept: Concept{
				PrefLabel:    "Test Pref Label",
				Type:         "",
				SchemaObject: "Test Concept Schema",
			},
			expectedError: true,
		},
		{
			name: "invalid concept - no concept schema",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/token" {
					handleTokenRequest(t, w)
				}
			}),
			concept: Concept{
				PrefLabel:    "Test Pref Label",
				Type:         "Test Type",
				SchemaObject: "",
			},
			expectedError: true,
		},
		{
			name: "success with 201",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/token" {
					handleTokenRequest(t, w)
				}
			}),
			concept: Concept{
				PrefLabel:    "Test Pref Label",
				Type:         "Test Type",
				SchemaObject: "Test Concept Schema",
			},
			expectedError: false,
		},
		{
			name: "non 200 response",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/token" {
					handleTokenRequest(t, w)
				}
				w.WriteHeader(http.StatusInternalServerError)
			}),
			concept: Concept{
				PrefLabel:    "Test Pref Label",
				Type:         "Test Type",
				SchemaObject: "Test Concept Schema",
			},
			expectedError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testServer := httptest.NewServer(test.serverHandler)
			serverURL, err := url.Parse(testServer.URL)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.TODO()

			client, err := NewClient(ctx, testServer.Client(), serverURL, "test", "test", "test")
			if err != nil {
				t.Fatalf("failed creating Smartlogic client: %v", err)
			}
			err = client.CreateConcept(ctx, test.concept, "testTask")
			if err != nil && !test.expectedError {
				t.Errorf("unexpected error adding new concept: %v", err)
			}
			if err == nil && test.expectedError {
				t.Errorf("expected error adding new concept")
			}
			testServer.Close()
		})
	}
}

func handleTokenRequest(t *testing.T, w http.ResponseWriter) {
	token := struct {
		AccessToken string `json:"access_token"`
	}{"test_token"}
	tokenResp, err := json.Marshal(token)
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.Write(tokenResp)
	if err != nil {
		t.Fatal(t)
	}
}
