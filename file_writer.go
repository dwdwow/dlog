package dlog

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

const (
	maxQueueLen    int64 = 20
	maxLogFileSize int64 = 10485760 // 10 * 1024 * 1024
)

type SingleFileWriter struct {
	mux         sync.Mutex
	fileName    string
	file        *os.File
	muxQueueLen sync.Mutex
	queueLen    int64
	maxQueueLen int64
}

func NewSingleFileWriter(fileName string) *SingleFileWriter {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		panic("file writer: " + err.Error())
	}
	return &SingleFileWriter{
		mux:         sync.Mutex{},
		fileName:    fileName,
		file:        file,
		maxQueueLen: maxQueueLen,
	}
}

func (fw *SingleFileWriter) Write(bs []byte) (n int, err error) {
	fw.muxQueueLen.Lock()
	defer fw.muxQueueLen.Unlock()
	if fw.queueLen > fw.maxQueueLen {
		slog.Error("SingleFileWriter queue is too long", "len", fw.queueLen, "maxLen", fw.maxQueueLen)
		return 0, errors.New("file writer queue is too long")
	}
	fw.queueLen++
	go fw.write(bs)
	return len(bs), nil
}

func (fw *SingleFileWriter) write(bs []byte) {
	defer func() {
		fw.muxQueueLen.Lock()
		defer fw.muxQueueLen.Unlock()
		fw.queueLen--
	}()
	fw.mux.Lock()
	defer fw.mux.Unlock()
	_, err := fw.file.Write(bs)
	if err != nil {
		slog.Error("Can not write log to file", "err", err, "file", fw.fileName, "log", string(bs))
		return
	}
}

func (fw *SingleFileWriter) Close() error {
	return fw.file.Close()
}

type FileWriter struct {
	mux      sync.Mutex
	filePath string

	maxFileSize int64

	muxFile  sync.Mutex
	fileName string
	file     *os.File
	fileSize int64
	// the numbers of created file
	fileNum int

	muxQueueLen sync.Mutex
	queueLen    int64
	maxQueueLen int64

	logger *slog.Logger
}

// NewFileWriter
// If startFilePath is "", will new file.
// If maxFileSize is 0, it will be set to 10MB.
// If logger is nil, it will be set to os.Stdout.
// New file will be created with time.RFC3339Nano name automatically.
func NewFileWriter(fileParentPath string, startFileName string, maxFileSize int64, logger *slog.Logger) *FileWriter {
	writer := &FileWriter{
		filePath:    fileParentPath,
		maxQueueLen: maxQueueLen,
	}
	if maxFileSize != 0 {
		writer.maxFileSize = maxFileSize
	}
	if logger != nil {
		writer.logger = logger
	} else {
		writer.logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	if startFileName == "" {
		if err := writer.newFile(); err != nil {
			panic("new FileWriter: new file, " + err.Error())
		}
	} else {
		fileName := fileParentPath + "/" + startFileName
		file, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
		if err != nil {
			panic("new FileWriter: open file, " + err.Error())
		}
		info, err := os.Stat(fileName)
		if err != nil {
			panic("new FileWriter: get file status, " + err.Error())
		}
		writer.fileName = fileName
		writer.file = file
		writer.fileSize = info.Size()
		writer.fileNum = 1
	}
	return writer
}

func (aw *FileWriter) newFile() error {
	fileName := aw.filePath + "/" + time.Now().Format(time.RFC3339Nano)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return fmt.Errorf("auto split file writer: %w", err)
	}
	aw.fileName = fileName
	aw.file = file
	aw.fileSize = 0
	aw.fileNum++
	return nil
}

func (aw *FileWriter) Write(bs []byte) (n int, err error) {
	go func() {
		_, err := aw.write(bs)
		if err != nil {
			aw.logger.Error("Can not write", "err", err)
		}
	}()
	return len(bs), nil
}

func (aw *FileWriter) write(bs []byte) (int, error) {
	aw.mux.Lock()
	defer aw.mux.Unlock()

	// check queue
	aw.muxQueueLen.Lock()
	queueLen := aw.queueLen
	aw.muxQueueLen.Unlock()
	if queueLen > aw.maxQueueLen {
		return 0, fmt.Errorf("queue is full, %v/%v", aw.queueLen, aw.maxQueueLen)
	}
	aw.queueLen++
	defer func() {
		aw.muxQueueLen.Lock()
		aw.queueLen--
		aw.muxQueueLen.Unlock()
	}()

	aw.muxFile.Lock()
	defer aw.muxFile.Unlock()

	if aw.file == nil {
		return 0, fmt.Errorf("file is nil")
	}

	// check file size
	if aw.fileSize+int64(len(bs)) >= aw.maxFileSize {
		err := aw.newFile()
		if err != nil {
			return 0, fmt.Errorf("check file, %w", err)
		}
	}

	// write
	n, err := aw.file.Write(bs)
	if err != nil {
		return 0, fmt.Errorf("write %v to %v, %w", string(bs), aw.fileName, err)
	}
	aw.fileSize += int64(n)
	return n, err
}

func (aw *FileWriter) Close() error {
	aw.muxFile.Lock()
	defer aw.muxFile.Unlock()
	if aw.file == nil {
		return nil
	}
	return aw.file.Close()
}
