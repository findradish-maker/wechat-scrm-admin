package service

import (
	"encoding/xml"
	"fmt"
	"html"
	"strings"
	"wechat-enterprise-backend/internal/domain"
	"wechat-enterprise-backend/internal/wechat"
)

type MessageArticle struct {
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	URL         string `json:"url"`
	Cover       string `json:"cover"`
	Publisher   string `json:"publisher"`
	PublishTime int64  `json:"publishTime"`
	ExtraCount  int    `json:"extraCount"`
}

type MessageQuote struct {
	Title       string `json:"title"`
	QuotedTitle string `json:"quotedTitle"`
	QuotedBy    string `json:"quotedBy"`
}

type MessageVoice struct {
	DurationMs int64 `json:"durationMs"`
	Length     int64 `json:"length"`
}

type MessageVideo struct {
	DurationSec int64 `json:"durationSec"`
	Length      int64 `json:"length"`
	Width       int64 `json:"width"`
	Height      int64 `json:"height"`
}

type MessageImage struct {
	URL      string `json:"url"`
	ThumbURL string `json:"thumbUrl"`
	Base64   string `json:"base64"`
	Length   int64  `json:"length"`
	Width    int64  `json:"width"`
	Height   int64  `json:"height"`
}

type MessageCard struct {
	Wxid     string `json:"wxid"`
	Nickname string `json:"nickname"`
	Alias    string `json:"alias"`
}

type MessageEmoji struct {
	MD5    string `json:"md5"`
	Length int64  `json:"length"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
}

type MessageSystem struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Action  string `json:"action"`
}

type appMessageEnvelope struct {
	AppMsg       appMessageXML `xml:"appmsg"`
	FromUserName string        `xml:"fromusername"`
	AppInfo      struct {
		AppName string `xml:"appname"`
	} `xml:"appinfo"`
}

type appMessageXML struct {
	Title    string      `xml:"title"`
	Desc     string      `xml:"des"`
	Type     int         `xml:"type"`
	URL      string      `xml:"url"`
	ThumbURL string      `xml:"thumburl"`
	MMReader mmReaderXML `xml:"mmreader"`
	ReferMsg referMsgXML `xml:"refermsg"`
}

type mmReaderXML struct {
	Category  mmReaderCategoryXML `xml:"category"`
	Publisher struct {
		UserName string `xml:"username"`
		NickName string `xml:"nickname"`
	} `xml:"publisher"`
}

type mmReaderCategoryXML struct {
	Name   string            `xml:"name"`
	Count  int               `xml:"count,attr"`
	TopNew mmReaderTopNewXML `xml:"topnew"`
	Items  []mmReaderItemXML `xml:"item"`
}

type mmReaderTopNewXML struct {
	Cover  string `xml:"cover"`
	Digest string `xml:"digest"`
}

type mmReaderItemXML struct {
	Title   string `xml:"title"`
	TitleV2 string `xml:"title_v2"`
	URL     string `xml:"url"`
	Summary string `xml:"summary"`
	Cover   string `xml:"cover"`
	PubTime int64  `xml:"pub_time"`
}

type referMsgXML struct {
	Content     string `xml:"content"`
	CreateTime  int64  `xml:"createtime"`
	DisplayName string `xml:"displayname"`
	FromUsr     string `xml:"fromusr"`
	Type        int    `xml:"type"`
}

type voiceMessageEnvelope struct {
	Voice voiceMessageXML `xml:"voicemsg"`
}

type voiceMessageXML struct {
	VoiceLength int64 `xml:"voicelength,attr"`
	Length      int64 `xml:"length,attr"`
}

type videoMessageEnvelope struct {
	Video videoMessageXML `xml:"videomsg"`
}

type videoMessageXML struct {
	Length      int64 `xml:"length,attr"`
	PlayLength  int64 `xml:"playlength,attr"`
	ThumbWidth  int64 `xml:"cdnthumbwidth,attr"`
	ThumbHeight int64 `xml:"cdnthumbheight,attr"`
}

type emojiMessageEnvelope struct {
	Emoji emojiMessageXML `xml:"emoji"`
}

type imageMessageEnvelope struct {
	XMLName        xml.Name `xml:"img"`
	Length         int64    `xml:"length,attr"`
	MD5            string   `xml:"md5,attr"`
	ThumbURL       string   `xml:"cdnthumburl,attr"`
	MidImageURL    string   `xml:"cdnmidimgurl,attr"`
	ThumbWidth     int64    `xml:"cdnthumbwidth,attr"`
	ThumbHeight    int64    `xml:"cdnthumbheight,attr"`
	MidImageWidth  int64    `xml:"cdnmidwidth,attr"`
	MidImageHeight int64    `xml:"cdnmidheight,attr"`
	HDImageWidth   int64    `xml:"cdnhdwidth,attr"`
	HDImageHeight  int64    `xml:"cdnhdheight,attr"`
}

type imageMessageWrapper struct {
	Image imageMessageEnvelope `xml:"img"`
}

type contactCardEnvelope struct {
	XMLName  xml.Name `xml:"msg"`
	UserName string   `xml:"username,attr"`
	NickName string   `xml:"nickname,attr"`
	Alias    string   `xml:"alias,attr"`
}

type emojiMessageXML struct {
	MD5    string `xml:"md5,attr"`
	Length int64  `xml:"len,attr"`
	Width  int64  `xml:"width,attr"`
	Height int64  `xml:"height,attr"`
}

type handoffMessageEnvelope struct {
	Op handoffOpXML `xml:"op"`
}

type handoffOpXML struct {
	ID   int           `xml:"id,attr"`
	Name string        `xml:"name"`
	Arg  handoffArgXML `xml:"arg"`
}

type handoffArgXML struct {
	HandoffLists []handoffListXML `xml:"handofflist"`
}

type handoffListXML struct {
	Opcode   int               `xml:"opcode,attr"`
	Handoffs []handoffEntryXML `xml:"handoff"`
}

type handoffEntryXML struct {
	ID                string `xml:"id,attr"`
	AppEntryPage      string `xml:"appentrypage"`
	Title             string `xml:"title"`
	DisplaySourceName string `xml:"displaySourceName"`
	AppID             string `xml:"appid"`
}

type sysMessageEnvelope struct {
	Type       string `xml:"type,attr"`
	GameCenter struct {
		Entrance struct {
			Text string `xml:"entrance_text"`
		} `xml:"entrance"`
	} `xml:"gamecenter"`
	SysmsgTemplate struct {
		ContentTemplate struct {
			Plain    string `xml:"plain"`
			Template string `xml:"template"`
		} `xml:"content_template"`
	} `xml:"sysmsgtemplate"`
}

func parseMessageSummary(
	item wechat.SyncMessage,
	ownerWxid string,
	account *domain.WechatAccount,
	contactIndex map[string]domain.WechatContact,
) MessageSummary {
	rawContent := strings.TrimSpace(item.Content.String)
	groupSenderWxid, strippedContent := splitGroupSender(item.FromUserName.String, item.ToUserName.String, rawContent)
	chatWxid, conversationType := resolveConversation(item.FromUserName.String, item.ToUserName.String, ownerWxid)
	senderWxid := resolveSenderWxid(item.FromUserName.String, conversationType, groupSenderWxid)

	summary := MessageSummary{
		MsgID:            item.MsgID,
		NewMsgID:         item.NewMsgID,
		MsgType:          item.MsgType,
		FromWxid:         item.FromUserName.String,
		ToWxid:           item.ToUserName.String,
		ChatWxid:         chatWxid,
		ChatDisplay:      resolveEntityDisplay(chatWxid, account, contactIndex),
		ConversationType: conversationType,
		SenderWxid:       senderWxid,
		SenderDisplay:    resolveEntityDisplay(senderWxid, account, contactIndex),
		Content:          strippedContent,
		CreateTime:       item.CreateTime,
		MsgSeq:           item.MsgSeq,
		IsSelf:           item.FromUserName.String == ownerWxid,
		Kind:             "unknown",
	}

	switch item.MsgType {
	case 1:
		summary.Kind = "text"
		summary.Content = normalizePlainText(strippedContent)
		summary.Preview = truncateText(summary.Content, 120)
	case 34:
		summary.Kind = "voice"
		summary.Voice = parseVoiceMessage(strippedContent)
		if summary.Voice != nil {
			summary.Preview = fmt.Sprintf("语音 %.1fs", float64(summary.Voice.DurationMs)/1000)
		} else {
			markMessageUnparsed(&summary, "voice_xml_decode_failed", strippedContent)
			summary.Preview = "语音消息"
		}
	case 3:
		summary.Kind = "image"
		summary.Image = parseImageMessage(item, strippedContent)
		if summary.Image == nil {
			markMessageUnparsed(&summary, "image_xml_decode_failed", strippedContent)
		}
		summary.Preview = "[图片]"
	case 43:
		summary.Kind = "video"
		summary.Video = parseVideoMessage(strippedContent)
		if summary.Video != nil && summary.Video.DurationSec > 0 {
			summary.Preview = fmt.Sprintf("视频 %ds", summary.Video.DurationSec)
		} else {
			if summary.Video == nil {
				markMessageUnparsed(&summary, "video_xml_decode_failed", strippedContent)
			}
			summary.Preview = "视频消息"
		}
	case 42:
		summary.Kind = "card"
		summary.Card = parseCardMessage(strippedContent)
		if summary.Card != nil {
			summary.Preview = firstNonEmpty(summary.Card.Nickname, summary.Card.Alias, summary.Card.Wxid, "名片消息")
		} else {
			markMessageUnparsed(&summary, "card_xml_decode_failed", strippedContent)
			summary.Preview = "名片消息"
		}
	case 47:
		summary.Kind = "emoji"
		summary.Emoji = parseEmojiMessage(strippedContent)
		if summary.Emoji == nil {
			markMessageUnparsed(&summary, "emoji_xml_decode_failed", strippedContent)
		}
		summary.Preview = "表情消息"
	case 49:
		parseAppMessage(&summary, strippedContent)
	case 51:
		summary.Kind = "handoff"
		summary.System = parseHandoffMessage(strippedContent)
		if summary.System != nil {
			summary.Content = summary.System.Summary
			summary.Preview = firstNonEmpty(summary.System.Title, summary.System.Summary, "系统消息")
		} else {
			summary.Kind = "system"
			markMessageUnparsed(&summary, "handoff_xml_decode_failed", strippedContent)
			summary.Preview = "系统消息"
		}
	case 10002:
		summary.Kind = "system_notice"
		summary.System = parseSystemNotice(strippedContent)
		if summary.System != nil {
			summary.Content = summary.System.Summary
			summary.Preview = firstNonEmpty(summary.System.Title, summary.System.Summary, "系统通知")
		} else {
			markMessageUnparsed(&summary, "system_notice_xml_decode_failed", strippedContent)
			summary.Preview = "系统通知"
		}
	default:
		summary.Content = normalizePlainText(strippedContent)
		summary.Preview = truncateText(summary.Content, 120)
		if xmlContent := extractXMLContent(strippedContent); xmlContent != "" {
			markMessageUnparsed(&summary, fmt.Sprintf("unsupported_msg_type_%d", item.MsgType), xmlContent)
		}
	}

	if summary.Preview == "" {
		summary.Preview = truncateText(normalizePlainText(summary.Content), 120)
	}
	if summary.ChatDisplay == "" {
		summary.ChatDisplay = summary.ChatWxid
	}
	if summary.SenderDisplay == "" {
		summary.SenderDisplay = summary.SenderWxid
	}

	return summary
}

func parseAppMessage(summary *MessageSummary, content string) {
	xmlContent := extractXMLContent(content)
	if xmlContent == "" {
		summary.Kind = "app"
		summary.Content = normalizePlainText(content)
		summary.Preview = truncateText(summary.Content, 120)
		return
	}

	var envelope appMessageEnvelope
	if err := xml.Unmarshal([]byte(xmlContent), &envelope); err != nil {
		markMessageUnparsed(summary, "app_xml_decode_failed", xmlContent)
		summary.Kind = "app"
		summary.Content = normalizePlainText(content)
		summary.Preview = truncateText(summary.Content, 120)
		return
	}

	switch envelope.AppMsg.Type {
	case 5:
		summary.Kind = "article"
		summary.Article = buildArticleCard(envelope)
		if summary.Article != nil {
			summary.Content = firstNonEmpty(summary.Article.Summary, summary.Article.Title)
			summary.Preview = firstNonEmpty(summary.Article.Title, truncateText(summary.Content, 120))
		}
	case 57:
		summary.Kind = "quote"
		summary.Quote = buildQuoteCard(envelope)
		if summary.Quote != nil {
			summary.Content = firstNonEmpty(summary.Quote.Title, summary.Quote.QuotedTitle)
			summary.Preview = firstNonEmpty(summary.Quote.Title, summary.Quote.QuotedTitle, "引用消息")
		}
	default:
		markMessageUnparsed(summary, fmt.Sprintf("unsupported_app_type_%d", envelope.AppMsg.Type), xmlContent)
		summary.Kind = "app"
		summary.Content = normalizePlainText(firstNonEmpty(envelope.AppMsg.Desc, envelope.AppMsg.Title))
		summary.Preview = firstNonEmpty(normalizePlainText(envelope.AppMsg.Title), truncateText(summary.Content, 120), "应用消息")
	}
}

func buildArticleCard(envelope appMessageEnvelope) *MessageArticle {
	item := mmReaderItemXML{}
	extraCount := 0
	if len(envelope.AppMsg.MMReader.Category.Items) > 0 {
		item = envelope.AppMsg.MMReader.Category.Items[0]
		if len(envelope.AppMsg.MMReader.Category.Items) > 1 {
			extraCount = len(envelope.AppMsg.MMReader.Category.Items) - 1
		}
	}

	title := firstNonEmpty(item.TitleV2, item.Title, envelope.AppMsg.Title)
	summary := firstNonEmpty(item.Summary, envelope.AppMsg.Desc, envelope.AppMsg.MMReader.Category.TopNew.Digest)
	cover := firstNonEmpty(item.Cover, envelope.AppMsg.ThumbURL, envelope.AppMsg.MMReader.Category.TopNew.Cover)
	url := firstNonEmpty(item.URL, envelope.AppMsg.URL)
	publisher := firstNonEmpty(envelope.AppMsg.MMReader.Publisher.NickName, envelope.AppInfo.AppName, envelope.AppMsg.MMReader.Category.Name)

	if title == "" && summary == "" && url == "" {
		return nil
	}

	return &MessageArticle{
		Title:       normalizePlainText(title),
		Summary:     normalizePlainText(summary),
		URL:         strings.TrimSpace(url),
		Cover:       strings.TrimSpace(cover),
		Publisher:   normalizePlainText(publisher),
		PublishTime: item.PubTime,
		ExtraCount:  extraCount,
	}
}

func buildQuoteCard(envelope appMessageEnvelope) *MessageQuote {
	title := normalizePlainText(envelope.AppMsg.Title)
	quotedTitle := ""
	if strings.TrimSpace(envelope.AppMsg.ReferMsg.Content) != "" {
		quotedTitle = parseNestedAppTitle(envelope.AppMsg.ReferMsg.Content)
	}
	if title == "" && quotedTitle == "" {
		return nil
	}
	return &MessageQuote{
		Title:       title,
		QuotedTitle: quotedTitle,
		QuotedBy:    normalizePlainText(envelope.AppMsg.ReferMsg.DisplayName),
	}
}

func parseNestedAppTitle(content string) string {
	xmlContent := html.UnescapeString(strings.TrimSpace(content))
	xmlContent = extractXMLContent(xmlContent)
	if xmlContent == "" {
		return truncateText(normalizePlainText(content), 80)
	}
	var envelope appMessageEnvelope
	if err := xml.Unmarshal([]byte(xmlContent), &envelope); err == nil {
		return normalizePlainText(firstNonEmpty(envelope.AppMsg.Title, envelope.AppMsg.Desc))
	}
	return truncateText(normalizePlainText(content), 80)
}

func parseVoiceMessage(content string) *MessageVoice {
	var envelope voiceMessageEnvelope
	if err := xml.Unmarshal([]byte(extractXMLContent(content)), &envelope); err != nil {
		return nil
	}
	return &MessageVoice{
		DurationMs: envelope.Voice.VoiceLength,
		Length:     envelope.Voice.Length,
	}
}

func parseVideoMessage(content string) *MessageVideo {
	var envelope videoMessageEnvelope
	if err := xml.Unmarshal([]byte(extractXMLContent(content)), &envelope); err != nil {
		return nil
	}
	return &MessageVideo{
		DurationSec: envelope.Video.PlayLength,
		Length:      envelope.Video.Length,
		Width:       envelope.Video.ThumbWidth,
		Height:      envelope.Video.ThumbHeight,
	}
}

func parseEmojiMessage(content string) *MessageEmoji {
	var envelope emojiMessageEnvelope
	if err := xml.Unmarshal([]byte(extractXMLContent(content)), &envelope); err != nil {
		return nil
	}
	return &MessageEmoji{
		MD5:    envelope.Emoji.MD5,
		Length: envelope.Emoji.Length,
		Width:  envelope.Emoji.Width,
		Height: envelope.Emoji.Height,
	}
}

func parseImageMessage(item wechat.SyncMessage, content string) *MessageImage {
	xmlContent := extractXMLContent(content)
	envelope, ok := decodeImageMessageEnvelope(xmlContent)
	if ok {
		base64 := strings.TrimSpace(item.ImgBuf.Buffer)
		if base64 != "" && !strings.HasPrefix(base64, "data:") {
			base64 = "data:image/jpeg;base64," + base64
		}

		width := envelope.MidImageWidth
		height := envelope.MidImageHeight
		if width <= 0 {
			width = envelope.ThumbWidth
		}
		if height <= 0 {
			height = envelope.ThumbHeight
		}

		return &MessageImage{
			URL:      strings.TrimSpace(envelope.MidImageURL),
			ThumbURL: strings.TrimSpace(envelope.ThumbURL),
			Base64:   base64,
			Length:   envelope.Length,
			Width:    width,
			Height:   height,
		}
	}

	base64 := strings.TrimSpace(item.ImgBuf.Buffer)
	if base64 == "" {
		return nil
	}
	if !strings.HasPrefix(base64, "data:") {
		base64 = "data:image/jpeg;base64," + base64
	}
	return &MessageImage{
		Base64: base64,
		Length: 0,
	}
}

func decodeImageMessageEnvelope(content string) (imageMessageEnvelope, bool) {
	var envelope imageMessageEnvelope
	if err := xml.Unmarshal([]byte(content), &envelope); err == nil && (envelope.Length > 0 || envelope.MD5 != "" || envelope.ThumbURL != "") {
		return envelope, true
	}

	var wrapper imageMessageWrapper
	if err := xml.Unmarshal([]byte(content), &wrapper); err == nil {
		if wrapper.Image.Length > 0 || wrapper.Image.MD5 != "" || wrapper.Image.ThumbURL != "" {
			return wrapper.Image, true
		}
	}

	return imageMessageEnvelope{}, false
}

func parseCardMessage(content string) *MessageCard {
	var envelope contactCardEnvelope
	if err := xml.Unmarshal([]byte(extractXMLContent(content)), &envelope); err != nil {
		return nil
	}
	return &MessageCard{
		Wxid:     strings.TrimSpace(envelope.UserName),
		Nickname: normalizePlainText(envelope.NickName),
		Alias:    strings.TrimSpace(envelope.Alias),
	}
}

func parseHandoffMessage(content string) *MessageSystem {
	var envelope handoffMessageEnvelope
	if err := xml.Unmarshal([]byte(extractXMLContent(content)), &envelope); err != nil {
		return nil
	}

	for _, handoffList := range envelope.Op.Arg.HandoffLists {
		for _, handoff := range handoffList.Handoffs {
			return &MessageSystem{
				Title:   firstNonEmpty(normalizePlainText(handoff.Title), normalizePlainText(handoff.DisplaySourceName), "多端接力"),
				Summary: normalizePlainText(firstNonEmpty(handoff.DisplaySourceName, handoff.AppEntryPage)),
				Action:  strings.TrimSpace(handoff.AppEntryPage),
			}
		}
	}

	return &MessageSystem{
		Title:   firstNonEmpty(normalizePlainText(envelope.Op.Name), "系统消息"),
		Summary: "多端接力状态更新",
	}
}

func parseSystemNotice(content string) *MessageSystem {
	var envelope sysMessageEnvelope
	if err := xml.Unmarshal([]byte(extractXMLContent(content)), &envelope); err != nil {
		return nil
	}

	switch envelope.Type {
	case "gamecenter":
		return &MessageSystem{
			Title:   "游戏通知",
			Summary: normalizePlainText(envelope.GameCenter.Entrance.Text),
		}
	case "sysmsgtemplate":
		return &MessageSystem{
			Title:   "群系统通知",
			Summary: normalizePlainText(firstNonEmpty(envelope.SysmsgTemplate.ContentTemplate.Template, envelope.SysmsgTemplate.ContentTemplate.Plain)),
		}
	default:
		return &MessageSystem{
			Title:   "系统通知",
			Summary: truncateText(normalizePlainText(content), 120),
		}
	}
}

func resolveConversation(fromWxid, toWxid, ownerWxid string) (string, string) {
	switch {
	case fromWxid == ownerWxid && toWxid == ownerWxid:
		return ownerWxid, "self"
	case strings.HasSuffix(fromWxid, "@chatroom"):
		return fromWxid, "group"
	case strings.HasSuffix(toWxid, "@chatroom"):
		return toWxid, "group"
	case fromWxid == ownerWxid:
		return toWxid, "direct"
	default:
		return fromWxid, "direct"
	}
}

func resolveSenderWxid(fromWxid, conversationType, groupSenderWxid string) string {
	if conversationType == "group" && strings.TrimSpace(groupSenderWxid) != "" {
		return groupSenderWxid
	}
	return fromWxid
}

func splitGroupSender(fromWxid, toWxid, content string) (string, string) {
	if !strings.HasSuffix(fromWxid, "@chatroom") && !strings.HasSuffix(toWxid, "@chatroom") {
		return "", strings.TrimSpace(content)
	}

	trimmed := strings.TrimSpace(content)
	index := strings.Index(trimmed, ":\n")
	if index <= 0 {
		return "", trimmed
	}

	candidate := strings.TrimSpace(trimmed[:index])
	if !looksLikeWechatIdentifier(candidate) || candidate == fromWxid || candidate == toWxid {
		return "", trimmed
	}

	return candidate, strings.TrimSpace(trimmed[index+2:])
}

func looksLikeWechatIdentifier(value string) bool {
	if value == "" || strings.Contains(value, " ") {
		return false
	}
	return strings.HasPrefix(value, "wxid_") ||
		strings.HasPrefix(value, "gh_") ||
		strings.Contains(value, "@chatroom") ||
		strings.Contains(value, "@openim")
}

func resolveEntityDisplay(wxid string, account *domain.WechatAccount, contacts map[string]domain.WechatContact) string {
	if wxid == "" {
		return ""
	}
	if account != nil && wxid == account.Wxid {
		return firstNonEmpty(account.Nickname, account.Alias, account.Wxid)
	}
	if contact, exists := contacts[wxid]; exists {
		return contactDisplayName(contact)
	}
	return wxid
}

func extractXMLContent(content string) string {
	trimmed := strings.TrimSpace(content)
	index := strings.Index(trimmed, "<")
	if index < 0 {
		return ""
	}
	return strings.TrimSpace(trimmed[index:])
}

func normalizePlainText(content string) string {
	text := html.UnescapeString(strings.TrimSpace(content))
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.TrimSpace(text)
}

func truncateText(content string, max int) string {
	if max <= 0 {
		return ""
	}
	trimmed := normalizePlainText(content)
	runes := []rune(trimmed)
	if len(runes) <= max {
		return trimmed
	}
	return string(runes[:max]) + "..."
}

func markMessageUnparsed(summary *MessageSummary, reason, xmlContent string) {
	if summary == nil {
		return
	}
	if strings.TrimSpace(summary.ParseStatus) == "unparsed" {
		return
	}
	xmlContent = strings.TrimSpace(xmlContent)
	if xmlContent == "" {
		xmlContent = extractXMLContent(summary.Content)
	}
	summary.ParseStatus = "unparsed"
	summary.ParseError = strings.TrimSpace(reason)
	summary.DecodeXML = xmlContent
}
