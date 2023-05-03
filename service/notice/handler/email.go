package handler

import (
	"NetDesk/common/conf"
	"NetDesk/service/notice/logic"
	"NetDesk/service/notice/proto/notice"
	"context"

	"github.com/sirupsen/logrus"
)

type noticeService struct{}

var NoticeService = noticeService{}

func (n noticeService) SendEmail(ctx context.Context, req *notice.SendEmailReq) (*notice.SendEmailResp, error) {
	// 校验参数
	if req.GetUserEmail() == "" || req.GetContent() == "" {
		logrus.Error("[NoticeService] SendSignUpEmailHandler empty email ", req)
		return &notice.SendEmailResp{
			Resp: &notice.RespBody{
				Code:    conf.HTTP_INVALID_PARAMS_CODE,
				RespMsg: conf.HTTP_INVALID_PARAMS_MESSAGE,
			},
		}, conf.ParamError
	}
	// 调用logic
	resp, err := logic.NewNoticeLogic().SendSignUpEmailLogic(req)
	if err != nil {
		logrus.Error("[NoticeService] SendSignUpEmailHandler err ", err)
		return nil, err
	}
	// 成功
	logrus.Info("[NoticeService] SendSignUpEmailHandler success ", req)
	return resp, nil
}
