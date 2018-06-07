package main

import (
	"fmt"
	"math/rand"
	"os"

	"strconv"

	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/ext/pubsub"
	"github.com/xiaonanln/typeconv"
)

// Avatar entity which is the player itself
type Avatar struct {
	entity.Entity // Entity type should always inherit entity.Entity
}

func (a *Avatar) DescribeEntityType(desc *entity.EntityTypeDesc) {
	desc.SetPersistent(true).SetUseAOI(true)
	desc.DefineAttr("name", "AllClients", "Persistent")
	desc.DefineAttr("level", "AllClients", "Persistent")
	desc.DefineAttr("prof", "AllClients", "Persistent")
	desc.DefineAttr("exp", "Client", "Persistent")
	desc.DefineAttr("mails", "Client", "Persistent")
	desc.DefineAttr("spaceKind", "Persistent")
	desc.DefineAttr("lastMailID", "Persistent")
	desc.DefineAttr("testListField", "AllClients")
	desc.DefineAttr("enteringNilSpace")
}

func (a *Avatar) OnInit() {

}

func (a *Avatar) OnAttrsReady() {
	a.setDefaultAttrs()
	gwlog.Debugf("Avatar %s is ready: client=%s, mails=%d", a, a.GetClient(), a.Attrs.GetMapAttr("mails").Size())
	//a.Msgbox.SetMsgHandler(a.handleMsgboxMsg)
}

// OnCreated is called when avatar is created
func (a *Avatar) OnCreated() {

	a.SetFilterProp("spaceKind", strconv.Itoa(int(a.GetInt("spaceKind"))))
	a.SetFilterProp("level", strconv.Itoa(int(a.GetInt("level"))))
	a.SetFilterProp("prof", strconv.Itoa(int(a.GetInt("prof"))))

	//gwlog.Debugf("Found OnlineService: %s", onlineServiceEid)
	a.CallService("OnlineService", "CheckIn", a.ID, a.Attrs.GetStr("name"), a.Attrs.GetInt("level"))
	for _, subject := range _TEST_PUBLISH_SUBSCRIBE_SUBJECTS { // subscribe all subjects
		a.CallService(pubsub.ServiceName, "Subscribe", a.ID, subject)
	}

	//a.AddTimer(time.Second, "PerSecondTick", 1, "")
}

// PerSecondTick is ticked per second, if timer is setup
func (a *Avatar) PerSecondTick(arg1 int, arg2 string) {
	fmt.Fprint(os.Stderr, "!")
}

func (a *Avatar) setDefaultAttrs() {
	a.Attrs.SetDefaultStr("name", "无名")
	a.Attrs.SetDefaultInt("level", 1)
	a.Attrs.SetDefaultInt("exp", 0)
	a.Attrs.SetDefaultInt("prof", int64(1+rand.Intn(4)))
	a.Attrs.SetDefaultInt("spaceKind", int64(1+rand.Intn(100)))
	a.Attrs.SetDefaultInt("lastMailID", 0)
	a.Attrs.SetDefaultMapAttr("mails", goworld.MapAttr())
	a.Attrs.SetDefaultListAttr("testListField", goworld.ListAttr())
	a.Attrs.SetDefaultBool("enteringNilSpace", false)
}

// TestListField_Client is a test RPC for client
func (a *Avatar) TestListField_Client() {
	testListField := a.GetListAttr("testListField")
	if testListField.Size() > 0 && rand.Float32() < 0.3333333333 {
		testListField.PopInt()
	} else if testListField.Size() > 0 && rand.Float32() < 0.5 {
		testListField.SetInt(rand.Intn(testListField.Size()), int64(rand.Intn(100)))
	} else {
		testListField.AppendInt(int64(rand.Intn(100)))
	}
	a.CallClient("OnTestListField", testListField.ToList())
}

func (a *Avatar) enterSpace(spaceKind int) {
	if a.Space.Kind == spaceKind {
		return
	}
	if consts.DEBUG_SPACES {
		gwlog.Infof("%s enter space from %d => %d", a, a.Space.Kind, spaceKind)
	}
	a.CallService("SpaceService", "EnterSpace", a.ID, spaceKind)
}

// OnClientConnected is called when client is connected
func (a *Avatar) OnClientConnected() {
	//gwlog.Infof("%s.OnClientConnected: current space = %s", a, a.Space)
	//a.Attrs.Set("exp", a.Attrs.GetInt("exp")+1)
	//a.Attrs.Set("testpop", 1)
	//v := a.Attrs.Pop("testpop")
	//gwlog.Infof("Avatar pop testpop => %v", v)
	//
	//a.Attrs.Set("subattr", goworld.MapAttr())
	//subattr := a.Attrs.GetMapAttr("subattr")
	//subattr.Set("a", 1)
	//subattr.Set("b", 1)
	//subattr = a.Attrs.PopMapAttr("subattr")
	//a.Attrs.Set("subattr", subattr)

	a.SetFilterProp("online", "0")
	a.SetFilterProp("online", "1")
	a.enterSpace(int(a.GetInt("spaceKind")))
}

// OnClientDisconnected is called when client is lost
func (a *Avatar) OnClientDisconnected() {
	gwlog.Infof("%s client disconnected", a)
	a.Destroy()
}

// EnterSpace_Client is enter space RPC for client
func (a *Avatar) EnterSpace_Client(kind int) {
	a.enterSpace(kind)
}

// DoEnterSpace is called by SpaceService to notify avatar entering specified space
func (a *Avatar) DoEnterSpace(kind int, spaceID common.EntityID) {
	// let the avatar enter space with spaceID
	a.EnterSpace(spaceID, a.randomPosition())
}

func (a *Avatar) randomPosition() entity.Vector3 {
	minCoord, maxCoord := -400, 400
	return entity.Vector3{
		X: entity.Coord(minCoord + rand.Intn(maxCoord-minCoord)),
		Y: 0,
		Z: entity.Coord(minCoord + rand.Intn(maxCoord-minCoord)),
	}
}

// OnEnterSpace is called when avatar enters a space
func (a *Avatar) OnEnterSpace() {
	if consts.DEBUG_SPACES {
		gwlog.Infof("%s ENTER SPACE %s", a, a.Space)
	}
}

// GetSpaceID is a server RPC to query avatar space ID
func (a *Avatar) GetSpaceID(callerID common.EntityID) {
	a.Call(callerID, "OnGetAvatarSpaceID", a.ID, a.Space.ID)
}

func (a *Avatar) EnterRandomNilSpace_Client() {
	gameIDs := goworld.ListGameIDs()
	gameid := gameIDs[rand.Intn(len(gameIDs))]
	nilSpaceID := goworld.GetNilSpaceID(gameid)
	gwlog.Debugf("%s EnterRandomNilSpace: %s on game%d", a, nilSpaceID, gameid)
	a.Attrs.SetBool("enteringNilSpace", true)

	//goworld.GetEntity()
	if goworld.GetSpace(nilSpaceID) != nil {
		// the nil space is local
		a.Attrs.SetBool("enteringNilSpace", false)
		a.EnterSpace(nilSpaceID, goworld.Vector3{})
		a.CallClient("OnEnterRandomNilSpace")
	} else {
		a.EnterSpace(nilSpaceID, goworld.Vector3{})
	}
}

// OnDestroy is called when avatar is destroying
func (a *Avatar) OnMigrateIn() {
	gwlog.Debugf("%s OnMigrateIn ...", a)
	if a.Attrs.GetBool("enteringNilSpace") {
		a.Attrs.Del("enteringNilSpace")
		a.CallClient("OnEnterRandomNilSpace")
	}
}

// OnDestroy is called when avatar is destroying
func (a *Avatar) OnDestroy() {
	a.CallService("OnlineService", "CheckOut", a.ID)
	// unsubscribe all subjects
	a.CallService(pubsub.ServiceName, "UnsubscribeAll", a.ID)
}

// SendMail_Client is a client RPC to send mail to others
func (a *Avatar) SendMail_Client(targetID common.EntityID, mail MailData) {
	a.CallService("MailService", "SendMail", a.ID, a.GetStr("name"), targetID, mail)
}

// OnSendMail is called by MailService to notify sending mail succeed
func (a *Avatar) OnSendMail(ok bool) {
	a.CallClient("OnSendMail", ok)
}

// NotifyReceiveMail is called by MailService to notify Avatar of receiving any mail
func (a *Avatar) NotifyReceiveMail() {
	//a.CallService("MailService", "GetMails", a.ID)
}

// GetMails_Client is a RPC for clients to retrive mails
func (a *Avatar) GetMails_Client() {
	a.CallService("MailService", "GetMails", a.ID, a.GetInt("lastMailID"))
}

// OnGetMails is called by MailService to send mails to avatar
func (a *Avatar) OnGetMails(lastMailID int, mails []interface{}) {
	//gwlog.Infof("%s.OnGetMails: lastMailID=%v/%v, mails=%v", a, a.GetInt("lastMailID"), lastMailID, mails)
	if lastMailID != int(a.GetInt("lastMailID")) {
		gwlog.Warnf("%s.OnGetMails: lastMailID mismatch: local=%v, return=%v", a, a.GetInt("lastMailID"), lastMailID)
		a.CallClient("OnGetMails", false)
		return
	}

	mailsAttr := a.Attrs.GetMapAttr("mails")
	for _, _item := range mails {
		item := _item.([]interface{})
		mailId := int(typeconv.Int(item[0]))
		if mailId <= int(a.GetInt("lastMailID")) {
			gwlog.Panicf("mail ID should be increasing")
		}
		if mailsAttr.HasKey(strconv.Itoa(mailId)) {
			gwlog.Errorf("mail %d received multiple times", mailId)
			continue
		}
		mail := typeconv.String(item[1])
		mailsAttr.SetStr(strconv.Itoa(mailId), mail)
		a.Attrs.SetInt("lastMailID", int64(mailId))
	}

	a.CallClient("OnGetMails", true)
}

// Say_Client is client RPC for chatting
func (a *Avatar) Say_Client(channel string, content string) {
	gwlog.Debugf("Say @%s: %s", channel, content)
	if channel == "world" {
		a.CallFilteredClients("", "=", "", "OnSay", a.ID, a.GetStr("name"), channel, content)
	} else if channel == "prof" {
		profStr := strconv.Itoa(int(a.GetInt("prof")))
		a.CallFilteredClients("prof", "=", profStr, "OnSay", a.ID, a.GetStr("name"), channel, content)
	} else {
		gwlog.Panicf("%s.Say_Client: invalid channel: %s", a, channel)
	}
}

// Move_Client is client RPC for moving
func (a *Avatar) Move_Client(pos entity.Vector3) {
	gwlog.Debugf("Move from %s -> %s", a.GetPosition(), pos)
	a.SetPosition(pos)
}

var _TEST_PUBLISH_SUBSCRIBE_SUBJECTS = []string{"monster", "npc", "item", "avatar", "boss_*"}

// TestPublish_Client is client RPC for Publish/Subscribe testing
func (a *Avatar) TestPublish_Client() {
	subject := _TEST_PUBLISH_SUBSCRIBE_SUBJECTS[rand.Intn(len(_TEST_PUBLISH_SUBSCRIBE_SUBJECTS))]
	if subject[len(subject)-1] == '*' {
		subject = subject[:len(subject)-1] + strconv.Itoa(rand.Intn(100))
	}
	a.CallService(pubsub.ServiceName, "Publish", subject, fmt.Sprintf("%s: hello %s, this is a test publish message", a.ID, subject))
}

func (a *Avatar) OnPublish(subject string, content string) {
	var publisher common.EntityID
	publisher = common.EntityID(content[:common.ENTITYID_LENGTH])
	gwlog.Debugf("OnPublish: publisher=%s, subject=%s, content=%s", publisher, subject, content)
	a.CallClient("OnTestPublish", publisher, subject, content)
}
