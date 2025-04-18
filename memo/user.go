package memo

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	userFile = "/app/data/.user.txt"
)

type User struct {
	ID    int64
	Token string
}

var (
	userMap = sync.Map{}
)

// LoadUsers 加载所有用户数据
func LoadUsers() error {
	users, err := loadUsers()
	if err != nil {
		return err
	}

	for _, u := range users {
		userMap.LoadOrStore(u.ID, &u)
	}
	return nil
}

func loadUsers() ([]User, error) {
	data, err := os.ReadFile(userFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []User{}, nil // 文件不存在返回空列表
		}
		return nil, err
	}

	var users []User
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue // 跳过格式错误的行
		}
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue // 跳过ID解析失败的行
		}
		users = append(users, User{
			ID:    id,
			Token: parts[1],
		})
	}
	return users, nil
}

// SaveUsers 保存用户数据
func saveUsers() error {
	var lines []string
	userMap.Range(func(_, value interface{}) bool {
		user := value.(*User)
		lines = append(lines, fmt.Sprintf("%d:%s", user.ID, user.Token))
		return true
	})

	data := strings.Join(lines, "\n")
	return os.WriteFile(userFile, []byte(data), 0644)
}

// AddOrUpdateUser 添加或更新用户
func AddOrUpdateUser(user *User) error {
	userMap.Swap(user.ID, user)
	return saveUsers()
}

func FindUser(uid int64) (user *User, ok bool) {
	value, ok := userMap.Load(uid)
	if !ok {
		return nil, ok
	}
	return value.(*User), ok
}

// DeleteUser 删除用户
func DeleteUser(uid int64) error {
	userMap.Delete(uid)
	return saveUsers()
}
