package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"merge_aac_and_ts/utils"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/antlabs/strsim"
)

type VideoInfo struct {
	TS  string
	AAC string
}

const outputFolderName = "output"

var extNames = []string{"ts", "aac", "avc", "mp4", "mp3"}
var extRegex = regexp.MustCompile(`\.(ts|aac|avc|mp3|mp4)$`)

func main() {
	defer func() {
		log.Println("----------------------")
		log.Println("10秒後該程序會自動關閉)")
		time.Sleep(time.Second * 10)
	}()
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

	allFileNames := GetAllFileNames(files)
	shouldCombineFiles := FilerFileNames(allFileNames, func(fileName string) bool {
		for _, extName := range extNames {
			if strings.HasSuffix(fileName, "."+extName) {
				return true
			}
		}
		return false
	})
	if len(shouldCombineFiles) == 0 {
		log.Println("資料夾中找不到可合併的檔案格式。格式需要是以下幾種:\n-ts檔案\n-aac檔案\n-mp4檔案\n-mp3檔案")
		return
	}

	fileMap := GroupVideoInfo(shouldCombineFiles)
	// for key, value := range fileMap {
	// 	log.Println("---")
	// 	log.Println(key)
	// 	log.Println(value.AAC, value.TS)
	// }
	// 組合所有ts和aac成mp4檔案
	errList = CombineAllFiles(fileMap)
	ShowMultiErrors(errList)

	log.Println("轉換成功!! 請查看output資料夾\n\n")
}

func FilerFileNames(fileNames []string, check func(string) bool) []string {
	resultList := make([]string, 0)
	for _, fileName := range fileNames {
		if check(fileName) {
			resultList = append(resultList, fileName)
		}
	}
	return resultList
}

func CombineAllFiles(fileMap map[string]*VideoInfo) []error {
	errList := []error{}
	wg := sync.WaitGroup{}
	for _, value := range fileMap {
		wg.Add(1)
		go func(v *VideoInfo) {
			defer func() {
				wg.Done()
			}()

			aacFile := v.AAC
			tsFile := v.TS
			mp4FileName, err := GenerateMp4FileName(tsFile)
			if err != nil {
				errList = append(errList, err)
				return
			}
			if err = CombineTsAndAAC(tsFile, aacFile, mp4FileName); err != nil {
				errList = append(errList, err)
			}

			os.Rename(mp4FileName, outputFolderName+"/"+mp4FileName)
		}(value)
	}
	wg.Wait()
	return errList
}

func GenerateMp4FileName(fileName string) (string, error) {
	if extRegex.MatchString(fileName) {
		return extRegex.ReplaceAllString(fileName, "_output.mp4"), nil
	}
	return "", fmt.Errorf("%s 無法正確更新成.mp4檔名!!請確認檔案是否為.ts,aac,avc,mp3,mp4格式之一\n", fileName)
}

func GroupVideoInfo(fileNames []string) map[string]*VideoInfo {
	videoMap := make(map[string]*VideoInfo)
	noNeedCheckIndexList := []int{}
	for i := 0; i < len(fileNames); i++ {
		// 確認是否已經有配對
		shouldContinue := false
		for _, index := range noNeedCheckIndexList {
			if index == i {
				shouldContinue = true
			}
		}
		if shouldContinue {
			continue
		}

		scoreList := []float64{}
		for j := 0; j < len(fileNames); j++ {
			if i == j {
				// 同一個不做比較
				scoreList = append(scoreList, -1)
				continue
			}
			score := strsim.Compare(fileNames[i], fileNames[j])
			scoreList = append(scoreList, score)
		}
		maxScore := utils.Highest(scoreList)
		anotherIndex := 0
		for i := 0; i < len(scoreList); i++ {
			if scoreList[i] == maxScore {
				anotherIndex = i
			}
		}

		noNeedCheckIndexList = append(noNeedCheckIndexList, anotherIndex)

		currentKey := strconv.Itoa(i)
		var currentVideoMap *VideoInfo
		if _, ok := videoMap[currentKey]; !ok {
			videoMap[currentKey] = &VideoInfo{}
		}
		currentVideoMap = videoMap[currentKey]
		if IsAAC(fileNames[i]) {
			currentVideoMap.AAC = fileNames[i]
			if IsAVC(fileNames[anotherIndex]) {
				currentVideoMap.TS = fileNames[anotherIndex]
			} else {
				log.Fatalf("%s找不到對應的avc影片檔案!! 請確認", fileNames[i])
			}
		}
		if IsAVC(fileNames[i]) {
			currentVideoMap.TS = fileNames[i]
			if IsAAC(fileNames[anotherIndex]) {
				currentVideoMap.AAC = fileNames[anotherIndex]
			} else {
				log.Fatalf("%s找不到對應的aac音訊檔案!! 請確認", fileNames[i])
			}
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
