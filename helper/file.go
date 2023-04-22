package helper

import (
	"io"
	"io/ioutil"
	"os"
)

// 打开本地文件
// OpenFile 判断文件是否存在  存在则OpenFile 不存在则Create
func OpenFile(filename string) (*os.File, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return os.Create(filename) //创建文件
	}
	return os.OpenFile(filename, os.O_APPEND, 0666) //打开文件
}

func WriteFile(path string, data io.Reader) error {
	body, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, body, 0666) //写入文件(字节数组)
	if err != nil {
		return err
	}
	return nil
}
