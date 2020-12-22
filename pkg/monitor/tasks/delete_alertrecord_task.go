package tasks

import (
	"context"
	"fmt"

	"yunion.io/x/jsonutils"
	"yunion.io/x/pkg/errors"

	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/taskman"
	"yunion.io/x/onecloud/pkg/monitor/models"
	"yunion.io/x/onecloud/pkg/util/logclient"
)

type DeleteAlertRecordTask struct {
	taskman.STask
}

func init() {
	taskman.RegisterTask(&DeleteAlertRecordTask{})
}

func (self *DeleteAlertRecordTask) OnInit(ctx context.Context, obj db.IStandaloneModel, body jsonutils.JSONObject) {
	alert := obj.(*models.SCommonAlert)
	errs := alert.DeleteAttachAlertRecords(ctx, self.GetUserCred())
	if len(errs) != 0 {
		msg := jsonutils.NewString(fmt.Sprintf("fail to DeleteAttachAlertRecords:%s.err:%v", alert.Name, errors.NewAggregate(errs)))
		self.taskFail(ctx, alert, msg)
		return
	}

	err := alert.RealDelete(ctx, self.UserCred)
	if err != nil {
		msg := fmt.Sprintf("delete SCommonAlert err:%v", err)
		self.taskFail(ctx, alert, jsonutils.NewString(msg))
		return
	}
	db.OpsLog.LogEvent(alert, db.ACT_DELETE, nil, self.GetUserCred())
	logclient.AddActionLogWithStartable(self, alert, logclient.ACT_DELETE, nil, self.UserCred, true)
	self.SetStageComplete(ctx, nil)
}

func (self *DeleteAlertRecordTask) taskFail(ctx context.Context, alert *models.SCommonAlert, msg jsonutils.JSONObject) {
	db.OpsLog.LogEvent(alert, db.ACT_DELETE_FAIL, msg, self.GetUserCred())
	logclient.AddActionLogWithStartable(self, alert, logclient.ACT_DELETE, msg, self.UserCred, false)
	self.SetStageFailed(ctx, msg)
	return
}
