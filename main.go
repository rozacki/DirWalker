package main

/*
havealook:gerkin goconvey

Todo:unit tests and benchmarks
Todo:error handling
Todo:graceful shutdown via signals
Todo: error messages
Todo: fileviewer
*/
import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//all types of resources we can be dealing with
const (
	AppName                    string = "DirWalker"
	DefaultTemplateString      string = "<td>{{.Mode}}</td><td>{{.Name}}</td><td>{{.Size}}</td><td>{{.ModTime}}</td>"
	DefaultTemplateName        string = "x" //unique name, if user want he can override default template with x template name
	FileTemplateName           string = "-"
	ModeDirTemplateName        string = "d"
	ModeExclusiveTemplateName  string = "l"
	ModeTemporaryTemplateName  string = "T"
	ModeSymlinkTemplateName    string = "L"
	ModeDeviceTemplateName     string = "D"
	ModeNamedPipeTemplateName  string = "p"
	ModeSocketTemplateName     string = "S"
	ModeSetuidTemplateName     string = "u"
	ModeSetgidTemplateName     string = "g"
	ModeCharDeviceTemplateName string = "c"
	ModeStickyTemplateName     string = "t"

	StartingFolder string = "/Users/chrisrozacki/Desktop/music/brian  eno/brian eno - 1973 here come the warm jets"

	ModTimelayout = "Jan 2, 2006 at 3:04pm"

	//error code
	//log file path
	LogFileName string = "log.txt"
)

//error codes
const ()

//error type
type Error struct {
	Op   string
	Err  error
	Port int
	Path string
	Code int
}

//error type implements Error() method
func (err Error) Error() string {
	return fmt.Sprintf("error: %s while listen and server for port %d", err.Err.Error(), err.Port)
}

//walker mainly keeps
type DirWalker struct {
	//
	DebugMode bool
	//
	WriterA Writer
	//var SearchJobWrapper SearchJob=SearchJob{}
	SearchJob SearchJob
}

func CreateDirWalker(debug bool, format string) *DirWalker {
	walker := DirWalker{}
	log.Printf("created walker: %p\n", &walker)

	switch format {
	case "html":
		walker.WriterA = &TemplateWriter{}
		walker.WriterA.Init()
	case "json":
		walker.WriterA = &JSONWriter{}
	}

	walker.SearchJob	=	SearchJob{ProcessChangeState:make(chan int)}
	log.Printf("created walker %+v", walker)

	return &walker
}

func (self* DirWalker) Start(urlPath string, nic string, port int) error {
	http.Handle(urlPath, self)

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)

	if err != nil {
		return Error{Op: "listen and serve", Err: err, Port: port}
	}
	return nil
}

func (self* DirWalker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var ok bool
	dir := "/"
	dirs := r.URL.Query()["dir"]
	if dirs != nil {
		dir = filepath.Clean(dirs[0])
		//log.Printf("dir=%s\n", dir)
	}
	items, err := ioutil.ReadDir(dir)
	if err != nil {
		if items, err = ioutil.ReadDir("/"); err != nil {
			panic(err)
		}
	}

	//define search parameters, if search param is not provided then no search action is triggered
	searchterm := ""
	if _, ok = r.URL.Query()["search"]; ok {
		searchterm = r.URL.Query()["search"][0]
		items, code, msg := self.SearchJob.Search(searchterm, dir,true)

		self.WriterA.WriteHeader(w, nil, code, msg)
		for _, item := range items {
			self.WriterA.WriteItem(w, item)
		}
		self.WriterA.WriteFooter(w)
		return
	}

	//define sort parameters, always sort, by default use name field
	sortBy := "name"
	if _, ok = r.URL.Query()["sort"]; ok {
		sortBy = r.URL.Query()["sort"][0]
	}

	sorting := false

	sortDir := false
	if _, ok = r.URL.Query()["sd"]; ok {
		if "asc" == strings.ToLower(r.URL.Query()["sd"][0]) {
			sortDir = true
			sorting = true
		} else if "desc" == strings.ToLower(r.URL.Query()["sd"][0]) {
			sortDir = false
			sorting = true
		}
	}

	if sorting {
		var sortableFileInfo SortableFileInfo = SortableFileInfo{Data: items, SortBy: sortBy, Dir: sortDir}
		sort.Sort(sortableFileInfo)
	}

	//split paths for breadcrumb
	dirs = nil
	SplitPath(dir, &dirs)
	infoMap := map[string]interface{}{
		"Path":       dir,
		"Breadcrumb": dirs,
	}
	//start writing response
	self.WriterA.WriteHeader(w, infoMap, 0, "ok")
	for _, info := range items {
		self.WriterA.WriteItem(w, MakeItem(info, dir))
	}
	self.WriterA.WriteFooter(w)
}

func MakeItem(info os.FileInfo, path string) Item {
	//convert struct to map and send it to the template
	return map[string]interface{}{
		"Name":    info.Name(),
		"Size":    info.Size(),
		"IsDir":   info.IsDir(),
		"Mode":    info.Mode(),
		"ModTime": info.ModTime(),
		"Path":    filepath.Join(path, info.Name()),
	}
}

func SplitPath(path string, dirs *[]string) {
	*dirs = append(*dirs, path)
	//log.Println(path, " ", len(path))
	//if nothing left then we stop
	if len(path) <= 1 {
		if len(path) == 0 {
			*dirs = append(*dirs, "/")
		}
		return
	}
	element := filepath.Base(path)
	fmt.Println(element)
	//starting index:ending indexes
	path = path[0 : (len(path)-len(element))-1]
	SplitPath(path, dirs)
}

func SetLogFile(logPath string) error {

	f, err := os.Create(logPath)
	if err != nil {
		return Error{Op: "file open", Err: err, Path: logPath}
	}
	defer f.Close()

	log.SetOutput(f)
	return nil
}

func main() {

	//set log file
	//SetLogFile(LogFileName)

	log.Println("staring ", AppName)

	//wait for signals
	//.....

	//parse command line parameters
	nic := flag.String("nic", "localhost", "")

	port := flag.Int("port", 8080, "")
	debug := flag.Bool("debug", false, "")
	urlPath := flag.String("path", "/", "")
	format := flag.String("format", "json", "")
	flag.Parse()

	//initialize the mani structure
	dw := CreateDirWalker(*debug, *format)
	log.Printf("returne walker: %p ", dw)
	//star the server
	err := dw.Start(*urlPath, *nic, *port)
	if err != nil {
		log.Println(err)
		log.Println("server will stop")
		return
	}

	log.Println("server started")
}
