package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Cert struct {
	Name       string //域名
	Starttime  string //证书签发时间
	Stoptime   string //证书过期时间
	IsCA       bool   //是不是根证书
	IPaddress  string //ip地址
	IPLocation string //地址所属地域
}

type IPLoc struct {
	Organization    string  `json:"organization"`
	Longitude       float64 `json:"longitude"`
	Timezone        string  `json:"timezone"`
	Isp             string  `json:"isp"`
	Offset          int     `json:"offset"`
	Asn             int     `json:"asn"`
	AsnOrganization string  `json:"asn_organization"`
	Country         string  `json:"country"`
	Ip              string  `json:"ip"`
	Latitude        float64 `json:"latitude"`
	ContinentCode   string  `json:"continent_code"`
	CountryCode     string  `json:"country_code"`
}

var d = make(map[string]*Cert)
var w sync.WaitGroup
var tempdir string
var mx sync.Mutex

var (
	dst = flag.String("d", "", "输入文件输出地址，默认是在系统的临时目录下 ")
	src = flag.String("s", "", "输入证书列表文件地址，")
)

func main() {
	flag.Parse()

	//设置ctx，主要用来设置超时时间
	ctx := context.Context(context.Background())

	//通过命令行输入src地址和目的地址
	if *src == "" {
		fmt.Println("请输入证书文件列表所在位置，example: readcert -c g:")
		return
	}

	if *dst == "" {
		tempdir, _ = os.MkdirTemp("", "kafka")
	} else {
		tempdir = *dst
	}

	//文件生成路径
	tempfile := filepath.Join(tempdir, "cert.csv")
	fmt.Println("存放文件的路径是： " + tempfile)

	//读取文件
	file, err := os.Open(*src)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF { //文件已经结束
				break
			}
			fmt.Println("文件读取错误", err)
			return
		}
		//上传httpsurl
		url := "https://" + string(line)
		w.Add(1)
		go GetCert(ctx, string(line), url, &w)
	}

	w.Wait()

	//将所有域名信息写入文件中
	readAndWriteLog(tempdir, nil)
	for _, v := range d {
		readAndWriteLog(tempdir, v)
	}

}

//获取证书
func GetCert(ctx context.Context, urlreal, url string, w *sync.WaitGroup) {
	defer w.Done()
	//证书
	var cc Cert

	//获取公网地址
	ip, location := iplocation(ctx, urlreal)
	cc.IPaddress = ip
	cc.IPLocation = location

	//设置过期时间，避免无法反应域名，直接报错
	ctx1, cancel := context.WithTimeout(ctx, time.Second*8)
	defer cancel()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}

	req = req.WithContext(ctx1)
	resp, err := client.Do(req)
	if err != nil {
		//无响应直接返回
		return
	}
	defer resp.Body.Close()
	//将证书信息拷贝下来
	if resp.TLS != nil && resp.TLS.PeerCertificates != nil {
		certs := resp.TLS.PeerCertificates
		if len(certs) > 0 {
			cc.Stoptime = TimeTrim(certs[0].NotAfter.String())
			cc.IsCA = certs[0].IsCA
		}
	}
	//丢弃内容，不然http底层复用会出问题
	io.Copy(io.Discard, resp.Body)

	cc.Name = urlreal

	//解决map并发锁
	mx.Lock()
	defer mx.Unlock()
	d[urlreal] = &cc

}

//截断日期,
func TimeTrim(s string) string {
	return strings.Trim(s, " +0000 UTC")
}

//读写一个文件追加日志
func readAndWriteLog(path string, v *Cert) {
	filepath := filepath.Join(path, "cert.csv")

	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644) //读写模式打开，写入追加
	if err != nil {
		log.Fatal("打开文件发胜错误， err", err)
	}
	defer f.Close()

	if v == nil {
		comm := "域名" + "," + "域名过期时间" + "," + "," + "域名对应IP地址" + "," + "域名地址所在地" + "\n"
		f.WriteString(comm)
		return
	}

	comm := v.Name + "," + v.Stoptime + "," + "," + v.IPaddress + "," + v.IPLocation + "\n"
	f.WriteString(comm)
}

//获取ip和geo信息
func iplocation(ctx context.Context, url string) (string, string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*7)

	defer cancel()
	var host, err = net.LookupHost(url)
	if err != nil {
		return "", ""
	}
	ip := host[0]
	urlx := "https://api.ip.sb/geoip/" + ip
	client := &http.Client{}
	req, err := http.NewRequest("GET", urlx, nil)
	if err != nil {
		return ip, ""
	}
	req = req.WithContext(ctx)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		//无响应直接返回
		return ip, ""
	}
	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ip, ""
	}

	var iplocation = &IPLoc{}
	err = json.Unmarshal(respbody, iplocation)
	if err != nil {
		return ip, ""
	}

	return ip, iplocation.Organization
}
