package main // import "tcw.im/picbed-cli"

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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
)

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
	if url == "" {
		url = os.Getenv("picbed_cli_apiurl")
		if url != "" && !strings.HasSuffix(url, "/api/upload") {
			url += "/api/upload"
		}
	}
	if token == "" {
		token = os.Getenv("picbed_cli_apitoken")
	}
	if url == "" || token == "" {
		fmt.Println("No valid picbed api url or token")
		usage()
		os.Exit(127)
	}
}

func post(stream io.ByteReader) (body string, err error) {
	client := &http.Client{}
	data := strings.NewReader("name=cjb")

	/*
	   dict(
	                   picbed=b64encode(stream),
	                   filename=filename,
	                   album=album,
	                   title=title,
	                   expire=expire,
	                   origin="cli/{}".format(__version__),
	               )
	*/
	req, err := http.NewRequest("POST", url, strings.NewReader("name=cjb"))
	if err != nil {
		// handle error
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "name=anny")

	resp, err := client.Do(req)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	fmt.Println(string(body))
}
