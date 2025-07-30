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

type GetDescendantsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	model model.OrganizationsModel
}

func NewGetDescendantsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDescendantsLogic {
	return &GetDescendantsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		model:  model.NewOrganizationsModel(svcCtx.SqlConn, svcCtx.CacheConf),
	}
}

// GetDescendants 获取后代树
func (l *GetDescendantsLogic) GetDescendants(in *organization.GetDescendantsRequest) (*organization.GetDescendantsResponse, error) {
	root, err := l.model.FindDescendantsById(l.ctx, in.Id)
	if err != nil {
		eInfo := "[GD001] 获取后代树失败"
		l.Logger.Errorf("%v: %v", eInfo, err)
		return nil, status.Error(codes.Internal, eInfo)
	}
	tree := ModelToProtoOrganizationTree(root)

	return &organization.GetDescendantsResponse{
		OrganizationTree: tree,
	}, nil
}
