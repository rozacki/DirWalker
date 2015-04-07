package main

import(
	"html/template"
	"net/http"
	"os"
	"log"
	"fmt"
)
const(
	HeaderTemplateName    string = "header"
	FooterTemplateName    string = "footer"
	ItemStartTemplateName string = "item_start"
	ItemEndTemplateName   string = "item_end"
)

type TemplateWriter struct{

	Templates map[string]*template.Template

	Writer	http.ResponseWriter

	//logger
	//

	Initiated bool

	Debug bool
}

func (self* TemplateWriter) Init() error{
	self.Initiated=true
	//
	self.Templates = map[string]*template.Template{
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
		return WriterError{Message: "default template parsing error",Err:err}
	}

	for key, _ := range self.Templates {
		t, err := template.New(key).ParseFiles(key)
		if err != nil {
			log.Print(err)
			log.Print("default template used")
			//reuse compiled template
			t = defaultTemplate
		} else {
			log.Print("templated loaded: ", key)
		}
		self.Templates[key] = t
	}
	//add default template
	self.Templates[DefaultTemplateName] = defaultTemplate

	log.Println(self.len())

	return nil
}

func (self TemplateWriter) SetSessionWriter(w http.ResponseWriter){

}

func (self TemplateWriter) WriteHeader(w http.ResponseWriter, item Item) error{
	w.Header().Set("Content-Type", "text/html")
	self.Templates[HeaderTemplateName].Execute(w, item)
	return nil
}

func (self TemplateWriter) WriteStartItem(w http.ResponseWriter){

}

func (self TemplateWriter) WriteItem(w http.ResponseWriter, item Item) error{

	var tmpl *template.Template = nil
	var key string = DefaultTemplateName
	var ok bool
	//if key exist
	//use key and template
	//else
	//use default
	if tmpl, ok = self.Templates[item["Mode"].(string)[:1]]; !ok {
		tmpl = self.Templates[key]
	}

	if self.Debug {
		w.Write([]byte(fmt.Sprintf("resource %s\n", tmpl.Name())))
	}

	if tmpl == nil {
		log.Print("missing template for FileMode type of ", item["Mode"].(string))
		log.Print("default template used")
		//assign the one that always should be available
		tmpl = self.Templates[DefaultTemplateName]
	}
	err := self.Templates[ItemStartTemplateName].Execute(w, nil)
	err = tmpl.Execute(w, item)
	err = self.Templates[ItemEndTemplateName].Execute(w, nil)

	return err
}

func (self TemplateWriter) WriteEndItem(w http.ResponseWriter){

}

func (self TemplateWriter) WriteFooter(w http.ResponseWriter, item Item){
	self.Templates[FooterTemplateName].Execute(w, item)
}

func (self TemplateWriter) len() int{
	return len(self.Templates)
}
