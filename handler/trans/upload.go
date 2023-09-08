package trans

import (
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
	"NetDesk/service"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 单个上传文件
// TODO 文件
func UploadHandler(c *gin.Context) {
	// 检查文件有效性时已经读取过文件，从ctx中获取文件
	v, exist := c.Get(conf.File_Form_Key)
	if !exist {
		log.Error("UploadHandler err: file data invaild")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	file, ok := v.([]byte)
	if !ok {
		log.Error("UploadHandler err: file data invaild")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 检查有效性性的中间件已经读取过了，因此从ctx中获取
	hash := c.GetString(conf.File_Hash_Key)
	// 文件夹名称
	fileName := c.PostForm(conf.File_Name_Key)
	name, ext, err := helper.SplitFileFullName(fileName)
	if err != nil {
		log.Error("UploadHandler invaild file name")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	fileKey := helper.GenFileKey(hash, ext)

	// 前端传入uuid后端查询id
	user_file_uuid_parent := c.PostForm(conf.Folder_Uuid_Key)
	if fileKey == "" || user_file_uuid_parent == "" {
		log.Error("UploadHandler empty file key")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	remotePath := c.PostForm(conf.File_Remote_Path_Key)
	if remotePath == "" {
		log.Error("UploadHandler empty remote path")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取用户ID
	var user_uuid string
	if idstr, f := c.Get(conf.UserID); f {
		user_uuid = helper.Strval(idstr)
	}
	if user_uuid == "" {
		log.Error("UploadHandler uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 生成ID
	file_uuid := helper.GenFid(fileKey)
	user_file_uuid := helper.GenUserFid(user_uuid, fileKey)
	uploadID := helper.GenUploadID(user_uuid, hash)
	// 打包参数
	param := &models.UploadObjectParams{
		UploadID:       uploadID,
		FileKey:        fileKey,
		User_Uuid:      user_uuid,
		Parent:         user_file_uuid_parent,
		RemotePath:     remotePath,
		Hash:           hash,
		Size:           len(file),
		Name:           name,
		Ext:            ext,
		File_Uuid:      file_uuid,
		User_File_Uuid: user_file_uuid,
	}
	// 查看是否秒传
	flag, err := service.QuickUpload(param)
	if err != nil {
		log.Error("UploadHandler quick upload err: ", err)
		// 同一个人上传同一个文件，返回错误前端走复制文件接口
		if err == conf.FileExistError {
			c.JSON(http.StatusOK, gin.H{
				"code": conf.FILE_EXIST_CODE,
				"msg":  conf.FILE_EXIST_MESSAGE,
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  fmt.Sprintf(conf.UPLOAD_FAIL_MESSAGE, fileName),
		})
		return
	}
	if flag {
		// 秒传直接返回
		log.Info("UploadHandler success: ", user_file_uuid)
		c.JSON(http.StatusOK, gin.H{
			"code":      conf.QUICK_UPLOAD_CODE,
			"msg":       fmt.Sprintf(conf.UPLOAD_SUCCESS_MESSAGE),
			"upload_id": uploadID,
		})
		return
	}
	// 非秒传
	err = service.UploadFileByStream(param, file)
	if err != nil {
		log.Error("UploadHandler upload err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_UPLOAD_CODE,
			"msg":  fmt.Sprintf(conf.UPLOAD_FAIL_MESSAGE, fileName),
		})
		return
	}

	log.Info("UploadHandler success: ", user_file_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":      conf.HTTP_SUCCESS_CODE,
		"msg":       fmt.Sprintf(conf.UPLOAD_SUCCESS_MESSAGE),
		"upload_id": uploadID,
	})
}

// param: hash: 文件md5值
//		  size: 文件大小
//		  parent_uuid: 父级文件夹的user_file_uuid
// 		  name: 文件全称 eg: aaa.txt
//		  local_path: 文件用户本地存储路径
// return: uploadID: 本次分块上传唯一标识
//		   chunk_size: 分块大小
// 		   chunk_count: 分块数量
// 		   chunk_list: 已经上传的分块列表
// user_file_uuid & file_uuid在这里生成
func InitUploadHandler(c *gin.Context) {
	// 获取文件大小
	fileSize := c.PostForm(conf.File_Size_Key)
	size, err := strconv.Atoi(fileSize)
	if err != nil {
		log.Error("InitUploadPartHandler invaild file size")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取用户ID
	var user_uuid string
	if idstr, f := c.Get(conf.UserID); f {
		user_uuid = helper.Strval(idstr)
	}
	if user_uuid == "" {
		log.Error("InitUploadHandler uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取文件哈希值
	hash := c.PostForm(conf.File_Hash_Key)
	if hash == "" {
		log.Error("InitUploadHandler invaild file hash")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 文件名称
	fileName := c.PostForm(conf.File_Name_Key)
	name, ext, err := helper.SplitFileFullName(fileName)
	if err != nil {
		log.Error("InitUploadHandler invaild file name")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 用户本地存储路径
	localPath := c.PostForm(conf.File_Local_Path_Key)
	if localPath == "" {
		log.Error("InitUploadHandler invaild local path")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 云存储路径
	remotePath := c.PostForm(conf.File_Remote_Path_Key)
	if remotePath == "" {
		log.Error("InitUploadHandler empty remote path")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// uploadID用于断点续传
	uploadID := c.PostForm(conf.File_Upload_ID_Key)
	fileKey := helper.GenFileKey(hash, ext)
	// 上传目录uuid
	user_file_uuid_parent := c.PostForm(conf.Folder_Uuid_Key)
	if fileKey == "" || user_file_uuid_parent == "" {
		log.Error("InitUploadHandler empty file key")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 生成ID
	file_uuid := helper.GenFid(fileKey)
	user_file_uuid := helper.GenUserFid(user_uuid, fileKey)
	// 打包参数
	param := &models.UploadObjectParams{
		UploadID:       uploadID,
		FileKey:        fileKey,
		LocalPath:      localPath,
		RemotePath:     remotePath,
		User_Uuid:      user_uuid,
		Parent:         user_file_uuid_parent,
		Hash:           hash,
		Size:           size,
		Name:           name,
		Ext:            ext,
		File_Uuid:      file_uuid,
		User_File_Uuid: user_file_uuid,
	}

	// 调用service层
	info, err := service.InitUpload(param)
	if err != nil {
		log.Error("InitUploadHandler quick upload err: ", err)
		// 同一个人上传同一个文件，返回错误前端走复制文件接口
		if err == conf.FileExistError {
			c.JSON(http.StatusOK, gin.H{
				"code": conf.FILE_EXIST_CODE,
				"msg":  conf.FILE_EXIST_MESSAGE,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	if info.Quick {
		// 秒传直接返回
		log.Info("InitUploadHandler success, file exist: ", user_file_uuid)
		c.JSON(http.StatusOK, gin.H{
			"code":      conf.QUICK_UPLOAD_CODE,
			"msg":       fmt.Sprintf(conf.UPLOAD_SUCCESS_MESSAGE),
			"upload_id": info.UploadID,
		})
		return
	}
	// 成功
	log.Info("InitUploadHandler success: ", info.UploadID)
	c.JSON(http.StatusOK, gin.H{
		"code":        conf.HTTP_SUCCESS_CODE,
		"msg":         conf.SUCCESS_RESP_MESSAGE,
		"upload_id":   info.UploadID,
		"chunk_size":  conf.File_Part_Size_Max,
		"chunk_count": info.ChunkCount,
		"chunk_list":  info.ChunkList,
	})
}

// param: files: 文件
//		  upload_id: 文件分块上传唯一ID
//		  chunk_num: 分块编号
// return: chunk_num: 分块编号
//		   upload_id: 文件分块上传唯一ID
func UploadPartHandler(c *gin.Context) {
	// 检查文件有效性时已经读取过文件，从ctx中获取文件
	file, err := c.FormFile(conf.File_Form_Key)
	if err != nil {
		log.Error("UploadPartHandler err: file data invaild")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	fd, err := file.Open()
	if err != nil {
		log.Error("UploadPartHandler file open err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		c.Abort()
		return
	}
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Error("FileCheck file open err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		c.Abort()
		return
	}
	// 获取chunknum
	chunkNum := c.PostForm(conf.File_Chunk_Num_Key)
	num, err := strconv.Atoi(chunkNum)
	if err != nil {
		log.Error("UploadPartHandler invaild chunk number")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取uploadID
	uploadID := c.PostForm(conf.File_Upload_ID_Key)
	if uploadID == "" {
		log.Error("UploadPartHandler invaild upload id")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 调用service
	err = service.UploadPart(uploadID, num, data)
	if err != nil {
		log.Error("UploadPartHandler service error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_UPLOAD_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 成功
	log.Info("UploadPartHandler success: uploadID: ", uploadID, " num: ", num)
	c.JSON(http.StatusOK, gin.H{
		"code":      conf.HTTP_SUCCESS_CODE,
		"msg":       conf.SUCCESS_RESP_MESSAGE,
		"upload_id": uploadID,
		"chunk_num": num,
	})
}

// param: upload_id: 文件分块上传唯一ID
// return: file_id: user_file_id
func CompleteUploadPartHandler(c *gin.Context) {
	// 获取uploadID
	uploadID := c.PostForm(conf.File_Upload_ID_Key)
	if uploadID == "" {
		log.Error("CompleteUploadPartHandler invaild upload id")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 调用service
	param, path, err := service.CompleteUploadPart(uploadID)
	if err != nil {
		log.Error("CompleteUploadPartHandler service error: ", err)
		if err == conf.InvaildFileHashError {
			c.JSON(http.StatusOK, gin.H{
				"code": conf.ERROR_FILE_HASH_CODE,
				"msg":  err.Error(),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_UPLOAD_CODE,
			"msg":  err.Error(),
		})
		return
	}
	// 上传cos
	err = service.UploadFileByPath(param, path)
	if err != nil {
		log.Error("CompleteUploadPartHandler service error: ", err)
		if errors.Is(err, conf.ChunkMissError) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": conf.CHUNK_MISS_CODE,
				"msg":  conf.SERVER_ERROR_MSG,
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  fmt.Sprintf(conf.UPLOAD_FAIL_MESSAGE, param.Name+"."+param.Ext),
		})
		return
	}
	// 成功
	log.Info("UploadPartHandler success: : ", param.User_File_Uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.HTTP_SUCCESS_CODE,
		"msg":     conf.SUCCESS_RESP_MESSAGE,
		"file_id": param.User_File_Uuid,
	})
}
