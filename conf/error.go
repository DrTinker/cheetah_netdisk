package conf

import "github.com/pkg/errors"

// 数据库错误
var DBSelectError = errors.New("DB select error")
var DBInsertError = errors.New("DB insert error")
var DBDeleteError = errors.New("DB delete error")
var DBUpdateError = errors.New("DB update error")
var DBNotFoundError = errors.New("DB not found error")
var DBPageOutOfRangeError = errors.New("The given page out of range")

// 数据处理错误
var JsonError = errors.New("JSON parse error")

// 参数校验
var ParamError = errors.New("Param check error")
var InvaildFileHashError = errors.New("File hash invaild")

// 文件上传
var VolumeError = errors.New("User volume run out")
var FilePathError = errors.New("Invaild file path")
var FileExistError = errors.New("File has already in this user's space")

// 分享
var FileDeletedError = errors.New("File has been deleted")

// 分块上传
var ChunkMissError = errors.New("Slice misses error")

// 文件操作
var NameRepeatError = errors.New("File name exist")
var ExtChangeError = errors.New("File ext name should not be changed")

// 文件删除
var EmptyDeleteError = errors.New("No such file to be deleted")

// 文件异步上传cos
var MQConnectionClosedError = errors.New("MQ connection has closed by error, please check")

// map
var MapNotHasError = errors.New("Map do not have such key")

// ws
var SocketNilError = errors.New("This socket is nil")

// 文件下载
var InvaildOwnerError = errors.New("User is not the owner of such file")

var TransFinishError = errors.New("This trans is finished")
var TransBrokenError = errors.New("This trans meta data missed")
