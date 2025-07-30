package organizationservicelogic

import (
	"context"
	"errors"
	"github.com/ziptako/organization/db/model"
	"github.com/ziptako/organization/internal/svc"
	"github.com/ziptako/organization/organization"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrganizationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	model model.OrganizationsModel
}

func NewGetOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrganizationLogic {
	return &GetOrganizationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		model:  model.NewOrganizationsModel(svcCtx.SqlConn, svcCtx.CacheConf),
	}
}

// GetOrganization 获取组织节点
func (l *GetOrganizationLogic) GetOrganization(in *organization.GetOrganizationRequest) (*organization.Organization, error) {
	organizations, err := l.model.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "[GO001] 组织节点不存在")
		}
		eInfo := "[GO002] 获取组织节点失败"
		l.Logger.Errorf("%v: %v", eInfo, err)
		return nil, status.Error(codes.Internal, eInfo)
	}

	return ModelToProtoOrganization(organizations), nil
}
