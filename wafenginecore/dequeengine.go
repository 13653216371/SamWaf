package wafenginecore

import (
	"SamWaf/global"
	"SamWaf/innerbean"
	"SamWaf/model"
	"SamWaf/utils"
	"SamWaf/utils/zlog"
	"encoding/json"
	"github.com/edwingeng/deque"
	uuid "github.com/satori/go.uuid"
	"time"
)

/*
*
初始化队列
*/
func InitDequeEngine() {
	global.GQEQUE_DB = deque.NewDeque()
	global.GQEQUE_LOG_DB = deque.NewDeque()
	global.GQEQUE_MESSAGE_DB = deque.NewDeque()
}

/*
*
处理队列信息
*/
func ProcessDequeEngine() {
	for {
		defer func() {
			e := recover()
			if e != nil {
				zlog.Info("ProcessErrorException", e)
			}
		}()
		for !global.GQEQUE_DB.Empty() {
			weblogbean := global.GQEQUE_DB.PopFront()
			if weblogbean != nil {
				global.GWAF_LOCAL_DB.Create(weblogbean)
			}
		}

		for !global.GQEQUE_LOG_DB.Empty() {
			weblogbean := global.GQEQUE_LOG_DB.PopFront()
			if weblogbean != nil {
				global.GWAF_LOCAL_LOG_DB.Create(weblogbean)
			}
		}

		for !global.GQEQUE_MESSAGE_DB.Empty() {
			messageinfo := global.GQEQUE_MESSAGE_DB.PopFront().(interface{})
			switch messageinfo.(type) {
			case innerbean.RuleMessageInfo:
				rulemessage := messageinfo.(innerbean.RuleMessageInfo)
				utils.NotifyHelperApp.SendRuleInfo(rulemessage)
				if rulemessage.BaseMessageInfo.OperaType == "命中保护规则" {
					//发送websocket
					for _, ws := range global.GWebSocket {
						if ws != nil {
							//写入ws数据
							msgBytes, err := json.Marshal(model.MsgPacket{
								MessageId:           uuid.NewV4().String(),
								MessageType:         "命中保护规则",
								MessageData:         rulemessage.RuleInfo + rulemessage.Ip,
								MessageAttach:       nil,
								MessageDateTime:     time.Now().Format("2006-01-02 15:04:05"),
								MessageUnReadStatus: true,
							})
							err = ws.WriteMessage(1, msgBytes)
							if err != nil {
								continue
							}
						}
					}
				}
				break
			case innerbean.OperatorMessageInfo:
				operatorMessage := messageinfo.(innerbean.OperatorMessageInfo)
				utils.NotifyHelperApp.SendNoticeInfo(operatorMessage)
				break
			}

			//zlog.Info("MESSAGE", messageinfo)
		}
		time.Sleep((100 * time.Millisecond))
	}
}