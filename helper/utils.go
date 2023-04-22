package helper

import (
	"NetDesk/conf"
	"math/rand"
	"os"
	"strings"
	"time"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

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
	// 判断时文件还是文件夹
	isFolder := strings.HasSuffix(path, "/")
	// 切分
	names := strings.Split(path, "/")
	if len(names) == 0 {
		return "", "", conf.FilePathError
	}
	name = names[len(names)-1]
	// 是文件
	if !isFolder {
		part := strings.Split(name, ".")
		name = part[0]
		ext = part[1]
	} else {
		// 是文件夹
		// 文件夹切分后最后一个是空字符串
		name = names[len(names)-2]
		ext = conf.Folder_Default_EXT
	}

	return name, ext, nil
}

// 切分name.ext -> name ext
func SplitFileFullName(fullName string) (name string, ext string, err error) {
	// 是文件夹则增加默认文件夹扩展名
	if !strings.Contains(fullName, ".") {
		fullName = fullName + "." + conf.Folder_Default_EXT
	}
	if len(fullName) == 0 {
		return "", "", conf.FilePathError
	}
	part := strings.Split(fullName, ".")
	if len(part) != 2 {
		return "", "", conf.FilePathError
	}
	name = part[0]
	ext = part[1]
	return name, ext, nil
}
