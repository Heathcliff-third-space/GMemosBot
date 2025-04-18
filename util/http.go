package util

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func HttpRequest(apiURL string, method string, token string, body io.Reader) ([]byte, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 创建新的请求
	req, err := http.NewRequest(method, apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求memos服务失败: %v", err)
	}

	data, err := io.ReadAll(resp.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("memos服务返回错误状态码: %v", err)
		}
	}(resp.Body)

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK || err != nil {
		return nil, fmt.Errorf("memos服务返回错误状态码: %d %v", resp.StatusCode, string(data))
	}

	return data, nil
}
