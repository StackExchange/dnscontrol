package cache

// FlushResult represent the result of a cache-flush.
type FlushResult struct {
	Count  int    `json:"count"`
	Result string `json:"result"`
}
