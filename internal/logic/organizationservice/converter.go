package organizationservicelogic

import (
	"database/sql"
	"github.com/ziptako/organization/db/model"
	"github.com/ziptako/organization/organization"
	"time"
)

// ModelToProtoOrganization 将model组织转换为proto组织
func ModelToProtoOrganization(source *model.Organizations) *organization.Organization {
	res := &organization.Organization{
		Id:        source.Id,
		ParentId:  source.ParentId.Int64,
		Name:      source.Name,
		CreatedAt: source.CreatedAt.Unix(),
		UpdatedAt: source.UpdatedAt.Unix(),
	}
	if source.DeletedAt.Valid {
		res.DeletedAt = source.DeletedAt.Time.Unix()
	}
	if source.DisabledAt.Valid {
		res.DisabledAt = source.DisabledAt.Time.Unix()
	}
	return res
}

// ModelToProtoOrganizationTree 将model的组织树转换为proto的组织树
func ModelToProtoOrganizationTree(source *model.OrganizationsTree) *organization.OrganizationTree {
	if source == nil {
		return nil
	}

	// 转换当前节点
	dst := &organization.OrganizationTree{
		Id:        source.Organizations.Id,
		Name:      source.Organizations.Name,
		ParentId:  source.Organizations.ParentId.Int64,
		CreatedAt: source.Organizations.CreatedAt.Unix(),
		UpdatedAt: source.Organizations.UpdatedAt.Unix(),
		DeletedAt: 0,
		Children:  make([]*organization.OrganizationTree, 0, len(source.Children)),
	}

	// 处理删除时间
	if source.Organizations.DeletedAt.Valid {
		dst.DeletedAt = source.Organizations.DeletedAt.Time.Unix()
	}

	// 处理禁用时间
	if source.Organizations.DisabledAt.Valid {
		dst.DisabledAt = source.Organizations.DisabledAt.Time.Unix()
	}

	// 递归转换子节点
	for _, child := range source.Children {
		if childTree := ModelToProtoOrganizationTree(child); childTree != nil {
			dst.Children = append(dst.Children, childTree)
		}
	}

	return dst
}

// ProtoToModelOrganization 将proto组织转换为model组织
func ProtoToModelOrganization(source *organization.Organization) *model.Organizations {
	return &model.Organizations{
		Id: source.Id,
		ParentId: sql.NullInt64{
			Valid: source.ParentId != 0,
			Int64: source.ParentId,
		},
		Name:      source.Name,
		CreatedAt: time.Unix(source.CreatedAt, 0),
		UpdatedAt: time.Unix(source.UpdatedAt, 0),
		DeletedAt: sql.NullTime{
			Valid: source.DeletedAt != 0,
			Time:  time.Unix(source.DeletedAt, 0),
		},
		DisabledAt: sql.NullTime{
			Valid: source.DisabledAt != 0,
			Time:  time.Unix(source.DisabledAt, 0),
		},
	}
}

// ProtoToModelOrganizationTree 将proto的组织树转换为model的组织树
func ProtoToModelOrganizationTree(source *organization.OrganizationTree) *model.OrganizationsTree {
	if source == nil {
		return nil
	}

	// 转换当前节点，复用单个组织转换函数
	modelOrg := ProtoToModelOrganization(&organization.Organization{
		Id:         source.Id,
		ParentId:   source.ParentId,
		Name:       source.Name,
		CreatedAt:  source.CreatedAt,
		UpdatedAt:  source.UpdatedAt,
		DeletedAt:  source.DeletedAt,
		DisabledAt: source.DisabledAt,
	})

	// 创建组织树节点
	tree := &model.OrganizationsTree{
		Organizations: modelOrg,
		Children:      make([]*model.OrganizationsTree, 0, len(source.Children)),
	}

	// 递归转换子节点
	for _, child := range source.Children {
		if childTree := ProtoToModelOrganizationTree(child); childTree != nil {
			tree.Children = append(tree.Children, childTree)
		}
	}

	return tree
}
