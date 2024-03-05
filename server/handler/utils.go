package handler

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"math/rand"
	"net/http"
	"reflect"
	"sync"
	"time"
)

var Upgrader = websocket.Upgrader{
	HandshakeTimeout: time.Hour * 2,
	ReadBufferSize:   10240,
	WriteBufferSize:  10240,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var CfgMutex sync.Mutex

func findRoom(roomId string) (*model.Room, int) {
	cfg := model.GetConfig()
	for i, room := range cfg.Rooms {
		if room.Id == roomId {
			return &cfg.Rooms[i], i
		}
	}
	return nil, 0
}

func Shuffle(strSlice []string) []string {
	// 使用 Fisher-Yates 算法随机排列切片中的元素
	for i := len(strSlice) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		strSlice[i], strSlice[j] = strSlice[j], strSlice[i]
	}
	return strSlice
}

func Contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func HasPassedGivenHours(timeStr3339 string, hours int) (bool, error) {
	now := time.Now()
	createdAt, err := time.Parse(time.RFC3339, timeStr3339)
	if err != nil {
		fmt.Println("解析时间字符串时出错：", err)
		return false, err
	}
	duration := now.Sub(createdAt).Hours()
	if int(duration) > hours {
		return true, nil
	}
	return false, nil
}

func RemoveElement(arr []model.Room, index int) []model.Room {
	if index < 0 || index >= len(arr) {
		return arr // 无效的索引，返回原始切片
	}

	return append(arr[:index], arr[index+1:]...)
}

func MapToStruct(m map[string]interface{}, s interface{}) error {
	sValue := reflect.ValueOf(s).Elem()
	// 遍历 map 中的键值对，将其赋值给结构体对应的字段
	for k, v := range m {
		field := sValue.FieldByName(k)
		if !field.IsValid() {
			// 如果字段不存在，忽略该键值对
			continue
		}
		// 判断字段的类型是否可以直接赋值
		if field.CanSet() {
			// 如果字段可以直接赋值，将值转换为字段的类型并赋值
			val := reflect.ValueOf(v)
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
			}
		}
	}
	return nil
}
