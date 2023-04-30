package handler

import (
	"NetDesk/common/conf"
	"NetDesk/service/user/logic"
	"NetDesk/service/user/proto/user"
	"context"

	"github.com/sirupsen/logrus"
)

type userService struct{}

var UserService = userService{}

func (u userService) UserSignIn(ctx context.Context, req *user.UserSignInReq) (*user.UserSignInResp, error) {
	// 参数校验
	email := req.GetEmail()
	pwd := req.GetPassword()
	if email == "" || pwd == "" {
		logrus.Error("[UserService] SignInHandler empty param ", req)
		return &user.UserSignInResp{
			UserInfo: nil,
			Resp: &user.RespBody{
				Code:    conf.HTTP_INVALID_PARAMS_CODE,
				RespMsg: conf.HTTP_INVALID_PARAMS_MESSAGE,
			},
		}, conf.ParamError
	}
	// 调用logic
	l := logic.NewUserLogic()
	resp, err := l.SignInLogic(req)
	if err != nil {
		logrus.Error("[UserService] UserSignIn err ", err)
		return nil, err
	}
	// 成功
	logrus.Error("[UserService] UserSignIn sucess ", req)
	return resp, nil
}

func (u userService) UserSignUp(ctx context.Context, req *user.UserSignUpReq) (*user.UserSignUpResp, error) {
	// 参数校验
	if req.GetCode() == "" || req.GetUserInfo() == nil {
		logrus.Error("[UserService] SignUpHandler empty param ", req)
		return &user.UserSignUpResp{
			UserUuid: "",
			Resp: &user.RespBody{
				Code:    conf.HTTP_INVALID_PARAMS_CODE,
				RespMsg: conf.HTTP_INVALID_PARAMS_MESSAGE,
			},
		}, conf.ParamError
	}
	// 调用logic
	l := logic.NewUserLogic()
	resp, err := l.SignUpLogic(req)
	if err != nil {
		logrus.Error("[UserService] UserSignUp err ", err)
		return nil, err
	}
	// 成功
	logrus.Error("[UserService] UserSignUp sucess ", req)
	return resp, nil
}
