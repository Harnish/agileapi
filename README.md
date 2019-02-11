This is an Object for working with LimeLight "Agile" Cloud Storage Platform via the jsonrpc and post api calls.


```golang
package main

import (
    "fmt"
    "github.llnw.net/jharnish/agileapi.v2"
       )

func main() {
    AgileUser := "myname"
    AgilePassword := "mypassword"
    Uplaodhost := "http://listen-l.upload.llnw.net"
    debug := true
    path := "/agileapi-test/"
    filename := "test.txt"

    agileapi := agileapi.New(AgileUser, AgilePassword, UploadHost, debug)

    //this is no longer needed with the recursive flag being set.
    //agileapi.MkDir2(path)

    data, err := os.Open(filename)
    if err != nil {
        fmt.Println(err)
    }
    fi, _ := data.Stat()
    err := agileapi.UploadFileStream(path, filename, data)
    if err != nil {
        fmt.Println("Error Uploading")
        fmt.Println(err)
    }
}
```
