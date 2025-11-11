package worker

var workers = make(map[string]Worker)

func init() {
	workers["ExampleWorker"] = &ExampleWorker{}
}
