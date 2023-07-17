package helper

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

// 打开本地文件
// OpenFile 判断文件是否存在  存在则OpenFile 不存在则Create
func OpenFile(filename string) (*os.File, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return os.Create(filename) //创建文件
	}
	return os.OpenFile(filename, os.O_APPEND, 0666) //打开文件
}

func WriteFile(path, name string, data []byte) error {
	// 目录不存在则创建目录
	os.MkdirAll(path, os.ModePerm)
	// 文件存在时清空文件再写
	des := fmt.Sprintf("%s/%s", path, name)
	err := ioutil.WriteFile(des, data, 0666) //写入文件(字节数组)
	if err != nil {
		return err
	}
	return nil
}

// 删除文件
func DelFile(path string) error {
	err := os.Remove(path)
	return err
}

// 返回目录下全部文件列表，按照数字大小排序
func ReadDir(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	res := make([]string, len(list))
	for _, f := range list {
		n, _ := strconv.Atoi(f.Name())
		// 分片从1开始
		res[n-1] = f.Name()
	}
	return res, nil
}

// 将src路径下的全部文件合成一个文件写入des
// src example: /tmp/aaa/
func MergeFile(src, des string) (*os.File, error) {
	// 打开目标文件，不存在则创建
	desFile, err := OpenFile(des)
	if err != nil {
		return nil, err
	}
	defer desFile.Close()
	writer := bufio.NewWriter(desFile)
	// 读取src路径下全部文件
	files, err := ReadDir(src)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		f, err := OpenFile(src + f)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		writer.Write(data)
		//Flush将缓存的文件真正写入到文件中
		writer.Flush()
		f.Close()
	}

	return desFile, nil
}

// 文件是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { //文件或者目录存在
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func RemoveDir(path string) error {
	exist, err := PathExists(path)
	if err != nil {
		return err
	}
	if exist {
		err = os.RemoveAll(path)
		if err != nil {
			return err
		}
	}
	return nil
}
