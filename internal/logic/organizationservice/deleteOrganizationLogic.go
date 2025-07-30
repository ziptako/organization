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

type DeleteOrganizationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	model model.OrganizationsModel
}

func NewDeleteOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOrganizationLogic {
	return &DeleteOrganizationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		model:  model.NewOrganizationsModel(svcCtx.SqlConn, svcCtx.CacheConf),
	}
}

// DeleteOrganization 删除组织节点
func (l *DeleteOrganizationLogic) DeleteOrganization(in *organization.DeleteOrganizationRequest) (*organization.DeleteOrganizationResponse, error) {
	_, err := l.model.FindOne(l.ctx, in.Id)
	if err != nil {
		eInfo := "[DO001] 查询失败"
		l.Logger.Errorf("%v: %v", eInfo, err)
		return nil, status.Error(codes.Internal, eInfo)
	}
	err = l.model.SoftDelete(l.ctx, in.Id)
	if err != nil {
		eInfo := "[DO002] 删除失败"
		l.Logger.Errorf("%v: %v", eInfo, err)
		return nil, status.Error(codes.Internal, eInfo)
	}

	return &organization.DeleteOrganizationResponse{
		Success: true,
	}, nil
}
