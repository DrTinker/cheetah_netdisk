package logic

import (
	"NetDesk/common/client"
	"NetDesk/common/conf"
	"NetDesk/common/helper"
	"NetDesk/common/models"
	"fmt"

	"github.com/pkg/errors"

	"NetDesk/service/user/proto/user"
)

type UserLogic struct {
	dbClient client.DBClient
}

func NewUserLogic() *UserLogic {
	logic := &UserLogic{
		dbClient: client.GetDBClient(),
	}
	return logic
}

func (u *UserLogic) SignInLogic(req *user.UserSignInReq) (*user.UserSignInResp, error) {
	// 获取参数
	email := req.GetEmail()
	pwd := req.GetPassword()
	// 获取用户数据
	info, err := client.GetDBClient().GetUserByEmail(email)
	// 未找到或密码不对
	if err == conf.DBNotFoundError || info.Password != pwd {
		return &user.UserSignInResp{
			UserInfo: nil,
			Resp: &user.RespBody{
				Code:    conf.ERROR_LOGIN_CODE,
				RespMsg: conf.LOGIN_ERROR_MESSAGE,
			},
		}, errors.Wrap(err, "[UserService] SignInLogic wrong pwd err: ")
		// 服务器内部错误
	} else if err != nil {
		return &user.UserSignInResp{
			UserInfo: nil,
			Resp: &user.RespBody{
				Code:    conf.SERVER_ERROR_CODE,
				RespMsg: conf.SERVER_ERROR_MSG,
			},
		}, errors.Wrap(err, "[UserService] SignInLogic service err: ")
	}
	// 成功
	return &user.UserSignInResp{
		UserInfo: &user.User{
			Uuid:        info.Uuid,
			Name:        info.Name,
			Password:    info.Password,
			Email:       info.Email,
			Phone:       info.Phone,
			Level:       int32(info.Level),
			StartUuid:   info.Start_Uuid,
			NowVolume:   info.Now_Volume,
			TotalVolume: info.Total_Volume,
		},
		Resp: &user.RespBody{
			Code:    conf.RPC_SUCCESS_CODE,
			RespMsg: conf.SUCCESS_RESP_MESSAGE,
		},
	}, nil
}

func (u *UserLogic) SignUpLogic(req *user.UserSignUpReq) (*user.UserSignUpResp, error) {
	// 获取参数
	user_info := req.GetUserInfo()
	// 判断是否存在用户
	info, err := client.GetDBClient().GetUserByEmail(user_info.Email)
	if err != nil && err != conf.DBNotFoundError {
		return &user.UserSignUpResp{
			UserUuid: "",
			Resp: &user.RespBody{
				Code:    conf.SERVER_ERROR_CODE,
				RespMsg: conf.SERVER_ERROR_MSG,
			},
		}, errors.Wrap(err, "[UserService] SignUpLogic err: ")
	}
	// 存在则报错
	if info != nil {
		return &user.UserSignUpResp{
			UserUuid: "",
			Resp: &user.RespBody{
				Code:    conf.ERROR_REGISTER_CODE,
				RespMsg: conf.REGISTER_REPEAT_MESSAGE,
			},
		}, nil
	}

	// 判断验证码是否有效
	src := req.GetCode()
	key := helper.GenVerifyCodeKey(conf.Code_Cache_Key, user_info.Email)
	code, err := client.GetCacheClient().Get(key)
	if err != nil || code == "" || src != code {
		return &user.UserSignUpResp{
			UserUuid: "",
			Resp: &user.RespBody{
				Code:    conf.ERROR_VERIFY_CODE,
				RespMsg: conf.VERIFY_CODE_ERROR_MESSAGE,
			},
		}, errors.Wrap(err, "[UserService] SignUpLogic verify code error ")
	}
	// 生成用户ID
	id := helper.GenUid(user_info.Name, user_info.Email)
	// 生成用户空间根目录uuid
	folderName := fmt.Sprintf("%s-%s", user_info.Name, id)
	user_file_uuid := helper.GenUserFid(user_info.Uuid, folderName)
	// 生成用户db结构
	user_db := &models.User{
		Uuid:         id,
		Name:         user_info.Name,
		Password:     user_info.Password,
		Email:        user_info.Email,
		Phone:        user_info.Phone,
		Level:        conf.User_Level_normal,
		Start_Uuid:   user_file_uuid,
		Now_Volume:   0,
		Total_Volume: conf.User_Normal_Volume,
	}
	// 生成user_file结构体
	user_file := &models.UserFile{
		Uuid:      user_file_uuid,
		User_Uuid: id,
		Parent_Id: conf.Default_System_parent,
		Name:      folderName,
		Ext:       conf.Folder_Default_EXT,
	}

	// 创建用户记录，同时创建用户空间根目录
	err = client.GetDBClient().CreateUser(user_db, user_file)
	if err != nil {
		return &user.UserSignUpResp{
			UserUuid: "",
			Resp: &user.RespBody{
				Code:    conf.SERVER_ERROR_CODE,
				RespMsg: conf.SERVER_ERROR_MSG,
			},
		}, errors.Wrap(err, "[UserService] SignUpLogic service err ")
	}

	// 返回成功
	return &user.UserSignUpResp{
		UserUuid: id,
		Resp: &user.RespBody{
			Code:    conf.RPC_SUCCESS_CODE,
			RespMsg: conf.SUCCESS_RESP_MESSAGE,
		},
	}, nil
}
