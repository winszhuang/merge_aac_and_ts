package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type VideoInfo struct {
	TS  string
	AAC string
}

const outputFolderName = "output"

func main() {
	// 指定資料夾路徑(當前資料夾)
	dir := "."

	// 建立output資料夾
	outputDir := filepath.Join(dir, outputFolderName)
	err := os.Mkdir(outputDir, 0755)
	if err != nil {
		if strings.Contains(err.Error(), "Cannot create a file when that file already exists") {
			log.Println("輸出資料夾已經存在瞜 ~ ")
		} else {
			log.Fatal(err)
		}
	}

	// 尋找符合條件的檔案
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	// 清除檔案名稱中不必要的空白
	errList := TrimAllSpace(files)
	ShowMultiErrors(errList)

	// 組合所有ts和aac成mp4檔案
	errList = CombineAllFiles(files)
	ShowMultiErrors(errList)

	log.Println("轉換成功!! 請查看output資料夾。(10秒後該程序會自動關閉)")
	time.Sleep(time.Second * 10)
}

func CombineAllFiles(files []fs.FileInfo) []error {
	allFileNames := GetAllFileNames(files)
	fileMap := GroupVideoInfo(allFileNames)
	errList := []error{}
	wg := sync.WaitGroup{}
	for key, value := range fileMap {
		wg.Add(1)
		go func(k string, v *VideoInfo) {
			defer func() {
				wg.Done()
			}()

			aacFile := v.AAC
			tsFile := v.TS
			mp4FileName, err := GenerateMp4FileName(aacFile, k)
			if err != nil {
				errList = append(errList, fmt.Errorf("從 %s 生成新的mp4檔案有問題!!", aacFile))
				return
			}
			CombineTsAndAAC(tsFile, aacFile, mp4FileName)
			os.Rename(mp4FileName, outputFolderName+"/"+mp4FileName)
		}(key, value)
	}
	wg.Wait()
	return errList
}

func GenerateMp4FileName(fileName string, index string) (string, error) {
	chunks := strings.Split(fileName, "Press")
	if len(chunks) != 2 {
		return "", fmt.Errorf("%s組成新的mp4檔案錯誤!!", fileName)
	}
	re := regexp.MustCompile(`-+$`)
	str := re.ReplaceAllString(chunks[0], "")
	str = strings.Replace(str, index, index+" ", 1)

	return str + ".mp4", nil
}

func GroupVideoInfo(fileNames []string) map[string]*VideoInfo {
	videoMap := make(map[string]*VideoInfo)
	re := regexp.MustCompile(`^(\d+-\d+)`)
	for _, fileName := range fileNames {
		index := re.FindString(fileName)
		if index == "" {
			continue
		}
		var currentVideoMap *VideoInfo
		if _, ok := videoMap[index]; !ok {
			videoMap[index] = &VideoInfo{}
		}
		currentVideoMap = videoMap[index]
		if IsAAC(fileName) {
			currentVideoMap.AAC = fileName
		}
		if IsAVC(fileName) {
			currentVideoMap.TS = fileName
		}
	}
	return videoMap
}

func IsAAC(fileName string) bool {
	return strings.Contains(fileName, "aac")
}

func IsAVC(fileName string) bool {
	return strings.Contains(fileName, "avc")
}

func TrimAllSpace(files []fs.FileInfo) []error {
	wg := sync.WaitGroup{}
	errList := []error{}
	for _, file := range files {
		wg.Add(1)
		go func(f fs.FileInfo) {
			originalName := f.Name()
			hasSpace := strings.Contains(originalName, " ")
			if hasSpace {
				newName := strings.ReplaceAll(originalName, " ", "")
				err := os.Rename(originalName, newName)
				if err != nil {
					errList = append(errList, fmt.Errorf("檔案%s重新命名出錯", originalName))
				}
			}
			wg.Done()
		}(file)
	}

	wg.Wait()
	return errList
}

func GetAllFileNames(files []fs.FileInfo) []string {
	result := []string{}
	for _, file := range files {
		result = append(result, file.Name())
	}
	return result
}

func CombineTsAndAAC(tsFile, aacFile, outputFile string) error {
	cmd := exec.Command("ffmpeg", "-i", tsFile, "-i", aacFile, "-map", "0:V:0", "-map", "1:a:0", "-c", "copy", "-f", "mp4", "-movflags", "+faststart", outputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := string(output)
		return fmt.Errorf("ffmpeg command failed: %s", errorMsg)
	}
	return nil
}

func ShowMultiErrors(errList []error) {
	if len(errList) > 0 {
		var text string
		for _, err := range errList {
			text += err.Error() + "\n"
		}
		log.Fatal(text)
	}
}
