package agileapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/rpc/v2/json2"
	"gopkg.in/cheggaaa/pb.v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type ListObject struct {
	FileType int    `json:"type"`
	Filename string `json:"name"`
}

type ListResult struct {
	Object []ListObject `json:"list"`
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

func DoAction(url, method, action string, args []string, debug bool) (output bool) {
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
	}
	output = false
	return
}

func SetMTime(url, token, path, mtime string, debug bool) (output bool) {
	args := []string{token, path, mtime}
	output = DoAction(url, "setMTime", "POST", args, debug)
	return
}

func RenameFile(url, token, originpath, destpath string, debug bool) (output bool) {
	args := []string{token, originpath, destpath}
	output = DoAction(url, "renameFile", "POST", args, debug)
	return
}

func RmFile(url, token, path string, debug bool) (output bool) {
	args := []string{token, path}
	output = DoAction(url, "deleteFile", "POST", args, debug)
	return
}

func RmDir(url, token, path string, debug bool) (output bool) {
	args := []string{token, path}
	output = DoAction(url, "deleteDir", "POST", args, debug)
	return
}

func MkDir2(url, token, path string, debug bool) (output bool) {
	args := []string{token, path}
	output = DoAction(url, "makeDir2", "POST", args, debug)
	return
}

func MkDir(url, token, path string, debug bool) (output bool) {
	args := []string{token, path}
	output = DoAction(url, "makeDir", "POST", args, debug)
	return
}

func StatFile(url, token, path string, debug bool) (output StatResult) {
	args := []string{token, path}
	outputjson := jsonrpcCallNoDecode(url, "stat", "POST", args, debug)
	var dec StatResponse
	err := json.Unmarshal([]byte(outputjson), &dec)
	if err != nil {
		fmt.Println(err)
	}
	output = dec.Result
	return
}

func ListFiles(url, token, path string, debug bool) (output ListResult) {
	args := []string{token, path}
	outputjson := jsonrpcCallNoDecode(url, "listFile", "POST", args, debug)
	var dec ListResponse
	err := json.Unmarshal([]byte(outputjson), &dec)
	if err != nil {
		fmt.Println(err)

	}
	output = dec.Result
	//fmt.Println(dec.Result)
	return
}

func ListDirs(url, token, path string, debug bool) (output ListResult) {
	args := []string{token, path}
	outputjson := jsonrpcCallNoDecode(url, "listDir", "POST", args, debug)
	var dec ListResponse
	err := json.Unmarshal([]byte(outputjson), &dec)
	if err != nil {
		fmt.Println(err)
	}
	output = dec.Result
	return
}

func UploadFile(url, token, path, file, localfilepath string, debug, progress bool) (err error) {
	params := map[string]string{
		"X-Agile-Authorization":  token,
		"X-Agile-Directory":      path,
		"X-Agile-Basename":       file,
		"X-Agile-Expose-Egress":  "COMPLETE",
		"X-Agile-Content-Detect": "auto",
	}
	urlbits := strings.Split(url, "/")
	host := urlbits[2]
	uri := fmt.Sprintf("http://%s:8080/post/raw", host)
	if debug {
		fmt.Println(uri)
		fmt.Println("curl -X POST -H \"X-Agile-Authorization: " + token + "\" -H \"X-Agile-Basename: " + file + "\" -H \"X-Agile-Directory: " + path + "\" --data-binary @" + localfilepath + " " + uri)
	}
	data, err := os.Open(localfilepath)
	client := &http.Client{}
	if progress {
		fi, err := data.Stat()
		bar := pb.New64(fi.Size()).SetUnits(pb.U_BYTES)
		bar.Start()
		r := bar.NewProxyReader(data)
		req, _ := http.NewRequest("POST", uri, r)
		for k, v := range params {
			req.Header.Add(k, v)
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return errors.New(fmt.Sprintf("PostRawFail:%d", resp.StatusCode))
		}
		bar.Finish()
	} else {

		req, _ := http.NewRequest("POST", uri, data)

		for k, v := range params {
			req.Header.Add(k, v)

		}
		resp, err := client.Do(req)
		if err != nil {
			return err

		}

		if resp.StatusCode != 200 {
			return errors.New(fmt.Sprintf("PostRawFail:%d", resp.StatusCode))

		}
	}
	return nil

}

func jsonrpcCallNoDecode(url, method, action string, args []string, debug bool) (output string) {

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
