package async

import "context"

type AsyncRequest struct {
}

func NewAsyncRequest() {
	bg := context.Background()
	context.WithCancel()
}
