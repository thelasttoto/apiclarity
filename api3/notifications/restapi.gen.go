// Package notifications provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.1-0.20220609223533-7da811e1cf30 DO NOT EDIT.
package notifications

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	externalRef0 "github.com/openclarity/apiclarity/api3/common"
	externalRef1 "github.com/openclarity/apiclarity/api3/global"
)

// APIClarityNotification defines model for APIClarityNotification.
type APIClarityNotification struct {
	union json.RawMessage
}

// A group of findings
type APIFindings struct {
	// A list of findings
	Items *[]externalRef0.APIFinding `json:"items,omitempty"`
}

// ApiFindingsNotification defines model for ApiFindingsNotification.
type ApiFindingsNotification struct {
	// A list of findings
	Items            *[]externalRef0.APIFinding `json:"items,omitempty"`
	NotificationType string                     `json:"notificationType"`
}

// ApiInfo defines model for ApiInfo.
type ApiInfo struct {
	DestinationNamespace *string `json:"destinationNamespace,omitempty"`
	HasProvidedSpec      *bool   `json:"hasProvidedSpec,omitempty"`
	HasReconstructedSpec *bool   `json:"hasReconstructedSpec,omitempty"`
	Id                   *uint32 `json:"id,omitempty"`

	// API name
	Name *string `json:"name,omitempty"`
	Port *int    `json:"port,omitempty"`

	// Trace Source UUID which created this API. empty means it has been created by APIClarity (from the UI for example)
	TraceSourceId *openapi_types.UUID `json:"traceSourceId,omitempty"`
}

// AuthorizationModel defines model for AuthorizationModel.
type AuthorizationModel struct {
	Learning   bool                                       `json:"learning"`
	Operations []externalRef1.AuthorizationModelOperation `json:"operations"`
	SpecType   externalRef1.SpecType                      `json:"specType"`
}

// AuthorizationModelNotification defines model for AuthorizationModelNotification.
type AuthorizationModelNotification struct {
	Learning         bool                                       `json:"learning"`
	NotificationType string                                     `json:"notificationType"`
	Operations       []externalRef1.AuthorizationModelOperation `json:"operations"`
	SpecType         externalRef1.SpecType                      `json:"specType"`
}

// Base Notification all APIClarity notifications must extend
type BaseNotification struct {
	NotificationType string `json:"notificationType"`
}

// NewDiscoveredAPINotification defines model for NewDiscoveredAPINotification.
type NewDiscoveredAPINotification struct {
	DestinationNamespace *string `json:"destinationNamespace,omitempty"`
	HasProvidedSpec      *bool   `json:"hasProvidedSpec,omitempty"`
	HasReconstructedSpec *bool   `json:"hasReconstructedSpec,omitempty"`
	Id                   *uint32 `json:"id,omitempty"`

	// API name
	Name             *string `json:"name,omitempty"`
	NotificationType string  `json:"notificationType"`
	Port             *int    `json:"port,omitempty"`

	// Trace Source UUID which created this API. empty means it has been created by APIClarity (from the UI for example)
	TraceSourceId *openapi_types.UUID `json:"traceSourceId,omitempty"`
}

// Describes the progress of an ongoing test
type ShortTestProgress struct {
	ApiID *externalRef0.ApiID `json:"apiID,omitempty"`

	// Progress of the test
	Progress int `json:"progress"`

	// Timestamp of the start of the test
	Starttime int64 `json:"starttime"`
}

// Short Test Report
type ShortTestReport struct {
	ApiID *externalRef0.ApiID `json:"apiID,omitempty"`

	// Severity of a finding
	HighestSeverity *externalRef0.Severity `json:"highestSeverity,omitempty"`

	// Timestamp of the start of the test
	Starttime int64 `json:"starttime"`

	// An enumeration.
	Status externalRef1.FuzzingStatusEnum `json:"status"`

	// Message for status details, if any
	StatusMessage *string                          `json:"statusMessage,omitempty"`
	Tags          *[]externalRef1.FuzzingReportTag `json:"tags,omitempty"`
}

// SpecDiffs defines model for SpecDiffs.
type SpecDiffs struct {
	Diffs externalRef1.APIDiffs `json:"diffs"`
}

// SpecDiffsNotification defines model for SpecDiffsNotification.
type SpecDiffsNotification struct {
	Diffs            externalRef1.APIDiffs `json:"diffs"`
	NotificationType string                `json:"notificationType"`
}

// TestProgressNotification defines model for TestProgressNotification.
type TestProgressNotification struct {
	ApiID            *externalRef0.ApiID `json:"apiID,omitempty"`
	NotificationType string              `json:"notificationType"`

	// Progress of the test
	Progress int `json:"progress"`

	// Timestamp of the start of the test
	Starttime int64 `json:"starttime"`
}

// TestReportNotification defines model for TestReportNotification.
type TestReportNotification struct {
	ApiID *externalRef0.ApiID `json:"apiID,omitempty"`

	// Severity of a finding
	HighestSeverity  *externalRef0.Severity `json:"highestSeverity,omitempty"`
	NotificationType string                 `json:"notificationType"`

	// Timestamp of the start of the test
	Starttime int64 `json:"starttime"`

	// An enumeration.
	Status externalRef1.FuzzingStatusEnum `json:"status"`

	// Message for status details, if any
	StatusMessage *string                          `json:"statusMessage,omitempty"`
	Tags          *[]externalRef1.FuzzingReportTag `json:"tags,omitempty"`
}

// PostNotificationApiIDJSONBody defines parameters for PostNotificationApiID.
type PostNotificationApiIDJSONBody = APIClarityNotification

// PostNotificationApiIDJSONRequestBody defines body for PostNotificationApiID for application/json ContentType.
type PostNotificationApiIDJSONRequestBody = PostNotificationApiIDJSONBody

func (t APIClarityNotification) AsApiFindingsNotification() (ApiFindingsNotification, error) {
	var body ApiFindingsNotification
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *APIClarityNotification) FromApiFindingsNotification(v ApiFindingsNotification) error {
	v.NotificationType = "ApiFindingsNotification"
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t APIClarityNotification) AsAuthorizationModelNotification() (AuthorizationModelNotification, error) {
	var body AuthorizationModelNotification
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *APIClarityNotification) FromAuthorizationModelNotification(v AuthorizationModelNotification) error {
	v.NotificationType = "AuthorizationModelNotification"
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t APIClarityNotification) AsNewDiscoveredAPINotification() (NewDiscoveredAPINotification, error) {
	var body NewDiscoveredAPINotification
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *APIClarityNotification) FromNewDiscoveredAPINotification(v NewDiscoveredAPINotification) error {
	v.NotificationType = "NewDiscoveredAPINotification"
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t APIClarityNotification) AsSpecDiffsNotification() (SpecDiffsNotification, error) {
	var body SpecDiffsNotification
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *APIClarityNotification) FromSpecDiffsNotification(v SpecDiffsNotification) error {
	v.NotificationType = "SpecDiffsNotification"
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t APIClarityNotification) AsTestProgressNotification() (TestProgressNotification, error) {
	var body TestProgressNotification
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *APIClarityNotification) FromTestProgressNotification(v TestProgressNotification) error {
	v.NotificationType = "TestProgressNotification"
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t APIClarityNotification) AsTestReportNotification() (TestReportNotification, error) {
	var body TestReportNotification
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *APIClarityNotification) FromTestReportNotification(v TestReportNotification) error {
	v.NotificationType = "TestReportNotification"
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t APIClarityNotification) Discriminator() (string, error) {
	var discriminator struct {
		Discriminator string `json:"notificationType"`
	}
	err := json.Unmarshal(t.union, &discriminator)
	return discriminator.Discriminator, err
}

func (t APIClarityNotification) ValueByDiscriminator() (interface{}, error) {
	discriminator, err := t.Discriminator()
	if err != nil {
		return nil, err
	}
	switch discriminator {
	case "ApiFindingsNotification":
		return t.AsApiFindingsNotification()
	case "AuthorizationModelNotification":
		return t.AsAuthorizationModelNotification()
	case "NewDiscoveredAPINotification":
		return t.AsNewDiscoveredAPINotification()
	case "SpecDiffsNotification":
		return t.AsSpecDiffsNotification()
	case "TestProgressNotification":
		return t.AsTestProgressNotification()
	case "TestReportNotification":
		return t.AsTestReportNotification()
	default:
		return nil, errors.New("unknown discriminator value: " + discriminator)
	}
}

func (t APIClarityNotification) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

func (t *APIClarityNotification) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	return err
}

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// PostNotificationApiID request with any body
	PostNotificationApiIDWithBody(ctx context.Context, apiID int64, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostNotificationApiID(ctx context.Context, apiID int64, body PostNotificationApiIDJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) PostNotificationApiIDWithBody(ctx context.Context, apiID int64, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostNotificationApiIDRequestWithBody(c.Server, apiID, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostNotificationApiID(ctx context.Context, apiID int64, body PostNotificationApiIDJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostNotificationApiIDRequest(c.Server, apiID, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewPostNotificationApiIDRequest calls the generic PostNotificationApiID builder with application/json body
func NewPostNotificationApiIDRequest(server string, apiID int64, body PostNotificationApiIDJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostNotificationApiIDRequestWithBody(server, apiID, "application/json", bodyReader)
}

// NewPostNotificationApiIDRequestWithBody generates requests for PostNotificationApiID with any type of body
func NewPostNotificationApiIDRequestWithBody(server string, apiID int64, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "apiID", runtime.ParamLocationPath, apiID)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/notification/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// PostNotificationApiID request with any body
	PostNotificationApiIDWithBodyWithResponse(ctx context.Context, apiID int64, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostNotificationApiIDResponse, error)

	PostNotificationApiIDWithResponse(ctx context.Context, apiID int64, body PostNotificationApiIDJSONRequestBody, reqEditors ...RequestEditorFn) (*PostNotificationApiIDResponse, error)
}

type PostNotificationApiIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *string
}

// Status returns HTTPResponse.Status
func (r PostNotificationApiIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostNotificationApiIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// PostNotificationApiIDWithBodyWithResponse request with arbitrary body returning *PostNotificationApiIDResponse
func (c *ClientWithResponses) PostNotificationApiIDWithBodyWithResponse(ctx context.Context, apiID int64, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostNotificationApiIDResponse, error) {
	rsp, err := c.PostNotificationApiIDWithBody(ctx, apiID, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostNotificationApiIDResponse(rsp)
}

func (c *ClientWithResponses) PostNotificationApiIDWithResponse(ctx context.Context, apiID int64, body PostNotificationApiIDJSONRequestBody, reqEditors ...RequestEditorFn) (*PostNotificationApiIDResponse, error) {
	rsp, err := c.PostNotificationApiID(ctx, apiID, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostNotificationApiIDResponse(rsp)
}

// ParsePostNotificationApiIDResponse parses an HTTP response from a PostNotificationApiIDWithResponse call
func ParsePostNotificationApiIDResponse(rsp *http.Response) (*PostNotificationApiIDResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostNotificationApiIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest string
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Provide to Apiclarity list of raw input for a given API ID associated with a given timestamp
	// (POST /notification/{apiID})
	PostNotificationApiID(w http.ResponseWriter, r *http.Request, apiID int64)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// PostNotificationApiID operation middleware
func (siw *ServerInterfaceWrapper) PostNotificationApiID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "apiID" -------------
	var apiID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "apiID", runtime.ParamLocationPath, chi.URLParam(r, "apiID"), &apiID)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "apiID", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostNotificationApiID(w, r, apiID)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/notification/{apiID}", wrapper.PostNotificationApiID)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/8RYTW/bOBP+KwTf97ALCLbbLvbgW9psAR/aGnV6KnygpZE0hfix5MipG/i/L0hRsmzJ",
	"jpNturdYnHk4H898MA881dJoBYocnz9wl5YgRfjzZrl4VwmLtPuoCXNMBaFW/iRDl1qUqARp6z9IYQyq",
	"ImgZfI8qQ1W4YzX+v+nhqmm8Z3pOPOE3NZXa4o/w+4POoLoK77JWwj/C/S26VG/BQnazXFwDelEn4SsD",
	"6S3m+VUOjwsn/A4cLa0uLLircM7KN1CfwWhL1wKNSO8Tbqw2YGn3UUjgc656x3c7A15EK/iU8/nXB/5/",
	"C/lTk7xPHtG7nMzH1C+m7THl8Tw9pnU2K9cojmVhvU98IbbhC9UHvvpMzOkNK6yuDdM5y1uhLnUIQQMJ",
	"5KhqhY5ONDvZaO1k4s2VWk21ASUMTnZCVuPp6uz01CBPkTkX1ood3x8+6M03SMlLnG0VD1xU1RW8eisc",
	"PI1QvUg2oTW4ULn2Vx6HLANHvr+hVp7+zogU/PfohCMb3SyFW1q9xQwyT5lGNxd1RXyei8pB5/hG6wqE",
	"ikqfIdXKka1TeoomZl4u11YK4nNeo6I3r3kniYqgAOslVajaQc6XCxZOkqErnn09H3tYZEUKK13bFBbZ",
	"EPTOH7PmnH35srhl9yWmJUstCIKMUYmO3SwXEwbS0I5JEMoxJFYKxzYAqpPc7Nhh7LDfcqsloxLYlwXL",
	"tWXwXUhTwe886YWgxmzozSjfBu1kmPcKhFVxlA2j7yWD9nFZHUqlqPRGVF2pFKAulMvAnE8t/LB+Eu4M",
	"pKHtPvvCVYvgo2Ph7xotZHz+9QB95GFyCMb6qmj+qiIepjHU8gBowFMvwfoiTFRVn3D9EeeYrB0x+E6g",
	"skFLHQzDYW84CfJAYyyol7eTlwtpbIMhjqtSW+oPsmEgb8OvDbhQnCbK+UkiFNOq0KgKRuBoEDdhcHH7",
	"zPESVJu95Ixdy54l3rJoghTfUdaSz1/NZgmXqJpfs4QTUgWdQyYuUMMG6EhYIhzrqHcowZGQpr00yJ5Y",
	"0HUrVPTnH/xw8SoIB+Thvadl2hnRi8G6h+UTx3zmWBeIZqfgI1Tr0hxFBo714DqUn5/MEosSHK1gC74E",
	"n4fWaf9nqQr3Uv0vxsH7+scPVMUqwPylankA/QDOiWLEoXgQZmMjyjIggZVLGPpi3B3b7wVk1ImCY4sA",
	"ieJnDLjoUUOeO9FAR2PuROHC+slHNsVzpI8hHqf8Baa3q/zImtd+fuYIXy4a4FOrG9z1JWN+UWs/OB+a",
	"+/ln5stZMJgnnSVjr9SXtyMyZb/e+7Rh3P9bSr3TFtgnvyf7quotByev7C1Y11Thq8ksLoeeMXzO30xm",
	"E7+UG0FlYNe0P/ynD6Fv7gMbtQudt1u7/HrNl9odBaVplh7PCgkE1oXooL/c38HbbT925D4XydaQxH/r",
	"HD0duvZ2OnTWjTo4equz0JBTrQgUNS3fVK0f31yTrwP4I2+vsX8nhRwcd7WTQB/7EgrNGa1cU8GvZ7Mn",
	"mXi6pA1uX9VpGkjq228tpbC7ZrPwbzxGmt0YTCMl2vezFfcMlakpcEawAregPHfY4pYJ53SK4Xlzj1R2",
	"x9ROo8YKB3bbZra2FZ/zqafoPwEAAP//YvClFqETAAA=",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	pathPrefix := path.Dir(pathToFile)

	for rawPath, rawFunc := range externalRef0.PathToRawSpec(path.Join(pathPrefix, "../common/openapi.yaml")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
	}
	for rawPath, rawFunc := range externalRef1.PathToRawSpec(path.Join(pathPrefix, "../global/openapi.gen.yaml")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
	}
	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
