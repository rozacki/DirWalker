package main

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

//search interface provides partial search results until the stream is exhausted
//it does not provide sorting because it would have to cache all results before sorting but we don't understand how much time it will take
type Search interface {

	//the search job always responds within 1 sec with results that have been found between 0 and 1 sec of searching
	//it responds error=0 which means "no more result"
	//or, error=1 which is "more results"
	//having fetched results to the client it spins the timer and waits 20 sec, in the meantime it keeps searching
	//if the client does not initiate a new request withing 20 sec the job is cancelled
	//if any new non-search request is sent then search cancels too
	//if new search term is used the previous one if cancelled and a new search job is started
	//if the same search term is requested without waiting for the previous on (<1 sec) the the previous one returns immediately and the new on
	//is issued
	//if having finished the search, the same search is reissued a new job is spinned as there is no caching
	//
	//proposed additions: caching, user sessions, typeahed, filters, sorting, use channels to terminate the goroutine
	// and to notify parent process that goroutine is about to finish
	//

	//true is search is working
	Search(searchterm string) ([]os.FileInfo, int, string)
}

const (
	Idle        = 0
	Searching   = 1
	Terminating = 2
	Timeout     = 3
)

//search job searched recursively folders using pattern
//each matching item is sent via channel
//when job is done it sends notification via channel
//it can be cancelled
type SearchJob struct {
	Pattern string
	//int is a command sent to the goroutine: "cancel"
	Cancelled bool
	//search status
	SearchState SearchState
	//
	Mutex sync.Mutex
	//
	State int
	//goroutine state channel
	ProcessChangeState chan int
	//Code int
	Msg string
	//found items
	Items []Item
	//
	LastRead time.Time
	//Do we timeout if user does not request items, if recursive is false then we don't timeout
	Timeout bool
}

type SearchState struct {
}

//although the client may be using single http requests, this function should be reentrant
//
func (self *SearchJob) Search(searchterm, root string, istimeout bool) (items []Item, code int, msg string) {

	log.SetPrefix("search")
	defer log.SetPrefix("")

	log.Println("searchterm:", searchterm)
	log.Println("root:", root)
	cwd, _:=os.Getwd()
	log.Println("cwd:",cwd)
	log.Println("state:", self.State)

	//read the latest status change and unbloxk the procs
	select{
	case self.State=<-self.ProcessChangeState:
	default:
	}


	if self.State == Searching {
		if searchterm == "" {
			log.Println("terminating")
			//finish the process by passing anything to the channel
			self.ProcessChangeState <- Terminating
			//wait for response
			self.State=<-self.ProcessChangeState
			//I'm blocking here waiting for the goroutine to finalise and send its new status over channel
			return self.GetItems(),self.State , ""
		}

		//check if the term is different than before
		if searchterm != self.Pattern {
			log.Println("restarting")
			self.ProcessChangeState <- Terminating
			//wait for response
			self.State=<-self.ProcessChangeState
			log.Println("process stopped")
			//start new process
			go self.SearchProc(searchterm, root, istimeout)
			//block client, wait 1 sec and return what is in buffer
			<-time.After(time.Second * 1)
			//
			return self.GetItems(), self.State, ""
		} else {
			log.Println("fetch items")
			//return what is already cached immediately
			return self.GetItems(), self.State, ""
		}
	}

	if self.State == Idle {
		if searchterm	==	self.Pattern{
			self.Pattern	=	""
			log.Println("fetch last items")
			return self.GetItems(), self.State, ""
		}else {
			log.Println("starting")
			//start new process
			self.SearchProc(searchterm, root, istimeout)
			//wait one second
			<-time.After(time.Second * 1)
			//return what's in buffer
			return self.GetItems(), self.State, ""
		}
	}
	return nil, self.State, ""
}

//write to the writer and clear the list
//tbc:what if there are multiple requests? mutex?
func (self *SearchJob) GetItems() (items []Item) {
	self.Mutex.Lock()
	items = self.Items
	//update last access time
	self.LastRead = time.Now()
	//write to the writer
	self.Items = make([]Item, 0)
	self.Mutex.Unlock()
	log.Println("items:",len(items))
	return items
}

//goroutine: just search recursively
func (self* SearchJob) SearchProc(searchterm, root string,istimeout bool) {
	go func() {
		//
		self.State = Searching
		//
		self.Timeout = istimeout
		//
		self.Pattern=searchterm

		log.Println("proc start")
		//
		filepath.Walk(root, self.WalkFn)
		//process just finished so can be set to idle
		//it will block this goroutine

		log.Println("proc stop")
		self.ProcessChangeState <- Idle
	}()
}

func (self *SearchJob) WalkFn(path string, info os.FileInfo, err error) error {

	//<-time.After(time.Second*1)
	log.Println("sub cwd:",path)

	//why is this true?
	if info== nil{
		return nil
	}
	log.Println("item:",info.Name())

	select {
	//
	case code := <-self.ProcessChangeState:
		log.Println("received signal to terminate: ", code)
		return Error{Code: Terminating}
	default:
		//log.Println("nothing received")
	}

	var matched bool
	if matched, err = filepath.Match(self.Pattern, info.Name()); err != nil {
		//todo, shall I cancel ?
	}

	if matched {
		//add to the list
		self.Mutex.Lock()
		self.Items = append(self.Items, MakeItem(info, path))
		self.Mutex.Unlock()
	}

	return nil
}
