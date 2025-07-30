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

type GetAncestorsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	model model.OrganizationsModel
}

func NewGetAncestorsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAncestorsLogic {
	return &GetAncestorsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		model:  model.NewOrganizationsModel(svcCtx.SqlConn, svcCtx.CacheConf),
	}
}

// GetAncestors 获取祖先链
func (l *GetAncestorsLogic) GetAncestors(in *organization.GetAncestorsRequest) (*organization.GetAncestorsResponse, error) {
	modelOrganizations, err := l.model.FindAncestorsById(l.ctx, in.Id)

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "[GA001] 组织节点不存在")
		}
		eInfo := "[GA002] 获取祖先链失败"
		l.Logger.Errorf("%v: %v", eInfo, err)
		return nil, status.Error(codes.Internal, eInfo)
	}
	var ancestor []*organization.Organization
	for _, org := range modelOrganizations {
		ancestor = append(ancestor, ModelToProtoOrganization(org))
	}
	return &organization.GetAncestorsResponse{
		Ancestors: ancestor,
	}, nil
}
