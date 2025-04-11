package util

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
	"unicode/utf16"
)

// ParseMarkdown 解析Markdown格式
func ParseMarkdown(text string, entities []tgbotapi.MessageEntity) string {
	// 将文本转换为UTF-16序列
	utf16Text := utf16.Encode([]rune(text))
	var result strings.Builder
	lastPos := 0

	for _, entity := range entities {
		// 跳过超出范围的实体
		if entity.Offset < 0 || entity.Offset+entity.Length > len(utf16Text) {
			continue
		}

		// 添加前面的普通文本
		if entity.Offset > lastPos {
			runes := utf16.Decode(utf16Text[lastPos:entity.Offset])
			result.WriteString(string(runes))
		}

		// 获取实体对应的UTF-16片段
		entityUTF16 := utf16Text[entity.Offset : entity.Offset+entity.Length]
		entityText := string(utf16.Decode(entityUTF16))

		// 根据实体类型添加格式
		switch {
		case entity.IsBold():
			result.WriteString("*" + entityText + "*")
		case entity.IsItalic():
			result.WriteString("_" + entityText + "_")
		case entity.IsCode():
			result.WriteString("`" + entityText + "`")
		case entity.IsPre():
			language := entity.Language
			result.WriteString("```" + language + "\n" + entityText + "\n```")
		case entity.IsTextLink():
			result.WriteString("[" + entityText + "](" + entity.URL + ")")
		default:
			result.WriteString(entityText)
		}

		lastPos = entity.Offset + entity.Length
	}

	// 添加剩余文本
	if lastPos < len(utf16Text) {
		runes := utf16.Decode(utf16Text[lastPos:])
		result.WriteString(string(runes))
	}

	return result.String()
}
