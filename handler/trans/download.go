package trans

import (
	"NetDesk/client"
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
	"NetDesk/service"
	"errors"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

// 客户端从服务端下载完整文件
func DownloadFileHandler(c *gin.Context) {
	// 获取user_file_uuid
	user_file_uuid := c.Query(conf.File_Uuid_Key)
	if user_file_uuid == "" {
		log.Error("DownloadFileHandler err: user file uuid rmpty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 用户本地存储路径
	local_path := c.Query(conf.File_Local_Path_Key)
	if local_path == "" {
		log.Error("DownloadFileHandler invaild local path")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 云存储路径
	remotePath := c.Query(conf.File_Remote_Path_Key)
	if remotePath == "" {
		log.Error("InitDownloadHandler empty remote path")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 用户本地存储路径
	parent_uuid := c.Query(conf.Folder_Uuid_Key)
	if parent_uuid == "" {
		log.Error("DownloadFileHandler invaild parent")
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
		log.Error("DownloadFileHandler uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 生成ID
	downloadID := helper.GenDownloadID(user_uuid, user_file_uuid)
	param := &models.DownloadObjectParam{
		DownloadID:     downloadID,
		User_File_Uuid: user_file_uuid,
		User_Uuid:      user_uuid,
		Parent_Uuid:    parent_uuid,
		LocalPath:      local_path,
		RemotePath:     remotePath,
	}

	// 调用service获取COS token
	fileToken, err := service.DownloadTotal(param)
	if fileToken == "" || err != nil {
		log.Error("DownloadFileHandler download err ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}

	// 返回签名
	log.Info("DownloadFileHandler success: ", user_file_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":        conf.HTTP_SUCCESS_CODE,
		"msg":         conf.SUCCESS_RESP_MESSAGE,
		"file_token":  fileToken,
		"download_id": downloadID,
	})
}

// websocket实现下载
// 弃用
func DownloadFileBySocketHandler(c *gin.Context) {
	// 获取user_file_uuid
	user_file_uuid := c.Query(conf.File_Uuid_Key)
	if user_file_uuid == "" {
		log.Error("DownloadFileBySocketHandler err: user file uuid rmpty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 用户本地存储路径
	local_path := c.Query(conf.File_Local_Path_Key)
	if local_path == "" {
		log.Error("DownloadFileBySocketHandler invaild local path")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 用户本地存储路径
	parent_uuid := c.Query(conf.Folder_Uuid_Key)
	if parent_uuid == "" {
		log.Error("DownloadFileBySocketHandler invaild parent")
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
		log.Error("DownloadFileBySocketHandler uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// uploadID用于断点续传
	downloadID := c.Query(conf.File_Download_ID_Key)
	continueFlag := true
	if downloadID == "" {
		downloadID = helper.GenDownloadID(user_uuid, user_file_uuid)
		continueFlag = false
	}

	param := &models.DownloadObjectParam{
		Req:            *c.Request,
		Resp:           c.Writer,
		DownloadID:     downloadID,
		User_File_Uuid: user_file_uuid,
		User_Uuid:      user_uuid,
		Parent_Uuid:    parent_uuid,
		LocalPath:      local_path,
		Continue:       continueFlag,
	}
	// 建立ws连接
	err := client.GetSocketClient().AddConn(param.Resp, &param.Req, downloadID)
	if err != nil {
		log.Error("DownloadFileBySocketHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 之后的通信通过ws完成
	res, err := service.InitDownload(param)
	if err != nil {
		log.Error("InitDownloadHandler err: ", err)
		client.GetSocketClient().SendMsg(downloadID, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  err,
		})
		// 关闭通道
		client.GetSocketClient().DeleteConn(downloadID)
		return
	}
	// 初始化成功返回数据
	client.GetSocketClient().SendMsg(downloadID, gin.H{
		"code":        conf.HTTP_SUCCESS_CODE,
		"msg":         conf.SUCCESS_RESP_MESSAGE,
		"download_id": res.DownloadID,
		"chunk_size":  conf.File_Part_Size_Max,
		"chunk_count": res.ChunkCount,
		"chunk_list":  res.ChunkList,
		"hash":        res.Hash,
	})
}

// param: file_uuid: 文件的uer_file_uuid
//		  local_path: 存储文件的客户端本地路径
//		  parent_uuid: 父级文件夹的user_file_uuid
// 		  download_id: 下载ID
// return: download_id: 本次分块上传唯一标识
//		   chunk_size: 分块大小
// 		   chunk_count: 分块数量
// 		   chunk_list: 已经上传的分块列表
// 		   hash: 文件的hash值
func InitDownloadHandler(c *gin.Context) {
	// 获取user_file_uuid
	user_file_uuid := c.Query(conf.File_Uuid_Key)
	if user_file_uuid == "" {
		log.Error("InitDownloadHandler err: user file uuid rmpty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 用户本地存储路径
	local_path := c.Query(conf.File_Local_Path_Key)
	if local_path == "" {
		log.Error("InitDownloadHandler invaild local path")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 云存储路径
	remotePath := c.Query(conf.File_Remote_Path_Key)
	if remotePath == "" {
		log.Error("InitDownloadHandler empty remote path")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 用户本地存储路径
	parent_uuid := c.Query(conf.Folder_Uuid_Key)
	if parent_uuid == "" {
		log.Error("InitDownloadHandler invaild parent")
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
		log.Error("InitDownloadHandler uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// downloadID用于断点续传
	downloadID := c.Query(conf.File_Download_ID_Key)
	continueFlag := true
	if downloadID == "" {
		downloadID = helper.GenDownloadID(user_uuid, user_file_uuid)
		continueFlag = false
	}
	param := &models.DownloadObjectParam{
		DownloadID:     downloadID,
		User_File_Uuid: user_file_uuid,
		User_Uuid:      user_uuid,
		Parent_Uuid:    parent_uuid,
		LocalPath:      local_path,
		RemotePath:     remotePath,
		Continue:       continueFlag,
	}
	res, err := service.InitDownloadCOS(param)
	if err != nil {
		log.Error("InitDownloadHandler err: ", err)
		if errors.Is(err, conf.InvaildOwnerError) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": conf.ERROR_FILE_OWNER_CODE,
				"msg":  conf.SERVER_ERROR_MSG,
			})
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}

	// 成功
	log.Info("InitDownloadHandler success: ", res.DownloadID)
	c.JSON(http.StatusOK, gin.H{
		"code":        conf.HTTP_SUCCESS_CODE,
		"msg":         conf.SUCCESS_RESP_MESSAGE,
		"download_id": res.DownloadID,
		"chunk_size":  conf.File_Part_Size_Max,
		"chunk_count": res.ChunkCount,
		"chunk_list":  res.ChunkList,
		"hash":        res.Hash,
		"url":         res.Url,
	})
}

// 轮询接口，查看服务端请求下载的文件是否准备好
func CheckDownloadReadyHandler(c *gin.Context) {
	// 获取dowanloadID
	downloadID := c.Query(conf.File_Download_ID_Key)
	if downloadID == "" {
		log.Error("CheckDownloadReadyHandler err: downloadID empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 调用service
	flag, err := service.CheckDownloadReady(downloadID)
	if err != nil {
		log.Error("CheckDownloadReadyHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 返回
	log.Info("CheckDownloadReadyHandler success: ", downloadID)
	c.JSON(http.StatusOK, gin.H{
		"code":  conf.HTTP_SUCCESS_CODE,
		"msg":   conf.SUCCESS_RESP_MESSAGE,
		"ready": flag,
	})
}

// 客户端从服务端下载文件分片
func DownloadPartHandler(c *gin.Context) {
	// 获取dowanloadID
	downloadID := c.Query(conf.File_Download_ID_Key)
	if downloadID == "" {
		log.Error("DownloadPartHandler err: downloadID empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 调用service下载COS文件至tmp
	// 获取chunknum
	chunkNum := c.Query(conf.File_Chunk_Num_Key)
	num, err := strconv.Atoi(chunkNum)
	if err != nil {
		log.Error("DownloadPartHandler invaild chunk number")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 调用service
	err = service.DownloadPartCOS(downloadID, num)
	if err != nil {
		log.Error("DownloadPartHandler empty chunk err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}

	// 写入文件至body
	// fileName := chunkNum
	// c.Header("Content-Type", "application/octet-stream")
	// c.Header("Content-Disposition", "attachment; filename="+fileName)
	// c.Header("Content-Transfer-Encoding", "binary")
	// c.Header("Accept-Length", fmt.Sprintf("%d", len(content)))
	// c.Header("Cache-Control", "no-cache")

	// c.Writer.Write(content)
	// log.Info("DownloadPartHandler success: ", downloadID)
	log.Info("DownloadPartHandler success: ", downloadID, " ", chunkNum)
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
	})
}

func CompleteDownloadPartHandler(c *gin.Context) {
	// 获取dowanloadID
	downloadID := c.Query(conf.File_Download_ID_Key)
	if downloadID == "" {
		log.Error("CompleteDownloadPartHandler err: downloadID empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 调用service
	err := service.CompleteDownloadPartCOS(downloadID)
	if err != nil {
		log.Error("CompleteDownloadPartHandler empty chunk err: ", err)
		if errors.Is(err, conf.ChunkMissError) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": conf.CHUNK_MISS_CODE,
				"msg":  conf.SERVER_ERROR_MSG,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}

	log.Info("CompleteDownloadPartHandler success: ", downloadID)
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
	})
}
