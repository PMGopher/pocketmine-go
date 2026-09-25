package scheduler

import "pocketmine-go/pocketmine/utils"

// BulkCurlTaskOperation is a port of pocketmine\scheduler\BulkCurlTaskOperation.
type BulkCurlTaskOperation struct {
	Page           string
	TimeoutSeconds float64
	ExtraHeaders   map[string]string
}

// BulkCurlTaskResult is one BulkCurlTask result: the response, or the error (PHP's
// InternetException) that occurred.
type BulkCurlTaskResult struct {
	Result *utils.InternetRequestResult
	Err    error
}

// BulkCurlTask is a port of pocketmine\scheduler\BulkCurlTask: executes a consecutive list of
// cURL (HTTP GET) operations. The result of this AsyncTask is a list of results, one per
// operation, handed to onCompletion on the main thread.
type BulkCurlTask struct {
	AsyncTaskBase

	operations   []BulkCurlTaskOperation
	onCompletion func(results []BulkCurlTaskResult)
}

func NewBulkCurlTask(operations []BulkCurlTaskOperation, onCompletion func(results []BulkCurlTaskResult)) *BulkCurlTask {
	return &BulkCurlTask{operations: operations, onCompletion: onCompletion}
}

func (t *BulkCurlTask) OnRun() {
	results := make([]BulkCurlTaskResult, 0, len(t.operations))
	for _, op := range t.operations {
		res, err := utils.GetURL(op.Page, op.TimeoutSeconds, op.ExtraHeaders)
		results = append(results, BulkCurlTaskResult{Result: res, Err: err})
	}
	t.SetResult(results)
}

func (t *BulkCurlTask) OnCompletion() {
	results, _ := t.GetResult().([]BulkCurlTaskResult)
	t.onCompletion(results)
}
