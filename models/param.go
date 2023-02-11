package models

type MultiFileUploadOptions struct {
	ThreadPoolSize int // 使用线程数
	PartSize       int
	CheckPoint     bool // 使用断点续传
}
