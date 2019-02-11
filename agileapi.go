package agileapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/rpc/v2/json2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"strings"
)

type AgileApi struct {
	Token      string
	Url        string
	Username   string
	Password   string
	Debug      bool
	Secure     bool
	TokenCache string
}

type ListObject struct {
	FileType int    `json:"type"`
	Filename string `json:"name"`
}

type ListResult struct {
	Object []ListObject `json:"list"`
}

type ListFullResult struct {
	Object []ListFullObject `json:"list"`
}

type ListFullObject struct {
	FileType int            `json:"type"`
	Filename string         `json:"name"`
	Stat     FullStatResult `json:"stat"`
}
type FullStatResult struct {
	Code   int    `json:"code"`
	Mtime  int    `json:"mtime"`
	Size   int    `json:"size"`
	Type   int    `json:"type"`
	Ctime  int    `json:"ctime"`
	Sha256 string `json:"checksum"`
}

type ListFullResponse struct {
	Version string         `json:"jsonrpc"`
	Id      int            `json:"id"`
	Result  ListFullResult `json:"result"`
	Code    int            `json:"code"`
	Cookie  int            `json:"cookie"`
}

type ListResponse struct {
	Version string     `json:"jsonrpc"`
	Id      int        `json:"id"`
	Result  ListResult `json:"result"`
	Code    int        `json:"code"`
	Cookie  int        `json:"cookie"`
}

type StatResult struct {
	Code  int `json:"code"`
	Mtime int `json:"mtime"`
	Size  int `json:"size"`
	Type  int `json:"type"`
	Ctime int `json:"ctime"`
}

type StatResponse struct {
	Version string     `json:"jsonrpc"`
	Id      int        `json:"id"`
	Result  StatResult `json:"result"`
}

type AuthenticateResponse struct {
	Code  int    `json:"code"`
	Token string `json:"token"`
	Uid   int    `json:"uid"`
	Gid   int    `json:"gid"`
	Path  string `json:"path"`
}

type ActionResponse struct {
	Version string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  int    `json:"result"`
}

type CodeResult struct {
	Code int `json:"code"`
}

type NoOpResponse struct {
	Version string     `json:"jsonrpc"`
	Id      int        `json:"id"`
	Result  CodeResult `json:"result"`
}

//{"jsonrpc": "2.0", "id": 6183213937838991992, "result": {"code": -10001}}
func New(username, password, url string, debug bool) *AgileApi {
	usr, _ := user.Current()
	dir := usr.HomeDir

	agiletokenfile := dir + "/.agiletoken"
	_, err := os.Stat(agiletokenfile)
	exists := false
	if err == nil {
		exists = true
	}
	if exists {
		tokenbyte, _ := ioutil.ReadFile(agiletokenfile)
		mytoken := string(tokenbyte)
		if TestToken(mytoken, url, debug) {
			me := &AgileApi{
				Username:   username,
				Password:   password,
				Token:      mytoken,
				Url:        url,
				Debug:      debug,
				Secure:     true,
				TokenCache: agiletokenfile,
			}
			return me
		}
	}
	// If the saved token is no longer valid, do this.
	mytoken, err := Authenticate(username, password, url, debug)
	if err != nil {
		fmt.Println("Authentication Failed.  Trying Again.")

		mytoken, err = Authenticate(username, password, url, debug)
		if err != nil {
			log.Fatal("Authentication Failed.  Exiting")
		}
	}
	err = ioutil.WriteFile(agiletokenfile, []byte(mytoken), 0644)
	if err != nil {
		fmt.Println("Can't write cache file")
	}
	me := &AgileApi{
		Username:   username,
		Password:   password,
		Url:        url,
		Token:      mytoken,
		Debug:      debug,
		Secure:     true,
		TokenCache: agiletokenfile,
	}
	return me
}

func (me *AgileApi) ReAuth() {
	mytoken, err := Authenticate(me.Username, me.Password, me.Url, me.Debug)
	if err != nil {
		fmt.Println("Auth Failed")
	}
	me.Token = mytoken
	err = ioutil.WriteFile(me.TokenCache, []byte(mytoken), 0644)
	if err != nil {
		fmt.Println("Can't write cacehfile")
	}
	return
}

func Authenticate(username, password, url string, debug bool) (token string, err error) {
	args := []string{username, password, "true"}
	output := jsonrpcCall(url, "login", "POST", args, debug)
	if output == nil {
		err = errors.New("Login Failed")
		return
	}
	token = output[0].(string)
	return
}

func jsonrpcCall(url, method, action string, args []string, debug bool) (output []interface{}) {

	message, err := json2.EncodeClientRequest(method, args)
	if err != nil {
		log.Fatalf("%s", err)

	}
	if debug {
		fmt.Println(string(message))
	}
	req, err := http.NewRequest(action, url, bytes.NewBuffer(message))
	if err != nil {
		log.Fatalf("%s", err)

	}
	req.Header.Set("Content-Type", "application/json")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error in sending request to %s. %s", url, err)

	}
	defer resp.Body.Close()
	err = json2.DecodeClientResponse(resp.Body, &output)
	if err != nil {
		log.Fatalf("Couldn't decode response. %s", err)

	}
	return
}

func DoAction(url, method, action string, args []interface{}, debug bool) (output bool) {
	outputjson := jsonrpcCallNoDecode(url, method, action, args, debug)
	var dec ActionResponse
	err := json.Unmarshal([]byte(outputjson), &dec)
	if err != nil {
		fmt.Println(err)
		output = false
		return
	}
	if dec.Result == 0 {
		output = true
		return
	} else if dec.Result == -10001 {
		//FIXME
		fmt.Println("Token Expired, retry auth.")
	}
	output = false
	return
}

func (me *AgileApi) CheckAuth() {
	if !TestToken(me.Token, me.Url, me.Debug) {
		me.ReAuth()
	}
}

func TestToken(token, url string, debug bool) (output bool) {
	if debug {
		fmt.Println("TESTING TOKEN")
	}
	args := []interface{}{token}
	outputf := jsonrpcCallNoDecode(url, "noop", "POST", args, debug)
	var dec NoOpResponse
	err := json.Unmarshal([]byte(outputf), &dec)
	if err != nil {
		return false
	}
	if dec.Result.Code == 0 {
		if debug {
			fmt.Println("Cached Credentials still valid.  Using.")
		}
		return true
	}
	if debug {
		fmt.Println("Cached Credentials are not valid.")
	}
	return false
}

func (me *AgileApi) SetMTime(path, mtime string) (output bool) {
	me.CheckAuth()
	args := []interface{}{me.Token, path, mtime}
	output = DoAction(me.Url, "setMTime", "POST", args, me.Debug)
	return
}

func (me *AgileApi) RenameFile(originpath, destpath string) (output bool) {
	me.CheckAuth()
	args := []interface{}{me.Token, originpath, destpath}
	output = DoAction(me.Url, "renameFile", "POST", args, me.Debug)
	return
}

func (me *AgileApi) RmFile(path string) (output bool) {
	me.CheckAuth()
	args := []interface{}{me.Token, path}
	output = DoAction(me.Url, "deleteFile", "POST", args, me.Debug)
	return
}

func (me *AgileApi) RmDir(path string) (output bool) {
	me.CheckAuth()
	args := []interface{}{me.Token, path}
	output = DoAction(me.Url, "deleteDir", "POST", args, me.Debug)
	return
}

func (me *AgileApi) MkDir2(path string) (output bool) {
	me.CheckAuth()
	args := []interface{}{me.Token, path}
	output = DoAction(me.Url, "makeDir2", "POST", args, me.Debug)
	return
}

func (me *AgileApi) MkDir(path string) (output bool) {
	me.CheckAuth()
	args := []interface{}{me.Token, path}
	output = DoAction(me.Url, "makeDir", "POST", args, me.Debug)
	return
}

func (me *AgileApi) StatFile(path string) (output StatResult) {
	me.CheckAuth()
	args := []interface{}{me.Token, path}
	outputjson := jsonrpcCallNoDecode(me.Url, "stat", "POST", args, me.Debug)
	var dec StatResponse
	err := json.Unmarshal([]byte(outputjson), &dec)
	if err != nil {
		fmt.Println(err)
	}
	output = dec.Result
	return
}

func (me *AgileApi) ListAllFilesDetails(path string) (output []ListFullObject) {
	pagesize := 10000
	pageoffset := 0
	includestat := true
	mylen := 1
	for mylen >= 0 {
		me.CheckAuth()
		args := []interface{}{me.Token, path, pagesize, pageoffset, includestat}
		outputjson := jsonrpcCallNoDecode(me.Url, "listFile", "POST", args, me.Debug)
		var dec ListFullResponse
		err := json.Unmarshal([]byte(outputjson), &dec)
		if err != nil {
			fmt.Println(err)

		}
		loutput := dec.Result.Object
		mylen = len(loutput)
		output = append(output, loutput[:mylen]...)
		pageoffset = dec.Cookie
		if pageoffset == 0 {
			mylen = -1
		}
	}
	return

}

func (me *AgileApi) ListFiles(path string) (output []ListObject) {
	me.CheckAuth()
	args := []interface{}{me.Token, path}
	fmt.Println(args)
	outputjson := jsonrpcCallNoDecode(me.Url, "listFile", "POST", args, me.Debug)
	var dec ListResponse
	err := json.Unmarshal([]byte(outputjson), &dec)
	if err != nil {
		fmt.Println(err)

	}
	//fmt.Println(dec.Cookie)
	output = dec.Result.Object
	//fmt.Println("Version 2")
	//fmt.Println(dec.Result)
	return
}

func (me *AgileApi) ListDirs(path string) (output []ListObject) {
	me.CheckAuth()
	args := []interface{}{me.Token, path}
	outputjson := jsonrpcCallNoDecode(me.Url, "listDir", "POST", args, me.Debug)
	var dec ListResponse
	err := json.Unmarshal([]byte(outputjson), &dec)
	if err != nil {
		fmt.Println(err)
	}
	output = dec.Result.Object
	return
}

func (me *AgileApi) ListAllDirsDetails(path string) (output []ListFullObject) {
	pagesize := 10000
	pageoffset := 0
	includestat := true
	mylen := 1
	for mylen >= 0 {
		me.CheckAuth()
		args := []interface{}{me.Token, path, pagesize, pageoffset, includestat}
		outputjson := jsonrpcCallNoDecode(me.Url, "listDir", "POST", args, me.Debug)
		var dec ListFullResponse
		err := json.Unmarshal([]byte(outputjson), &dec)
		if err != nil {
			fmt.Println(err)

		}
		loutput := dec.Result.Object
		mylen = len(loutput)
		output = append(output, loutput[:mylen]...)
		pageoffset = dec.Cookie
		if pageoffset == 0 {
			mylen = -1
		}

	}
	return

}

func (me *AgileApi) UploadFileStream(path, file string, filereader io.Reader) (err error) {
	me.CheckAuth()
	params := map[string]string{
		"X-Agile-Authorization":  me.Token,
		"X-Agile-Directory":      path,
		"X-Agile-Basename":       file,
		"X-Agile-Expose-Egress":  "COMPLETE",
		"X-Agile-Content-Detect": "auto",
		"X-Agile-Recursive":      "true",
	}
	urlbits := strings.Split(me.Url, "/")
	host := urlbits[2]
	uri_template := "https://%s/post/raw"
	if !me.Secure {
		uri_template = "http://%s:8080/post/raw"
	}
	uri := fmt.Sprintf(uri_template, host)
	client := &http.Client{}

	req, _ := http.NewRequest("POST", uri, filereader)
	for k, v := range params {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("PostRawFail:%s", resp.StatusCode))
	}
	return nil
}

func (me *AgileApi) UploadFile(path, file, localfilepath string, progress bool) (err error) {
	data, err := os.Open(localfilepath)
	err = me.UploadFileStream(path, file, data)
	return err
}

func jsonrpcCallNoDecode(url, method, action string, args []interface{}, debug bool) (output string) {

	message, err := json2.EncodeClientRequest(method, args)
	if err != nil {
		log.Fatalf("%s", err)

	}
	if debug {
		fmt.Println(string(message))

	}
	req, err := http.NewRequest(action, url, bytes.NewBuffer(message))
	if err != nil {
		log.Fatalf("%s", err)

	}
	req.Header.Set("Content-Type", "application/json")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error in sending request to %s. %s", url, err)

	}
	defer resp.Body.Close()
	outputbs, err := ioutil.ReadAll(resp.Body)
	output = string(outputbs)
	if err != nil {
		log.Fatalf("Error reading response %s", output)
	}
	return
}
