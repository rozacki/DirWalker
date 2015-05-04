package main

import (
	"os"
	"strings"
)

type SortableFileInfo struct {
	Data   []os.FileInfo
	SortBy string
	Dir    bool
}

func (fi SortableFileInfo) Len() int {
	return len(fi.Data)
}
func (fi SortableFileInfo) Less(i, j int) bool {
	//default field is name
	if fi.SortBy == "" {
		fi.SortBy = "name"
	}else{
		fi.SortBy=strings.ToLower(fi.SortBy)
	}

	//
	switch fi.SortBy {
	default:
	case "name":
		if !fi.Dir {
			return fi.Data[i].Name() < fi.Data[j].Name()
		} else {
			return fi.Data[i].Name() > fi.Data[j].Name()
		}
	case "size":
		if !fi.Dir {
			return fi.Data[i].Size() < fi.Data[j].Size()
		} else {
			return fi.Data[i].Size() > fi.Data[j].Size()
		}
	case "modtime":
		if !fi.Dir {
			return fi.Data[i].ModTime().Before(fi.Data[j].ModTime())
		} else {
			return fi.Data[i].ModTime().After(fi.Data[j].ModTime())
		}
	}
	return false
}

func (fi SortableFileInfo) Swap(i, j int) {
	fi.Data[i], fi.Data[j] = fi.Data[j], fi.Data[i]
}

func (fi *SortableFileInfo) SetField(field string) {
	fi.SortBy = field
}
