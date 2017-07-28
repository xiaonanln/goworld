package main

import (
	"fmt"

	"strconv"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/entity"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/kvdb"
	"github.com/xiaonanln/goworld/kvdb/types"
	"github.com/xiaonanln/goworld/netutil"
)

const (
	END_MAIL_ID = 9999999999
)

// MailService to handle mail sending & receiving
type MailService struct {
	entity.Entity // Entity type should always inherit entity.Entity
	mailPacker    netutil.MsgPacker
}

func (s *MailService) OnInit() {
	s.mailPacker = netutil.MessagePackMsgPacker{}
}

func (s *MailService) OnCreated() {
	gwlog.Info("Registering MailService ...")
	s.Attrs.SetDefault("lastMailID", 0)
	s.DeclareService("MailService")
}

func (s *MailService) SendMail(senderID common.EntityID, senderName string, targetID common.EntityID, data MailData) {
	gwlog.Debug("%s.SendMail: sender=%s,%s, target=%s, mail=%v", s, senderID, senderName, targetID, data)

	mailID := s.genMailID()
	mailKey := s.getMailKey(mailID, targetID)

	mail := map[string]interface{}{
		"senderID":   senderID,
		"senderName": senderName,
		"targetID":   targetID,
		"data":       data,
	}
	mailBytes, err := s.mailPacker.PackMsg(mail, nil)
	if err != nil {
		gwlog.Panicf("Pack mail failed: %s", err)
		s.Call(senderID, "OnSendMail", false)
	}

	kvdb.Put(mailKey, string(mailBytes), func(err error) {
		if err != nil {
			gwlog.Panicf("Put mail to kvdb failed: %s", err)
			s.Call(senderID, "OnSendMail", false)
		}
		gwlog.Debug("Put mail %s to KVDB succeed", mailKey)
		s.Call(senderID, "OnSendMail", true)
		// tell the target that you have got a mail
		s.Call(targetID, "NotifyReceiveMail")
	})
}

// GetMails request from Avatar
func (s *MailService) GetMails(avatarID common.EntityID, lastMailID int) {
	beginMailKey := s.getMailKey(lastMailID+1, avatarID)
	endMailKey := s.getMailKey(END_MAIL_ID, avatarID)

	kvdb.GetRange(beginMailKey, endMailKey, func(items []kvdb_types.KVItem, err error) {
		s.PanicOnError(err)

		var mails []interface{}
		for _, item := range items { // Parse the mails
			_, mailId := s.parseMailKey(item.Key) // eid should always equal to avatarID
			mails = append(mails, []interface{}{
				mailId, item.Val, // val is the marshalled mail
			})
		}

		s.Call(avatarID, "OnGetMails", lastMailID, mails)
	})
}

func (s *MailService) genMailID() int {
	lastMailID := s.Attrs.GetInt("lastMailID") + 1
	s.Attrs.Set("lastMailID", lastMailID)
	return lastMailID
}

func (s *MailService) getMailKey(mailID int, targetID common.EntityID) string {
	return fmt.Sprintf("mail$%s$%010d", targetID, mailID)
}

func (s *MailService) parseMailKey(mailKey string) (common.EntityID, int) {
	//	mail$WVKLioYW8i5wAAD9$0000020969
	eid := common.EntityID(mailKey[5 : 5+common.ENTITYID_LENGTH])
	mailIdStr := mailKey[5+common.ENTITYID_LENGTH+1:]
	mailId, err := strconv.Atoi(mailIdStr)
	s.PanicOnError(err)
	return eid, mailId
}
