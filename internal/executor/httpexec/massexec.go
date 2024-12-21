package httpexec

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/logger"
	"github.com/ablankz/bloader/internal/utils"
	"gopkg.in/yaml.v3"
)

// RequestCountLimit represents the request count limit
type RequestCountLimit struct {
	Enabled bool
	Count   int
}

// MassRequestContent represents the request content
type MassRequestContent[Req ExecReq] struct {
	Req          Req
	Interval     time.Duration
	ResponseWait bool
	ResChan      chan<- ResponseContent
	CountLimit   RequestCountLimit
	ResponseType ResponseType
}

// MassRequestExecute executes the request
func (q MassRequestContent[Req]) MassRequestExecute(
	ctx context.Context,
	ctr *container.Container,
) error {
	req, err := q.Req.CreateRequest(ctx, ctr)
	if err != nil {
		ctr.Logger.Error(ctx, "failed to create request",
			logger.Value("error", err), logger.Value("on", "RequestContent.QueryExecute"))
		return err
	}

	go func() {
		// defer close(q.ResChan) // TODO: close channel

		ticker := time.NewTicker(q.Interval)
		defer ticker.Stop()
		var waitForResponse = q.ResponseWait
		var count int
		var countLimitOver bool
		chanForWait := make(chan struct{})
		defer close(chanForWait)

		for {
			select {
			case <-ctx.Done():
				ctr.Logger.Info(ctx, "request processing is interrupted due to context termination",
					logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
				return
			case <-ticker.C:
				if count > 0 && waitForResponse {
					select {
					case <-ctx.Done():
						ctr.Logger.Info(ctx, "request processing is interrupted due to context termination",
							logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))

						return
					case <-chanForWait:
					}
				}

				count++
				if q.CountLimit.Enabled && count >= q.CountLimit.Count {
					ctr.Logger.Info(ctx, "request processing is interrupted due to count limit",
						logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
					countLimitOver = true
				}

				reqClone := cloneRequest(req)

				go func(asyncReq *http.Request, countOver bool) {
					defer func() {
						if waitForResponse {
							chanForWait <- struct{}{}
						}
					}()

					client := &http.Client{
						Timeout: 10 * time.Minute,
						Transport: &utils.DelayedTransport{
							Transport: http.DefaultTransport,
							// Delay:     2 * time.Second,
						},
					}

					ctr.Logger.Debug(ctx, "sending request",
						logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL), logger.Value("count", count))
					startTime := time.Now()
					resp, err := client.Do(asyncReq)
					endTime := time.Now()
					ctr.Logger.Debug(ctx, "received response",
						logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL), logger.Value("count", count))
					if err != nil {
						ctr.Logger.Error(ctx, "response error",
							logger.Value("error", err), logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
						select {
						case <-ctx.Done():
							ctr.Logger.Info(ctx, "request processing is interrupted due to context termination",
								logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
							return
						case q.ResChan <- ResponseContent{
							Success:        false,
							StartTime:      startTime,
							EndTime:        endTime,
							Count:          count,
							ResponseTime:   endTime.Sub(startTime).Milliseconds(),
							HasSystemErr:   true,
							WithCountLimit: countOver,
						}: // do nothing
						}

						return
					}
					defer resp.Body.Close()

					statusCode := resp.StatusCode
					var response any
					responseByte, err := io.ReadAll(resp.Body)
					if err != nil {
						ctr.Logger.Error(ctx, "failed to read response",
							logger.Value("error", err), logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
						select {
						case <-ctx.Done():
							ctr.Logger.Info(ctx, "request processing is interrupted due to context termination",
								logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))

							return
						case q.ResChan <- ResponseContent{
							Success:        false,
							Res:            response,
							StartTime:      startTime,
							EndTime:        endTime,
							Count:          count,
							ResponseTime:   endTime.Sub(startTime).Milliseconds(),
							StatusCode:     statusCode,
							ParseResHasErr: true,
							WithCountLimit: countOver,
						}: // do nothing
						}
						return
					}
					switch ResponseType(q.ResponseType) {
					case ResponseTypeJSON:
						err = json.Unmarshal(responseByte, &response)
					case ResponseTypeXML:
						err = xml.Unmarshal(responseByte, &response)
					case ResponseTypeYAML:
						err = yaml.Unmarshal(responseByte, &response)
					case ResponseTypeText:
						response = string(responseByte)
					case ResponseTypeHTML:
						response = string(responseByte)
					default:
						err = fmt.Errorf("invalid response type: %s", q.ResponseType)
					}
					if err != nil {
						ctr.Logger.Error(ctx, "failed to parse response",
							logger.Value("error", err), logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
						select {
						case <-ctx.Done():
							ctr.Logger.Info(ctx, "request processing is interrupted due to context termination",
								logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
							return
						case q.ResChan <- ResponseContent{
							Success:        false,
							Res:            response,
							ByteResponse:   responseByte,
							StartTime:      startTime,
							EndTime:        endTime,
							Count:          count,
							ResponseTime:   endTime.Sub(startTime).Milliseconds(),
							StatusCode:     statusCode,
							ParseResHasErr: true,
							WithCountLimit: countOver,
						}: // do nothing
						}
						return
					}

					ctr.Logger.Debug(ctx, "response OK",
						logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
					responseContent := ResponseContent{
						Success:        true,
						ByteResponse:   responseByte,
						Res:            response,
						StartTime:      startTime,
						EndTime:        endTime,
						Count:          count,
						ResponseTime:   endTime.Sub(startTime).Milliseconds(),
						StatusCode:     statusCode,
						WithCountLimit: countOver,
					}
					select {
					case q.ResChan <- responseContent:
					case <-ctx.Done():
						ctr.Logger.Info(ctx, "request processing is interrupted due to context termination",
							logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
						return
					}
				}(reqClone, countLimitOver)

				if countLimitOver {
					<-ctx.Done()
					ctr.Logger.Info(ctx, "request processing is interrupted due to count limit",
						logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
					return
				}
			}
		}
	}()

	return nil
}

func cloneRequest(req *http.Request) *http.Request {
	clone := req.Clone(req.Context())
	if req.Body != nil {
		body, _ := req.GetBody()
		clone.Body = body
	}
	return clone
}

var _ MassRequestExecutor = MassRequestContent[HTTPRequest]{}
