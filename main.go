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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	neturl "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	version = "0.6.0"

	// apiV1: POST /api/upload, Authorization: LinkToken <token>
	// apiV2: POST /api/v2/upload, Authorization: Bearer <PAT>
	apiV1 = "v1"
	apiV2 = "v2"
)

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
	apiVer string

	wg sync.WaitGroup

	// apiVersion: minimum compatible api interface version
	apiVersion   string = "1.12.0"
	apiVersionV2 string = "2.0.0"
	goVersion    string = strings.TrimLeft(runtime.Version(), "go")
	commitID     string = "N/A" // git commit id when building
	built        string = "N/A" // UTC time when building
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

// apiV2Result is the v2 response envelope, flattened into apiResult later.
type apiV2Result struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      struct {
		Image struct {
			Sha      string            `json:"sha"`
			Filename string            `json:"filename"`
			Src      string            `json:"src"`
			Tpl      map[string]string `json:"tpl"`
		} `json:"image"`
	} `json:"data"`
}

// uploadTarget holds what differs between the two api versions.
type uploadTarget struct {
	api     string // apiV1 or apiV2
	root    string // site root, without any api path
	url     string // upload endpoint
	token   string
	albumID string // v2 only, resolved from -a
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

	flag.StringVar(&apiVer, "A", "auto", "")
	flag.StringVar(&apiVer, "api-version", "auto", "")

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
		fmt.Printf("Api version: %s (%s), %s (%s)\n", apiVersion, apiV1, apiVersionV2, apiV2)
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
                  [-A {auto,v1,v2}]
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
                        The sapic site url, the api path is appended
                        automatically.
                        Or use environment variable: sapicli_apiurl
  -t, --sapic-token SAPI_TOKEN
                        The sapic token: a LinkToken for v1, a PAT(sp_...)
                        for v2.
                        Or use environment variable: sapicli_apitoken
  -a, --album ALBUM     Set image album, its name or its id(v2 only)
  -d, --desc DESC       Set image title(description)
  -e, --expire EXPIRE   Set image expire(seconds)
  -s, --style STYLE     The upload output style: { default, typora, line, empty, <MOD> }.
                        <MOD> allows to pass in a python module name, and use
                        "python -m py-mod-name" to customize the output style.
  -c, --copy {url,md,rst}
                        Copy the uploaded image url type to the clipboard
                        for Windows/macOS/Linux.
                        By the way, md=Markdown, rst=reStructuredText
  -A, --api-version {auto,v1,v2}
                        The sapic api version to talk to, auto detects it from
                        the site and the token. v1 = POST /api/upload with
                        "Authorization: LinkToken <token>",
                        v2 = POST /api/v2/upload with
                        "Authorization: Bearer <PAT>".
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
	switch strings.ToLower(strings.TrimSpace(apiVer)) {
	case "", "auto", "v1", "v2", "1", "2":
	default:
		fmt.Printf("No valid api version option\n\n")
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
	target := newUploadTarget(url, token)
	if target.api == apiV2 && album != "" {
		// v2 only accepts album_id, so -a may be a numeric id or an album name.
		target.albumID = resolveAlbumID(target, album)
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
		go apiUpload(target, f, &result, i)
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

func apiUpload(target uploadTarget, f string, result *[]apiResult, index int) {
	defer wg.Done()

	pic, err := ioutil.ReadFile(f)
	if err != nil {
		log.Fatal("read image failed for " + f)
		return
	}

	var req *http.Request
	if target.api == apiV2 {
		req, err = newV2Request(target, f, pic)
	} else {
		req, err = newV1Request(target, f, pic)
	}
	if err != nil {
		log.Fatal(err)
		return
	}
	req.Header.Set("User-Agent", userAgent())

	client := &http.Client{Timeout: 30 * time.Second}
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

	if target.api == apiV2 {
		(*result)[index] = parseV2Body(body, res.StatusCode)
		return
	}

	var data apiResult
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal(err)
		return
	}
	(*result)[index] = data
}

// newUploadTarget picks the api version and the upload endpoint to use.
func newUploadTarget(rawURL, token string) uploadTarget {
	root := siteRoot(rawURL)
	api := detectAPI(rawURL, token)
	return uploadTarget{
		api:   api,
		root:  root,
		url:   uploadURL(root, api),
		token: token,
	}
}

// siteRoot strips a trailing api path so "-u site/api/v2" works as well as the
// bare site url the doc recommends.
func siteRoot(rawURL string) string {
	root := strings.TrimRight(rawURL, "/")
	for _, suffix := range []string{"/api/v2/upload", "/api/upload", "/api/v2", "/api"} {
		if strings.HasSuffix(root, suffix) {
			return strings.TrimSuffix(root, suffix)
		}
	}
	return root
}

// detectAPI resolves the api version: -A wins, then an explicit api path in -u,
// then the site probe (v2 answers /api/v2/site), then the PAT prefix.
func detectAPI(rawURL, token string) string {
	switch strings.ToLower(strings.TrimSpace(apiVer)) {
	case apiV2, "2":
		return apiV2
	case apiV1, "1":
		return apiV1
	}

	lowerURL := strings.ToLower(strings.TrimRight(rawURL, "/"))
	if strings.Contains(lowerURL, "/api/v2") {
		return apiV2
	}
	if strings.HasSuffix(lowerURL, "/api/upload") {
		return apiV1
	}
	if siteIsV2(siteRoot(rawURL)) {
		return apiV2
	}
	// The site probe may fail (blocked, offline); a v2 PAT still decides it.
	if strings.HasPrefix(token, "sp_") {
		return apiV2
	}
	return apiV1
}

// siteIsV2 reports whether the site serves the v2 api.
func siteIsV2(root string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Get(root + "/api/v2/site")
	if err != nil {
		return false
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return false
	}

	var payload struct {
		Code string `json:"code"`
		Data struct {
			Version string `json:"version"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return false
	}
	return strings.HasPrefix(payload.Data.Version, "2.")
}

func uploadURL(root, api string) string {
	if api == apiV2 {
		return root + "/api/v2/upload"
	}
	return root + "/api/upload"
}

// resolveAlbumID maps the -a value to a v2 album id: digits are taken as the id,
// anything else is looked up by name. A miss falls back to the token's own
// album instead of blocking the upload.
func resolveAlbumID(target uploadTarget, album string) string {
	name := strings.TrimSpace(album)
	if name == "" {
		return ""
	}
	if _, err := strconv.Atoi(name); err == nil {
		return name
	}

	req, err := http.NewRequest("GET", target.root+"/api/v2/albums", nil)
	if err != nil {
		log.Printf("resolve album %q failed: %s", name, err)
		return ""
	}
	if target.token != "" {
		req.Header.Set("Authorization", "Bearer "+target.token)
	}
	req.Header.Set("User-Agent", userAgent())

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("resolve album %q failed: %s", name, err)
		return ""
	}
	defer res.Body.Close()

	var payload struct {
		Code string `json:"code"`
		Data struct {
			Items []struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		log.Printf("resolve album %q failed: %s", name, err)
		return ""
	}
	for _, item := range payload.Data.Items {
		if item.Name == name {
			return strconv.Itoa(item.ID)
		}
	}
	log.Printf("album %q not found, using the token default album", name)
	return ""
}

// newV1Request posts the base64 payload the v1 api expects.
func newV1Request(target uploadTarget, f string, pic []byte) (*http.Request, error) {
	form := neturl.Values{}
	form.Set("origin", "cli/"+version)
	form.Set("_upload_field", "picbed")
	form.Set("picbed", base64.StdEncoding.EncodeToString(pic))
	form.Set("filename", filepath.Base(f))
	form.Set("album", album)
	form.Set("title", desc)
	if expire > 0 {
		form.Set("expire", strconv.Itoa(int(expire)))
	}

	req, err := http.NewRequest(
		"POST", target.url, strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if target.token != "" {
		req.Header.Set("Authorization", "LinkToken "+target.token)
	}
	return req, nil
}

// newV2Request posts multipart/form-data with album_id, title and expire.
func newV2Request(target uploadTarget, f string, pic []byte) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	header := textproto.MIMEHeader{}
	header.Set(
		"Content-Disposition",
		fmt.Sprintf(
			`form-data; name="file"; filename="%s"`,
			strings.Replace(filepath.Base(f), `"`, "", -1),
		),
	)
	header.Set("Content-Type", http.DetectContentType(pic))
	part, err := writer.CreatePart(header)
	if err != nil {
		return nil, err
	}
	if _, err = part.Write(pic); err != nil {
		return nil, err
	}

	if desc != "" {
		writer.WriteField("title", desc)
	}
	if expire > 0 {
		writer.WriteField("expire", strconv.Itoa(int(expire)))
	}
	if target.albumID != "" {
		writer.WriteField("album_id", target.albumID)
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", target.url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if target.token != "" {
		req.Header.Set("Authorization", "Bearer "+target.token)
	}
	return req, nil
}

// parseV2Body flattens the v2 envelope {code,message,data:{image}} into the
// v1-shaped result the output styles and the copy logic already understand.
func parseV2Body(body []byte, status int) apiResult {
	var payload apiV2Result
	if err := json.Unmarshal(body, &payload); err != nil {
		return apiResult{Code: status, Msg: strings.TrimSpace(string(body))}
	}

	image := payload.Data.Image
	if payload.Code == "OK" && image.Src != "" {
		return apiResult{
			Code:     0,
			Filename: image.Filename,
			Src:      image.Src,
			Tpl:      image.Tpl,
		}
	}

	code := status
	if code == http.StatusOK || code == http.StatusCreated {
		code = -1
	}
	msg := payload.Message
	if msg == "" {
		msg = payload.Code
	}
	if payload.RequestID != "" {
		msg = fmt.Sprintf("%s (request_id: %s)", msg, payload.RequestID)
	}
	if msg == "" {
		msg = strings.TrimSpace(string(body))
	}
	return apiResult{Code: code, Msg: msg}
}

func userAgent() string {
	return fmt.Sprintf(
		"sapicli/%s go/%s %s %s",
		version,
		goVersion,
		runtime.GOOS,
		runtime.GOARCH,
	)
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
