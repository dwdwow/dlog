package dlog

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestFileWriter_Write(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fileName := home + "/test/file_writer.log"

	_, err = os.Stat(fileName)
	if !os.IsNotExist(err) {
		err = os.Remove(fileName)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	writer := NewSingleFileWriter(fileName)
	text := "file writer test\n"
	n, err := writer.Write([]byte(text))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	time.Sleep(time.Second)
	t.Log("1: write to file successfully", n)
	err = writer.Close()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	writer = NewSingleFileWriter(fileName)
	n, err = writer.Write([]byte(text))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	time.Sleep(time.Second)
	t.Log("2: write to file successfully", n)
	err = writer.Close()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	t.Log("check text written")

	data, err := os.ReadFile(fileName)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if string(data) != "file writer test\nfile writer test\n" {
		t.Error("file text != text written", "file text", string(data))
		t.FailNow()
	}

	t.Log("file text is same as text written")
}

func TestAutoSplitFileWriter_write(t *testing.T) {
	const isSync = false
	const text = "hello, my logger\n"
	const totalLines = 19 // async: 19, maxQueueNum-1
	const maxFileSize = 110

	home, err := os.UserHomeDir()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// this dir should be empty
	fileParentPath := "/test/auto_split_file_logs"
	filePath := home + fileParentPath

	t.Log("Start to test in", filePath)

	entries, err := os.ReadDir(filePath)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if len(entries) != 0 {
		t.Log("Dir is not empty, try to clear.")
		for _, entry := range entries {
			err := os.Remove(filePath + "/" + entry.Name())
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
		}
	}

	writer := NewFileWriter(filePath, "", int64(maxFileSize), nil)

	var writerFunc func([]byte) (int, error)
	if isSync {
		writerFunc = writer.write
	} else {
		writerFunc = writer.Write
	}

	textSize := len(text)

	idealSizeOneFile := maxFileSize / textSize * textSize

	t.Log("Writing logs")
	var totalBytes int
	for i := 0; i < totalLines; i++ {
		n, err := writerFunc([]byte(text))
		if err != nil {
			t.Error("Can not write", "err", err)
			t.FailNow()
		}
		totalBytes += n
	}

	if !isSync {
		time.Sleep(time.Millisecond * time.Duration(totalLines))
	}

	lineNumOneFile := maxFileSize / textSize
	idealFileNum := totalLines / lineNumOneFile
	if totalLines%lineNumOneFile != 0 {
		idealFileNum++
	}

	t.Log("Checking logs")

	entries, err = os.ReadDir(filePath)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if len(entries) != idealFileNum {
		t.Error("File num is not correct", "fileNum", len(entries), "ideal written num", idealFileNum)
		t.FailNow()
	}

	idealLastFileSize := (totalLines % (maxFileSize / textSize)) * textSize

	for i, entry := range entries {
		data, err := os.ReadFile(filePath + "/" + entry.Name())
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		num := strings.Count(string(data), text)
		info, err := entry.Info()
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		ideaSize := int64(num * textSize)
		if info.Size() != ideaSize {
			t.Error("file size is not right", "fileSize", info.Size(), "idealSize", ideaSize)
			t.FailNow()
		}
		isLast := i == len(entries)-1
		if !isLast && info.Size() != int64(idealSizeOneFile) {
			t.Error("file size is not right", "fileSize", info.Size(), "idealSizeOneFile", idealSizeOneFile)
			t.FailNow()
		}
		if isLast && info.Size() != int64(idealLastFileSize) {
			t.Error("last file size is not right", "fileSize", info.Size(), "idealLastFileSize", idealLastFileSize)
			t.FailNow()
		}
	}

	t.Log("PASS")
}
