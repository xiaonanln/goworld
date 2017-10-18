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
	subscribers         entity.EntityIDSet
	wildcardSubscribers entity.EntityIDSet
}

func newSubscribing() *subscribing {
	return &subscribing{
		subscribers:         entity.EntityIDSet{},
		wildcardSubscribers: entity.EntityIDSet{},
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
	tree trietst.TST
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
	for _, c := range subject {
		if c == '*' {
			gwlog.Panicf("subject should not contains '*' while publishing")
		}
	}

	s.publishInTree(subject, content, &s.tree, 0)
	//subs := s.getSubscribing(subject)
	//for eid := range subs.subscribers {
	//	s.Call(eid, "OnPublish", subject, content)
	//}
}

func (s *PublishSubscribeService) publishInTree(subject string, content string, st *trietst.TST, idx int) {
	// call all wildcard subscribers
	subs := s.getSubscribingOfTree(st, false)
	if subs != nil {
		for eid := range subs.wildcardSubscribers {
			//gwlog.Debugf("subject %s matches subscribe %s", subject, subject[:idx]+"*")
			s.Call(eid, "OnPublish", subject, content)
		}
	}
	if idx < len(subject) {
		s.publishInTree(subject, content, st.Child(subject[idx]), idx+1)
	} else {
		// exact match
		if subs != nil {
			for eid := range subs.subscribers {
				//gwlog.Debugf("subject %s matches subscribe %s", subject, subject)
				s.Call(eid, "OnPublish", subject, content)
			}
		}
	}
}

func (s *PublishSubscribeService) getSubscribing(subject string, newIfNotExists bool) *subscribing {
	t := s.tree.Sub(subject)
	return s.getSubscribingOfTree(t, newIfNotExists)
}

func (s *PublishSubscribeService) getSubscribingOfTree(t *trietst.TST, newIfNotExists bool) *subscribing {
	var subs *subscribing
	if t.Val == nil {
		if newIfNotExists {
			subs = newSubscribing()
			t.Val = subs
		}
	} else {
		subs = t.Val.(*subscribing)
	}
	return subs
}

// Subscribe subscribe to the specified subject
// subject can endswith '*' which matches any zero or more characters
// for example, if an entity subscribe to 'apple.*', it will receive published message on 'apple.', 'apple.1', 'apple.2', etc
// There can be only one '*' at the end of subject while subscribing, same for unsubscribing
func (s *PublishSubscribeService) Subscribe(subscriber common.EntityID, subject string) {
	gwlog.Debugf("Subscribe: subject=%s, subscriber=%s", subject, subscriber)

	for i, c := range subject {
		if c == '*' && i != len(subject)-1 {
			gwlog.Panicf("'*' can only be used at the end of subject while subscribing")
		}
	}

	wildcard := false
	if subject != "" && subject[len(subject)-1] == '*' {
		// subject ends with *
		wildcard = true
		subject = subject[:len(subject)-1]
	}
	subs := s.getSubscribing(subject, true)
	if !wildcard {
		subs.subscribers.Add(subscriber)
	} else {
		subs.wildcardSubscribers.Add(subscriber)
	}
}

// Unsubscribe subscribe to the specified subject
func (s *PublishSubscribeService) Unsubscribe(subscriber common.EntityID, subject string) {
	gwlog.Debugf("Unsubscribe: subject=%s, subscriber=%s", subject, subscriber)
	for i, c := range subject {
		if c == '*' && i != len(subject)-1 {
			gwlog.Panicf("'*' can only be used at the end of subject while unsubscribing")
		}
	}

	wildcard := false
	if subject != "" && subject[len(subject)-1] == '*' {
		// subject ends with *
		wildcard = true
		subject = subject[:len(subject)-1]
	}
	subs := s.getSubscribing(subject, false)
	if subs == nil {
		return
	}

	if !wildcard {
		subs.subscribers.Del(subscriber)
	} else {
		subs.wildcardSubscribers.Del(subscriber)
	}
}

// OnFreeze converts all subscribings to entity attrs
func (s *PublishSubscribeService) OnFreeze() {

}

// OnRestored restores subscribings from entity attrs
func (s *PublishSubscribeService) OnRestored() {

}

// RegisterService registeres PublishSubscribeService to goworld
func RegisterService() {
	goworld.RegisterEntity(ServiceName, &PublishSubscribeService{}, false, false).DefineAttrs(map[string][]string{})
}
