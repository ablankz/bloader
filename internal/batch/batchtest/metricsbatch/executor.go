package metricsbatch

type MetricsThreadExecutor struct {
	writer  MetricsWriter
	fetcher MetricsFetcher
	closer  func()
}
