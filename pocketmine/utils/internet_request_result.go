package utils

// InternetRequestResult is a port of pocketmine\utils\InternetRequestResult.
//
// The PHP original is a private-fields-plus-getters DTO; in Go the idiomatic equivalent for
// a plain immutable data holder is just exported fields, so the getters are dropped.
type InternetRequestResult struct {
	Headers []map[string]string
	Body    string
	Code    int
}
