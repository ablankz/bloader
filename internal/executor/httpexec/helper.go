package httpexec

// type WriteData struct {
// 	Success          bool
// 	SendDatetime     string
// 	ReceivedDatetime string
// 	Count            int
// 	ResponseTime     int
// 	StatusCode       string
// 	Data             any
// 	RawData          any
// }

// func (d WriteData) ToSlice() []string {
// 	return []string{
// 		strconv.FormatBool(d.Success),
// 		d.SendDatetime,
// 		d.ReceivedDatetime,
// 		strconv.Itoa(d.Count),
// 		strconv.Itoa(d.ResponseTime),
// 		d.StatusCode,
// 		fmt.Sprintf("%v", d.Data),
// 	}
// }

// type writeSendData struct {
// 	uid       uuid.UUID
// 	writeData WriteData
// }

// type ResponseDataConsumer func(
// 	ctx context.Context,
// 	ctr *container.Container,
// 	id int,
// 	data WriteData,
// ) error

// func runResponseHandler[Res any](
// 	ctx context.Context,
// 	ctr *container.Container,
// 	id int,
// 	request *ValidatedExecRequest,
// 	termChan chan<- TerminateType,
// 	writeErrChan <-chan struct{},
// 	uidChan <-chan uuid.UUID,
// 	resChan <-chan ResponseContent,
// 	writeChan chan<- writeSendData,
// ) {
// 	defer close(termChan)
// 	var count int
// 	var timeout <-chan time.Time
// 	if request.Break.Time > 0 {
// 		timeout = time.After(request.Break.Time)
// 	}
// 	sentUid := make(map[uuid.UUID]struct{})
// 	for {
// 		count++
// 		select {
// 		case uid := <-uidChan:
// 			delete(sentUid, uid)
// 			count--
// 		case <-writeErrChan:
// 			ctr.Logger.Warn(ctx, "Term Condition: Write Error",
// 				logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 			for len(sentUid) > 0 {
// 				uid := <-uidChan
// 				delete(sentUid, uid)
// 			}
// 			termChan <- ByWriteError
// 			return
// 		case <-timeout:
// 			sentLen := len(sentUid)
// 			writeErr := false
// 			for sentLen > 0 {
// 				select {
// 				case uid := <-uidChan:
// 					delete(sentUid, uid)
// 					sentLen--
// 				case <-writeErrChan:
// 					ctr.Logger.Warn(ctx, "write error occurred",
// 						logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 					writeErr = true
// 				}
// 			}
// 			if writeErr {
// 				ctr.Logger.Warn(ctx, "Term Condition: Write Error",
// 					logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 				termChan <- ByWriteError
// 				return
// 			}
// 			ctr.Logger.Info(ctx, "Term Condition: Time",
// 				logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 			termChan <- ByTimeout
// 			return
// 		case <-ctx.Done():
// 			sentLen := len(sentUid)
// 			writeErr := false
// 			for sentLen > 0 {
// 				select {
// 				case uid := <-uidChan:
// 					delete(sentUid, uid)
// 					sentLen--
// 				case <-writeErrChan:
// 					ctr.Logger.Warn(ctx, "write error occurred",
// 						logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 					writeErr = true
// 				}
// 			}
// 			if writeErr {
// 				ctr.Logger.Warn(ctx, "Term Condition: Write Error",
// 					logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 				termChan <- ByWriteError
// 				return
// 			}
// 			ctr.Logger.Info(ctx, "Term Condition: Context Done",
// 				logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 			termChan <- ByContext
// 			return
// 		case v := <-resChan:
// 			var mustWrite bool
// 			var response any
// 			err := json.Unmarshal(v.ByteResponse, &response)
// 			if err != nil {
// 				ctr.Logger.Error(ctx, "The response is not a valid JSON",
// 					logger.Value("error", err), logger.Value("on", "runResponseHandler"))
// 			}

// 			if request.DataOutputFilter.HasValue {
// 				jmesPathQuery := request.DataOutputFilter.JMESPath
// 				result, err := jmespath.Search(jmesPathQuery, response)
// 				if err != nil {
// 					ctr.Logger.Error(ctx, "failed to search jmespath",
// 						logger.Value("error", err), logger.Value("on", "runResponseHandler"))
// 				}
// 				if result != nil {
// 					if v, ok := result.(bool); ok {
// 						if v {
// 							mustWrite = true
// 						}
// 					} else {
// 						ctr.Logger.Warn(ctx, "The result of the jmespath query is not a boolean",
// 							logger.Value("on", "runResponseHandler"))
// 					}
// 				}
// 			} else {
// 				mustWrite = true
// 			}

// 			if request.ExcludeStatusFilter(v.StatusCode) {
// 				ctr.Logger.Info(ctx, "Status output filter found",
// 					logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 				mustWrite = false
// 			}

// 			if mustWrite {
// 				var data any = v.Res
// 				if request.DataOutput.HasValue {
// 					jmesPathQuery := request.DataOutput.JMESPath
// 					result, err := jmespath.Search(jmesPathQuery, response)
// 					if err != nil {
// 						ctr.Logger.Error(ctx, "failed to search jmespath",
// 							logger.Value("error", err), logger.Value("on", "runResponseHandler"))
// 					}
// 					data = result
// 				}
// 				uid := uuid.New()
// 				writeData := WriteData{
// 					Success:          v.Success,
// 					SendDatetime:     v.StartTime.Format(time.RFC3339Nano),
// 					ReceivedDatetime: v.EndTime.Format(time.RFC3339Nano),
// 					Count:            count,
// 					ResponseTime:     int(v.ResponseTime),
// 					StatusCode:       strconv.Itoa(v.StatusCode),
// 					Data:             data,
// 					RawData:          response,
// 				}
// 				sentUid[uid] = struct{}{}
// 				// for {
// 				// 	select {
// 				// 	case <-ctx.Done():
// 				// 		sentLen := len(sentUid)
// 				// 		writeErr := false
// 				// 		for sentLen > 0 {
// 				// 			select {
// 				// 			case uid := <-uidChan:
// 				// 				delete(sentUid, uid)
// 				// 				sentLen--
// 				// 			case <-writeErrChan:
// 				// 				ctr.Logger.Warn(ctx, "write error occurred",
// 				// 					logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 				// 				writeErr = true
// 				// 			}
// 				// 		}
// 				// 		if writeErr {
// 				// 			ctr.Logger.Warn(ctx, "Term Condition: Write Error",
// 				// 				logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 				// 			termChan <- ByWriteError
// 				// 			return
// 				// 		}
// 				// 		ctr.Logger.Info(ctx, "Term Condition: Context Done",
// 				// 			logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 				// 		termChan <- ByContext
// 				// 		return
// 				// 	case uid := <-uidChan:
// 				// 		delete(sentUid, uid)
// 				// 		continue
// 				// 	case writeChan <- writeSendData{
// 				// 		uid:       uid,
// 				// 		writeData: writeData,
// 				// 	}:
// 				// 	}
// 				// 	break
// 				// }
// 				go func() {
// 					writeChan <- writeSendData{
// 						uid:       uid,
// 						writeData: writeData,
// 					}
// 				}()
// 			}

// 			if v.ReqCreateHasErr {
// 				sentLen := len(sentUid)
// 				// writeErr := false
// 				for sentLen > 0 {
// 					select {
// 					case uid := <-uidChan:
// 						delete(sentUid, uid)
// 						sentLen--
// 					case <-writeErrChan:
// 						ctr.Logger.Warn(ctx, "write error occurred",
// 							logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 						// writeErr = true
// 					}
// 				}
// 				// if writeErr {
// 				// 	ctr.Logger.Warn(ctx, "Term Condition: Write Error",
// 				// 		logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 				// 	termChan <- ByWriteError
// 				// 	return
// 				// }

// 				ctr.Logger.Warn(ctx, "Term Condition: Request Creation Error",
// 					logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 				termChan <- ByCreateRequestError
// 				return
// 			}
// 			if v.HasSystemErr {
// 				if request.Break.SysError {
// 					sentLen := len(sentUid)
// 					// writeErr := false
// 					for sentLen > 0 {
// 						select {
// 						case uid := <-uidChan:
// 							delete(sentUid, uid)
// 							sentLen--
// 						case <-writeErrChan:
// 							ctr.Logger.Warn(ctx, "write error occurred",
// 								logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 							// writeErr = true
// 						}
// 					}
// 					// if writeErr {
// 					// 	ctr.Logger.Warn(ctx, "Term Condition: Write Error",
// 					// 		logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 					// 	termChan <- ByWriteError
// 					// 	return
// 					// }

// 					ctr.Logger.Warn(ctx, "Term Condition: System Error",
// 						logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 					termChan <- BySystemError
// 					return
// 				} else {
// 					ctr.Logger.Warn(ctx, "System error occurred",
// 						logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 				}
// 			}
// 			if v.ParseResHasErr {
// 				if request.Break.ParseError {
// 					sentLen := len(sentUid)
// 					// writeErr := false
// 					for sentLen > 0 {
// 						select {
// 						case uid := <-uidChan:
// 							delete(sentUid, uid)
// 							sentLen--
// 						case <-writeErrChan:
// 							ctr.Logger.Warn(ctx, "write error occurred",
// 								logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 							// writeErr = true
// 						}
// 					}
// 					// if writeErr {
// 					// 	ctr.Logger.Warn(ctx, "Term Condition: Write Error",
// 					// 		logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 					// 	termChan <- ByWriteError
// 					// 	return
// 					// }
// 					ctr.Logger.Warn(ctx, "Term Condition: Response Parse Error",
// 						logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 					termChan <- ByParseResponseError
// 					return
// 				} else {
// 					ctr.Logger.Warn(ctx, "Parse error occurred",
// 						logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 				}
// 			}
// 			if v.WithCountLimit {
// 				sentLen := len(sentUid)
// 				writeErr := false
// 				for sentLen > 0 {
// 					select {
// 					case uid := <-uidChan:
// 						delete(sentUid, uid)
// 						sentLen--
// 					case <-writeErrChan:
// 						ctr.Logger.Warn(ctx, "write error occurred",
// 							logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 						writeErr = true
// 					}
// 				}
// 				if writeErr {
// 					ctr.Logger.Warn(ctx, "Term Condition: Write Error",
// 						logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 					termChan <- ByWriteError
// 					return
// 				}

// 				ctr.Logger.Info(ctx, "Term Condition: Count Limit",
// 					logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 				termChan <- ByCount
// 				return
// 			}
// 			if request.Break.ResponseBody.HasValue {
// 				jmesPathQuery := request.Break.ResponseBody.JMESPath
// 				result, err := jmespath.Search(jmesPathQuery, response)
// 				if err != nil {
// 					ctr.Logger.Error(ctx, "failed to search jmespath",
// 						logger.Value("error", err), logger.Value("on", "runResponseHandler"))
// 				}
// 				if result != nil {
// 					if v, ok := result.(bool); ok {
// 						if v {
// 							sentLen := len(sentUid)
// 							writeErr := false
// 							for sentLen > 0 {
// 								select {
// 								case uid := <-uidChan:
// 									delete(sentUid, uid)
// 									sentLen--
// 								case <-writeErrChan:
// 									ctr.Logger.Warn(ctx, "write error occurred",
// 										logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 									writeErr = true
// 								}
// 							}
// 							if writeErr {
// 								ctr.Logger.Warn(ctx, "Term Condition: Write Error",
// 									logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 								termChan <- ByWriteError
// 								return
// 							}

// 							ctr.Logger.Info(ctx, "Term Condition: Response Body",
// 								logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 							termChan <- ByResponseBodyMatch
// 							return
// 						}
// 					} else {
// 						ctr.Logger.Warn(ctx, "The result of the jmespath query is not a boolean",
// 							logger.Value("on", "runResponseHandler"))
// 					}
// 				}
// 			}
// 			if request.Break.StatusCodeMatcher(v.StatusCode) {
// 				sentLen := len(sentUid)
// 				writeErr := false
// 				for sentLen > 0 {
// 					select {
// 					case <-ctx.Done():
// 						ctr.Logger.Info(ctx, "Term Condition: Context Done",
// 							logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 						termChan <- ByContext
// 						return
// 					case uid := <-uidChan:
// 						delete(sentUid, uid)
// 						sentLen--
// 					case <-writeErrChan:
// 						ctr.Logger.Warn(ctx, "write error occurred",
// 							logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 						writeErr = true
// 					}
// 				}
// 				if writeErr {
// 					ctr.Logger.Warn(ctx, "Term Condition: Write Error",
// 						logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 					termChan <- ByWriteError
// 					return
// 				}

// 				ctr.Logger.Info(ctx, "Term Condition: Status Code",
// 					logger.Value("id", id), logger.Value("count", count), logger.Value("on", "runResponseHandler"))
// 				termChan <- ByStatusCodeMatch
// 				return
// 			}
// 		}
// 	}
// }

// func RunAsyncProcessing[Res any](
// 	ctx context.Context,
// 	ctr *container.Container,
// 	id int,
// 	request *ValidatedExecRequest,
// 	termChan chan<- TerminateType,
// 	resChan <-chan ResponseContent,
// 	consumer ResponseDataConsumer,
// ) {
// 	writeChan := make(chan writeSendData)
// 	wroteUidChan := make(chan uuid.UUID)
// 	writeErrChan := make(chan struct{})
// 	go func() {
// 		runResponseHandler(ctx, ctr, id, request, termChan, writeErrChan, wroteUidChan, resChan, writeChan)
// 	}()

// 	go func() {
// 		defer close(wroteUidChan)
// 		defer close(writeErrChan)
// 		defer close(writeChan)
// 		for {
// 			d := <-writeChan
// 			ctr.Logger.Debug(ctx, "Writing data",
// 				logger.Value("id", id), logger.Value("data", d), logger.Value("on", "runAsyncProcessing"))
// 			if err := consumer(ctx, ctr, id, d.writeData); err != nil {
// 				ctr.Logger.Error(ctx, "failed to write data",
// 					logger.Value("error", err), logger.Value("on", "runAsyncProcessing"))
// 				if request.Break.WriteError {
// 					writeErrChan <- struct{}{}
// 					wroteUidChan <- d.uid
// 					continue
// 				}
// 			}
// 			wroteUidChan <- d.uid
// 			// select {
// 			// case d := <-writeChan:
// 			// 	ctr.Logger.Debug(ctx, "Writing data",
// 			// 		logger.Value("id", id), logger.Value("data", d), logger.Value("on", "runAsyncProcessing"))
// 			// 	if err := consumer(ctx, ctr, id, d.writeData); err != nil {
// 			// 		ctr.Logger.Error(ctx, "failed to write data",
// 			// 			logger.Value("error", err), logger.Value("on", "runAsyncProcessing"))
// 			// 		if request.Break.WriteError {
// 			// 			writeErrChan <- struct{}{}
// 			// 			wroteUidChan <- d.uid
// 			// 			continue
// 			// 		}
// 			// 	}
// 			// 	wroteUidChan <- d.uid
// 			// case <-ctx.Done():
// 			// 	// return
// 			// }
// 		}
// 	}()
// }
