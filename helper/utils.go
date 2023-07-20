package helper

import (
	"NetDesk/conf"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

func GenRandCode() string {
	s := "1234567890QWERTYUIOPASDFGHJKLZXCVBNM"
	code := ""
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < conf.Code_Len; i++ {
		flag := rand.Intn(len(s))
		code += string(s[flag])
	}

	return code
}

// 数组切分
func ArrayInGroupsOf(arr []byte, num int64) [][]byte {
	max := int64(len(arr))
	//判断数组大小是否小于等于指定分割大小的值，是则把原数组放入二维数组返回
	if max <= num {
		return [][]byte{arr}
	}
	//获取应该数组分割为多少份
	var quantity int64
	if max%num == 0 {
		quantity = max / num
	} else {
		quantity = (max / num) + 1
	}
	//声明分割好的二维数组
	var segments = make([][]byte, 0)
	//声明分割数组的截止下标
	var start, end, i int64
	for i = 1; i <= quantity; i++ {
		end = i * num
		if i != quantity {
			segments = append(segments, arr[start:end])
		} else {
			segments = append(segments, arr[start:])
		}
		start = i * num
	}
	return segments
}

// TODO 更加严格的路径有效性检测
// 切分文件路径获取名称和扩展名
func SplitFilePath(path string) (name, ext string, err error) {
	// 判断空字符串，且必须有/
	if len(path) == 0 || !strings.Contains(path, "/") {
		return "", "", conf.FilePathError
	}
	// 切分
	names := strings.Split(path, "/")
	if len(names) == 0 {
		return "", "", conf.FilePathError
	}
	fullName := names[len(names)-1]
	name, ext, err = SplitFileFullName(fullName)

	return name, ext, err
}

// 切分name.ext -> name ext
func SplitFileFullName(fullName string) (name string, ext string, err error) {
	// TODO 文件夹判断不准确 com.example.aaa也可是文件夹名称，需要前端传入ext
	// 是文件夹则增加默认文件夹扩展名
	if !strings.Contains(fullName, ".") {
		fullName = fullName + "." + conf.Folder_Default_EXT
	}
	if len(fullName) == 0 {
		return "", "", conf.FilePathError
	}
	part := strings.Split(fullName, ".")
	if len(part) == 0 {
		return "", "", conf.FilePathError
	}
	name = ""
	ext = part[len(part)-1]
	for i := 0; i < len(part)-1; i++ {
		name += part[i]
	}
	return name, ext, nil
}

// 计算md5值
// mod 0 读取path路径下的文件计算md5值；1 计算data的md5值
func CountMD5(path string, data []byte, mod int) string {
	if mod == 0 {
		fd, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer fd.Close()
		data, _ := ioutil.ReadAll(fd)
		return fmt.Sprintf("%x", md5.Sum(data))
	} else {
		return fmt.Sprintf("%x", md5.Sum(data))
	}
}
