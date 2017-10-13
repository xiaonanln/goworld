package async

import "github.com/xiaonanln/goworld/engine/post"

type AsyncCallback func(res interface{}, err error)

func (ac AsyncCallback) Callback(res interface{}, err error) {
	if ac != nil {
		post.Post(func() {
			ac(res, err)
		})
	}
}

//type AsyncRequest struct {
//	callback func(err error, res ...interface{})
//	finished xnsyncutil.AtomicBool
//}
//
//func NewAsyncRequest(callback func(err error, res ...interface{})) *AsyncRequest {
//	ar := &AsyncRequest{
//		callback: callback,
//	}
//	return ar
//}
//
//func (ar *AsyncRequest) Done(res ...interface{}) {
//	ar.Error(nil, res...)
//}
//
//func (ar *AsyncRequest) Error(err error, res ...interface{}) {
//	if ar == nil {
//		return
//	}
//
//	if ar.finished.Load() {
//		// request is already finished
//		gwlog.Errorf("async request is finished multiple times: error=%v, res=%v", err, res)
//		return
//	}
//
//	ar.finished.Store(true)
//
//	post.Post(func() { // post to main game routine
//		if err != nil {
//			gwlog.Errorf("async request failed with error: %v", err)
//		}
//
//		if ar.callback != nil {
//			ar.callback(err, res...)
//		}
//	})
//}
