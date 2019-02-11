package agileapi

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type File struct {
	Url    string
	Mtime  time.Time
	Ctime  time.Time
	Size   uint64
	Sha256 string
	Path   string
	UUID   string
	Inode  uint64
	af     *AgileFiles
}

func (me *File) NewReader() (io.Reader, error) {
	//FIXME doesn't work
	// may need to make into a buffio
	req, err := http.Get(me.Url)
	if err != nil {
		return nil, fmt.Errorf("NewReader failed "+me.Url+" Error: ", err)
	}
	defer req.Body.Close()
	if req.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("File Not found")
	}
	return req.Body, nil
}

func (me *File) Contents() ([]byte, error) {
	req, err := http.Get(me.Url)
	if err != nil {
		return nil, fmt.Errorf("Contents failed "+me.Url+" Error: ", err)
	}
	if req.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("File Not Found : " + me.Url)
	}
	output, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body.Close()
	return output, nil
}

func (me *File) Delete() error {
	err := me.af.AgileApi.RmFile(me.Path)
	return fmt.Errorf("failed to perform Delete on "+me.Path+" Error: ", err)
}

func (me *File) Rename(newname string) error {
	fmt.Println("Rename " + me.Path + " to " + newname)
	err := me.af.AgileApi.RenameFile(me.Path, newname)
	if err != nil {
		return fmt.Errorf("failed to rename "+me.Path+" to "+newname+" Error: ", err)
	}
	me.Path = newname
	return nil
}

func (me *File) Move(newname string) error {
	return me.Rename(newname)
}

func (me *File) SetMtime(mtime time.Time) error {
	err := me.af.AgileApi.SetMTime(me.Path, string(mtime.Unix()))
	return fmt.Errorf("Failed to set Mtime on "+me.Path+" Error: ", err)
}
