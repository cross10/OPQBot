package events

import (
	"context"
	"encoding/json"
	"github.com/cross10/OPQBot/v2/apiBuilder"
	"strings"
)

type EventName string

const (
	EventNameGroupMsg             EventName = "ON_EVENT_GROUP_NEW_MSG"
	EventNameFriendMsg            EventName = "ON_EVENT_FRIEND_NEW_MSG"
	EventNameGroupJoin            EventName = "ON_EVENT_GROUP_JOIN"
	EventNameGroupExit            EventName = "ON_EVENT_GROUP_EXIT"
	EventNameGroupInvite          EventName = "ON_EVENT_GROUP_INVITE"
	EventNameGroupSystemMsgNotify EventName = "ON_EVENT_GROUP_SYSTEM_MSG_NOTIFY"
	EventNameLoginSuccess         EventName = "ON_EVENT_LOGIN_SUCCESS"
	EventNameNetworkChange        EventName = "ON_EVENT_NETWORK_CHANGE"
)

type MsgType int

const (
	// MsgTypeGroupMsg 	收到群消息
	MsgTypeGroupMsg MsgType = 82
	// MsgTypeFriendsMsg 	收到好友私聊消息
	MsgTypeFriendsMsg MsgType = 166
	// MsgTypeGroupJoin 有人进群了
	MsgTypeGroupJoin MsgType = 33
	// MsgTypeGroupExit 有人退群了
	MsgTypeGroupExit MsgType = 34
	/*
	  1. 发出去消息的回应
	  2. 有人撤回消息
	  3. 自己被邀请入群
	*/
	MsgTypeMsgSent MsgType = 732
	// MsgTypeGroupChange 自己的群名片被改了
	MsgTypeGroupChange MsgType = 528
)

type EventCallbackFunc func(ctx context.Context, event IEvent)

type IEvent interface {
	ICommonMsg
	GetCurrentQQ() int64
	GetRawBytes() []byte
	GetEventName() EventName
	ParseGroupMsg() IGroupMsg
	ParseLoginSuccessEvent() ILoginSuccess
	ParseNetworkChangeEvent() INetworkChange
	GetApiBuilder() apiBuilder.IMainFunc
	ExcludeBot() IEvent
}
type IGroupMsg interface {
	ICommonMsg
	ExcludeAtInfo() IGroupMsg
	AtBot() bool
	GetGroupUin() int64
	GetSenderUin() int64
	ParseTextMsg() ITextMsg
	ContainedPic() bool
	ContainedAt() bool
}
type ITextMsg interface {
	GetTextContent() string
}
type ILoginSuccess interface {
	GetLoginSuccessBot() (nick string, uin int64)
}
type INetworkChange interface {
	GetNetworkChangeBot() (nick string, uin int64, content string)
}
type ICommonMsg interface {
	GetMsgUid() int64
	GetMsgType() MsgType
	GetMsgTime() int64
}

func New(apiUrl string, data []byte, middleCallFunc ...func(*apiBuilder.Builder) bool) (IEvent, error) {
	event := &EventStruct{apiUrl: apiUrl, middleCallFunc: middleCallFunc}
	event.rawEvent = data
	return event, json.Unmarshal(data, event)
}

type EventStruct struct {
	middleCallFunc []func(builder *apiBuilder.Builder) bool
	apiUrl         string
	rawEvent       []byte
	CurrentPacket  struct {
		EventData struct {
			Nick    *string `json:"Nick,omitempty"`
			Uin     *int64  `json:"Uin,omitempty"`
			Content *string `json:"Content,omitempty"`
			MsgHead *struct {
				FromUin    int64  `json:"FromUin"`
				ToUin      int64  `json:"ToUin"`
				FromType   int    `json:"FromType"`
				SenderUin  int64  `json:"SenderUin"`
				SenderNick string `json:"SenderNick"`
				MsgType    int    `json:"MsgType"`
				C2CCmd     int    `json:"C2cCmd"`
				MsgSeq     int64  `json:"MsgSeq"`
				MsgTime    int64  `json:"MsgTime"`
				MsgRandom  int64  `json:"MsgRandom"`
				MsgUid     int64  `json:"MsgUid"`
				GroupInfo  struct {
					GroupCard    string `json:"GroupCard"`
					GroupCode    int    `json:"GroupCode"`
					GroupInfoSeq int    `json:"GroupInfoSeq"`
					GroupLevel   int    `json:"GroupLevel"`
					GroupRank    int    `json:"GroupRank"`
					GroupType    int    `json:"GroupType"`
					GroupName    string `json:"GroupName"`
				} `json:"GroupInfo"`
				C2CTempMessageHead interface{} `json:"C2CTempMessageHead"`
			} `json:"MsgHead,omitempty"`
			MsgBody *struct {
				SubMsgType int    `json:"SubMsgType"`
				Content    string `json:"Content"`
				Images     []struct {
					FileId   int64  `json:"FileId"`
					FileMd5  string `json:"FileMd5"`
					FileSize int    `json:"FileSize"`
					Url      string `json:"Url"`
				} `json:"Images"`
				AtUinLists []struct {
					Nick string `json:"Nick"`
					Uin  int64  `json:"Uin"`
				} `json:"AtUinLists"`
				Video interface{} `json:"Video"`
				Voice interface{} `json:"Voice"`
			} `json:"MsgBody,omitempty"`
		} `json:"EventData"`
		EventName string `json:"EventName"`
	} `json:"CurrentPacket"`
	CurrentQQ int64 `json:"CurrentQQ"`
}

func (e *EventStruct) ParseNetworkChangeEvent() INetworkChange {
	return e
}
func (e *EventStruct) ParseLoginSuccessEvent() ILoginSuccess {
	return e
}
func (e *EventStruct) GetNetworkChangeBot() (nick string, uin int64, content string) {
	return *e.CurrentPacket.EventData.Nick, *e.CurrentPacket.EventData.Uin, *e.CurrentPacket.EventData.Content
}
func (e *EventStruct) GetAtList() (list []int64) {
	for _, v := range e.CurrentPacket.EventData.MsgBody.AtUinLists {
		list = append(list, v.Uin)
	}
	return list
}
func (e *EventStruct) ContainedPic() bool {
	return e.CurrentPacket.EventData.MsgBody.Images != nil
}
func (e *EventStruct) ContainedAt() bool {
	return e.CurrentPacket.EventData.MsgBody.AtUinLists != nil
}
func (e *EventStruct) AtBot() (at bool) {
	for _, v := range e.CurrentPacket.EventData.MsgBody.AtUinLists {
		if v.Uin == e.CurrentQQ {
			at = true
			break
		}
	}
	return at
}
func (e *EventStruct) ExcludeAtInfo() IGroupMsg {
	for _, v := range e.CurrentPacket.EventData.MsgBody.AtUinLists {
		e.CurrentPacket.EventData.MsgBody.Content = strings.ReplaceAll(e.CurrentPacket.EventData.MsgBody.Content, "@"+v.Nick, "")
	}
	e.CurrentPacket.EventData.MsgBody.Content = strings.TrimSpace(e.CurrentPacket.EventData.MsgBody.Content)
	return e
}

func (e *EventStruct) ExcludeBot() IEvent {
	if e.CurrentQQ == e.CurrentPacket.EventData.MsgHead.FromUin {
		return nil
	} else {
		return e
	}
}
func (e *EventStruct) GetSenderUin() int64 {
	return e.CurrentPacket.EventData.MsgHead.SenderUin
}
func (e *EventStruct) GetApiBuilder() apiBuilder.IMainFunc {
	return apiBuilder.NewApi(e.apiUrl, e.CurrentQQ, e.middleCallFunc...)
}
func (e *EventStruct) GetMsgType() MsgType {
	return MsgType(e.CurrentPacket.EventData.MsgHead.MsgType)
}
func (e *EventStruct) GetMsgTime() int64 {
	return e.CurrentPacket.EventData.MsgHead.MsgTime
}

func (e *EventStruct) GetTextContent() string {
	return e.CurrentPacket.EventData.MsgBody.Content
}
func (e *EventStruct) ParseTextMsg() ITextMsg {
	return e
}
func (e *EventStruct) ParseGroupMsg() IGroupMsg {
	return e
}
func (e *EventStruct) GetMsgUid() int64 {
	return e.CurrentPacket.EventData.MsgHead.MsgUid
}
func (e *EventStruct) GetGroupUin() int64 {
	return e.CurrentPacket.EventData.MsgHead.FromUin
}
func (e *EventStruct) GetCurrentQQ() int64 {
	return e.CurrentQQ
}
func (e *EventStruct) GetRawBytes() []byte {
	return e.rawEvent
}
func (e *EventStruct) GetEventName() EventName {
	return EventName(e.CurrentPacket.EventName)
}
func (e *EventStruct) GetLoginSuccessBot() (nick string, uin int64) {
	nick = *e.CurrentPacket.EventData.Nick
	uin = *e.CurrentPacket.EventData.Uin
	return
}
