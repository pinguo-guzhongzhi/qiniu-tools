package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/qiniu/api.v7/v7/auth"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	_ "golang.org/x/image/bmp"
)

const QiNiuRefreshCDNUrl = "http://fusion.qiniuapi.com/v2/tune/refresh"

var qiNiuZoneMaps = map[string]*storage.Zone{
	"ZoneHuadong": &storage.ZoneHuadong,
	"ZoneHuabei":  &storage.ZoneHuabei,
	"ZoneHuanan":  &storage.ZoneHuanan,
	"ZoneBeimei":  &storage.ZoneBeimei,
}

type QiNiu struct {
	accessKey       string
	secretKey       string
	bucket          string
	zone            string
	useHTTPS        bool
	useCdnDomains   bool
	url             string
	DeleteAfterDays int
}

func (s *QiNiu) getRemoteName(localFileName string) string {
	temp := strings.Split(localFileName, "/")
	temp = strings.Split(temp[len(temp)-1], ".")

	h := md5.New()
	h.Write([]byte(localFileName + time.Now().String()))
	return hex.EncodeToString(h.Sum(nil)) + "." + strings.Join(temp[1:], ".")
}

func (s *QiNiu) RefreshCDN(urls []string, dirs []string) error {
	if dirs == nil {
		dirs = []string{}
	}
	if urls == nil {
		urls = []string{}
	}
	if len(urls) == 0 && len(dirs) == 0 {
		return fmt.Errorf("dirs or urls are can not both be null")
	}
	data := map[string]interface{}{
		"urls": urls,
		"dirs": dirs,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", QiNiuRefreshCDNUrl, bytes.NewReader(body))
	if err != nil {
		return err
	}
	mac := qbox.NewMac(s.accessKey, s.secretKey)
	req.Header.Set("Content-Type", "application/json")
	err = mac.AddToken(auth.TokenQBox, req)
	if err != nil {
		return err
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	body, err = io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	var rspData struct {
		Code      int    `json:"code"`
		Error     string `json:"error"`
		RequestID string `json:"requestId"`
	}
	fmt.Println(string(body))
	err = json.Unmarshal(body, &rspData)
	if err != nil {
		return err
	}
	if rspData.Code != 200 {
		return fmt.Errorf(rspData.Error)
	}
	return nil
}
