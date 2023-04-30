package logic

import (
	"NetDesk/common/client"
	"NetDesk/common/conf"
	"NetDesk/common/models"
	"NetDesk/service/apigw/proto/user"
	"context"
)

type UserLogic struct {
	UserClient user.UserServiceClient
}

func NewUserLogic() (*UserLogic, error) {
	// 建立链接
	userConnClient, err := client.NewRpcConnect(conf.User_Service_Name)
	if err != nil {
		return nil, err
	}
	conn, err := userConnClient.GetConnect()
	if err != nil {
		return nil, err
	}
	return &UserLogic{
		UserClient: user.NewUserServiceClient(conn),
	}, nil
}

func (u *UserLogic) UserSignIn(user *user.UserSignInReq, flag bool) (*user.UserSignInResp, string, error) {
	// 调用user服务
	resp, err := u.UserClient.UserSignIn(context.Background(), user)
	if err != nil {
		return nil, "", err
	}
	// 解析resp
	if resp == nil {
		return nil, "", conf.RPCRespEmptyError
	}
	// 不成功
	if resp.Resp.Code != conf.RPC_SUCCESS_CODE {
		return resp, "", nil
	}
	// 成功则获取user_info
	info := resp.UserInfo
	var token string
	// 未携带token则生成新token
	if !flag {
		token, _ = client.EncryptionClient.JwtEncode(models.Token{
			ID:       info.Uuid,
			Email:    info.Email,
			Password: info.Password,
			Expire:   0,
		})
	}
	return resp, token, nil
}

func (u *UserLogic) UserSignUp(info *models.User, code string) (*user.UserSignUpResp, error) {
	// 调用user服务
	resp, err := u.UserClient.UserSignUp(context.Background(), &user.UserSignUpReq{
		UserInfo: &user.User{
			Uuid:     info.Uuid,
			Name:     info.Name,
			Password: info.Password,
			Email:    info.Email,
			Phone:    info.Phone,
		},
		Code: code,
	})
	// 解析resp
	if resp == nil {
		return nil, conf.RPCRespEmptyError
	}
	if err != nil {
		return nil, err
	}

	return resp, nil
}
