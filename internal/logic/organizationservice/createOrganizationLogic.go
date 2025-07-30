package organizationservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ziptako/organization/db/model"
	"github.com/ziptako/organization/internal/svc"
	"github.com/ziptako/organization/organization"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrganizationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	model model.OrganizationsModel
}

func NewCreateOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrganizationLogic {
	return &CreateOrganizationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		model:  model.NewOrganizationsModel(svcCtx.SqlConn, svcCtx.CacheConf),
	}
}

// CreateOrganization 创建组织节点
func (l *CreateOrganizationLogic) CreateOrganization(in *organization.CreateOrganizationRequest) (*organization.CreateOrganizationResponse, error) {
	// 检查祖先节点
	if in.ParentId != 0 {
		_, err := l.model.FindOne(l.ctx, in.ParentId)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, status.Error(codes.NotFound, "[CO002] 祖先节点不存在")
			}
			eInfo := "[CO003] 查询失败"
			l.Logger.Errorf("%v: %v", eInfo, err)
			return nil, status.Error(codes.Internal, eInfo)
		}
	}
	newOrg := &model.Organizations{
		ParentId: sql.NullInt64{
			Valid: in.ParentId != 0,
			Int64: in.ParentId,
		},
		Name: in.Name,
	}
	insert, err := l.model.Insert(l.ctx, newOrg)
	if err != nil {
		eInfo := "[CO001] 创建组织失败"
		l.Logger.Errorf("%v: %v", eInfo, err)
		return nil, status.Error(codes.Internal, eInfo)
	}
	id, _ := insert.LastInsertId()
	return &organization.CreateOrganizationResponse{
		Id: id,
	}, nil
}
