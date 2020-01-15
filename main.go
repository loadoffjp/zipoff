package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/yeka/zip"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func main() {
	var (
		password = flag.String("p", "", "password")
		name     = flag.String("n", "./default.zip", "zip file name")
	)

	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("zipにする対象のディレクトリを指定してください")
		fmt.Println("zipoff -n zipname.zip path/to/dir")
		return
	}
	list := getFiles(args[0])
	if list == nil {
		return
	}
	dir, _ := filepath.Split(args[0])
	createZip(*name, list, dir, *password)
	fmt.Println("zipを作成しました。[", *name, "]")
}

// getFiles dir以下に存在するファイルの一覧を取得する
func getFiles(dir string) []string {
	fileList := []string{}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("エラーが発生しました[", err, "]")
		return nil
	}
	for _, f := range files {
		if f.IsDir() == true {
			list := getFiles(path.Join(dir, f.Name()))
			fileList = append(fileList, list...)
			continue
		}
		fileList = append(fileList, path.Join(dir, f.Name()))
	}
	return fileList
}

func createZip(zipPath string, fileList []string, replace string, password string) {
	var zipfile *os.File
	var err error
	if zipfile, err = os.Create(zipPath); err != nil {
		log.Fatalln(err)
	}
	defer zipfile.Close()
	w := zip.NewWriter(zipfile)
	for _, file := range fileList {
		read, err := os.Open(file)
		defer read.Close()
		if err != nil {
			fmt.Println(err)
			continue
		}
		var f io.Writer
		name := strings.Replace(file, replace, "", 1)
		name, err = UTF8toSJIS(name)
		if err != nil {
			fmt.Println(file)
			fmt.Println(err)
			continue
		}
		if password == "" {
			f, err = w.Create(name)
		} else {
			f, err = w.Encrypt(name, password, zip.StandardEncryption)
		}
		if err != nil {
			fmt.Println(err)
			continue
		}
		if _, err = io.Copy(f, read); err != nil {
			fmt.Println(err)
			continue
		}
	}
	w.Close()
}

// UTF8toSJIS UTF-8 から ShiftJIS
func UTF8toSJIS(str string) (string, error) {
	str = norm.NFC.String(str)
	ret, err := ioutil.ReadAll(transform.NewReader(strings.NewReader(str), japanese.ShiftJIS.NewEncoder()))
	if err != nil {
		return "", err
	}
	return string(ret), err
}
