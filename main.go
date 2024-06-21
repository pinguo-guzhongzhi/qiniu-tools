package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var url string
var dir string
var prefetch bool

func init() {
	flag.StringVar(&url, "url", "", "需要刷新的Url地址")
	flag.StringVar(&dir, "dir", "", "需要刷新的Dir目录")
	flag.BoolVar(&prefetch, "prefetch", false, "生成新的缓存")
}

func main() {
	flag.Parse()
	ak := os.Getenv("QINIU_AK")
	sk := os.Getenv("QINIU_SK")
	qiniu := &QiNiu{
		accessKey: ak,
		secretKey: sk,
	}

	urls := []string{}
	dirs := []string{}
	if dir != "" {
		tmp := strings.Split(dir, ",")
		for _, u := range tmp {
			u = strings.TrimSpace(u)
			if u == "" {
				continue
			}
			dirs = append(dirs, u)
		}
	}

	if url != "" {
		tmp := strings.Split(url, ",")
		for _, u := range tmp {
			u = strings.TrimSpace(u)
			if u == "" {
				continue
			}
			urls = append(urls, u)
		}
	}

	if len(urls) == 0 && len(dirs) == 0 {
		panic("url/dir不能为空")
	}
	var err error
	if prefetch {
		err = qiniu.Prefetch(urls)
	} else {
		err = qiniu.RefreshCDN(urls, dirs)
	}
	if err != nil {
		fmt.Println(err)
	}
}
