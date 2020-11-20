package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const version = "0.4.0"

var (
	h bool
	v bool

	url    string
	token  string
	album  string
	desc   string
	expire uint
	style  string
	copy   string

	wg sync.WaitGroup
)

type apiResult struct {
	Code     int               `json:"code"`
	Msg      string            `json:"msg"`
	Filename string            `json:"filename"`
	Sender   string            `json:"sender"`
	API      string            `json:"api"`
	Src      string            `json:"src"`
	Tpl      map[string]string `json:"tpl"`
}

func init() {
	flag.BoolVar(&h, "h", false, "show help")
	flag.BoolVar(&h, "help", false, "show help")

	flag.BoolVar(&v, "v", false, "show version and exit")
	flag.BoolVar(&v, "version", false, "show version and exit")

	flag.StringVar(&url, "u", "", "")
	flag.StringVar(&url, "picbed-url", "", "")

	flag.StringVar(&token, "t", "", "")
	flag.StringVar(&token, "picbed-token", "", "")

	flag.StringVar(&album, "a", "", "")
	flag.StringVar(&album, "album", "", "")

	flag.StringVar(&desc, "d", "", "")
	flag.StringVar(&desc, "desc", "", "")

	flag.UintVar(&expire, "e", 0, "")
	flag.UintVar(&expire, "expire", 0, "")

	flag.StringVar(&style, "s", "default", "")
	flag.StringVar(&style, "style", "default", "")

	flag.StringVar(&copy, "c", "", "")
	flag.StringVar(&copy, "copy", "", "")

	flag.Usage = usage
}

func main() {
	flag.Parse()
	if h {
		flag.Usage()
	} else if v {
		fmt.Println(version)
	} else {
		handle()
	}
}

func usage() {
	helpStr := `usage: picbed-cli [-h] [-v] [-u PICBED_URL] [-t PICBED_TOKEN] [-a ALBUM]
                  [-d DESC] [-e EXPIRE] [-s STYLE] [-c {url,md,rst}]
                  file [file ...]

Doc to https://picbed.rtfd.vip/cli.html
Git to https://github.com/staugur/picbed-cli

positional arguments:
  file                  local image file

optional arguments:
  -h, --help            show this help message and exit
  -v, --version         show cli version
  -u, --picbed-url PICBED_URL
                        The picbed upload api url.
                        Or use environment variable: picbed_cli_apiurl
  -t, --picbed-token PICBED_TOKEN
                        The picbed LinkToken.
                        Or use environment variable: picbed_cli_apitoken
  -a, --album ALBUM     Set image album
  -d, --desc DESC       Set image title(description)
  -e, --expire EXPIRE   Set image expire(seconds)
  -s, --style STYLE     The upload output style: { default, typora, line }.
                        And, allows the use of "module.function" to customize
                        the output style.
  -c, --copy {url,md,rst}
                        Copy the uploaded image url type to the clipboard
                        for win/mac/linux.
                        By the way, md=markdown, rst=reStructuredText
    `
	fmt.Println(helpStr)
}

func handle() {
	fmt.Println("URL", url)
	fmt.Println("Token", token)
	fmt.Println("Album", album)
	fmt.Println("Desc", desc)
	fmt.Println("Expire", expire)
	fmt.Println("Style", style)
	fmt.Println("Copy", copy)
	fmt.Println(flag.Args())
	fmt.Println(flag.NArg())
	if flag.NArg() == 0 {
		usage()
		return
	}
	if url == "" {
		url = os.Getenv("picbed_cli_apiurl")
	}
	if token == "" {
		token = os.Getenv("picbed_cli_apitoken")
	}
	if url != "" && !strings.HasSuffix(url, "/api/upload") {
		url += "/api/upload"
	}
	if url == "" || token == "" {
		fmt.Println("No valid picbed api url or token")
		usage()
		return
	}
	fmt.Println("before post")
	fmt.Println("URL", url)
	fmt.Println("Token", token)

	files := flag.Args()
	result := make([]apiResult, len(files))
	for _, f := range files {
		f, _ = filepath.Abs(f)
		if !isFile(f) {
			continue
		}
		wg.Add(1)
		go post(f, &result)
	}
	wg.Wait()
	fmt.Println(result)
}

func post(f string, result *[]apiResult) {
	pic, err := ioutil.ReadFile(f)
	if err != nil {
		log.Print("read image failed for " + f)
		return
	}

	var post http.Request
	post.ParseForm()
	post.Form.Add("picbed", base64.StdEncoding.EncodeToString(pic))
	post.Form.Add("filename", filepath.Base(f))
	post.Form.Add("album", album)
	post.Form.Add("title", desc)
	post.Form.Add("expire", string(strconv.Itoa(int(expire))))
	post.Form.Add("origin", "cli/"+version)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(post.Form.Encode()))
	if err != nil {
		log.Print(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "LinkToken "+token)
	req.Header.Set("User-Agent", "picbed-cli/ "+version)

	res, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
		return
	}

	var data apiResult
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Print(err)
	}
	*result = append(*result, data)
	wg.Done()
}

func isExists(path string) (os.FileInfo, bool) {
	f, err := os.Stat(path)
	return f, err == nil || os.IsExist(err)
}

func isFile(path string) bool {
	f, flag := isExists(path)
	return flag && !f.IsDir()
}
