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
const QiNiuURL = "http://fusion.qiniuapi.com"

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

func (s *QiNiu) Prefetch(urls []string) error {
	params := make(map[string]interface{})
	params["urls"] = urls
	body, err := json.Marshal(params)
	if err != nil {
		return err
	}
	// /v2/tune/prefetch
	body, err = s.sendRequest(body, QiNiuURL+"/v2/tune/prefetch")
	if err != nil {
		return err
	}

	var rspData struct {
		Code      int               `json:"code"`
		Error     string            `json:"error"`
		RequestID string            `json:"requestId"`
		TaskIds   map[string]string `json:"taskIds"`
	}
	err = json.Unmarshal(body, &rspData)
	if err != nil {
		return err
	}

	if rspData.Code != 200 {
		return fmt.Errorf(rspData.Error)
	}
	fmt.Println("start to check result")
	params = map[string]interface{}{
		"urls": urls,
	}
	taskStatus := make(map[string]bool)
	for _, taskID := range rspData.TaskIds {
		taskStatus[taskID] = false
	}
	paramsBody, _ := json.Marshal(params)
	maxNumber := 30
	for {
		fmt.Println("=========")
		time.Sleep(time.Second * 2)
		if maxNumber <= 0 {
			break
		}
		maxNumber--
		rspBody, err := s.sendRequest(paramsBody, QiNiuURL+"/v2/tune/prefetch/list")
		if err != nil {
			return err
		}
		fmt.Println("rspBody", string(rspBody))
		var rspData QueryResponse
		err = json.Unmarshal(rspBody, &rspData)
		if err != nil {
			continue
		}
		for _, item := range rspData.Items {
			if _, ok := taskStatus[item.TaskID]; ok {
				taskStatus[item.TaskID] = item.State != "processing"
			}
		}

		isAllSucceed := true
		for _, v := range taskStatus {
			isAllSucceed = isAllSucceed && v
		}
		if !isAllSucceed {
			continue
		}
		break
	}

	return nil
}

func (s *QiNiu) sendRequest(body []byte, requestURL string) ([]byte, error) {
	req, err := http.NewRequest("POST", requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	mac := qbox.NewMac(s.accessKey, s.secretKey)
	req.Header.Set("Content-Type", "application/json")
	err = mac.AddToken(auth.TokenQBox, req)
	if err != nil {
		return nil, err
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()
	body, err = io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	var rspData struct {
		Code      int    `json:"code"`
		Error     string `json:"error"`
		RequestID string `json:"requestId"`
	}
	fmt.Println(string(body))
	err = json.Unmarshal(body, &rspData)
	if err != nil {
		return nil, err
	}
	if rspData.Code != 200 {
		return nil, fmt.Errorf(rspData.Error)
	}
	return body, nil
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
		Code      int               `json:"code"`
		Error     string            `json:"error"`
		RequestID string            `json:"requestId"`
		TaskIds   map[string]string `json:"taskIds"`
	}
	fmt.Println(string(body))
	err = json.Unmarshal(body, &rspData)
	if err != nil {
		return err
	}

	if rspData.Code != 200 {
		return fmt.Errorf(rspData.Error)
	}
	for _, taskID := range rspData.TaskIds {
		fmt.Println("taskID", taskID)
		maxNumber := 30
		for {
			if maxNumber <= 0 {
				break
			}
			maxNumber--
			time.Sleep(time.Second * 2)
			params := map[string]interface{}{
				"taskId": taskID,
			}
			body, _ := json.Marshal(params)
			rspBody, err := s.sendRequest(body, QiNiuURL+"/v2/tune/refresh/list")
			fmt.Println("rspBody", string(rspBody))
			if err != nil {
				break
			}
			var rspData QueryResponse
			err = json.Unmarshal(rspBody, &rspData)
			if err != nil {
				continue
			}
			if rspData.Items[0].State == "processing" {
				continue
			}
			break
		}
	}
	return nil
}

type QueryResponse struct {
	Code  int `json:"code"`
	Items []struct {
		State  string `json:"state"`
		TaskID string `json:"taskId"`
	} `json:"items"`
}
