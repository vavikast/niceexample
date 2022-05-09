package main

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type FileNode struct {
	FullPath string
	Info os.FileInfo
}

func insertSorted(fileList *list.List,fileNode FileNode)  {
	if fileList.Len()==0{
		fileList.PushFront(fileNode)
		return
	}
	for element := fileList.Front(); element != nil; element = element.Next(){
		if fileNode.Info.Size() < element.Value.(FileNode).Info.Size(){
			fileList.InsertBefore(fileNode,element)
			return
		}
	}
	fileList.PushBack(fileNode)
}
func getFilesInDirRecursivelyBySize(fileList *list.List,path string)  {
	dirFiles,err := ioutil.ReadDir(path)
	if err != nil {
		log.Println("Error reading directory:",err.Error())
	}
	for _,dirFile := range dirFiles{
		fullpath := filepath.Join(path,dirFile.Name())
		if dirFile.IsDir(){
			getFilesInDirRecursivelyBySize(fileList,filepath.Join(path,dirFile.Name()))
		}else if dirFile.Mode().IsRegular() {
			insertSorted(fileList,FileNode{FullPath: fullpath,Info:dirFile})
		}
	}
}
func main()  {
	fileList := list.New()
	getwd, err := os.Getwd()
	if err !=nil{
		fmt.Println(err)
		return
	}
	getFilesInDirRecursivelyBySize(fileList,getwd)
	for element := fileList.Front(); element != nil; element =element.Next(){
		fmt.Printf("%d ",element.Value.(FileNode).Info.Size())
		fmt.Printf("%s\n ",element.Value.(FileNode).FullPath)
	}
	element := fileList.Back()
	fmt.Printf("%s ",GetMemSize(element.Value.(FileNode).Info.Size()))
	fmt.Printf("%s\n ",element.Value.(FileNode).FullPath)
}
func GetMemSize(u int64) (size string) {
	if u < 1024 {
		size = fmt.Sprintf("%.2fB", float64(u))
		return
	} else if float64(u) < 1024*1024 {
		size = fmt.Sprintf("%.2fKB", float64(u)/float64(1024))
		return
	} else if float64(u) < 1024*1024*1024 {
		size = fmt.Sprintf("%.2fMB", float64(u)/float64(1024*1024))
		return
	} else if float64(u) < 1024*1024*1024*1024 {
		size = fmt.Sprintf("%.2fGB", float64(u)/float64(1024*1024*1024))
		return
	} else {
		size = fmt.Sprintf("%.2fTB", float64(u)/float64(1024*1024*1024))
		return
	}

}

