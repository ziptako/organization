package organizationservicelogic

import (
	"context"
	"github.com/ziptako/organization/db/model"
	"github.com/ziptako/organization/internal/svc"
	"github.com/ziptako/organization/organization"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrganizationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	model model.OrganizationsModel
}

func NewListOrganizationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrganizationsLogic {
	return &ListOrganizationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		model:  model.NewOrganizationsModel(svcCtx.SqlConn, svcCtx.CacheConf),
	}
}

// ListOrganizations 分页查询子节点
func (l *ListOrganizationsLogic) ListOrganizations(in *organization.ListOrganizationsRequest) (*organization.ListOrganizationsResponse, error) {
	// todo: add your logic here and delete this line

	return &organization.ListOrganizationsResponse{}, nil
}
