package agileapi

import (
	"sort"
	"strings"
)

type Files []Filestruct

type By func(f1, f2 *Filestruct) bool

func (by By) Sort(files []Filestruct) {
	ps := &filesSorter{
		files: files,
		by:    by, // The Sort method's receiver is the function (closure) that defines the sort order.

	}
	sort.Sort(ps)

}

type filesSorter struct {
	files []Filestruct
	by    func(p1, p2 *Filestruct) bool
}

func (f *filesSorter) Len() int {
	return len(f.files)

}

func (f *filesSorter) Swap(i, j int) {
	f.files[i], f.files[j] = f.files[j], f.files[i]

}

func (f *filesSorter) Less(i, j int) bool {
	return f.by(&f.files[i], &f.files[j])

}

func (me *AgileFiles) SortFiles(files []Filestruct, mysortby, sortorder string, caseinsensitive bool) (sortedlist []Filestruct) {
	sortby := mysortby
	sorters := make(map[string]func(*Filestruct, *Filestruct) bool)

	sorters["name"] = func(p1, p2 *Filestruct) bool {
		return p1.Filename < p2.Filename
	}
	sorters["namelc"] = func(p1, p2 *Filestruct) bool {
		return strings.ToLower(p1.Filename) < strings.ToLower(p2.Filename)
	}
	sorters["namedesc"] = func(p1, p2 *Filestruct) bool {
		return p1.Filename > p2.Filename
	}
	sorters["namedesclc"] = func(p1, p2 *Filestruct) bool {
		return strings.ToLower(p1.Filename) > strings.ToLower(p2.Filename)
	}
	sorters["mtime"] = func(p1, p2 *Filestruct) bool {
		//FIXME convert mtime to number
		return p1.Mtime.Before(p2.Mtime)
	}
	sorters["mtimedesc"] = func(p1, p2 *Filestruct) bool {
		//FIXME convert Mtime to number
		return p1.Mtime.After(p2.Mtime)
	}
	sorters["size"] = func(p1, p2 *Filestruct) bool {
		return p1.Size < p2.Size
	}
	sorters["sizedesc"] = func(p1, p2 *Filestruct) bool {
		return p1.Size > p2.Size
	}
	if sortorder == "desc" {
		sortby = sortby + "desc"
	}
	if caseinsensitive {
		sortby = sortby + "lc"
	}
	By(sorters[sortby]).Sort(files)
	sortedlist = files
	return
}
