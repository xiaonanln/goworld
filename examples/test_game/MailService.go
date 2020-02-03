package main

import (
	"fmt"

	"strconv"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/kvdb"
	"github.com/xiaonanln/goworld/engine/kvdb/types"
	"github.com/xiaonanln/goworld/engine/netutil"
)

const (
	// END_MAIL_ID is the max possible Mail ID
	END_MAIL_ID = 9999999999
)

// MailService to handle mail sending & receiving
type MailService struct {
	entity.Entity // Entity type should always inherit entity.Entity
	mailPacker    netutil.MsgPacker
	lastMailID    int
}

func (s *MailService) DescribeEntityType(desc *entity.EntityTypeDesc) {
}

// OnInit is called when initializing MailService
func (s *MailService) OnInit() {
	s.mailPacker = netutil.MessagePackMsgPacker{}
	s.lastMailID = -1
}

// OnCreated is called when MailService is created
func (s *MailService) OnCreated() {
	gwlog.Infof("Registering MailService ...")
	kvdb.GetOrPut("MailService:lastMailID", "0", func(oldVal string, err error) {
		if oldVal == "" {
			s.lastMailID = 0
		} else {
			var err error
			s.lastMailID, err = strconv.Atoi(oldVal)
			if err != nil {
				gwlog.Panicf("MailService: lastMailID is invalid: %#v", oldVal)
			}
		}
	})
}

// SendMail handles send mail requests from avatars
func (s *MailService) SendMail(senderID common.EntityID, senderName string, targetID common.EntityID, data MailData) {
	gwlog.Debugf("%s.SendMail: sender=%s,%s, target=%s, mail=%v", s, senderID, senderName, targetID, data)

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
		gwlog.Debugf("Put mail %s to KVDB succeed", mailKey)
		s.Call(senderID, "OnSendMail", true)
		// tell the target that you have got a mail
		s.Call(targetID, "NotifyReceiveMail")
	})
}

// GetMails handle get mails requests from avatars
func (s *MailService) GetMails(avatarID common.EntityID, lastMailID int) {
	beginMailKey := s.getMailKey(lastMailID+1, avatarID)
	endMailKey := s.getMailKey(END_MAIL_ID, avatarID)

	kvdb.GetRange(beginMailKey, endMailKey, func(items []kvdbtypes.KVItem, err error) {
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
	if s.lastMailID < 0 {
		gwlog.Panicf("MailService: lastMailId=%v (not loaded successfully)", s.lastMailID)
	}

	s.lastMailID += 1
	kvdb.Put("MailService:lastMailID", strconv.Itoa(s.lastMailID), func(err error) {
		if err != nil {
			gwlog.Panicf("MailService: save lastMailID failed: %+v", err)
		} else {
			gwlog.Debugf("MailService: save lastMailID = %+v", s.lastMailID)
		}
	})
	return s.lastMailID
}

func (s *MailService) getMailKey(mailID int, targetID common.EntityID) string {
	return fmt.Sprintf("MailService:mail$%s$%010d", targetID, mailID)
}

func (s *MailService) parseMailKey(mailKey string) (common.EntityID, int) {
	//	mail$WVKLioYW8i5wAAD9$0000020969
	eid := common.EntityID(mailKey[5 : 5+common.ENTITYID_LENGTH])
	mailIdStr := mailKey[5+common.ENTITYID_LENGTH+1:]
	mailId, err := strconv.Atoi(mailIdStr)
	s.PanicOnError(err)
	return eid, mailId
}
