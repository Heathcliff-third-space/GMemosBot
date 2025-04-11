package bot

import (
	"fmt"
	"github.com/Heathcliff-third-space/GMemosBot/memo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

func handleStartCmd(msg *tgbotapi.Message, reply *tgbotapi.MessageConfig) {
	reply.Text = "这是一个简单的基于Memo的Bot"
}

func handleInfoCmd(msg *tgbotapi.Message, reply *tgbotapi.MessageConfig) {
	userName, err := memo.UserInfo(msg.From.ID)
	if err != nil {
		reply.Text = fmt.Sprintf("❌ 用户信息获取失败 %v", err)
	} else {
		reply.Text = fmt.Sprintf("✅ 欢迎你：%s", userName)
	}
}

func handleTokenCmd(msg *tgbotapi.Message, reply *tgbotapi.MessageConfig) {
	args := strings.Fields(msg.Text)
	if len(args) < 2 {
		reply.Text = "使用方法: /token <您的token>"
		return
	}

	token := args[1]
	result, err := memo.ValidateToken(token, msg.From.ID)
	if err != nil {
		reply.Text = "验证token时出错: " + err.Error()
	} else if result.Code == 0 {
		reply.Text = fmt.Sprintf("✅ Token验证成功！欢迎你：%s", result.UserName)
	} else {
		reply.Text = "❌ Token无效或已过期"
	}
}

func handleDefaultCmd(msg *tgbotapi.Message, reply *tgbotapi.MessageConfig) {
	reply.Text = "未知命令，发送 /help 查看可用命令"
}
