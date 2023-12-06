package main

import (
	"flag"
	"fmt"
	"os"
)

var url string
var dir string

func init() {
	flag.StringVar(&url, "url", "", "需要刷新的Url地址")
	flag.StringVar(&dir, "dir", "", "需要刷新的Dir目录")
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

	dirs := []string{}
	if dir != "" {
		dirs = append(dirs, dir)
	}

	err := qiniu.RefreshCDN([]string{url}, dirs)
	if err != nil {
		fmt.Println(err)
	}
}
