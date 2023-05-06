/*
   Copyright 2021 Hiroshi.tao

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

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
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const version = "0.5.4"

var (
	h    bool
	v    bool
	info bool

	url    string
	token  string
	album  string
	desc   string
	expire uint
	style  string
	copy   string

	wg sync.WaitGroup

	// apiVersion: minimum compatible api interface version
	apiVersion string = "1.12.0"
	goVersion  string = strings.TrimLeft(runtime.Version(), "go")
	commitID   string = "N/A" // git commit id when building
	built      string = "N/A" // UTC time when building
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
	log.SetFlags(log.LstdFlags)

	flag.BoolVar(&h, "h", false, "")
	flag.BoolVar(&h, "help", false, "")

	flag.BoolVar(&v, "v", false, "")
	flag.BoolVar(&v, "version", false, "")

	flag.BoolVar(&info, "i", false, "")
	flag.BoolVar(&info, "info", false, "")

	flag.StringVar(&url, "u", "", "")
	flag.StringVar(&url, "picbed-url", "", "")
	flag.StringVar(&url, "sapic-url", "", "")

	flag.StringVar(&token, "t", "", "")
	flag.StringVar(&token, "picbed-token", "", "")
	flag.StringVar(&token, "sapic-token", "", "")

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
	} else if info {
		fmt.Printf("Cli version: %s\n", version)
		fmt.Printf("Api version: %s\n", apiVersion)
		fmt.Printf("Git commit:  %s\n", commitID)
		fmt.Printf("Go version:  %s\n", goVersion)
		fmt.Printf("OS/Arch:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Built:       %s\n", built)
	} else {
		handle()
	}
}

func usage() {
	helpStr := `usage: sapicli [-h] [-v] [-i] [-u SAPIC_URL] [-t SAPIC_TOKEN] [-a ALBUM]
                  [-d DESC] [-e EXPIRE] [-s STYLE] [-c {url,md,rst}]
                  file [file ...]

Doc to https://sapic.rtfd.vip/cli.html
Git to https://github.com/sapicd/cli

positional arguments:
  file                  local image file

optional arguments:
  -h, --help            show this help message and exit
  -v, --version         show cli version and exit
  -i, --info            show full info and exit
  -u, --sapic-url SAPIC_URL
                        The sapic upload api url.
                        Or use environment variable: sapicli_apiurl
  -t, --sapic-token SAPI_TOKEN
                        The sapic LinkToken.
                        Or use environment variable: sapicli_apitoken
  -a, --album ALBUM     Set image album
  -d, --desc DESC       Set image title(description)
  -e, --expire EXPIRE   Set image expire(seconds)
  -s, --style STYLE     The upload output style: { default, typora, line, empty, <MOD> }.
                        <MOD> allows to pass in a python module name, and use
                        "python -m py-mod-name" to customize the output style.
  -c, --copy {url,md,rst}
                        Copy the uploaded image url type to the clipboard
                        for Windows/macOS/Linux.
                        By the way, md=Markdown, rst=reStructuredText
    `
	fmt.Println(helpStr)
}

func handle() {
	if flag.NArg() == 0 {
		fmt.Printf("Please choose file to upload\n\n")
		usage()
		return
	}
	allowCopies := map[string]bool{
		"url": true,
		"rst": true,
		"md":  true,
		"":    true, //allow empty copy
	}
	if _, ok := allowCopies[copy]; !ok {
		fmt.Printf("No valid copy option\n\n")
		usage()
		return
	}
	if url == "" {
		url = os.Getenv("picbed_cli_apiurl")
		if url == "" {
			url = os.Getenv("sapicli_apiurl")
		}
	}
	if token == "" {
		token = os.Getenv("picbed_cli_apitoken")
		if token == "" {
			token = os.Getenv("sapicli_apitoken")

		}
	}
	if url == "" {
		fmt.Printf("No valid sapic(picbed) api url\n\n")
		usage()
		return
	}
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}
	if !strings.HasSuffix(url, "/api/upload") {
		url = strings.TrimRight(url, "/") + "/api/upload"
	}

	var files []string
	for _, f := range flag.Args() {
		f, _ = filepath.Abs(f)
		if isFile(f) {
			files = append(files, f)
		}
	}

	result := make([]apiResult, len(files))
	for i, f := range files {
		wg.Add(1)
		go apiUpload(f, &result, i)
	}
	wg.Wait()

	// show upload result
	switch style {
	case "typora", "line":
		if style == "typora" {
			fmt.Println("Upload Success:")
		}
		for _, res := range result {
			if res.Code == 0 && res.Src != "" {
				fmt.Println(res.Src)
			}
		}
	case "default": //the style default here
		output, _ := json.Marshal(result)
		fmt.Println(string(output))
	case "empty": //ignore output
	default: //handle with python module
		output, _ := json.Marshal(result)
		cmd := exec.Command("python", "-m", style, string(output))
		modout, err := cmd.Output()
		if err != nil {
			log.Print(err)
		} else {
			fmt.Println(string(modout))
		}
	}

	// auto copy
	if copy == "" {
		return
	}
	contents := make([]map[string]string, len(result))
	for i, res := range result {
		if res.Code == 0 && res.Src != "" {
			url := res.Tpl["URL"]
			rst := res.Tpl["rST"]
			md := res.Tpl["Markdown"]
			contents[i] = map[string]string{
				"name": res.Filename, "url": url, "rst": rst, "md": md,
			}
		}
	}
	autoCopy(getCopyContent(contents, copy))
}

func apiUpload(f string, result *[]apiResult, index int) {
	pic, err := ioutil.ReadFile(f)
	if err != nil {
		log.Fatal("read image failed for " + f)
		return
	}

	var post http.Request
	post.ParseForm()
	post.Form.Add("origin", "cli/"+version)
	post.Form.Add("_upload_field", "picbed")
	post.Form.Add("picbed", base64.StdEncoding.EncodeToString(pic))
	post.Form.Add("filename", filepath.Base(f))
	post.Form.Add("album", album)
	post.Form.Add("title", desc)
	if expire > 0 {
		post.Form.Add("expire", string(strconv.Itoa(int(expire))))
	}

	ua := fmt.Sprintf(
		"sapicli/%s go/%s %s %s",
		version,
		goVersion,
		runtime.GOOS,
		runtime.GOARCH,
	)
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(
		"POST", url, strings.NewReader(post.Form.Encode()),
	)
	if err != nil {
		log.Fatal(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "LinkToken "+token)
	req.Header.Set("User-Agent", ua)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	var data apiResult
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal(err)
	}
	(*result)[index] = data
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

func getCopyContent(contents []map[string]string, copyType string) string {
	var content strings.Builder
	for _, res := range contents {
		content.WriteString(res[copyType])
		content.WriteString("\n")
	}
	return strings.TrimRight(content.String(), "\n")
}

func autoCopy(content string) {
	if content == "" {
		return
	}
	switch runtime.GOOS {
	case "windows":
		cmd := fmt.Sprintf(`echo %s | clip`, strings.ReplaceAll(content, "\n", "\\n"))
		err := exec.Command("cmd.exe", "/C", cmd).Run()
		if err != nil {
			log.Print(err)
			return
		}
		sf := genTmpPS1()
		if isFile(sf) {
			defer os.Remove(sf)
			err = exec.Command(
				"powershell", "-ExecutionPolicy", "Unrestricted", sf,
				"上传成功", "已复制到剪贴板",
			).Run()
			if err != nil {
				log.Print(err)
			}
		}

	case "darwin":
		cmd1 := fmt.Sprintf(`echo "%s" | pbcopy`, content)
		err := exec.Command("bash", "-c", cmd1).Run()
		if err != nil {
			log.Print(err)
			return
		}
		cmd2 := fmt.Sprintf("display notification \"%s\" with title \"%s\" sound name \"%s\"", "已复制到剪贴板", "上传成功", "default")
		err = exec.Command("osascript", "-e", cmd2).Run()
		if err != nil {
			log.Print(err)
		}

	case "linux":
		cmd := fmt.Sprintf(`echo "%s" | xclip -selection clipboard`, content)
		err := exec.Command("bash", "-c", cmd).Run()
		if err != nil {
			log.Print(err)
		}
	}
}

func genTmpPS1() (filepath string) {
	tpl := []byte(`
param(
    [String] $Title,
    [String] $SubTitle
)

[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.UI.Notifications.ToastNotification, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

$APP_ID = '110366bd-56e2-47ed-9bdf-3ce1fa408b6c'

$template = @"
<toast>
    <visual>
        <binding template="ToastText02">
            <text id="1">$($Title)</text>
            <text id="2">$($SubTitle)</text>
        </binding>
    </visual>
</toast>
"@

$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml($template)
$toast = New-Object Windows.UI.Notifications.ToastNotification $xml
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($APP_ID).Show($toast)
`)

	tmpfile, err := ioutil.TempFile(os.TempDir(), "*.ps1")
	if err != nil {
		log.Print(err)
		return
	}

	_, err = tmpfile.Write(tpl)
	if err != nil {
		log.Print(err)
		return
	}
	tmpfile.Close()

	return tmpfile.Name()
}
