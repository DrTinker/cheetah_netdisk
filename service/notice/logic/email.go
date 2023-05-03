package logic

import (
	"NetDesk/common/client"
	"NetDesk/common/conf"
	"NetDesk/common/helper"
	"NetDesk/service/notice/proto/notice"

	"github.com/pkg/errors"
)

type NoticeLogic struct {
	cache client.CacheClient
}

func NewNoticeLogic() *NoticeLogic {
	return &NoticeLogic{
		cache: client.GetCacheClient(),
	}
}

func (n *NoticeLogic) SendSignUpEmailLogic(req *notice.SendEmailReq) (*notice.SendEmailResp, error) {
	// 获取配置文件
	cfg, err := client.GetConfigClient().GetEmailConfig()
	if err != nil {
		return &notice.SendEmailResp{
			Resp: &notice.RespBody{
				Code:    conf.SERVER_ERROR_CODE,
				RespMsg: conf.SERVER_ERROR_MSG,
			},
		}, errors.Wrap(err, "[NoticeService] SendSignUpEmailLogic get cfg err: ")
	}
	// 获取参数
	to := req.GetUserEmail()
	// 生成验证码
	code := helper.GenRandCode()
	// 生成rediskey
	key := helper.GenVerifyCodeKey(conf.Code_Cache_Key, to)
	// 上一个验证码过期后才能set
	flag, err := client.GetCacheClient().SetNX(key, code, conf.Code_Expire)
	if err != nil {
		return &notice.SendEmailResp{
			Resp: &notice.RespBody{
				Code:    conf.ERROR_VERIFY_CODE,
				RespMsg: conf.VERIFY_CODE_GEN_ERROR_MESSAGE,
			},
		}, errors.Wrap(err, "[NoticeService] SendSignUpEmailLogic cache err: ")
	}
	// key存在
	if !flag {
		return &notice.SendEmailResp{
			Resp: &notice.RespBody{
				Code:    conf.VERIFY_CODE_EXIST,
				RespMsg: conf.VERIFY_CODE_GEN_ERROR_MESSAGE,
			},
		}, nil
	}
	// 发送邮件
	content := req.GetContent()
	err = client.GetMsgClient().SendHTMLWithTls(cfg, to, content, conf.Email_Verify_MSG)
	if err != nil {
		return &notice.SendEmailResp{
			Resp: &notice.RespBody{
				Code:    conf.SERVER_ERROR_CODE,
				RespMsg: conf.SERVER_ERROR_MSG,
			},
		}, errors.Wrap(err, "[NoticeService] SendSignUpEmailLogic send err: ")
	}

	// 返回信息
	return &notice.SendEmailResp{
		Resp: &notice.RespBody{
			Code:    conf.RPC_SUCCESS_CODE,
			RespMsg: conf.SUCCESS_RESP_MESSAGE,
		},
	}, nil
}
