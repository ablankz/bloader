package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ablankz/bloader/internal/executor/httpexec"
	"github.com/ablankz/bloader/internal/logger"
	"github.com/ablankz/bloader/internal/runner/matcher"
	"github.com/google/uuid"
)

// WriteData represents the write data
type WriteData struct {
	Success          bool
	SendDatetime     string
	ReceivedDatetime string
	Count            int
	ResponseTime     int
	StatusCode       string
	RawData          any
}

// ToSlice converts WriteData to slice
func (d WriteData) ToSlice() []string {
	return []string{
		strconv.FormatBool(d.Success),
		d.SendDatetime,
		d.ReceivedDatetime,
		strconv.Itoa(d.Count),
		strconv.Itoa(d.ResponseTime),
		d.StatusCode,
	}
}

type writeSendData struct {
	uid       uuid.UUID
	writeData WriteData
}

// ResponseDataConsumer represents the response data consumer
type ResponseDataConsumer func(
	ctx context.Context,
	log logger.Logger,
	id int,
	data WriteData,
) error

func runResponseHandler(
	ctx context.Context,
	log logger.Logger,
	id int,
	request ValidMassExecRequest,
	termChan chan<- termChanType,
	writeErrChan <-chan struct{},
	uidChan <-chan uuid.UUID,
	resChan <-chan httpexec.ResponseContent,
	writeChan chan<- writeSendData,
) {
	defer close(termChan)
	var timeout <-chan time.Time
	if request.Break.Time.Enabled && request.Break.Time.Time > 0 {
		timeout = time.After(request.Break.Time.Time)
	}
	sentUid := make(map[uuid.UUID]struct{})
	for {
		select {
		case uid := <-uidChan:
			delete(sentUid, uid)
		case <-writeErrChan:
			log.Warn(ctx, "Term Condition: Write Error",
				logger.Value("id", id), logger.Value("on", "runResponseHandler"))
			for len(sentUid) > 0 {
				uid := <-uidChan
				delete(sentUid, uid)
			}
			termChan <- NewTermChanType(matcher.TerminateTypeByWriteError, "")
			return
		case <-timeout:
			sentLen := len(sentUid)
			writeErr := false
			for sentLen > 0 {
				select {
				case uid := <-uidChan:
					delete(sentUid, uid)
					sentLen--
				case <-writeErrChan:
					log.Warn(ctx, "write error occurred",
						logger.Value("id", id), logger.Value("on", "runResponseHandler"))
					writeErr = true
				}
			}
			if writeErr {
				log.Warn(ctx, "Term Condition: Write Error",
					logger.Value("id", id), logger.Value("on", "runResponseHandler"))
				termChan <- NewTermChanType(matcher.TerminateTypeByWriteError, "")
				return
			}
			log.Info(ctx, "Term Condition: Time",
				logger.Value("id", id), logger.Value("on", "runResponseHandler"))
			termChan <- NewTermChanType(matcher.TerminateTypeByTimeout, "")
			return
		case <-ctx.Done():
			sentLen := len(sentUid)
			writeErr := false
			for sentLen > 0 {
				select {
				case uid := <-uidChan:
					delete(sentUid, uid)
					sentLen--
				case <-writeErrChan:
					log.Warn(ctx, "write error occurred",
						logger.Value("id", id), logger.Value("on", "runResponseHandler"))
					writeErr = true
				}
			}
			if writeErr {
				log.Warn(ctx, "Term Condition: Write Error",
					logger.Value("id", id), logger.Value("on", "runResponseHandler"))
				termChan <- NewTermChanType(matcher.TerminateTypeByWriteError, "")
				return
			}
			log.Info(ctx, "Term Condition: Context Done",
				logger.Value("id", id), logger.Value("on", "runResponseHandler"))
			termChan <- NewTermChanType(matcher.TerminateTypeByContext, "")
			return
		case v := <-resChan:
			mustWrite := true
			var response any
			err := json.Unmarshal(v.ByteResponse, &response)
			if err != nil {
				log.Error(ctx, "The response is not a valid JSON",
					logger.Value("error", err), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
			}
			_, isMatch := request.RecordExcludeFilter.CountFilter(v.Count)
			if isMatch {
				log.Debug(ctx, "Count output filter found",
					logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
				fmt.Println("Count output filter found")
				mustWrite = false
			}
			_, isMatch = request.RecordExcludeFilter.StatusCodeFilter(v.StatusCode)
			if isMatch {
				log.Debug(ctx, "Status output filter found",
					logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
				fmt.Println("Status output filter found")
				mustWrite = false
			}
			var matchID string
			matchID, isMatch, err = request.RecordExcludeFilter.ResponseBodyFilter(response)
			if err != nil {
				log.Error(ctx, "failed to search jmespath",
					logger.Value("error", err), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
				sentLen := len(sentUid)
				writeErr := false
				for sentLen > 0 {
					select {
					case uid := <-uidChan:
						delete(sentUid, uid)
						sentLen--
					case <-writeErrChan:
						log.Warn(ctx, "write error occurred",
							logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
						writeErr = true
					}
				}
				if writeErr {
					log.Warn(ctx, "Term Condition: Write Error",
						logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
					termChan <- NewTermChanType(matcher.TerminateTypeByWriteError, "")
					return
				}
				log.Info(ctx, "Term Condition: Response Body Write Filter Error",
					logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
				termChan <- NewTermChanType(matcher.TerminateTypeByResponseBodyWriteFilterError, matchID)
				return
			}
			if isMatch {
				log.Debug(ctx, "Response output filter found",
					logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
				fmt.Println("Response output filter found")
				mustWrite = false
			}

			if mustWrite {
				uid := uuid.New()
				writeData := WriteData{
					Success:          v.Success,
					SendDatetime:     v.StartTime.Format(time.RFC3339Nano),
					ReceivedDatetime: v.EndTime.Format(time.RFC3339Nano),
					Count:            v.Count,
					ResponseTime:     int(v.ResponseTime),
					StatusCode:       strconv.Itoa(v.StatusCode),
					RawData:          response,
				}
				sentUid[uid] = struct{}{}
				go func() {
					writeChan <- writeSendData{
						uid:       uid,
						writeData: writeData,
					}
				}()
			}

			if v.ReqCreateHasErr {
				sentLen := len(sentUid)
				for sentLen > 0 {
					select {
					case uid := <-uidChan:
						delete(sentUid, uid)
						sentLen--
					case <-writeErrChan:
						log.Warn(ctx, "write error occurred",
							logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
					}
				}
				log.Warn(ctx, "Term Condition: Request Creation Error",
					logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
				termChan <- NewTermChanType(matcher.TerminateTypeByCreateRequestError, "")
				return
			}
			if v.HasSystemErr {
				if request.Break.SysError {
					sentLen := len(sentUid)
					for sentLen > 0 {
						select {
						case uid := <-uidChan:
							delete(sentUid, uid)
							sentLen--
						case <-writeErrChan:
							log.Warn(ctx, "write error occurred",
								logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
						}
					}
					log.Warn(ctx, "Term Condition: System Error",
						logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
					termChan <- NewTermChanType(matcher.TerminateTypeBySystemError, "")
					return
				} else {
					log.Warn(ctx, "System error occurred",
						logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
				}
			}
			if v.ParseResHasErr {
				if request.Break.ParseError {
					sentLen := len(sentUid)
					for sentLen > 0 {
						select {
						case uid := <-uidChan:
							delete(sentUid, uid)
							sentLen--
						case <-writeErrChan:
							log.Warn(ctx, "write error occurred",
								logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
						}
					}
					log.Warn(ctx, "Term Condition: Response Parse Error",
						logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
					termChan <- NewTermChanType(matcher.TerminateTypeByParseResponseError, "")
					return
				} else {
					log.Warn(ctx, "Parse error occurred",
						logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
				}
			}
			if v.WithCountLimit {
				sentLen := len(sentUid)
				writeErr := false
				for sentLen > 0 {
					select {
					case uid := <-uidChan:
						delete(sentUid, uid)
						sentLen--
					case <-writeErrChan:
						log.Warn(ctx, "write error occurred",
							logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
						writeErr = true
					}
				}
				if writeErr {
					log.Warn(ctx, "Term Condition: Write Error",
						logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
					termChan <- NewTermChanType(matcher.TerminateTypeByWriteError, "")
					return
				}

				log.Info(ctx, "Term Condition: Count Limit",
					logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
				termChan <- NewTermChanType(matcher.TerminateTypeByCount, "")
				return
			}
			matchID, isMatch, err = request.Break.ResponseBodyMatcher(response)
			if err != nil {
				log.Error(ctx, "failed to search jmespath",
					logger.Value("error", err), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
				sentLen := len(sentUid)
				writeErr := false
				for sentLen > 0 {
					select {
					case uid := <-uidChan:
						delete(sentUid, uid)
						sentLen--
					case <-writeErrChan:
						log.Warn(ctx, "write error occurred",
							logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
						writeErr = true
					}
				}
				if writeErr {
					log.Warn(ctx, "Term Condition: Write Error",
						logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
					termChan <- NewTermChanType(matcher.TerminateTypeByWriteError, "")
					return
				}

				log.Info(ctx, "Term Condition: Response Body Break Filter Error",
					logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
				termChan <- NewTermChanType(matcher.TerminateTypeByResponseBodyBreakFilterError, matchID)
				return
			}
			if isMatch {
				sentLen := len(sentUid)
				writeErr := false
				for sentLen > 0 {
					select {
					case uid := <-uidChan:
						delete(sentUid, uid)
						sentLen--
					case <-writeErrChan:
						log.Warn(ctx, "write error occurred",
							logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
						writeErr = true
					}
				}
				if writeErr {
					log.Warn(ctx, "Term Condition: Write Error",
						logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
					termChan <- NewTermChanType(matcher.TerminateTypeByWriteError, "")
					return
				}

				log.Info(ctx, "Term Condition: Response Body",
					logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
				termChan <- NewTermChanType(matcher.TerminateTypeByResponseBody, matchID)
				return

			}
			matchID, isMatch = request.Break.StatusCodeMatcher(v.StatusCode)
			if isMatch {
				fmt.Println("Status code match")
				fmt.Println("matchID", matchID)
				fmt.Println("v.StatusCode", v.StatusCode)
				fmt.Println("v.ByteResponse", string(v.ByteResponse))
				sentLen := len(sentUid)
				writeErr := false
				for sentLen > 0 {
					select {
					case uid := <-uidChan:
						delete(sentUid, uid)
						sentLen--
					case <-writeErrChan:
						log.Warn(ctx, "write error occurred",
							logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
						writeErr = true
					}
				}
				if writeErr {
					log.Warn(ctx, "Term Condition: Write Error",
						logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
					termChan <- NewTermChanType(matcher.TerminateTypeByWriteError, "")
					return
				}

				log.Info(ctx, "Term Condition: Status Code",
					logger.Value("id", id), logger.Value("on", "runResponseHandler"), logger.Value("count", v.Count))
				termChan <- NewTermChanType(matcher.TerminateTypeByStatusCode, matchID)
				return
			}
		}
	}
}

func RunAsyncProcessing(
	ctx context.Context,
	log logger.Logger,
	id int,
	request ValidMassExecRequest,
	termChan chan<- termChanType,
	resChan <-chan httpexec.ResponseContent,
	consumer ResponseDataConsumer,
) {
	writeChan := make(chan writeSendData)
	wroteUidChan := make(chan uuid.UUID)
	writeErrChan := make(chan struct{})
	go func() {
		runResponseHandler(ctx, log, id, request, termChan, writeErrChan, wroteUidChan, resChan, writeChan)
	}()

	go func() {
		defer close(wroteUidChan)
		defer close(writeErrChan)
		defer close(writeChan)
		for {
			d := <-writeChan
			log.Debug(ctx, "Writing data",
				logger.Value("id", id), logger.Value("data", d), logger.Value("on", "runAsyncProcessing"))
			if err := consumer(ctx, log, id, d.writeData); err != nil {
				log.Error(ctx, "failed to write data",
					logger.Value("error", err), logger.Value("on", "runAsyncProcessing"))
				if request.Break.WriteError {
					writeErrChan <- struct{}{}
					wroteUidChan <- d.uid
					continue
				}
			}
			wroteUidChan <- d.uid
		}
	}()
}
