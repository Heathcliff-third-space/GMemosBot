package bot

import (
	"encoding/base64"
	"fmt"
	"github.com/Heathcliff-third-space/GMemosBot/memo"
	"github.com/Heathcliff-third-space/GMemosBot/util"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

func parseTextContent(msg *tgbotapi.Message) string {

	var builder strings.Builder

	write := func(s string) {
		if _, err := builder.WriteString(s); err != nil {
			log.Printf("写入字符串失败: %v", err)
		}
	}

	forwardInfo := processForwardInfo(msg)
	if forwardInfo != "" {
		write(forwardInfo)
	}

	if msg.Text != "" {
		if msg.Entities != nil && containsMarkdown(msg.Entities) {
			write(util.ParseMarkdown(msg.Text, msg.Entities) + "\n")
		} else {
			write(msg.Text + "\n")
		}
	}

	if msg.Caption != "" {
		if msg.CaptionEntities != nil && containsMarkdown(msg.CaptionEntities) {
			write(util.ParseMarkdown(msg.Caption, msg.CaptionEntities) + "\n")
		} else {
			write(msg.Caption + "\n")
		}
	}

	return builder.String()
}

func processForwardInfo(msg *tgbotapi.Message) string {
	if msg.ForwardFrom == nil && msg.ForwardFromChat == nil {
		return ""
	}

	var builder strings.Builder

	if msg.ForwardFrom != nil {
		// 从用户转发
		name := msg.ForwardFrom.FirstName + " " + msg.ForwardFrom.LastName
		name = strings.TrimSpace(name)
		if name == "" {
			name = msg.ForwardFrom.UserName
		}
		if name == "" {
			name = "用户" + strconv.FormatInt(msg.ForwardFrom.ID, 10)
		}

		// 构建用户链接
		link := fmt.Sprintf("tg://user?id=%d", msg.ForwardFrom.ID)
		if msg.ForwardFromMessageID != 0 {
			chatID := msg.ForwardFrom.ID
			link = buildMessageLink(chatID, msg.ForwardFromMessageID)
		}

		builder.WriteString("Forwarded from: [" + name + "](" + link + ")")

	} else if msg.ForwardFromChat != nil {
		// 从聊天/频道转发
		name := msg.ForwardFromChat.Title
		if name == "" {
			name = "聊天" + strconv.FormatInt(msg.ForwardFromChat.ID, 10)
		}

		// 构建聊天链接
		link := fmt.Sprintf("https://t.me/%s", msg.ForwardFromChat.UserName)
		if msg.ForwardFromChat.UserName == "" {
			link = fmt.Sprintf("tg://resolve?domain=c%s", getChannelID(msg.ForwardFromChat.ID))
		}

		if msg.ForwardFromMessageID != 0 {
			link = buildMessageLink(msg.ForwardFromChat.ID, msg.ForwardFromMessageID)
		}

		builder.WriteString("Forwarded from: [" + name + "](" + link + ")")
	}

	if builder.Len() > 0 {
		builder.WriteString("\n")
	}

	return builder.String()
}

// 构建消息链接
func buildMessageLink(chatID int64, messageID int) string {
	// 处理超级群组/频道的ID格式（去掉-100前缀）
	strChatID := strconv.FormatInt(chatID, 10)
	if strings.HasPrefix(strChatID, "-100") {
		strChatID = strChatID[4:]
	}
	return fmt.Sprintf("https://t.me/c/%s/%d", strChatID, messageID)
}

// 获取频道ID（处理超级群组ID）
func getChannelID(chatID int64) string {
	strChatID := strconv.FormatInt(chatID, 10)
	if strings.HasPrefix(strChatID, "-100") {
		return strChatID[4:]
	}
	return strChatID
}

// 获取消息中的所有附件
func parseAttachments(msg *tgbotapi.Message) []*memo.Resource {
	var resources []*memo.Resource

	// 处理图片组 (Photo数组)
	if len(msg.Photo) > 0 {
		bestPhoto := getBestQualityPhoto(msg.Photo)
		resources = append(resources, &memo.Resource{
			FileId:   bestPhoto.FileID,
			Filename: fmt.Sprintf("photo_%s_%d.jpg", bestPhoto.FileUniqueID, bestPhoto.FileSize),
		})
	}

	// 处理文档
	if msg.Document != nil {
		resources = append(resources, &memo.Resource{
			FileId:   msg.Document.FileID,
			Filename: msg.Document.FileName,
		})
	}

	// 处理音频
	if msg.Audio != nil {
		resources = append(resources, &memo.Resource{
			FileId:   msg.Audio.FileID,
			Filename: getFileName(msg.Audio.FileName, "audio", msg.MessageID, ".mp3"),
		})
	}

	// 处理视频
	if msg.Video != nil {
		resources = append(resources, &memo.Resource{
			FileId:   msg.Video.FileID,
			Filename: getFileName(msg.Video.FileName, "video", msg.MessageID, ".mp4"),
		})
	}

	// 处理语音
	if msg.Voice != nil {
		resources = append(resources, &memo.Resource{
			FileId:   msg.Voice.FileID,
			Filename: fmt.Sprintf("voice_%d.ogg", msg.MessageID),
		})
	}

	// 处理视频笔记
	if msg.VideoNote != nil {
		resources = append(resources, &memo.Resource{
			FileId:   msg.VideoNote.FileID,
			Filename: fmt.Sprintf("video_note_%d.mp4", msg.MessageID),
		})
	}

	// 下载文件，将其转换成base64编码
	if len(resources) > 0 {
		for _, att := range resources {
			parseAttachment(att)
		}
	}

	return resources
}

func getBestQualityPhoto(photos []tgbotapi.PhotoSize) *tgbotapi.PhotoSize {
	if len(photos) == 0 {
		return nil
	}

	// 选择文件尺寸最大的图片(通常是原图)
	var bestPhoto *tgbotapi.PhotoSize
	maxSize := 0

	for i := range photos {
		if photos[i].FileSize > maxSize {
			maxSize = photos[i].FileSize
			bestPhoto = &photos[i]
		}
	}

	return bestPhoto
}

// 生成文件名
func getFileName(originalName string, prefix string, msgID int, defaultExt string) string {
	if originalName != "" {
		return originalName
	}
	return fmt.Sprintf("%s_%d%s", prefix, msgID, defaultExt)
}

// 下载单个附件
func parseAttachment(resource *memo.Resource) {
	// 获取文件URL
	fileConfig := tgbotapi.FileConfig{FileID: resource.FileId}
	file, err := bot.GetFile(fileConfig)
	if err != nil {
		return
	}

	// 构建完整文件URL
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", bot.Token, file.FilePath)

	// 创建HTTP请求
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return
	}

	// 使用bot的客户端执行请求
	resp, err := bot.Client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("关闭响应体失败: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return
	}

	// 读取文件内容
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// 检测MIME类型
	mimeType := mime.TypeByExtension(filepath.Ext(resource.Filename))
	if mimeType == "" {
		// 使用内容检测作为后备方案
		mimeType = http.DetectContentType(data)
	}

	// Base64编码
	base64Str := base64.StdEncoding.EncodeToString(data)

	resource.Type = mimeType
	resource.Content = base64Str
}

// 检查是否包含Markdown格式
func containsMarkdown(entities []tgbotapi.MessageEntity) bool {
	for _, entity := range entities {
		if entity.IsBold() || entity.IsItalic() || entity.IsCode() || entity.IsPre() || entity.IsTextLink() {
			return true
		}
	}
	return false
}
