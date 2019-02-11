This is an Object for working with LimeLight "Agile" Cloud Storage Platform via the jsonrpc and post api calls.

Using the api directly.
```golang
package main

import (
    "fmt"
    "github.llnw.net/jharnish/agileapi.v3"
       )

func main() {
    AgileUser := "myname"
    AgilePassword := "mypassword"
    Uplaodhost := "labs-l.upload.llnw.net"
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

Using the Helpers:
```golang 
package main

import (
    "fmt"
    "github.llnw.net/jharnish/agileapi.v3"
       )

func main() {
    AgileUser := "myname"
    AgilePassword := "mypassword"
    Uplaodhost := "labs-l.upload.llnw.net"
    debug := true
    path := "/agileapi-test/"
    filename := "test.txt"
    egresspath := "http://mycompany.cdn.limelight.com/"

    agileapi := agileapi.New(AgileUser, AgilePassword, UploadHost, debug)
    agilefs := agileapi.NewFS(egresspath)

    //Upload File
    data, err := os.Open(filename)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    fi, _ := data.Stat()
    bar := pb.New64(fi.Size).SetUnits(pb.U_BYTES)
	bar.Start()
	progress_reader := bar.NewProxyReader(data)		
    err := agilefs.NewFile(filename, path, progress_reader)
    bar.Finish()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    //Get Contents
    myfile := agilefs.GetFile(path + filename)
    myreader := myfile.NewReader()
    

}

```
