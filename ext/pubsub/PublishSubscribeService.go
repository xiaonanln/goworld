package pubsub

import (
	"github.com/xiaonanln/go-trie-tst"
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

const (
	ServiceName = "PublishSubscribeService"
)

type subscribing struct {
	subscribers entity.EntityIDSet
	//wildcardSubscribers entity.EntityIDSet
}

func newSubscribing() *subscribing {
	return &subscribing{
		subscribers: entity.EntityIDSet{},
		//wildcardSubscribers: entity.EntityIDSet{},
	}
}

//func (subs *subscribing) forSubscribers(callback func(eid common.EntityID)) {
//	for eid := range subs.subscribers {
//		callback(eid)
//	}
//}

// PublishSubscribeService is the service entity for maintain total online avatar infos
type PublishSubscribeService struct {
	entity.Entity
	tree trie_tst.TST
}

// OnInit initialize PublishSubscribeService fields
func (s *PublishSubscribeService) OnInit() {
}

// OnCreated is called when PublishSubscribeService is created
func (s *PublishSubscribeService) OnCreated() {
	gwlog.Infof("Registering PublishSubscribeService ...")
	s.DeclareService("PublishSubscribeService")
}

// Publish is called when Avatars login
func (s *PublishSubscribeService) Publish(subject string, content string) {
	gwlog.Debugf("Publish: subject=%s, content=%s", subject, content)
	subs := s.getSubscribing(subject)
	for eid := range subs.subscribers {
		s.Call(eid, "OnPublish", subject, content)
	}
}

func (s *PublishSubscribeService) getSubscribing(subject string) *subscribing {
	t := s.tree.Sub(subject)
	var subs *subscribing
	if t.Val == nil {
		subs = newSubscribing()
		t.Val = subs
	} else {
		subs = t.Val.(*subscribing)
	}
	return subs
}

// Subscribe subscribe to the specified subject
func (s *PublishSubscribeService) Subscribe(subscriber common.EntityID, subject string) {
	gwlog.Debugf("Subscribe: subject=%s, subscriber=%s", subject, subscriber)
	subs := s.getSubscribing(subject)
	subs.subscribers.Add(subscriber)
}

// Unsubscribe subscribe to the specified subject
func (s *PublishSubscribeService) Unsubscribe(subscriber common.EntityID, subject string) {
	gwlog.Debugf("Unsubscribe: subject=%s, subscriber=%s", subject, subscriber)
	subs := s.getSubscribing(subject)
	subs.subscribers.Del(subscriber)
}

// RegisterService registeres PublishSubscribeService to goworld
func RegisterService() {
	goworld.RegisterEntity(ServiceName, &PublishSubscribeService{}, false, false).DefineAttrs(map[string][]string{})
}
