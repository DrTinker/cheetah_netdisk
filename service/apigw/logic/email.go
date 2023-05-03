package logic

import (
	"NetDesk/common/client"
	"NetDesk/common/conf"
	"NetDesk/common/helper"
	"NetDesk/service/apigw/proto/notice"
	"context"
	"fmt"
)

type NoticeLogic struct {
	NoticeClient notice.NoticeserviceClient
}

func NewNoticeLogic() (*NoticeLogic, error) {
	// 建立链接
	userConnClient, err := client.NewRpcConnect(conf.Notice_Service_Name)
	if err != nil {
		return nil, err
	}
	conn, err := userConnClient.GetConnect()
	if err != nil {
		return nil, err
	}
	return &NoticeLogic{
		NoticeClient: notice.NewNoticeserviceClient(conn),
	}, nil
}

func (n *NoticeLogic) SendSignUpEmailLogic(email string) (*notice.SendEmailResp, error) {
	// 生成验证码
	code := helper.GenRandCode()
	// 发送邮件
	content := fmt.Sprintf(conf.Email_Verify_Page, code)
	// 拼接req
	req := &notice.SendEmailReq{
		UserEmail: email,
		Content:   content,
	}
	// 调用notice服务
	resp, err := n.NoticeClient.SendEmail(context.Background(), req)
	if err != nil {
		return nil, err
	}
	// 解析resp
	if resp == nil {
		return nil, conf.RPCRespEmptyError
	}
	return resp, nil
}
