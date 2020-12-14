package smartlogic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

const (
	MaxAccessFailures = 3

	ConceptURIPrefix    = "http://www.ft.com/thing"
	MetadataFieldPrefix = "http://www.ft.com/ontology"
)

type Client struct {
	httpClient *http.Client

	baseAPIURL  url.URL
	apiTokenURL url.URL
	apiKey      string
	model       string

	accessToken string
}

func NewClient(ctx context.Context, httpClient *http.Client, baseCloudURL *url.URL, clientID, apiKey, model string) (*Client, error) {
	baseAPIURL := *baseCloudURL
	baseAPIURL.Path = path.Join(baseCloudURL.Path, fmt.Sprintf("/sw/client/%s/api", clientID))

	apiTokenURL := *baseCloudURL
	apiTokenURL.Path = path.Join(baseCloudURL.Path, "token")

	client := &Client{
		httpClient:  httpClient,
		baseAPIURL:  baseAPIURL,
		apiTokenURL: apiTokenURL,
		apiKey:      apiKey,
		model:       model,
	}

	accessToken, err := client.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	client.accessToken = accessToken

	return client, nil
}

// CreateConcept creates concept under given schema so the input concept should have schema defined.
func (c *Client) CreateConcept(ctx context.Context, concept Concept, task string) error {
	if concept.PrefLabel == "" {
		return errors.New("input concept should have prefLaber defined")
	}

	if concept.SchemaObject == "" && concept.Broader == "" {
		return errors.New("input concept should have either schema or broader relation defined")
	}

	if concept.Type == "" {
		return errors.New("input concept should have type defined")
	}

	// Construct the request url. It looks like smartlogicURL?path=task:MyModel:Mytask/skos:Concept/rdf:instance.
	reqURL := c.baseAPIURL
	path := fmt.Sprintf("task:%s:%s/skos:Concept/rdf:instance", c.model, task)
	// We don't want to encode the path param here.
	reqURL.RawQuery = fmt.Sprintf("path=%s", path)

	// Construct the request body, we really on the custom marshalling of the concept object.
	body, err := json.Marshal(concept)
	if err != nil {
		return fmt.Errorf("failed json encoding concept: %w", err)
	}
	resp, err := c.makeAuthorizedRequest(ctx, http.MethodPost, reqURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed creating new concept: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed creating new concept, returned status %v", resp.StatusCode)
	}

	return nil
}

func (c *Client) AddConceptMetadataField(ctx context.Context, conceptID, fieldName, fieldValue, task string) error {
	// Construct the request url. It looks like smartlogicURL?path=task:MyModel:Mytask/doubleEncodedConcept.
	reqURL := c.baseAPIURL

	conceptURI := ConceptURIPrefix + "/" + conceptID
	// Smartlogic API requires the conceptURI that is part of the path query param to be escaped twice and inside < >.
	encodedConceptURI := url.QueryEscape(url.QueryEscape(fmt.Sprintf("<%s>", conceptURI)))

	path := fmt.Sprintf("task:%s:%s/%s", c.model, task, encodedConceptURI)

	// We don't want to encode the path param here.
	reqURL.RawQuery = fmt.Sprintf("path=%s", path)

	// Construct the request body.
	fieldURI := MetadataFieldPrefix + "/" + fieldName
	bodyMap := map[string]string{
		"@id":    conceptURI,
		fieldURI: fieldValue,
	}
	body, err := json.Marshal(bodyMap)
	if err != nil {
		return fmt.Errorf("failed encoding metadata body: %w", err)
	}

	resp, err := c.makeAuthorizedRequest(ctx, http.MethodPost, reqURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed adding metadata to concept %s: %v", conceptID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed adding metadata to concept %s, returned status %v", conceptID, resp.StatusCode)
	}

	return nil
}

func (c *Client) GetConceptsWithCustomMetadata(ctx context.Context, task string, field string, value string) ([]interface{}, error) {
	params := url.Values{}
	params.Add("path", path.Join(
		fmt.Sprintf("task:%s:%s", c.model, task),
		"skos:Concept",
		"meta:transitiveInstance",
	))
	params.Add("properties", `rdf:type,meta:displayName,[]`)
	params.Add("filters", fmt.Sprintf(`subject(<%s>="%s")`, field, value))
	reqURL := c.baseAPIURL
	reqURL.RawQuery = params.Encode()

	resp, err := c.makeAuthorizedRequest(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make search request: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Graph []interface{} `json:"@graph"`
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("failed to read search response: %w", err)
	}

	return data.Graph, nil
}

func (c *Client) makeAuthorizedRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	for accessFailures := 0; accessFailures < MaxAccessFailures; accessFailures++ {
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, fmt.Errorf("failed creating authorized request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
		req.Header.Set("Content-Type", "application/ld+json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return resp, fmt.Errorf("failed making authorized request: %w", err)
		}

		// We're checking if we got a 401, which would be because the token had expired.
		// If it has, generate a new one and then make the request again.
		if resp.StatusCode == http.StatusUnauthorized {
			accessToken, err := c.getAccessToken(ctx)
			if err != nil {
				// close the body of the current request as it won't be read
				resp.Body.Close()
				// We got error 401 when making the request and we are not able to receive valid access token.
				return nil, errors.New("failed making request with valid access token")
			}
			c.accessToken = accessToken
			// close the body of the current request as it won't be read
			resp.Body.Close()
			// Try making the request with the fresh access token.
			continue
		}

		return resp, nil
	}
	return nil, errors.New("failed making request with valid access token")
}

func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "apikey")
	data.Set("key", c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiTokenURL.String(), bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return "", fmt.Errorf("failed creating access token request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed making access token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("access token request returned http status %v", resp.StatusCode)
	}

	tokenResp := struct {
		AccessToken string `json:"access_token"`
	}{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&tokenResp)
	if err != nil {
		return "", fmt.Errorf("failed decoding access token in response body: %w", err)
	}
	return tokenResp.AccessToken, nil
}
