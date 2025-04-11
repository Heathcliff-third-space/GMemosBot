package memo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Heathcliff-third-space/GMemosBot/util"
	"log"
	"os"
	"strings"
)

const (
	tag = "#tg"
)

type Base struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type AuthResult struct {
	Base
	UserName string `json:"username"`
	Nickname string `json:"nickname"`
}

type Memo struct {
	Base
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Resource struct {
	Base
	Name     string `json:"name"`
	Content  string `json:"content"`
	Type     string `json:"type"`
	Memo     string `json:"memo"`
	Filename string `json:"filename"`
	FileId   string `json:"fileId"`
}

var serverBaseUrl string

func Start() {
	serverBaseUrl = os.Getenv("MEMOS_SERVER_URL")
	if serverBaseUrl == "" {
		log.Fatal("MEMOS_SERVER_URL 环境变量未设置")
	}

	getMemoInfo()
}

func getMemoInfo() {
	apiURL := serverBaseUrl + "/api/v1/workspace/profile"
	respData, err := util.HttpRequest(apiURL, "GET", "", nil)
	if err != nil {
		log.Fatalf("memo初始化失败 %v", err)
	}

	log.Printf("memo服务信息：%s", string(respData))
}

func ValidateToken(token string, uid int64) (*AuthResult, error) {
	apiURL := serverBaseUrl + "/api/v1/auth/status"

	respData, err := util.HttpRequest(apiURL, "GET", token, nil)
	if err != nil {
		return nil, err
	}

	result := new(AuthResult)
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("用户保存失败 %v", result.Message)
	}

	err = AddOrUpdateUser(&User{
		ID:    uid,
		Token: token,
	})

	if err != nil {
		return nil, fmt.Errorf("用户保存失败 %v", err)
	}

	return result, nil
}

func UserInfo(uid int64) (string, error) {
	user, ok := FindUser(uid)
	if !ok {
		return "", fmt.Errorf("当前用户不存在，请通过 /auth TOKEN 命令创建用户")
	}

	authResult, err := ValidateToken(user.Token, user.ID)
	if err != nil {
		return "", err
	}

	return authResult.UserName, nil
}

func CreateMemo(content string, resources []*Resource, uid int64) (*Memo, error) {

	apiURL := serverBaseUrl + "/api/v1/memos"

	user, ok := FindUser(uid)
	if !ok {
		return nil, fmt.Errorf("当前用户不存在，请通过 /auth TOKEN 命令创建用户")
	}

	var builder strings.Builder
	builder.WriteString(tag)
	if !strings.HasPrefix(content, "#") {
		builder.WriteString("\n")
	}
	builder.WriteString(content)

	memo := &Memo{
		Content: builder.String(),
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(memo)
	if err != nil {
		log.Fatal(err)
	}

	respData, err := util.HttpRequest(apiURL, "POST", user.Token, &buf)
	result := new(Memo)
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, fmt.Errorf("memos创建失败: %v", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("memos创建失败: %v", result.Message)
	}

	err = createMemoResource(user, result.Name, resources)
	return result, err
}

func createMemoResource(user *User, memoName string, resources []*Resource) error {
	apiURL := serverBaseUrl + "/api/v1/resources"

	for _, resource := range resources {

		go func(res *Resource) {
			res.Memo = memoName

			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(res)
			if err != nil {
				log.Printf("memos：%s 附件: %s 上传失败 %v", memoName, res.Filename, err)
				return
			}

			respData, err := util.HttpRequest(apiURL, "POST", user.Token, &buf)
			if err != nil {
				log.Printf("memos：%s 附件: %s 上传失败 %v", memoName, res.Filename, err)
				return
			}

			result := new(Resource)
			if err := json.Unmarshal(respData, &result); err != nil {
				log.Printf("memos：%s 附件: %s 上传失败 %v", memoName, res.Filename, err)
				return
			}

			if result.Code != 0 {
				log.Printf("memos：%s 附件: %s 上传失败 %v", memoName, res.Filename, result.Message)
			}

			log.Printf("memos：%s 附件: %s 上传成功", memoName, res.Filename)
		}(resource)
	}

	return nil
}
