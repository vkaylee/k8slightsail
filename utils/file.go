package utils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)
type File struct {
	path		*string
	contents	*string
}

func InitFileWithPath(path *string) *File {
	return &File{
		path:     path,
		contents: nil,
	}
}
func (f *File) SetContents(contents *string)  {
	f.contents = contents
}
func (f *File) GetContents() *string {
	return f.contents
}
func (f *File) GetPath() *string {
	return f.path
}
func (f *File) Loadfile() error {
	data, err := ioutil.ReadFile(*f.path)
	if err != nil {
		return err
	}
	f.contents = String(string(data))
	return nil
}
func (f *File) WriteFile() {
	// Use os.Create to create a file for writing.
	file, _ := os.Create(*f.path)

	// Create a new writer.
	w := bufio.NewWriter(file)

	// Write a string to the file.
	_, _ = w.WriteString(*f.contents)

	// Flush.
	_ = w.Flush()
}
func (f *File) IsExistedFile() bool {
	if _, err := os.Stat(*f.path); os.IsNotExist(err) {
		return false
	}
	return true
}
func (f *File) RemoveFile()  {
	if f.IsExistedFile() {
		err := os.Remove(*f.path)
		if err != nil {
			fmt.Println(err)
			return
		}
		log.Printf("File %s successfully deleted", *f.path)
	}
}