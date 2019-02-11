package agileapi

import "log"

func (me *AgileApi) CreateMultipart(path, file string) (err error) {
	me.CheckAuth()
	params := map[string]string{
		"X-Agile-Authorization":  me.Token,
		"X-Agile-Directory":      path,
		"X-Agile-Basename":       file,
		"X-Agile-Expose-Egress":  "COMPLETE",
		"X-Agile-Content-Detect": "auto",
		"X-Agile-Recursive":      "true",
	}
	log.Println(params)
	return nil
}
