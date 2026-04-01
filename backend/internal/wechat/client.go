package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Envelope struct {
	Code     int             `json:"Code"`
	Success  bool            `json:"Success"`
	Message  string          `json:"Message"`
	Data     json.RawMessage `json:"Data"`
	Data62   string          `json:"Data62"`
	DeviceID string          `json:"DeviceId"`
	Debug    string          `json:"Debug"`
	ID       int64           `json:"ID"`
}

type QRPayload struct {
	QRBase64    string `json:"QrBase64"`
	UUID        string `json:"Uuid"`
	QRURL       string `json:"QrUrl"`
	ExpiredTime string `json:"ExpiredTime"`
}

type LoginCheckPayload struct {
	UUID                    string `json:"uuid"`
	Status                  int    `json:"status"`
	PushLoginURLExpiredTime int    `json:"pushLoginUrlexpiredTime"`
	ExpiredTime             int    `json:"expiredTime"`
}

type ProxyInfo struct {
	ProxyIP       string `json:"ProxyIp"`
	ProxyPassword string `json:"ProxyPassword"`
	ProxyUser     string `json:"ProxyUser"`
}

type AwakenLoginPayload struct {
	Uuid        string `json:"Uuid"`
	CheckTime   int    `json:"CheckTime"`
	ExpiredTime int    `json:"ExpiredTime"`
	NotifyKey   struct {
		ILen   int64  `json:"iLen"`
		Buffer string `json:"buffer"`
	} `json:"NotifyKey"`
}

type CacheInfo struct {
	Wxid       string `json:"Wxid"`
	Nickname   string `json:"NickName"`
	HeadURL    string `json:"HeadUrl"`
	Alias      string `json:"Alais"`
	Mobile     string `json:"Mobile"`
	DeviceID   string `json:"Deviceid_str"`
	DeviceName string `json:"DeviceName"`
}

type InitPayload struct {
	CurrentSyncKey struct {
		Buffer string `json:"buffer"`
	} `json:"CurrentSynckey"`
	MaxSyncKey struct {
		Buffer string `json:"buffer"`
	} `json:"MaxSynckey"`
	ModUserInfos []struct {
		NickName   BuiltinString `json:"NickName"`
		BindMobile BuiltinString `json:"BindMobile"`
		Alias      string        `json:"Alias"`
		Signature  string        `json:"Signature"`
	} `json:"ModUserInfos"`
}

type BuiltinString struct {
	String string `json:"string"`
}

type SyncMessage struct {
	MsgID        int64         `json:"MsgId"`
	NewMsgID     int64         `json:"NewMsgId"`
	MsgType      int64         `json:"MsgType"`
	FromUserName BuiltinString `json:"FromUserName"`
	ToUserName   BuiltinString `json:"ToUserName"`
	Content      BuiltinString `json:"Content"`
	CreateTime   int64         `json:"CreateTime"`
	MsgSeq       int64         `json:"MsgSeq"`
	Status       int64         `json:"Status"`
	ImgStatus    int64         `json:"ImgStatus"`
	ImgBuf       struct {
		ILen   int64  `json:"iLen"`
		Buffer string `json:"buffer"`
	} `json:"ImgBuf"`
	MsgSource string `json:"MsgSource"`
}

type SyncPayload struct {
	AddMsgs      []SyncMessage `json:"AddMsgs"`
	ContinueFlag int64         `json:"ContinueFlag"`
	Status       int64         `json:"Status"`
	KeyBuf       struct {
		ILen   int64  `json:"iLen"`
		Buffer string `json:"buffer"`
	} `json:"KeyBuf"`
}

type ContactListPayload struct {
	CurrentWxcontactSeq       int64    `json:"CurrentWxcontactSeq"`
	CurrentChatRoomContactSeq int64    `json:"CurrentChatRoomContactSeq"`
	ContinueFlag              int64    `json:"CountinueFlag"`
	ContactUsernameList       []string `json:"ContactUsernameList"`
}

type ContactDetail struct {
	UserName        BuiltinString `json:"UserName"`
	NickName        BuiltinString `json:"NickName"`
	Pyinitial       BuiltinString `json:"Pyinitial"`
	QuanPin         BuiltinString `json:"QuanPin"`
	Alias           string        `json:"Alias"`
	VerifyFlag      int64         `json:"VerifyFlag"`
	Remark          BuiltinString `json:"Remark"`
	RemarkPyinitial BuiltinString `json:"RemarkPyinitial"`
	RemarkQuanPin   BuiltinString `json:"RemarkQuanPin"`
	BigHeadImgURL   string        `json:"BigHeadImgUrl"`
	SmallHeadImgURL string        `json:"SmallHeadImgUrl"`
	Signature       string        `json:"Signature"`
	Province        string        `json:"Province"`
	City            string        `json:"City"`
	ChatRoomOwner   string        `json:"ChatRoomOwner"`
	NewChatroomData struct {
		MemberCount    int64 `json:"MemberCount"`
		ChatRoomMember []struct {
			UserName           string `json:"UserName"`
			NickName           string `json:"NickName"`
			ChatroomMemberFlag int64  `json:"ChatroomMemberFlag"`
			InviterUserName    string `json:"InviterUserName"`
		} `json:"ChatRoomMember"`
	} `json:"NewChatroomData"`
}

type ContactDetailPayload struct {
	ContactList []ContactDetail `json:"ContactList"`
	Ticket      []struct {
		Username       string `json:"Username"`
		AntispamTicket string `json:"Antispamticket"`
	} `json:"Ticket"`
}

type FriendSearchPayload struct {
	UserName        BuiltinString             `json:"UserName"`
	NickName        BuiltinString             `json:"NickName"`
	Alias           string                    `json:"Alias"`
	Signature       string                    `json:"Signature"`
	Province        string                    `json:"Province"`
	City            string                    `json:"City"`
	BigHeadImgURL   string                    `json:"BigHeadImgUrl"`
	SmallHeadImgURL string                    `json:"SmallHeadImgUrl"`
	AntispamTicket  string                    `json:"AntispamTicket"`
	ContactCount    uint32                    `json:"ContactCount"`
	ContactList     []FriendSearchContactItem `json:"Contactlist"`
}

type FriendSearchContactItem struct {
	UserName        BuiltinString `json:"UserName"`
	NickName        BuiltinString `json:"NickName"`
	Alias           string        `json:"alias"`
	Signature       string        `json:"signature"`
	Province        string        `json:"province"`
	City            string        `json:"city"`
	BigHeadImgURL   string        `json:"bigHeadImgUrl"`
	SmallHeadImgURL string        `json:"smallHeadImgUrl"`
	AntispamTicket  string        `json:"antispamTicket"`
	VerifyFlag      uint32        `json:"verifyFlag"`
	MatchType       uint32        `json:"matchType"`
}

type FriendRequestPayload struct {
	Username string `json:"Username"`
}

type FriendStatePayload struct {
	BaseResponse struct {
		Ret    int64         `json:"ret"`
		ErrMsg BuiltinString `json:"errMsg"`
	} `json:"BaseResponse"`
	OpenID         string `json:"Openid"`
	NickName       string `json:"NickName"`
	HeadImgURL     string `json:"HeadImgUrl"`
	Sign           string `json:"Sign"`
	FriendRelation uint32 `json:"FriendRelation"`
}

type ChatRoomInfoPayload struct {
	ContactList []struct {
		UserName        BuiltinString `json:"UserName"`
		NickName        BuiltinString `json:"NickName"`
		Remark          BuiltinString `json:"Remark"`
		SmallHeadImgURL string        `json:"SmallHeadImgUrl"`
		BigHeadImgURL   string        `json:"BigHeadImgUrl"`
		ChatRoomOwner   string        `json:"ChatRoomOwner"`
		NewChatroomData struct {
			MemberCount int64 `json:"MemberCount"`
		} `json:"NewChatroomData"`
	} `json:"ContactList"`
}

type ChatRoomInfoDetailPayload struct {
	Announcement            string `json:"Announcement"`
	AnnouncementEditor      string `json:"AnnouncementEditor"`
	AnnouncementPublishTime int64  `json:"AnnouncementPublishTime"`
	ChatRoomInfoVersion     int64  `json:"ChatRoomInfoVersion"`
	ChatRoomStatus          int64  `json:"ChatRoomStatus"`
	ChatRoomBusinessType    int64  `json:"ChatRoomBusinessType"`
}

type ChatRoomMemberDetailPayload struct {
	NewChatroomData struct {
		MemberCount    int64 `json:"MemberCount"`
		ChatRoomMember []struct {
			UserName           string `json:"UserName"`
			NickName           string `json:"NickName"`
			DisplayName        string `json:"DisplayName"`
			BigHeadImgURL      string `json:"BigHeadImgUrl"`
			SmallHeadImgURL    string `json:"SmallHeadImgUrl"`
			ChatroomMemberFlag int64  `json:"ChatroomMemberFlag"`
			InviterUserName    string `json:"InviterUserName"`
		} `json:"ChatRoomMember"`
	} `json:"NewChatroomData"`
}

type GroupAddFriendPayload struct {
	ChatRoom string `json:"ChatRoom"`
	Scene    int    `json:"Scene"`
	Opcode   int32  `json:"Opcode"`
	Items    []struct {
		TargetWxid string      `json:"TargetWxid"`
		Username   string      `json:"Username"`
		V1         string      `json:"V1"`
		V2         string      `json:"V2"`
		Success    bool        `json:"Success"`
		Message    string      `json:"Message"`
		VerifyData interface{} `json:"VerifyData"`
	} `json:"Items"`
}

type FinderUserPreparePayload struct {
	BaseResponse struct {
		Ret    int64         `json:"ret"`
		ErrMsg BuiltinString `json:"errMsg"`
	} `json:"baseResponse"`
	ActionType                      int64                           `json:"actionType"`
	UserFlag                        int64                           `json:"userFlag"`
	NicknameModifyWording           string                          `json:"nicknameModifyWording"`
	IsNonresidentRealtimeLocation   bool                            `json:"isNonresidentRealtimeLocation"`
	IsNonresidentWxacctLocation     bool                            `json:"isNonresidentWxacctLocation"`
	IsNonresidentFinderacctLocation bool                            `json:"isNonresidentFinderacctLocation"`
	VerifyInfo                      FinderNicknameVerifyInfoPayload `json:"verifyInfo"`
	SelfContact                     FinderContactPayload            `json:"selfContact"`
}

type FinderNicknameVerifyInfoPayload struct {
	VerifyPrefix   string `json:"verifyPrefix"`
	BannerWording  string `json:"bannerWording"`
	VerifyLink     string `json:"verifyLink"`
	AppName        string `json:"appname"`
	VerifyNickname string `json:"verifyNickname"`
	HeadImgURL     string `json:"headImgUrl"`
	ErrScene       int64  `json:"errScene"`
}

type FinderContactPayload struct {
	Username     string                      `json:"username"`
	Nickname     string                      `json:"nickname"`
	HeadURL      string                      `json:"headUrl"`
	Seq          uint64                      `json:"seq"`
	Signature    string                      `json:"signature"`
	FollowFlag   int64                       `json:"followFlag"`
	FollowTime   int64                       `json:"followTime"`
	CoverImgURL  string                      `json:"coverImgUrl"`
	SpamStatus   int64                       `json:"spamStatus"`
	ExtFlag      int64                       `json:"extFlag"`
	OriginalFlag int64                       `json:"originalFlag"`
	AuthInfo     FinderAuthInfoPayload       `json:"authInfo"`
	ExtInfo      FinderContactExtInfoPayload `json:"extInfo"`
}

type FinderAuthInfoPayload struct {
	RealName       string                `json:"realName"`
	AuthIconType   int64                 `json:"authIconType"`
	AuthProfession string                `json:"authProfession"`
	DetailLink     string                `json:"detailLink"`
	AppName        string                `json:"appName"`
	AuthGuarantor  *FinderContactPayload `json:"authGuarantor"`
}

type FinderContactExtInfoPayload struct {
	Country    string `json:"country"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Sex        int64  `json:"sex"`
	BirthYear  int64  `json:"birthYear"`
	BirthMonth int64  `json:"birthMonth"`
	BirthDay   int64  `json:"birthDay"`
}

type SendTextPayload struct {
	Count int64 `json:"Count"`
	List  []struct {
		Ret         int64         `json:"Ret"`
		ToUserName  BuiltinString `json:"ToUsetName"`
		MsgID       int64         `json:"MsgId"`
		ClientMsgID int64         `json:"ClientMsgid"`
		CreateTime  int64         `json:"Createtime"`
		ServerTime  int64         `json:"servertime"`
		Type        int64         `json:"Type"`
		NewMsgID    int64         `json:"NewMsgId"`
	} `json:"List"`
}

type SendMessagePayload struct {
	Count int64            `json:"Count"`
	List  []SendMessageAck `json:"List"`
}

type SendMessageAck struct {
	Ret         int64         `json:"Ret"`
	ToUserName  BuiltinString `json:"ToUserName"`
	ToUsetName  BuiltinString `json:"ToUsetName"`
	MsgID       int64         `json:"MsgId"`
	ClientMsgID int64         `json:"ClientMsgid"`
	CreateTime  int64         `json:"Createtime"`
	ServerTime  int64         `json:"servertime"`
	Type        int64         `json:"Type"`
	NewMsgID    int64         `json:"NewMsgId"`
}

type DownloadImagePayload struct {
	BaseResponse struct {
		Ret    int64         `json:"ret"`
		ErrMsg BuiltinString `json:"errMsg"`
	} `json:"BaseResponse"`
	MsgID        uint32 `json:"msgId"`
	NewMsgID     uint64 `json:"newMsgId"`
	TotalLen     uint32 `json:"totalLen"`
	StartPos     uint32 `json:"startPos"`
	DataLen      uint32 `json:"dataLen"`
	CompressType uint32 `json:"compressType"`
	Data         struct {
		ILen   int64  `json:"iLen"`
		Buffer string `json:"buffer"`
	} `json:"data"`
}

type DownloadImageInput struct {
	Wxid         string `json:"Wxid"`
	ToWxid       string `json:"ToWxid"`
	MsgID        uint32 `json:"MsgId"`
	DataLen      int    `json:"DataLen"`
	CompressType int    `json:"CompressType"`
	Section      struct {
		StartPos uint32 `json:"StartPos"`
		DataLen  uint32 `json:"DataLen"`
	} `json:"Section"`
}

type MomentStringPayload struct {
	ILen   int64  `json:"iLen"`
	Buffer string `json:"buffer"`
}

type MomentCommentPayload struct {
	Username       string `json:"Username"`
	Nickname       string `json:"Nickname"`
	Source         uint32 `json:"Source"`
	Type           uint32 `json:"Type"`
	Content        string `json:"Content"`
	CreateTime     uint32 `json:"CreateTime"`
	CommentID      int64  `json:"CommentId"`
	ReplyCommentID int64  `json:"ReplyCommentId"`
	ReplyUsername  string `json:"ReplyUsername"`
	DeleteFlag     uint32 `json:"DeleteFlag"`
}

type MomentObjectPayload struct {
	ID              uint64                 `json:"Id"`
	Username        string                 `json:"Username"`
	Nickname        string                 `json:"Nickname"`
	CreateTime      uint32                 `json:"CreateTime"`
	ObjectDesc      MomentStringPayload    `json:"ObjectDesc"`
	LikeFlag        uint32                 `json:"LikeFlag"`
	LikeCount       uint32                 `json:"LikeCount"`
	LikeUserList    []MomentCommentPayload `json:"LikeUserList"`
	CommentCount    uint32                 `json:"CommentCount"`
	CommentUserList []MomentCommentPayload `json:"CommentUserList"`
	WithUserCount   uint32                 `json:"WithUserCount"`
	BlackListCount  uint32                 `json:"BlackListCount"`
	DeleteFlag      uint32                 `json:"DeleteFlag"`
	ReferUsername   string                 `json:"ReferUsername"`
	ReferID         uint64                 `json:"ReferId"`
}

type MomentTimelinePayload struct {
	BaseResponse struct {
		Ret    int64         `json:"ret"`
		ErrMsg BuiltinString `json:"errMsg"`
	} `json:"BaseResponse"`
	FirstPageMd5   string                `json:"FirstPageMd5"`
	FristPageMd5   string                `json:"FristPageMd5"`
	ObjectCount    uint32                `json:"ObjectCount"`
	ObjectList     []MomentObjectPayload `json:"ObjectList"`
	NewRequestTime uint32                `json:"NewRequestTime"`
	ContinueID     []uint64              `json:"ContinueId"`
	RetTips        []string              `json:"RetTips"`
}

type MomentDetailPayload struct {
	BaseResponse struct {
		Ret    int64         `json:"ret"`
		ErrMsg BuiltinString `json:"errMsg"`
	} `json:"baseResponse"`
	Object MomentObjectPayload `json:"object"`
}

type MomentPostPayload struct {
	BaseResponse struct {
		Ret    int64         `json:"ret"`
		ErrMsg BuiltinString `json:"errMsg"`
	} `json:"BaseResponse"`
	SnsObjectID uint64 `json:"SnsObjectId"`
	ObjectDesc  struct {
		ID uint64 `json:"Id"`
	} `json:"ObjectDesc"`
}

type MomentOperationPayload struct {
	BaseResponse struct {
		Ret    int64         `json:"ret"`
		ErrMsg BuiltinString `json:"errMsg"`
	} `json:"BaseResponse"`
}

type FavoriteSyncItem struct {
	FavID      int32  `json:"FavId"`
	Type       int32  `json:"Type"`
	Flag       uint32 `json:"Flag"`
	UpdateTime uint32 `json:"UpdateTime"`
	UpdateSeq  uint32 `json:"UpdateSeq"`
}

type FavoriteSyncPayload struct {
	Ret    int32 `json:"Ret"`
	List   []FavoriteSyncItem
	KeyBuf struct {
		ILen   int64  `json:"iLen"`
		Buffer string `json:"buffer"`
	} `json:"KeyBuf"`
}

type FavoriteDetailObject struct {
	FavID      uint32 `json:"FavId"`
	Status     int32  `json:"Status"`
	Object     string `json:"Object"`
	Flag       uint32 `json:"Flag"`
	UpdateTime uint32 `json:"UpdateTime"`
	UpdateSeq  uint32 `json:"UpdateSeq"`
}

type FavoriteDetailPayload struct {
	BaseResponse struct {
		Ret    int64         `json:"ret"`
		ErrMsg BuiltinString `json:"errMsg"`
	} `json:"BaseResponse"`
	Count      uint32                 `json:"Count"`
	ObjectList []FavoriteDetailObject `json:"ObjectList"`
}

type FavoriteDeletePayload struct {
	BaseResponse struct {
		Ret    int64         `json:"ret"`
		ErrMsg BuiltinString `json:"errMsg"`
	} `json:"BaseResponse"`
	Count uint32 `json:"Count"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *Client) CreateLoginQR(ctx context.Context, platform, deviceName string) (*Envelope, *QRPayload, []byte, error) {
	path := loginPath(platform)
	body := map[string]any{
		"Proxy":      map[string]any{},
		"DeviceID":   "",
		"DeviceName": deviceName,
		"LoginType":  "",
	}
	envelope, raw, err := c.post(ctx, path, nil, body)
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &QRPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) CheckLoginQR(ctx context.Context, uuid string) (*Envelope, []byte, error) {
	return c.post(ctx, "/Login/LoginCheckQR", map[string]string{"uuid": uuid}, nil)
}

func (c *Client) AwakenLogin(ctx context.Context, wxid string) (*Envelope, *AwakenLoginPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Login/LoginAwaken", nil, map[string]any{
		"Wxid": wxid,
		"Proxy": ProxyInfo{
			ProxyIP:       "",
			ProxyPassword: "",
			ProxyUser:     "",
		},
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &AwakenLoginPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) GetCacheInfo(ctx context.Context, wxid string) (*Envelope, *CacheInfo, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Login/GetCacheInfo", map[string]string{"wxid": wxid}, nil)
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &CacheInfo{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) Init(ctx context.Context, wxid string) (*Envelope, *InitPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Login/Newinit", map[string]string{
		"wxid":           wxid,
		"MaxSynckey":     "",
		"CurrentSynckey": "",
	}, nil)
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &InitPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) StartHeartbeat(ctx context.Context, wxid string) (*Envelope, []byte, error) {
	return c.post(ctx, "/Login/AutoHeartBeat", map[string]string{"wxid": wxid}, nil)
}

func (c *Client) StopHeartbeat(ctx context.Context, wxid string) (*Envelope, []byte, error) {
	return c.post(ctx, "/Login/CloseAutoHeartBeat", map[string]string{"wxid": wxid}, nil)
}

func (c *Client) Logout(ctx context.Context, wxid string) (*Envelope, []byte, error) {
	return c.post(ctx, "/Login/LogOut", map[string]string{"wxid": wxid}, nil)
}

func (c *Client) SyncMessages(ctx context.Context, wxid, syncKey string) (*Envelope, *SyncPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Msg/Sync", nil, map[string]any{
		"Wxid":    wxid,
		"Scene":   0,
		"Synckey": syncKey,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &SyncPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) GetContractList(ctx context.Context, wxid string, wxSeq, chatRoomSeq int64) (*Envelope, *ContactListPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Friend/GetContractList", nil, map[string]any{
		"Wxid":                      wxid,
		"CurrentWxcontactSeq":       wxSeq,
		"CurrentChatRoomContactSeq": chatRoomSeq,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &ContactListPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) GetContractDetail(ctx context.Context, wxid string, toWxids []string) (*Envelope, *ContactDetailPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Friend/GetContractDetail", nil, map[string]any{
		"Wxid":     wxid,
		"Towxids":  strings.Join(toWxids, ","),
		"ChatRoom": "",
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &ContactDetailPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) SearchFriend(ctx context.Context, wxid, keyword string, fromScene, searchScene int32) (*Envelope, *FriendSearchPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Friend/Search", nil, map[string]any{
		"Wxid":        wxid,
		"ToUserName":  keyword,
		"FromScene":   fromScene,
		"SearchScene": searchScene,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &FriendSearchPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) SendFriendRequest(
	ctx context.Context,
	wxid, v1, v2, verifyContent string,
	scene int64,
	opcode int32,
) (*Envelope, *FriendRequestPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Friend/SendRequest", nil, map[string]any{
		"Wxid":          wxid,
		"V1":            v1,
		"V2":            v2,
		"Opcode":        opcode,
		"Scene":         scene,
		"VerifyContent": verifyContent,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &FriendRequestPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			payload = &FriendRequestPayload{}
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) GetFriendState(ctx context.Context, wxid, targetWxid string) (*Envelope, *FriendStatePayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Friend/GetFriendRelation", nil, map[string]any{
		"Wxid":     wxid,
		"UserName": targetWxid,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &FriendStatePayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			payload = &FriendStatePayload{}
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) DownloadImage(ctx context.Context, input DownloadImageInput) (*Envelope, *DownloadImagePayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Tools/DownloadImg", nil, input)
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &DownloadImagePayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) GetMomentsList(ctx context.Context, wxid string, maxID uint64, firstPageMd5 string) (*Envelope, *MomentTimelinePayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/FriendCircle/GetList", nil, map[string]any{
		"Wxid":         wxid,
		"Maxid":        maxID,
		"Fristpagemd5": firstPageMd5,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &MomentTimelinePayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) GetMomentDetail(ctx context.Context, wxid, toWxid string, id uint64) (*Envelope, *MomentDetailPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/FriendCircle/GetIdDetail", nil, map[string]any{
		"Wxid":   wxid,
		"Towxid": toWxid,
		"Id":     id,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &MomentDetailPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) PublishMoment(ctx context.Context, wxid, content, blackList, withUserList string) (*Envelope, *MomentPostPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/FriendCircle/Messages", nil, map[string]any{
		"Wxid":         wxid,
		"Content":      content,
		"BlackList":    blackList,
		"WithUserList": withUserList,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &MomentPostPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) OperateMoment(ctx context.Context, wxid, id string, actionType, commentID uint32) (*Envelope, *MomentOperationPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/FriendCircle/Operation", nil, map[string]any{
		"Wxid":      wxid,
		"Id":        id,
		"Type":      actionType,
		"CommnetId": commentID,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &MomentOperationPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) SyncFavorites(ctx context.Context, wxid, keyBuf string) (*Envelope, *FavoriteSyncPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Favor/Sync", nil, map[string]any{
		"Wxid":   wxid,
		"Keybuf": keyBuf,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &FavoriteSyncPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) GetFavoriteItem(ctx context.Context, wxid string, favID int32) (*Envelope, *FavoriteDetailPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Favor/GetFavItem", nil, map[string]any{
		"Wxid":  wxid,
		"FavId": favID,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &FavoriteDetailPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) DeleteFavorite(ctx context.Context, wxid string, favID int32) (*Envelope, *FavoriteDeletePayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Favor/Del", nil, map[string]any{
		"Wxid":  wxid,
		"FavId": favID,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &FavoriteDeletePayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) DeleteFriend(ctx context.Context, wxid, targetWxid string) (*Envelope, []byte, error) {
	return c.post(ctx, "/Friend/Delete", nil, map[string]any{
		"Wxid":   wxid,
		"ToWxid": targetWxid,
	})
}

func (c *Client) BlacklistFriend(ctx context.Context, wxid, targetWxid string, val int32) (*Envelope, []byte, error) {
	return c.post(ctx, "/Friend/Blacklist", nil, map[string]any{
		"Wxid":   wxid,
		"ToWxid": targetWxid,
		"Val":    val,
	})
}

func (c *Client) GetChatRoomInfo(ctx context.Context, wxid, qid string) (*Envelope, *ChatRoomInfoPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Group/GetChatRoomInfo", nil, map[string]any{
		"Wxid": wxid,
		"QID":  qid,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &ChatRoomInfoPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) GetChatRoomInfoDetail(ctx context.Context, wxid, qid string) (*Envelope, *ChatRoomInfoDetailPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Group/GetChatRoomInfoDetail", nil, map[string]any{
		"Wxid": wxid,
		"QID":  qid,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &ChatRoomInfoDetailPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) GetChatRoomMemberDetail(ctx context.Context, wxid, qid string) (*Envelope, *ChatRoomMemberDetailPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Group/GetChatRoomMemberDetail", nil, map[string]any{
		"Wxid": wxid,
		"QID":  qid,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &ChatRoomMemberDetailPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) GetFinderUserPrepare(ctx context.Context, wxid string) (*Envelope, *FinderUserPreparePayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Finder/UserPrepare", map[string]string{
		"wxid": wxid,
	}, nil)
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &FinderUserPreparePayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) SetChatRoomName(ctx context.Context, wxid, qid, content string) (*Envelope, []byte, error) {
	return c.post(ctx, "/Group/SetChatRoomName", nil, map[string]any{
		"Wxid":    wxid,
		"QID":     qid,
		"Content": content,
	})
}

func (c *Client) SetChatRoomAnnouncement(ctx context.Context, wxid, qid, content string) (*Envelope, []byte, error) {
	return c.post(ctx, "/Group/SetChatRoomAnnouncement", nil, map[string]any{
		"Wxid":    wxid,
		"QID":     qid,
		"Content": content,
	})
}

func (c *Client) SetChatRoomRemarks(ctx context.Context, wxid, qid, content string) (*Envelope, []byte, error) {
	return c.post(ctx, "/Group/SetChatRoomRemarks", nil, map[string]any{
		"Wxid":    wxid,
		"QID":     qid,
		"Content": content,
	})
}

func (c *Client) MoveContractList(ctx context.Context, wxid, qid string, val uint32) (*Envelope, []byte, error) {
	return c.post(ctx, "/Group/MoveContractList", nil, map[string]any{
		"Wxid": wxid,
		"QID":  qid,
		"Val":  val,
	})
}

func (c *Client) AddChatRoomMember(ctx context.Context, wxid, qid, toWxids string) (*Envelope, []byte, error) {
	return c.post(ctx, "/Group/AddChatRoomMember", nil, map[string]any{
		"Wxid":         wxid,
		"ChatRoomName": qid,
		"ToWxids":      toWxids,
	})
}

func (c *Client) InviteChatRoomMember(ctx context.Context, wxid, qid, toWxids string) (*Envelope, []byte, error) {
	return c.post(ctx, "/Group/InviteChatRoomMember", nil, map[string]any{
		"Wxid":         wxid,
		"ChatRoomName": qid,
		"ToWxids":      toWxids,
	})
}

func (c *Client) DelChatRoomMember(ctx context.Context, wxid, qid, toWxids string) (*Envelope, []byte, error) {
	return c.post(ctx, "/Group/DelChatRoomMember", nil, map[string]any{
		"Wxid":         wxid,
		"ChatRoomName": qid,
		"ToWxids":      toWxids,
	})
}

func (c *Client) OperateChatRoomAdmin(ctx context.Context, wxid, qid, toWxids string, val int32) (*Envelope, []byte, error) {
	return c.post(ctx, "/Group/OperateChatRoomAdmin", nil, map[string]any{
		"Wxid":    wxid,
		"QID":     qid,
		"ToWxids": toWxids,
		"Val":     val,
	})
}

func (c *Client) QuitGroup(ctx context.Context, wxid, qid string) (*Envelope, []byte, error) {
	return c.post(ctx, "/Group/Quit", nil, map[string]any{
		"Wxid": wxid,
		"QID":  qid,
	})
}

func (c *Client) AddGroupFriend(ctx context.Context, wxid, qid, toWxids, verifyContent, accessVerifyTicket string, scene int, opcode int32, getContactScene int32) (*Envelope, *GroupAddFriendPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Group/AddGroupFriend", nil, map[string]any{
		"Wxid":                       wxid,
		"QID":                        qid,
		"ToWxids":                    toWxids,
		"VerifyContent":              verifyContent,
		"Scene":                      scene,
		"Opcode":                     opcode,
		"GetContactScene":            getContactScene,
		"ChatRoomAccessVerifyTicket": accessVerifyTicket,
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &GroupAddFriendPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) SendText(ctx context.Context, wxid, toWxid, content string) (*Envelope, *SendTextPayload, []byte, error) {
	envelope, raw, err := c.post(ctx, "/Msg/SendTxt", nil, map[string]any{
		"Wxid":    wxid,
		"ToWxid":  toWxid,
		"Content": content,
		"Type":    1,
		"At":      "",
	})
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &SendTextPayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			return nil, nil, raw, err
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) UploadImage(ctx context.Context, wxid, toWxid, base64 string) (*Envelope, *SendMessagePayload, []byte, error) {
	return c.sendMessage(ctx, "/Msg/UploadImg", map[string]any{
		"Wxid":   wxid,
		"ToWxid": toWxid,
		"Base64": base64,
	})
}

func (c *Client) SendEmoji(ctx context.Context, wxid, toWxid, md5 string, totalLen int64) (*Envelope, *SendMessagePayload, []byte, error) {
	return c.sendMessage(ctx, "/Msg/SendEmoji", map[string]any{
		"Wxid":     wxid,
		"ToWxid":   toWxid,
		"Md5":      md5,
		"TotalLen": totalLen,
	})
}

func (c *Client) ShareCard(ctx context.Context, wxid, toWxid, cardWxid, cardNickname, cardAlias string) (*Envelope, *SendMessagePayload, []byte, error) {
	return c.sendMessage(ctx, "/Msg/ShareCard", map[string]any{
		"Wxid":         wxid,
		"ToWxid":       toWxid,
		"CardWxId":     cardWxid,
		"CardNickName": cardNickname,
		"CardAlias":    cardAlias,
	})
}

func (c *Client) ShareLink(ctx context.Context, wxid, toWxid, xml string) (*Envelope, *SendMessagePayload, []byte, error) {
	return c.sendMessage(ctx, "/Msg/ShareLink", map[string]any{
		"Wxid":   wxid,
		"ToWxid": toWxid,
		"Type":   5,
		"Xml":    xml,
	})
}

func (c *Client) sendMessage(ctx context.Context, path string, body map[string]any) (*Envelope, *SendMessagePayload, []byte, error) {
	envelope, raw, err := c.post(ctx, path, nil, body)
	if err != nil {
		return nil, nil, raw, err
	}
	payload := &SendMessagePayload{}
	if len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, payload); err != nil {
			payload = &SendMessagePayload{}
		}
	}
	return envelope, payload, raw, nil
}

func (c *Client) post(ctx context.Context, path string, query map[string]string, body any) (*Envelope, []byte, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		reader = bytes.NewReader(payload)
	}

	requestURL := c.baseURL + path
	if len(query) > 0 {
		params := url.Values{}
		for key, value := range query {
			params.Set(key, value)
		}
		requestURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, reader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, raw, fmt.Errorf("wechatReal http status %d", resp.StatusCode)
	}

	envelope := &Envelope{}
	if err := json.Unmarshal(raw, envelope); err != nil {
		return nil, raw, err
	}
	return envelope, raw, nil
}

func loginPath(platform string) string {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "android_pad", "logingetqrpad", "pad_latest", "pad", "android":
		return "/Login/LoginGetQRPad"
	case "win_unified", "windows_unified", "winunified", "logingetqrwinunified":
		return "/Login/LoginGetQRWinUnified"
	case "car", "logingetqrcar":
		return "/Login/LoginGetQRCar"
	case "windows", "win":
		return "/Login/LoginGetQRWin"
	case "mac":
		return "/Login/LoginGetQRMac"
	case "ipad_bypass":
		return "/Login/LoginGetQRx"
	default:
		return "/Login/LoginGetQR"
	}
}
