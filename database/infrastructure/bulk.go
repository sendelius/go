package infrastructure

type BulkResult struct {
	Success int64 `json:"success"`
	Failed  int64 `json:"failed"`
}
