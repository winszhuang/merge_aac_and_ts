package utils

import (
	"math/rand"
	"time"
)

func GenerateRandomKey() string {
	rand.Seed(time.Now().UnixNano()) // 使用當前時間的納秒作為隨機種子

	keyLength := 10                                                           // 隨機字串的長度
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // 可用於生成隨機字串的字符集

	randomKey := make([]byte, keyLength)
	for i := 0; i < keyLength; i++ {
		randomKey[i] = chars[rand.Intn(len(chars))]
	}

	return string(randomKey)
}
