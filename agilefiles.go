package agileapi

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/Harnish/sha256proxy"
	pb "gopkg.in/cheggaaa/pb.v1"
)

type Filestruct struct {
	Filename string
	Url      string
	Mtime    time.Time
	Ctime    time.Time
	Size     uint64
	Sha256   string
	Path     string
	UUID     string
	Inode    uint64
}

type FilePath struct {
	Path  string
	af    *AgileFiles
	Dirs  []Filestruct
	Files []Filestruct
}

type AgileFiles struct {
	AgileApi  *AgileApi
	EgressURL string
	Debug     bool
}

func ReturnMime(mypath string) (mimename string) {
	ext := path.Ext(mypath)
	mimenamefull := mime.TypeByExtension(ext)
	mimenameparts := strings.Split(mimenamefull, "/")
	mimename = mimenameparts[0]
	return
}

func (me *AgileFiles) GetPath(path string) *FilePath {
	myreturn := &FilePath{
		Path:  path,
		af:    me,
		Dirs:  me.GetDirs(path),
		Files: me.GetFiles(path),
	}
	return myreturn
}

func (me *AgileFiles) GetFiles(path string) (files []Filestruct) {
	spacer := ""
	if path != "/" {
		spacer = "/"

	}
	me.writedebug(path)
	myfiles := me.AgileApi.ListAllFilesDetails(path)
	for myfile := range myfiles {
		myurl := path + spacer + myfiles[myfile].Filename
		mysize := uint64(myfiles[myfile].Stat.Size)
		mymtime := time.Unix(int64(myfiles[myfile].Stat.Mtime), 0)

		data := Filestruct{
			Filename: myfiles[myfile].Filename,
			Url:      myurl,
			Mtime:    mymtime,
			Size:     mysize,
			Sha256:   myfiles[myfile].Stat.Sha256,
			Path:     myurl,
		}
		files = append(files, data)
	}
	return
}

func (me *AgileFiles) GetDirs(path string) (temp []Filestruct) {
	mydirs := me.AgileApi.ListAllDirsDetails(path)
	for mydir := range mydirs {
		spacer := ""
		if path != "/" {
			spacer = "/"
		}
		myurl := path + spacer + mydirs[mydir].Filename
		mysize := uint64(0)
		mymtime := time.Unix(int64(mydirs[mydir].Stat.Mtime), 0)
		data := Filestruct{
			Filename: mydirs[mydir].Filename,
			Url:      myurl,
			Mtime:    mymtime,
			Size:     mysize,
			Path:     myurl,
		}
		temp = append(temp, data)
	}
	return
}

func (me *AgileFiles) GetFile(path string) (*File, error) {
	stat, err := me.AgileApi.StatFile(path)
	if err != nil {
		return nil, err
	}
	egresspath := me.EgressURL
	if strings.HasSuffix(me.EgressURL, "/") && strings.HasPrefix(path, "/") {
		egresspath = strings.TrimSuffix(egresspath, "/")
	}
	returnobj := &File{
		Url:    egresspath + path,
		Mtime:  time.Unix(int64(stat.Mtime), 0),
		Ctime:  time.Unix(int64(stat.Ctime), 0),
		Size:   uint64(stat.Size),
		Sha256: stat.Checksum,
		Path:   path,
		af:     me,
	}

	return returnobj, nil
}

func (me *AgileFiles) UploadFileStreamReturnSha(path, filename string, filereader io.Reader, size int64, progress bool) (string, error) {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	shain := shaproxy.New()
	sha_reader := shain.NewProxyReader(filereader)
	buf_reader := bufio.NewReader(sha_reader)
	if progress {
		bar := pb.New64(size).SetUnits(pb.U_BYTES)
		bar.Start()
		progress_reader := bar.NewProxyReader(buf_reader)
		err := me.AgileApi.UploadFileStream(path, filename, progress_reader)
		bar.Finish()
		if err != nil {
			return "", fmt.Errorf("AgileFiles.UploadFilesStreamReturnSha - Error: %s", err)
		}

	} else {
		err := me.AgileApi.UploadFileStream(path, filename, buf_reader)
		if err != nil {
			return "", fmt.Errorf("AgileFiles.UploadFilesStreamReturnSha - Error: %s", err)
		}
	}
	shain.Finish()
	shasum := shain.SumHex()
	return shasum, nil

}

func (me *AgileFiles) CheckAgileSHA(path, mysha256 string) (bool, error) {
	me.writedebug("CheckAgileSHA - path: " + path + " SHA256: " + mysha256)
	response, err := http.Head(me.EgressURL + path)
	if err != nil {
		return false, err
	}
	remoteSha256 := response.Header.Get("X-Agile-Checksum")
	me.writedebug("CheckAgileSHA - path: " + path + " SHA256: " + mysha256 + " Limelight's sha256: " + remoteSha256)
	if remoteSha256 == mysha256 {
		return true, nil
	}
	return false, nil

}

func (me *AgileFiles) IsFile(mypath string) (bool, error) {
	dirpath, filename := path.Split(mypath)
	var files []Filestruct
	files = me.GetFiles(dirpath)
	for _, file := range files {
		if file.Filename == filename {
			return true, nil
		}
	}
	return false, nil
}

func (me *AgileFiles) Type(mypath string) (int, error) {
	// Type 1 is dir
	// Type 2 is file
	// Type 0 doesn't exist
	isfile, err := me.IsFile(mypath)
	if isfile {
		return 2, err
	}

	isdir, err := me.IsDir(mypath)
	if isdir {
		return 1, err
	}
	return 0, nil
}

func (me *AgileFiles) IsDir(mypath string) (bool, error) {
	dirpath, filename := path.Split(mypath)
	var files []Filestruct
	files = me.GetDirs(dirpath)
	for _, file := range files {
		if file.Filename == filename {
			return true, nil
		}
	}
	return false, nil
}

func (me *AgileFiles) writedebug(message string) {
	if me.Debug {
		log.Println("DEBUG - agilefiles.go :  " + message)
	}

}

/*
func (me *AgileFiles) NewFileWriter(filepath string) io.Writer {
	return me.AgileApi.UploadFileWriter(filepath)
}
*/
func (me *AgileFiles) NewFile(filename, path string, data io.Reader) (*File, error) {
	err := me.AgileApi.UploadFileStream(path, filename, data)
	if err != nil {
		return nil, err
	}
	stat, err := me.AgileApi.StatFile(path + filename)
	if err != nil {
		return nil, err
	}
	egresspath := me.EgressURL
	if strings.HasSuffix(me.EgressURL, "/") && strings.HasPrefix(path, "/") {
		egresspath = strings.TrimSuffix(egresspath, "/")
	}
	returnobj := &File{
		Url:    egresspath + path + filename,
		Mtime:  time.Unix(int64(stat.Mtime), 0),
		Ctime:  time.Unix(int64(stat.Ctime), 0),
		Size:   uint64(stat.Size),
		Sha256: stat.Checksum,
		Path:   path + filename,
		af:     me,
	}

	return returnobj, nil
}
