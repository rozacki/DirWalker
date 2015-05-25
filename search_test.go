package main

import (
	"testing"
	"time"
	"os"
)
const FolderName string	=	"someguiqwer1234"
var searchJob SearchJob =	SearchJob{}

func TestSearch_basic(t* testing.T){
	//create empty folder
	os.Mkdir(FolderName,os.ModeDir)

	searchStep(1,"search " ,"a" , FolderName,t,1 ,"",0,1)

	searchStep(2,"search " ,"a" ,FolderName,t,1 ,"",0,0)

	//delete the empty folder
	os.Remove(FolderName)
}

func TestSearchFindFiles(t* testing.T){
	os.Mkdir(FolderName,os.ModeDir)

	//create 100 files starting with name a
	for i:=0;i<=100;i++{
		os.Create("a"+string(i))
	}


	os.Remove(FolderName)
}

func TestSearch_(t* testing.T){
	//create a few empty files
	//read folder
	//check items
	//check duration
	//delete the files created at the beginning
}

func searchStep(step int, prefix string, searchterm string, folder string, t* testing.T, c int, m string, l int, d int) (items []Item, code int,msg string, duration int){
	t.Log(step, " phase on ",prefix)
	start:=time.Now()
	items, code , msg = searchJob.Search(searchterm, folder, false)
	//check duration
	duration=int(time.Now().Sub(start).Seconds())

	t.Log("duration=",duration,"code=",code,"msg=",msg,"len(items)=",len(items))
	if duration!=d{
		t.Error("duration != 1 sec: ", duration)
	}

	if code!= c{
		t.Error("error code != 1:", code)
	}

	if msg!= m{
		t.Errorf("msg != \"\":", msg)
	}


	if len(items)!=l{
		t.Error("len(items) >0:",len(items))
	}

	t.Log("finish phase ", step)

	return
}
