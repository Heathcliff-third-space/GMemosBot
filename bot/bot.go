package bot

import (
	"fmt"
	"github.com/Heathcliff-third-space/GMemosBot/memo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

var bot *tgbotapi.BotAPI

func botInit() tgbotapi.UpdatesChannel {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN 环境变量未设置")
	}

	var err error
	err = memo.LoadUsers()
	if err != nil {
		log.Fatalf("初始化Bot失败: %v", err)
	}

	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("初始化Bot失败: %v", err)
	}

	// 设置菜单
	if err := setupMenu(); err != nil {
		log.Printf("菜单设置失败: %v", err)
	} else {
		log.Println("菜单设置成功")
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	return updates
}

func Start() {
	memo.Start()
	updates := botInit()

	for update := range updates {
		msg := update.Message
		if msg == nil {
			continue
		}

		log.Printf("[%s] 收到消息 (ID: %d)", msg.From.UserName, msg.MessageID)

		if msg.IsCommand() {
			handleCommand(msg)
			continue
		}

		textContent, attachments := parseMessage(msg)
		createdMemo, err := memo.CreateMemo(textContent, attachments, msg.From.ID)
		if err != nil {
			replyContent(msg, fmt.Sprintf("操作失败: %v", err))
		} else {
			replyContent(msg, fmt.Sprintf("创建成功！%v", createdMemo.Name))
		}
	}
}

func setupMenu() error {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "开始使用Bot"},
		{Command: "token", Description: "验证您的token"},
		{Command: "info", Description: "获取你的信息"},
	}

	_, err := bot.Request(tgbotapi.NewSetMyCommands(commands...))
	return err
}

func handleCommand(msg *tgbotapi.Message) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, "")
	reply.ReplyToMessageID = msg.MessageID

	switch msg.Command() {
	case "start":
		handleStartCmd(msg, &reply)
	case "token":
		handleTokenCmd(msg, &reply)
	case "info":
		handleInfoCmd(msg, &reply)
	default:
		handleDefaultCmd(msg, &reply)
	}

	if _, err := bot.Send(reply); err != nil {
		log.Printf("发送命令回复失败: %v", err)
	}
}

func parseMessage(msg *tgbotapi.Message) (string, []*memo.Resource) {
	messageInfo, attachments := parseTextContent(msg), parseAttachments(msg)
	return messageInfo, attachments
}

func replyContent(msg *tgbotapi.Message, content string) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, content)
	reply.ReplyToMessageID = msg.MessageID
	if _, err := bot.Send(reply); err != nil {
		log.Printf("发送确认消息失败: %v", err)
	}
}
