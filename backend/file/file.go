package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/libs4go/errors"

	"github.com/libs4go/scf4go"
	"github.com/libs4go/slf4go"
)

func getSize(path string) int64 {
	fileInfo, err := os.Stat(path)
	if err != nil {
		println(fmt.Sprintf("get file %s size error", path))
		return 0
	}
	fileSize := fileInfo.Size() //获取size

	return fileSize
}

type filebackendImpl struct {
	Path               string        `json:"path"`
	Name               string        `json:"name"`
	Extension          string        `json:"extension"`
	MaxSize            int64         `json:"maxsize"`
	RotationTime       time.Duration `json:"rotation_time"`
	TimestampFormatter string        `json:"timestamp"`
	currentPath        string
	currentTimestamp   time.Time
}

func new() *filebackendImpl {
	impl := &filebackendImpl{
		Path:               "./",
		Name:               "unknown",
		Extension:          "log",
		MaxSize:            1024 * 1024 * 10,
		RotationTime:       time.Hour * 24,
		TimestampFormatter: "2006-01-02T15:04:05Z07:00",
	}

	impl.newFilePath()

	return impl
}

func (filebackend *filebackendImpl) Send(entry *slf4go.EventEntry) {
	buff, err := json.Marshal(entry)
	if err != nil {
		println("marshal event entry error")
		return
	}

	file, err := os.OpenFile(filebackend.currentPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		println(fmt.Sprintf("open file %s error %s", filebackend.currentPath, err))
		return
	}

	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%s\n", string(buff)))

	if err != nil {
		println(fmt.Sprintf("write to file %s error %s", filebackend.currentPath, err))
		return
	}

	info, err := file.Stat()

	if err != nil {
		println(fmt.Sprintf("get file info %s error %s", filebackend.currentPath, err))
		return
	}

	if info.Size() > filebackend.MaxSize {
		filebackend.newFilePath()
		return
	}

	if filebackend.currentTimestamp.Add(filebackend.RotationTime).Unix() < time.Now().Unix() {
		println(filebackend.currentTimestamp.Add(filebackend.RotationTime).Unix(), time.Now().Unix())
		filebackend.newFilePath()
		return
	}
}

func (filebackend *filebackendImpl) Sync() {

}

func (filebackend *filebackendImpl) Config(config scf4go.Config) error {

	filebackend.Path = config.Get("path").String("./")
	filebackend.Name = config.Get("name").String("unknown")
	filebackend.Extension = config.Get("extension").String("log")
	filebackend.MaxSize = int64(config.Get("maxsize").Int(1024 * 1024 * 10))
	filebackend.RotationTime = config.Get("rotation_time").Duration(time.Hour * 24)

	return filebackend.checkConfig()
}

func (filebackend *filebackendImpl) checkConfig() error {
	err := os.MkdirAll(filebackend.Path, 0755)

	if err != nil {
		return errors.Wrap(err, "create dir %s error", filebackend.Path)
	}

	var lastFileTimestamp *time.Time
	var lastFilePath string

	filepath.Walk(filebackend.Path, func(path string, info os.FileInfo, err error) error {

		fileName := filepath.Base(path)

		fileName = strings.TrimSuffix(fileName, "."+strings.TrimPrefix(filebackend.Extension, "."))

		if strings.HasPrefix(fileName, filebackend.Name+"-") {

			suffix := strings.TrimPrefix(fileName, filebackend.Name+"-")

			timestamp, err := time.Parse(filebackend.TimestampFormatter, suffix)

			if err != nil {
				println(fmt.Sprintf("parse %s timestamp error skipped: %s", path, err))
				return nil
			}

			if lastFileTimestamp != nil && lastFileTimestamp.After(timestamp) {
				return nil
			}

			if info.Size() > filebackend.MaxSize {
				return nil
			}

			lastFilePath = path
			lastFileTimestamp = &timestamp
		}

		return nil
	})

	if lastFileTimestamp == nil {
		filebackend.newFilePath()
		return nil
	}

	if lastFileTimestamp.Add(filebackend.RotationTime).Unix() < time.Now().Unix() {
		filebackend.newFilePath()
		return nil
	}

	filebackend.currentPath = lastFilePath
	filebackend.currentTimestamp = *lastFileTimestamp

	return nil
}

func (filebackend *filebackendImpl) newFilePath() {
	filebackend.currentTimestamp = time.Now()
	fileName := fmt.Sprintf("%s-%s.%s",
		filebackend.Name,
		filebackend.currentTimestamp.Format(filebackend.TimestampFormatter),
		filebackend.Extension)

	filebackend.currentPath = filepath.Join(filebackend.Path, fileName)
}

func init() {
	slf4go.RegisterBackend("file", new())
}
