package organizationservicelogic

import (
	"context"
	"github.com/ziptako/organization/db/model"
	"github.com/ziptako/organization/internal/svc"
	"github.com/ziptako/organization/organization"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateOrganizationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	model model.OrganizationsModel
}

func NewUpdateOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrganizationLogic {
	return &UpdateOrganizationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		model:  model.NewOrganizationsModel(svcCtx.SqlConn, svcCtx.CacheConf),
	}
}

// UpdateOrganization 更新组织节点名称
func (l *UpdateOrganizationLogic) UpdateOrganization(in *organization.UpdateOrganizationRequest) (*organization.Organization, error) {
	organizations, err := l.model.FindActiveById(l.ctx, in.Id)
	if err != nil {
		eInfo := "[UO001] 未找到"
		l.Logger.Errorf("%v: %v", eInfo, err)
		return nil, status.Error(codes.Internal, eInfo)
	}
	organizations.Name = in.Name
	err = l.model.Update(l.ctx, organizations)
	if err != nil {
		eInfo := "[UO002] 更新失败"
		l.Logger.Errorf("%v: %v", eInfo, err)
		return nil, status.Error(codes.Internal, eInfo)
	}
	return ModelToProtoOrganization(organizations), nil
}
