package main
/*
Todo:unit tests and benchamrks

gerkin: goconvey
upload to github
issue tracker
 */
import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"reflect"
	"io/ioutil"
)

//all types of resources we can be dealing with
const (
	DefaultTemplateString      string = "<td>{{.Mode}}</td><td>{{.Name}}</td><td>{{.Size}}</td><td>{{.ModTime}}</td>"
	DefaultTemplateName        string = "x"//unique name, if user want he can override default template with x template name
	FileTemplateName		   string="-"
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
	HeaderTemplateName			string="header"
	FooterTemplateName			string="footer"
	ItemStartTemplateName		string="item_start"
	ItemEndTemplateName		string="item_end"
)

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
	walker.Templates=map[string]*template.Template{
		//use Stringer and provide template key
		// d---------
		FileTemplateName:	nil,
		os.ModeDir.String()[:1]:	nil,
		os.ModeAppend.String()[:1]: nil,
		os.ModeExclusive.String()[:1]:	nil,
		os.ModeTemporary.String()[:1]:  nil,
		os.ModeSymlink.String()[:1]:    nil,
		os.ModeDevice.String()[:1]:     nil,
		os.ModeNamedPipe.String()[:1]:  nil,
		os.ModeSocket.String()[:1]:     nil,
		os.ModeSetuid.String()[:1]:     nil,
		os.ModeSetgid.String()[:1]:     nil,
		os.ModeCharDevice.String()[:1]: nil,
		os.ModeSticky.String()[:1]:     nil,
		FooterTemplateName:nil,
		HeaderTemplateName:nil,
		ItemStartTemplateName:nil,
		ItemEndTemplateName:nil,
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
			t= defaultTemplate
		} else {
			log.Print("templated loaded: ", key)
		}
		walker.Templates[key] = t
	}
	//add default template
	walker.Templates[DefaultTemplateName] = defaultTemplate

	return walker
}

func (self *DirWalker) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	self.Templates[HeaderTemplateName].Execute(w,nil)

	items,err:=	ioutil.ReadDir("/")

	for _,info:= range	items {

		log.Printf("%+v\n", info)
		log.Println(err)
		debug := false
		if _, ok := r.URL.Query()["debug"]; ok {
			debug = true
		}

		if debug {
			w.Write([]byte(fmt.Sprintf("%+v\n", info)))
		}

		//convert struct to map and send it to the template
		infoMap := map[string]string{
		"Name":    info.Name(),
		"Size":    strconv.FormatInt(info.Size(), 10),
		"IsDir":   strconv.FormatBool(info.IsDir()),
		"Mode":    info.Mode().String(),
		"ModTime": info.ModTime().Format(ModTimelayout),
		}

		var tmpl *template.Template = nil
		var key string = DefaultTemplateName
		var ok    bool
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
		err=self.Templates[ItemStartTemplateName].Execute(w,nil)
		err = tmpl.Execute(w, infoMap)
		err=self.Templates[ItemEndTemplateName].Execute(w,nil)
	}
	self.Templates[FooterTemplateName].Execute(w,nil)
	log.Println(err)
}

const ModTimelayout = "Jan 2, 2006 at 3:04pm"

func WalkDirHtml(w http.ResponseWriter, r *http.Request) {
	err := filepath.Walk("/Users/chrisrozacki/Desktop/music/brian  eno/brian eno - 1973 here come the warm jets", func(path string, info os.FileInfo, err error) error {
		tmpl, err := template.New("test").Parse("<h3>Name</h3> - {{.Name}}	Size	-	{{.Size}}		IsDir	-	{{.IsDir}}	Mode	-	{{.Mode}}	ModTime	-	{{.ModTine}}</br>")
		/*
			Name() string       // base name of the file
			Size() int64        // length in bytes for regular files; system-dependent for others
			Mode() FileMode     // file mode bits
			ModTime() time.Time // modification time
			IsDir() bool        // abbreviation for Mode().IsDir()
			Sys() interface{}
		*/

		m := map[string]string{
			"Name":    info.Name(),
			"Size":    strconv.FormatInt(info.Size(), 10),
			"IsDir":   strconv.FormatBool(info.IsDir()),
			"Mode":    strconv.FormatUint(uint64(info.Mode()), 10),
			"ModTime": info.ModTime().Format(ModTimelayout),
		}
		w.Header().Set("Content-Type", "text/html")
		err = tmpl.Execute(w, m)

		return nil
	})
	log.Println(err)
}

func WalkFunc(path string, info os.FileInfo, err error) error {
	fmt.Println(path, info, err)
	return nil
}

func PrintFileModes(){
	/*fmt.Println(os.ModeAppend,os.ModeExclusive,os.ModeTemporary,
		os.ModeSymlink,os.ModeDevice,
		os.ModeNamedPipe,os.ModeSocket,os.ModeSetuid,
		os.ModeSetgid,os.ModeCharDevice,os.ModeSticky, os.ModeType, os.ModePerm)
		*/
	fmt.Print(os.ModeDir.String()[0:1])

	fmt.Printf("%+v\n",reflect.TypeOf(os.ModeDir))
	//fmt.Printf("%+v\n",reflect.TypeOf(os.FileMode))
}


func main() {

	PrintFileModes();

	log.Println("staring SYSInfo")

	dw := CreateDirWalker(true)


	http.HandleFunc("/a", WalkDirHtml)
	http.Handle("/b", &dw)

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Panic(err)
	}
}
