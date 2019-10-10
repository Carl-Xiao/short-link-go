package modules

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/carl-xiao/short-link-go/setting"
	"github.com/go-redis/redis"
	"log"
)

const (
	URLIDKEY           = "next.url.id"
	URLHASHKEY         = "urlhash:%s:url"
	SHORTLINKURL       = "shortlink:%s:url"
	SHORTLINKDETAILURL = "shortlinkdetail:%s:url"
)

const CODE62 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const CODE_LENTH = 62

var EDOC = map[string]int{"0": 0, "1": 1, "2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7, "8": 8, "9": 9, "a": 10, "b": 11, "c": 12, "d": 13, "e": 14, "f": 15, "g": 16, "h": 17, "i": 18, "j": 19, "k": 20, "l": 21, "m": 22, "n": 23, "o": 24, "p": 25, "q": 26, "r": 27, "s": 28, "t": 29, "u": 30, "v": 31, "w": 32, "x": 33, "y": 34, "z": 35, "A": 36, "B": 37, "C": 38, "D": 39, "E": 40, "F": 41, "G": 42, "H": 43, "I": 44, "J": 45, "K": 46, "L": 47, "M": 48, "N": 49, "O": 50, "P": 51, "Q": 52, "R": 53, "S": 54, "T": 55, "U": 56, "V": 57, "W": 58, "X": 59, "Y": 60, "Z": 61,}

/**
 * 编码 整数 为 base62 字符串
 */
func Encode(number int) string {
	result := make([]byte, 0)
	for number > 0 {
		round := number / CODE_LENTH
		remain := number % CODE_LENTH
		result = append(result, CODE62[remain])
		number = round
	}
	return string(result)
}

// Redis client
type RedisClient struct {
	Client *redis.Client
}

//Url detail info
type UrlDetail struct {
	Url    string
}

//Redis Client初始化
func RedisInit() *RedisClient {
	sec, err := setting.Cfg.GetSection("redis")
	if err != nil {
		log.Fatal(2, "Fail to get section 'database': %v", err)
	}
	url := sec.Key("URL").String()
	password := sec.Key("PASSWORD").String()
	database, _ := sec.Key("DATABASE").Int()
	Client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       database,
	})
	return &RedisClient{Client: Client}
}
func (r *RedisClient) ShortenUrl(url string) (string, error) {
	d, err := r.Client.Get(fmt.Sprintf(URLHASHKEY, url)).Result()
	fmt.Printf(d)
	if err == nil {
		//not existed
	}
	//自增
	err = r.Client.Incr(URLIDKEY).Err()
	if err != nil {
		//异常处理
	}
	//获取自增
	id, err := r.Client.Get(URLIDKEY).Int()
	if err != nil {
		//异常处理
	}
	hashId := Encode(id)
	//设置hash
	err = r.Client.Set(fmt.Sprintf(SHORTLINKURL, hashId), url, 0).Err()
	if err != nil {
		//异常处理
	}
	err = r.Client.Set(fmt.Sprintf(URLHASHKEY, hashId), url, 0).Err()
	if err != nil {
		//异常处理
	}
	//存放详细信息
	s := &UrlDetail{
		Url:    url,
	}
	detail, _ := json.Marshal(s)

	err = r.Client.Set(fmt.Sprintf(SHORTLINKDETAILURL, hashId), detail, 0).Err()
	if err != nil {
		//异常处理
	}
	return hashId, nil
}

//查看短地址详细信息
func (r *RedisClient) ShortLinkInfo(eid string, expire int64) (interface{}, error) {
	result, err := r.Client.Get(fmt.Sprintf(SHORTLINKDETAILURL, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{Code: 404, Err: errors.New("Not Found this detail")}
	} else {
		return result, nil
	}
}

//还原地址信息
func (r *RedisClient) UnShortUrl(eid string) (interface{}, error) {
	result, err := r.Client.Get(fmt.Sprintf(SHORTLINKURL, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{Code: 404, Err: errors.New("Not Found this detail")}
	} else {
		detail := &UrlDetail{}
		err = json.Unmarshal([]byte(result), &detail)
		return result, nil
	}
}