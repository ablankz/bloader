package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/logger"
	"github.com/ablankz/bloader/internal/output"
	"gopkg.in/yaml.v3"
)

// ExecReq represents the request executor
type ExecReq interface {
	// CreateRequest creates the http.Request object for the query
	CreateRequest(ctx context.Context, ctr *container.Container, count int) (*http.Request, error)
}

// HTTPRequestBodyType represents the HTTP request body type
type HTTPRequestBodyType string

const (
	// HTTPRequestBodyTypeJSON represents the JSON body type
	HTTPRequestBodyTypeJSON HTTPRequestBodyType = "json"
	// HTTPRequestBodyTypeForm represents the form body type
	HTTPRequestBodyTypeForm HTTPRequestBodyType = "form"
	// HTTPRequestBodyTypeMultipart represents the multipart body type
	HTTPRequestBodyTypeMultipart HTTPRequestBodyType = "multipart"

	// DefaultHTTPRequestBodyType represents the default HTTP request body type
	DefaultHTTPRequestBodyType = HTTPRequestBodyTypeJSON
)

// AttachRequestInfo represents the request info
type AttachRequestInfo func(ctx context.Context, ctr *container.Container, req *http.Request) error

// HTTPRequest represents the HTTP request
type HTTPRequest struct {
	Method            string
	URL               string
	Headers           map[string]any    // map[string]any or map[string][]any
	QueryParams       map[string]any    // map[string]any or map[string][]any
	PathVariables     map[string]string // /path/{variable} -> /path/value
	BodyType          HTTPRequestBodyType
	Body              any
	AttachRequestInfo AttachRequestInfo
	TmplStr           string
	ReplaceData       *sync.Map
	OCtr              output.OutputContainer
	IsMass            bool
	ReqIndex          int
}

func solvePathVariables(path string, pathVariables map[string]string) string {
	for key, value := range pathVariables {
		path = strings.ReplaceAll(path, "{"+key+"}", value)
	}
	return path
}

// CreateRequest creates the http.Request object for the query
func (r HTTPRequest) CreateRequest(ctx context.Context, ctr *container.Container, count int) (*http.Request, error) {
	if r.IsMass {
		// if v, ok := r.ReplaceData["Dynamic"]; ok {
		// 	if mapV, ok := v.(map[string]any); ok {
		// 		mapV["RequestLoopCount"] = count
		// 	}
		// }
		replaceData := make(map[string]any)
		dynamicData := make(map[string]any)
		r.ReplaceData.Range(func(key, value any) bool {
			if key.(string) == "Dynamic" {
				if mapV, ok := value.(map[string]any); ok {
					for k, v := range mapV {
						dynamicData[k] = v
					}
				}
			} else {
				replaceData[key.(string)] = value
			}
			return true
		})
		dynamicData["RequestLoopCount"] = count
		replaceData["Dynamic"] = dynamicData
		tmpl, err := template.New("yaml").Funcs(sprig.TxtFuncMap()).Parse(r.TmplStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse yaml: %w", err)
		}
		var buffer bytes.Buffer
		if err := tmpl.Execute(&buffer, replaceData); err != nil {
			return nil, fmt.Errorf("failed to execute template: %w", err)
		}
		var massExec MassExec
		if err := yaml.Unmarshal(buffer.Bytes(), &massExec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal json: %w", err)
		}
		validMassExec, err := massExec.Validate(ctr, r.OCtr, r.TmplStr, replaceData)
		if err != nil {
			return nil, fmt.Errorf("failed to validate mass exec: %w", err)
		}
		request := validMassExec.Requests[r.ReqIndex]

		r.URL = request.URL
		r.Method = request.Method
		r.Headers = request.Headers
		r.QueryParams = request.QueryParams
		r.PathVariables = request.PathVariables
		r.BodyType = request.BodyType
		r.Body = request.Body
	}

	reqUrl := solvePathVariables(r.URL, r.PathVariables)
	fullURL, err := url.Parse(reqUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to construct URL: %w", err)
	}
	queryParams := fullURL.Query()
	for key, value := range r.QueryParams {
		if arr, ok := value.([]any); ok {
			for _, v := range arr {
				queryParams.Add(key, fmt.Sprint(v))
			}
			continue
		}
		queryParams.Set(key, fmt.Sprint(value))
	}
	fullURL.RawQuery = queryParams.Encode()
	ctr.Logger.Debug(ctx, "GET request to file objects endpoint URL created",
		logger.Value("url", fullURL.String()), logger.Value("on", "GetFileObjectsReq.CreateRequest"))

	var body io.Reader = nil
	var header = http.Header{}
	switch r.BodyType {
	case HTTPRequestBodyTypeJSON:
		if r.Body == nil {
			break
		}
		bodyBytes, err := json.Marshal(r.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(bodyBytes)
		header.Set("Content-Type", "application/json")
	case HTTPRequestBodyTypeForm:
		form := url.Values{}
		if r.Body == nil {
			break
		}
		if arr, ok := r.Body.(map[string]any); ok {
			for key, value := range arr {
				form.Add(key, fmt.Sprint(value))
			}
		}
		body = strings.NewReader(form.Encode())
		header.Set("Content-Type", "application/x-www-form-urlencoded")
	case HTTPRequestBodyTypeMultipart:
		// Usually, string is entered in map[string]any of body,
		// but in the case of file, field_name, file_name and file_path are entered
		// in map[string]string.
		// In the case of a file, field_name, file_name, and file_path are entered
		// as map[string]string.
		if r.Body == nil {
			break
		}
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		if arr, ok := r.Body.(map[string]any); ok {
			for key, value := range arr {
				if str, ok := value.(string); ok {
					writer.WriteField(key, str)
					continue
				}
				if arr, ok := value.(map[string]string); ok {
					fieldName, fieldOk := arr["field_name"]
					fileName, fileOk := arr["file_name"]
					filePath, pathOk := arr["file_path"]
					if !fieldOk || !fileOk || !pathOk {
						return nil, fmt.Errorf("invalid multipart body: %v", value)
					}
					file, err := os.Open(filePath)
					if err != nil {
						return nil, fmt.Errorf("failed to open file: %w", err)
					}
					defer file.Close()
					part, err := writer.CreateFormFile(fieldName, fileName)
					if err != nil {
						return nil, fmt.Errorf("failed to create form file: %w", err)
					}
					_, err = io.Copy(part, file)
					if err != nil {
						return nil, fmt.Errorf("failed to copy file: %w", err)
					}
				}
			}
		}
		body = &buf
		header.Set("Content-Type", writer.FormDataContentType())
		writer.Close()
	}

	req, err := http.NewRequest(r.Method, fullURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header = header
	if r.AttachRequestInfo != nil {
		err = r.AttachRequestInfo(ctx, ctr, req)
		if err != nil {
			return nil, fmt.Errorf("failed to attach request info: %w", err)
		}
	}

	for key, value := range r.Headers {
		if arr, ok := value.([]any); ok {
			for _, v := range arr {
				req.Header.Add(key, fmt.Sprint(v))
			}
			continue
		}
		req.Header.Set(key, fmt.Sprint(value))
	}

	return req, nil
}

var _ ExecReq = (*HTTPRequest)(nil)