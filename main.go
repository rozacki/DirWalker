package main

/*
havealook:gerkin goconvey

Todo:unit tests and benchmarks
Todo:error handling
Todo:graceful shutdown via signals
*/
import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	_ "sort"
	"strconv"
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

	StartingFolder        string = "/Users/chrisrozacki/Desktop/music/brian  eno/brian eno - 1973 here come the warm jets"
	HeaderTemplateName    string = "header"
	FooterTemplateName    string = "footer"
	ItemStartTemplateName string = "item_start"
	ItemEndTemplateName   string = "item_end"

	ModTimelayout = "Jan 2, 2006 at 3:04pm"

	//error code
	//log file path
	LogFileName	string	="log.txt"
)
//error codes
const(

)
//error type
type Error struct{
	Op string
	Err error
	Port int
	Path string
}
//error type implemtns Error() method
func (err Error) Error() string{
	return fmt.Sprintf("error: %s while listen and server for port %d",err.Err.Error(),err.Port)
}

//walker mainly keeps
type DirWalker struct {
	//complied templates
	Templates map[string]*template.Template
	//
	DebugMode bool
}

//creates new instance of walker and initializes templates from files if template file is missing the default, hardcoded is used
func CreateDirWalker(debug bool) DirWalker {

	walker := DirWalker{}
	walker.DebugMode = debug

	//
	walker.Templates = make(map[string]*template.Template)
	walker.Templates = map[string]*template.Template{
		//use Stringer and provide template key
		// d---------
		FileTemplateName:               nil,
		os.ModeDir.String()[:1]:        nil,
		os.ModeAppend.String()[:1]:     nil,
		os.ModeExclusive.String()[:1]:  nil,
		os.ModeTemporary.String()[:1]:  nil,
		os.ModeSymlink.String()[:1]:    nil,
		os.ModeDevice.String()[:1]:     nil,
		os.ModeNamedPipe.String()[:1]:  nil,
		os.ModeSocket.String()[:1]:     nil,
		os.ModeSetuid.String()[:1]:     nil,
		os.ModeSetgid.String()[:1]:     nil,
		os.ModeCharDevice.String()[:1]: nil,
		os.ModeSticky.String()[:1]:     nil,
		FooterTemplateName:             nil,
		HeaderTemplateName:             nil,
		ItemStartTemplateName:          nil,
		ItemEndTemplateName:            nil,
	}

	//parse the default template of error then panic

	defaultTemplate, err := template.New(DefaultTemplateName).Parse(DefaultTemplateString)
	if err != nil {
		log.Panicf("default template parsing error, application will quit %+v\n", err)
	}

	for key, _ := range walker.Templates {
		t, err := template.New(key).ParseFiles(key)
		if err != nil {
			log.Print(err)
			log.Print("default template used")
			//reuse compiled template
			t = defaultTemplate
		} else {
			log.Print("templated loaded: ", key)
		}
		walker.Templates[key] = t
	}
	//add default template
	walker.Templates[DefaultTemplateName] = defaultTemplate

	return walker
}

func (self DirWalker) Start(urlPath string,nic string, port int) error{
	http.Handle(urlPath, &self)

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)

	if err != nil {
		return Error{Op:"listen and serve",Err:err,Port:port}
	}
	return nil
}

func (self *DirWalker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	debug := false
	if _, ok := r.URL.Query()["debug"]; ok {
		debug = true
	}
	dir := "/"
	dirs := r.URL.Query()["dir"]
	if dirs != nil {
		dir = filepath.Clean(dirs[0])
		log.Printf("dir=%s\n", dir)
	}
	items, err := ioutil.ReadDir(dir)
	if err != nil {
		if items, err = ioutil.ReadDir("/"); err != nil {
			panic(err)
		}
	}

	dirs = make([]string, 0)
	SplitPath(dir, &dirs)

	infoMap := map[string]interface{}{
		"Path":  dir,
		"Paths": dirs,
	}
	self.Templates[HeaderTemplateName].Execute(w, infoMap)

	for _, info := range items {
		log.Printf("%+v\n", info)
		log.Println(err)

		if debug {
			w.Write([]byte(fmt.Sprintf("%+v\n", info)))
		}

		//convert struct to map and send it to the template
		infoMap := map[string]interface{}{
			"Name":    info.Name(),
			"Size":    strconv.FormatInt(info.Size(), 10),
			"IsDir":   strconv.FormatBool(info.IsDir()),
			"Mode":    info.Mode().String(),
			"ModTime": info.ModTime().Format(ModTimelayout),
			"Path":    filepath.Join(dir, info.Name()),
		}

		var tmpl *template.Template = nil
		var key string = DefaultTemplateName
		var ok bool
		//if key exist
		//use key and template
		//else
		//use default
		if tmpl, ok = self.Templates[info.Mode().String()[:1]]; !ok {
			tmpl = self.Templates[key]
		}

		if debug {
			w.Write([]byte(fmt.Sprintf("resource %s\n", tmpl.Name())))
		}

		if tmpl == nil {
			log.Print("missing template for FileMode type of ", info.Mode())
			log.Print("default template used")
			//assign the one that always should be available
			tmpl = self.Templates[DefaultTemplateName]
		}
		err = self.Templates[ItemStartTemplateName].Execute(w, nil)
		err = tmpl.Execute(w, infoMap)
		err = self.Templates[ItemEndTemplateName].Execute(w, nil)
	}
	self.Templates[FooterTemplateName].Execute(w, nil)
	log.Println(err)
}

func SplitPath(path string, dirs *[]string) {
	*dirs = append(*dirs, path)
	log.Println(path, " ", len(path))
	//if nothing left then we stop
	if len(path) <= 1 {
		if len(path) == 0 {
			*dirs = append(*dirs, "/")
		}
		return
	}
	element := filepath.Base(path)
	fmt.Println(element)
	//starting index:ending index
	path = path[0 : (len(path)-len(element))-1]
	SplitPath(path, dirs)
}

func SetLogFile(logPath string ) error{

	f, err:=os.Create(logPath)
	if err!=nil{
		return Error{Op:"file open",Err:err,Path:logPath}
	}
	defer f.Close()

	log.SetOutput(f)
	return nil
}

func main() {

	//set log file
	SetLogFile(LogFileName)

	log.Println("staring ", AppName)

	//wait for signals
	//.....

	//parse command line parameters
	nic := flag.String("nic", "localhost", "")
	port := flag.Int("port", 8080, "")
	debug := flag.Bool("debug", false, "")
	urlPath:=flag.String("path","/","")
	flag.Parse()

	log.Println("interface=", *nic)
	log.Println("port=", *port)
	log.Println("debug=", *debug)
	log.Println("path=",*urlPath)

	//initialize the mani structure
	dw := CreateDirWalker(*debug)

	//star the server
	err:=	dw.Start(*urlPath ,*nic , *port)
	if err!=nil{
		log.Println(err)
		log.Println("server will stop")
		return
	}

	log.Println("server started")
}
