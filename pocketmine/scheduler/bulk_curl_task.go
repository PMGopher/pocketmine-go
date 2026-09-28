package scheduler

import (
	"strings"

	"pocketmine-go/pocketmine/utils"
)

// BulkCurlTaskOperation is a port of pocketmine\scheduler\BulkCurlTaskOperation.
//
// PHP's operation also takes raw cURL options; the only ones the server passes (TimingsCommand's
// upload: CURLOPT_POST + CURLOPT_POSTFIELDS) are PostFields here: a non-empty value makes the
// operation a POST with that body.
type BulkCurlTaskOperation struct {
	Page           string
	TimeoutSeconds float64
	ExtraHeaders   map[string]string
	PostFields     string
}

// BulkCurlTaskResult is one BulkCurlTask result: the response, or the error (PHP's
// InternetException) that occurred.
type BulkCurlTaskResult struct {
	Result *utils.InternetRequestResult
	Err    error
}

// BulkCurlTask is a port of pocketmine\scheduler\BulkCurlTask: executes a consecutive list of
// cURL (HTTP) operations. The result of this AsyncTask is a list of results, one per
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
		var res *utils.InternetRequestResult
		var err error
		if op.PostFields != "" {
			res, err = utils.PostURL(op.Page, strings.NewReader(op.PostFields), op.TimeoutSeconds, op.ExtraHeaders)
		} else {
			res, err = utils.GetURL(op.Page, op.TimeoutSeconds, op.ExtraHeaders)
		}
		results = append(results, BulkCurlTaskResult{Result: res, Err: err})
	}
	t.SetResult(results)
}

func (t *BulkCurlTask) OnCompletion() {
	results, _ := t.GetResult().([]BulkCurlTaskResult)
	t.onCompletion(results)
}
