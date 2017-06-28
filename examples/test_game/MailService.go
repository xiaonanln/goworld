package main

import (
	"fmt"

	"strconv"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/kvdb"
	"github.com/xiaonanln/goworld/netutil"
)

type MailService struct {
	entity.Entity
	lastMailID int64
	mailPacker netutil.MsgPacker
}

func (s *MailService) OnInit() {
	s.lastMailID = -1
	s.mailPacker = netutil.MessagePackMsgPacker{}
}

func (s *MailService) OnCreated() {
	gwlog.Info("Registering MailService ...")
	s.DeclareService("MailService")

	kvdb.Get("lastMailID", func(val string, err error) {
		if err != nil {
			gwlog.Panic(err)
		}
		var lastMailID int
		if val == "" {
			lastMailID = 0
		} else {
			lastMailID, err = strconv.Atoi(val)
			if err != nil {
				gwlog.Panic(err)
			}
		}
		s.lastMailID = int64(lastMailID)
	})
}

func (s *MailService) SendMail_Server(senderID common.EntityID, senderName string, targetID common.EntityID, data MailData) {
	gwlog.Debug("%s.SendMail: sender=%s,%s, target=%s, mail=%v", s, senderID, senderName, targetID, data)
	if s.lastMailID == -1 {
		// not ready
		gwlog.Warn("%s is not ready for send mail", s)
		return
	}

	mailID := s.genMailID()
	mailKey := s.getMailDataKeyPrefix(mailID, targetID)

	mail := map[string]interface{}{
		"senderID":   senderID,
		"senderName": senderName,
		"targetID":   targetID,
		"data":       data,
	}
	mailBytes, err := s.mailPacker.PackMsg(mail, nil)
	if err != nil {
		gwlog.Panicf("Pack mail failed: %s", err)
	}

	kvdb.Put(mailKey, string(mailBytes), func(err error) {
		if err != nil {
			gwlog.Panicf("Put mail to kvdb failed: %s", err)
		}
		gwlog.Debug("Put mail %s to KVDB succeed", mailID)
	})
}

func (s *MailService) genMailID() int64 {
	s.lastMailID += 1
	lastMailID := s.lastMailID
	kvdb.Put("lastMailID", strconv.Itoa(int(lastMailID)), func(err error) {
		gwlog.Info("Save lastMailID %d: error=%s", lastMailID, err)
	})
	return lastMailID
}

func (s *MailService) getMailDataKeyPrefix(mailID int64, targetID common.EntityID) string {
	return fmt.Sprintf("mailData$%s$%d", targetID, targetID)
}

//func (s *MailService) IsPersistent() bool {
//	return true
//}
//
//// Override the default GetPersistentData because we are not using Attrs
//func (s *MailService) GetPersistentData() map[string]interface{} {
//	return map[string]interface{}{
//		"lastMailID": s.lastMailID,
//	}
//}
//
//func (s *MailService) LoadPersistentData(data map[string]interface{}) {
//	if lastMailID, ok := data["lastMailID"]; ok {
//		s.lastMailID = typeconv.Int(lastMailID)
//	}
//}
