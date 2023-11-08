package main

import (
	"flag"
	"fmt"
	"os"
)

var url string

func init() {
	flag.StringVar(&url, "url", "", "需要刷新的Url地址")
}

func main() {
	flag.Parse()
	ak := os.Getenv("QINIU_AK")
	sk := os.Getenv("QINIU_SK")
	qiniu := &QiNiu{
		accessKey: ak,
		secretKey: sk,
	}
	if url == "" {
		panic("url不能为空")
	}
	err := qiniu.RefreshCDN([]string{url}, nil)
	if err != nil {
		fmt.Println(err)
	}
}
