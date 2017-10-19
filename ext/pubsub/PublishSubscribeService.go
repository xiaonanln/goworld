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
func (pss *PublishSubscribeService) OnInit() {
}

// OnCreated is called when PublishSubscribeService is created
func (pss *PublishSubscribeService) OnCreated() {
	gwlog.Infof("Registering PublishSubscribeService ...")
	pss.Attrs.SetDefault("subscribers", goworld.MapAttr())
	pss.Attrs.SetDefault("wildcardSubscribers", goworld.MapAttr())
	pss.DeclareService("PublishSubscribeService")
}

// Publish is called when Avatars login
func (pss *PublishSubscribeService) Publish(subject string, content string) {
	gwlog.Debugf("Publish: subject=%pss, content=%pss", subject, content)
	for _, c := range subject {
		if c == '*' {
			gwlog.Panicf("subject should not contains '*' while publishing")
		}
	}

	pss.publishInTree(subject, content, &pss.tree, 0)
	//subs := pss.getSubscribing(subject)
	//for eid := range subs.subscribers {
	//	pss.Call(eid, "OnPublish", subject, content)
	//}
}

func (pss *PublishSubscribeService) publishInTree(subject string, content string, st *trietst.TST, idx int) {
	// call all wildcard subscribers
	subs := pss.getSubscribingOfTree(st, false)
	if subs != nil {
		for eid := range subs.wildcardSubscribers {
			//gwlog.Debugf("subject %pss matches subscribe %pss", subject, subject[:idx]+"*")
			pss.Call(eid, "OnPublish", subject, content)
		}
	}
	if idx < len(subject) {
		pss.publishInTree(subject, content, st.Child(subject[idx]), idx+1)
	} else {
		// exact match
		if subs != nil {
			for eid := range subs.subscribers {
				//gwlog.Debugf("subject %pss matches subscribe %pss", subject, subject)
				pss.Call(eid, "OnPublish", subject, content)
			}
		}
	}
}

func (pss *PublishSubscribeService) getSubscribing(subject string, newIfNotExists bool) *subscribing {
	t := pss.tree.Sub(subject)
	return pss.getSubscribingOfTree(t, newIfNotExists)
}

func (pss *PublishSubscribeService) getSubscribingOfTree(t *trietst.TST, newIfNotExists bool) *subscribing {
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
func (pss *PublishSubscribeService) Subscribe(subscriber common.EntityID, subject string) {
	gwlog.Debugf("Subscribe: subject=%pss, subscriber=%pss", subject, subscriber)

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
	pss.subscribe(subscriber, subject, wildcard)
}

func (pss *PublishSubscribeService) subscribe(subscriber common.EntityID, subject string, wildcard bool) {
	subs := pss.getSubscribing(subject, true)
	if !wildcard {
		subs.subscribers.Add(subscriber)
	} else {
		subs.wildcardSubscribers.Add(subscriber)
	}
}

// Unsubscribe subscribe to the specified subject
func (pss *PublishSubscribeService) Unsubscribe(subscriber common.EntityID, subject string) {
	gwlog.Debugf("Unsubscribe: subject=%pss, subscriber=%pss", subject, subscriber)
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
	subs := pss.getSubscribing(subject, false)
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
func (pss *PublishSubscribeService) OnFreeze() {
	subscribersAttr := pss.GetMapAttr("subscribers")
	wildcardSubscribersAttr := pss.GetMapAttr("wildcardSubscribers")

	pss.tree.ForEach(func(s string, val interface{}) {
		subs := val.(*subscribing)
		subscribersAttr.Set(s, goworld.MapAttr())
		wildcardSubscribersAttr.Set(s, goworld.MapAttr())

		for eid := range subs.subscribers {
			subscribersAttr.GetMapAttr(s).Set(string(eid), 1)
		}
		for eid := range subs.wildcardSubscribers {
			wildcardSubscribersAttr.GetMapAttr(s).Set(string(eid), 1)
		}
	})
}

// OnRestored restores subscribings from entity attrs
func (pss *PublishSubscribeService) OnRestored() {
	subscribersAttr := pss.GetMapAttr("subscribers")
	wildcardSubscribersAttr := pss.GetMapAttr("wildcardSubscribers")
	restoreCounter := 0
	subscribersAttr.ForEach(func(subject string, val interface{}) {
		eids := val.(*entity.MapAttr)
		eids.ForEach(func(eidStr string, _ interface{}) {
			eid := common.EntityID(eidStr)
			pss.subscribe(eid, subject, false)
			restoreCounter += 1
			//gwlog.Infof("%s: restored subscribing: %s -> %s", pss, eid, subject)
		})
	})

	wildcardSubscribersAttr.ForEach(func(subject string, val interface{}) {
		eids := val.(*entity.MapAttr)
		eids.ForEach(func(eidStr string, _ interface{}) {
			eid := common.EntityID(eidStr)
			pss.subscribe(eid, subject, true)
			restoreCounter += 1
			//gwlog.Infof("%s: restored subscribing: %s -> %s*", pss, eid, subject)
		})
	})
	gwlog.Infof("%s: restored %d subscribings", pss, restoreCounter)
}

// RegisterService registeres PublishSubscribeService to goworld
func RegisterService() {
	goworld.RegisterEntity(ServiceName, &PublishSubscribeService{}, false, false).DefineAttrs(map[string][]string{})
}
