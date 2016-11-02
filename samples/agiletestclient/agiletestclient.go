package main

import (
	"flag"
	"fmt"
	"gitlab.harnish.local/jharnish/agileapi"
)

var username = flag.String("username", "jharnish", "Agile Username")
var password = flag.String("password", "", "Agile Password")
var url = flag.String("url", "https://listen-l.upload.llnw.net/jsonrpc", "Agile api url")
var debug = flag.Bool("debug", false, "Enable debug mode")

func main() {
	flag.Parse()
	if *debug {
		fmt.Println("Username " + *username)
		fmt.Println("Password " + *password)
		fmt.Println("URL: " + *url)
	}
	token := agileapi.Authenticate(*username, *password, *url, *debug)
	fmt.Println(token)
	fileinfo := agileapi.StatFile(*url, token, "/jeep.jpg", *debug)
	fmt.Println(fileinfo)
	dirlist := agileapi.ListFiles(*url, token, "/", *debug)
	fmt.Println(dirlist)
	dirlist = agileapi.ListDirs(*url, token, "/", *debug)
	fmt.Println(dirlist)
	if agileapi.MkDir(*url, token, "/golang", *debug) {
		fmt.Println("made a directory")
	} else {
		fmt.Println("failed to make a directory")
	}
	if agileapi.RmDir(*url, token, "/golang", *debug) {
		fmt.Println("Removed directory")
	} else {
		fmt.Println("Faild to remove directory")
	}
	err := agileapi.UploadFile(*url, token, "/", "golang2.txt", "agiletestclient.go", *debug)
	if err != nil {
		fmt.Println(err)
	}
}
