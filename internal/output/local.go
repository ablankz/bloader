package output

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/logger"
	"github.com/ablankz/bloader/internal/utils"
)

// LocalOutput represents the local output service
type LocalOutput struct {
	// Format of the output
	Format config.OutputFormat
	// BasePath of the output
	BasePath string
}

// NewLocalOutput creates a new LocalOutput
func NewLocalOutput(cfg config.ValidOutputRespectiveValueConfig) LocalOutput {
	return LocalOutput{
		Format:   cfg.Format,
		BasePath: cfg.BasePath,
	}
}

// HTTPDataWriteFactory returns the HTTPDataWrite function
func (o LocalOutput) HTTPDataWriteFactory(
	ctx context.Context,
	ctr *container.Container,
	enabled bool,
	uniqueName string,
	header []string,
) (HTTPDataWrite, Close, error) {
	var filePath string
	switch o.Format {
	case config.OutputFormatCSV:
		filePath = fmt.Sprintf("%s/%s.csv", o.BasePath, uniqueName)
	default:
		return nil, nil, fmt.Errorf("unsupported output format: %s", o.Format)
	}
	f, err := utils.CreateFileWithDir(filePath)
	if err != nil {
		ctr.Logger.Error(ctx, "failed to create file",
			logger.Value("error", err), logger.Value("on", "runAsyncProcessing"))
		return nil, nil, fmt.Errorf("failed to create file: %w", err)
	}
	switch o.Format {
	case config.OutputFormatCSV:
		writer := csv.NewWriter(f)
		if err := writer.Write(header); err != nil {
			ctr.Logger.Error(ctx, "failed to write header",
				logger.Value("error", err), logger.Value("on", "runAsyncProcessing"))
			return nil, nil, fmt.Errorf("failed to write header: %w", err)
		}
		writer.Flush()
	}
	return func(
			ctx context.Context,
			ctr *container.Container,
			data []string,
		) error {
			if !enabled {
				return nil
			}
			switch o.Format {
			case config.OutputFormatCSV:
				writer := csv.NewWriter(f)
				ctr.Logger.Debug(ctx, "Writing data to csv",
					logger.Value("data", data), logger.Value("on", "runAsyncProcessing"))
				if err := writer.Write(data); err != nil {
					ctr.Logger.Error(ctx, "failed to write data to csv",
						logger.Value("error", err), logger.Value("on", "runAsyncProcessing"))
				}
				writer.Flush()
				return nil
			}
			return fmt.Errorf("unsupported output format: %s", o.Format)
		}, func() error {
			return f.Close()
		}, nil
}

// MetricsDataWriteFactory returns the MetricsDataWrite function
// func (o LocalOutput) MetricsDataWriteFactory(
// 	ctx context.Context,
// 	ctr *container.Container,
// 	enabled bool,
// 	uniqueName string,
// ) (MetricsDataWrite, Close, error) {
// 	var filePath string
// 	switch o.Format {
// 	case config.OutputFormatCSV:
// 		filePath = fmt.Sprintf("%s/%s.csv", o.BasePath, uniqueName)
// 	default:
// 		return nil, nil, fmt.Errorf("unsupported output format: %s", o.Format)
// 	}
// 	f, err := os.Create(filePath)
// 	if err != nil {
// 		ctr.Logger.Error(ctx, "failed to create file",
// 			logger.Value("error", err), logger.Value("on", "runAsyncProcessing"))
// 		return nil, nil, fmt.Errorf("failed to create file: %w", err)
// 	}
// 	switch o.Format {
// 	case config.OutputFormatCSV:
// 		writer := csv.NewWriter(f)
// 		header := []string{"Success", "SendDatetime", "ReceivedDatetime", "Count", "ResponseTime", "StatusCode", "Data"}
// 		if err := writer.Write(header); err != nil {
// 			ctr.Logger.Error(ctx, "failed to write header",
// 				logger.Value("error", err), logger.Value("on", "runAsyncProcessing"))
// 			return nil, nil, fmt.Errorf("failed to write header: %w", err)
// 		}
// 		writer.Flush()
// 	}
// 	return func(
// 			ctx context.Context,
// 			ctr *container.Container,
// 			data WriteMetricsData,
// 		) error {
// 			if !enabled {
// 				return nil
// 			}
// 			switch o.Format {
// 			case config.OutputFormatCSV:
// 				writer := csv.NewWriter(f)
// 				ctr.Logger.Debug(ctx, "Writing data to csv",
// 					logger.Value("data", data), logger.Value("on", "runAsyncProcessing"))
// 				if err := writer.Write(data.ToSlice()); err != nil {
// 					ctr.Logger.Error(ctx, "failed to write data to csv",
// 						logger.Value("error", err), logger.Value("on", "runAsyncProcessing"))
// 				}
// 				writer.Flush()
// 				return nil
// 			}
// 			return fmt.Errorf("unsupported output format: %s", o.Format)
// 		}, func() error {
// 			return f.Close()
// 		}, nil
// }

// writeFunc := func(
// 	ctx context.Context,
// 	ctr *app.Container,
// 	data executor.ResponseContent[any],
// ) error {
// 	records := make([]string, 0, len(req.Data)+5)
// 	records = append(
// 		records,
// 		fmt.Sprintf("%v", data.Success),
// 		data.StartTime.Format(time.RFC3339Nano),
// 		data.EndTime.Format(time.RFC3339Nano),
// 		fmt.Sprintf("%v", data.ResponseTime),
// 		fmt.Sprintf("%v", data.StatusCode),
// 	)
// 	for _, d := range req.Data {
// 		jmesPathQuery := d.JMESPath
// 		result, err := jmespath.Search(jmesPathQuery, data.Res)
// 		if err != nil {
// 			ctr.Logger.Error(ctx, "failed to search jmespath",
// 				logger.Value("error", err), logger.Value("on", "metricsFetchBatch"))
// 			return fmt.Errorf("failed to search jmespath: %v", err)
// 		}
// 		if result == nil {
// 			switch d.OnNil {
// 			case "cancel":
// 				ctr.Logger.Warn(ctx, "cancel nil value",
// 					logger.Value("on", "metricsFetchBatch"))
// 				return fmt.Errorf("cancel nil value")
// 			default:
// 				ctr.Logger.Warn(ctx, "ignore nil value",
// 					logger.Value("on", "metricsFetchBatch"))
// 			}
// 		}

// 		records = append(records, fmt.Sprintf("%v", result))
// 	}

// 	writer := csv.NewWriter(file)
// 	if err := writer.Write(records); err != nil {
// 		ctr.Logger.Error(ctx, "failed to write data to csv",
// 			logger.Value("error", err), logger.Value("on", "metricsFetchBatch"))
// 	}
// 	writer.Flush()

// 	return nil
// }

var _ Output = LocalOutput{}
