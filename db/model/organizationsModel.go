package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// customResult 自定义Result实现，用于PostgreSQL的RETURNING子句
type customResult struct {
	insertedID int64
}

// LastInsertId 返回插入的ID
func (r *customResult) LastInsertId() (int64, error) {
	return r.insertedID, nil
}

// RowsAffected 返回受影响的行数
func (r *customResult) RowsAffected() (int64, error) {
	return 1, nil
}

var _ OrganizationsModel = (*customOrganizationsModel)(nil)

type (
	// OrganizationsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customOrganizationsModel.
	OrganizationsModel interface {
		organizationsModel
		FindAncestorsById(ctx context.Context, id int64) ([]*Organizations, error)
		FindActiveById(ctx context.Context, id int64) (*Organizations, error)               // 查询活跃组织（未删除且未禁用）
		FindById(ctx context.Context, id int64) (*Organizations, error)                     // 查询活跃组织（未删除且未禁用）
		FindByName(ctx context.Context, name string) ([]*Organizations, error)              // 按名称查询活跃组织
		FindActiveByName(ctx context.Context, name string) ([]*Organizations, error)        // 按名称查询活跃组织
		FindByParentId(ctx context.Context, parentId int64) ([]*Organizations, error)       // 查询子组织
		FindActiveByParentId(ctx context.Context, parentId int64) ([]*Organizations, error) // 查询活跃子组织

		FindDescendantsById(ctx context.Context, id int64) (*OrganizationsTree, error)       // 查询所有后代组织
		FindActiveDescendantsById(ctx context.Context, id int64) (*OrganizationsTree, error) // 查询所有活跃后代组织

		SoftDelete(ctx context.Context, id int64) error         // 软删除组织
		Restore(ctx context.Context, id int64) error            // 恢复已删除组织
		Disable(ctx context.Context, id int64) error            // 禁用组织
		Enable(ctx context.Context, id int64) error             // 启用组织
		BatchSoftDelete(ctx context.Context, ids []int64) error // 批量软删除
		BatchDisable(ctx context.Context, ids []int64) error    // 批量禁用
		/*
			TODO: 根据表结构和索引优化，添加以下业务方法


			// 树形结构方法

			BuildTree(ctx context.Context, rootId int64) (*OrganizationsTree, error)               // 构建组织树
			BuildActiveTree(ctx context.Context, rootId int64) (*OrganizationsTree, error)         // 构建活跃组织树
			FindRoots(ctx context.Context) ([]*Organizations, error)                               // 查询根组织
			FindActiveRoots(ctx context.Context) ([]*Organizations, error)                         // 查询活跃根组织

			// 状态管理方法


			// 统计和分页方法
			CountByParentId(ctx context.Context, parentId int64) (int64, error)                    // 统计子组织数量
			CountActiveByParentId(ctx context.Context, parentId int64) (int64, error)              // 统计活跃子组织数量
			FindWithPagination(ctx context.Context, page, pageSize int64) ([]*Organizations, int64, error) // 分页查询
			FindActiveWithPagination(ctx context.Context, page, pageSize int64) ([]*Organizations, int64, error) // 分页查询活跃组织

			// 验证方法
			ExistsByName(ctx context.Context, name string, excludeId int64) (bool, error)          // 检查名称是否存在（排除指定ID）
			IsAncestor(ctx context.Context, ancestorId, descendantId int64) (bool, error)          // 检查是否为祖先关系
			ValidateParent(ctx context.Context, id, parentId int64) error                          // 验证父级关系（防止循环引用）

			// 搜索方法
			SearchByName(ctx context.Context, keyword string, limit int64) ([]*Organizations, error) // 按名称模糊搜索
			FindByTimeRange(ctx context.Context, startTime, endTime string) ([]*Organizations, error) // 按时间范围查询

		*/
	}

	customOrganizationsModel struct {
		*defaultOrganizationsModel
	}
)
type (
	OrganizationsTree struct {
		*Organizations
		Children []*OrganizationsTree
	}
)

// NewOrganizationsModel returns a model for the database table.
func NewOrganizationsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) OrganizationsModel {
	return &customOrganizationsModel{
		defaultOrganizationsModel: newOrganizationsModel(conn, c, opts...),
	}
}

// FindAncestorsById 查找指定组织的所有祖先组织（从根到父级）
func (m *customOrganizationsModel) FindAncestorsById(ctx context.Context, id int64) ([]*Organizations, error) {
	var ancestors []*Organizations

	curID := id
	for {
		org, err := m.FindById(ctx, curID)
		if err != nil {
			return nil, err
		}

		// 追加到末尾，最后再反转即可
		ancestors = append(ancestors, org)

		if !org.ParentId.Valid {
			break
		}
		curID = org.ParentId.Int64
	}

	for i, j := 0, len(ancestors)-1; i < j; i, j = i+1, j-1 {
		ancestors[i], ancestors[j] = ancestors[j], ancestors[i]
	}
	return ancestors, nil
}

// FindActiveById 查询活跃组织（未删除且未禁用）
func (m *customOrganizationsModel) FindActiveById(ctx context.Context, id int64) (*Organizations, error) {
	query := fmt.Sprintf("select %s from %s where id = $1 and deleted_at IS NULL and disabled_at IS NULL limit 1", organizationsRows, m.table)
	var resp Organizations
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, id)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlc.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindById 查询组织 (未删除)
func (m *customOrganizationsModel) FindById(ctx context.Context, id int64) (*Organizations, error) {
	query := fmt.Sprintf("select %s from %s where id = $1 and deleted_at IS NULL limit 1", organizationsRows, m.table)
	var resp Organizations
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, id)
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlc.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByName 按名称查询组织 (未删除且未禁用)
func (m *customOrganizationsModel) FindByName(ctx context.Context, name string) ([]*Organizations, error) {
	query := fmt.Sprintf("select %s from %s where name = $1 and deleted_at IS NULL order by created_at", organizationsRows, m.table)
	var resp []*Organizations
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, name)
	return resp, err
}

// FindActiveByName 按名称查询组织 (未删除且未禁用)
func (m *customOrganizationsModel) FindActiveByName(ctx context.Context, name string) ([]*Organizations, error) {
	query := fmt.Sprintf("select %s from %s where name = $1 and deleted_at IS NULL and disabled_at IS NULL order by created_at", organizationsRows, m.table)
	var resp []*Organizations
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, name)
	return resp, err
}

// FindByParentId 查询子组织 (未删除)
func (m *customOrganizationsModel) FindByParentId(ctx context.Context, parentId int64) ([]*Organizations, error) {
	query := fmt.Sprintf("select %s from %s where parent_id = $1 and deleted_at IS NULL order by created_at", organizationsRows, m.table)
	var resp []*Organizations
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, parentId)
	return resp, err
}

// FindActiveByParentId 查询活跃子组织 (未删除且未禁用)
func (m *customOrganizationsModel) FindActiveByParentId(ctx context.Context, parentId int64) ([]*Organizations, error) {
	query := fmt.Sprintf("select %s from %s where parent_id = $1 and deleted_at IS NULL and disabled_at IS NULL order by created_at", organizationsRows, m.table)
	var resp []*Organizations
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, parentId)
	return resp, err
}

// FindDescendantsById 查询组织所有后代组织并构建树形结构
func (m *customOrganizationsModel) FindDescendantsById(ctx context.Context, id int64) (*OrganizationsTree, error) {
	// 首先根据id查询组织
	org, err := m.FindById(ctx, id)
	if err != nil {
		return nil, err
	}

	// 创建组织树根节点
	tree := &OrganizationsTree{
		Organizations: org,
		Children:      []*OrganizationsTree{},
	}

	// 递归构建子树
	var buildDescendantsTree func(*OrganizationsTree) error
	buildDescendantsTree = func(node *OrganizationsTree) error {
		// 根据当前节点ID查询子组织
		children, err := m.FindByParentId(ctx, node.Id)
		if err != nil {
			return err
		}

		// 为每个子组织创建树节点
		for _, child := range children {
			childNode := &OrganizationsTree{
				Organizations: child,
				Children:      []*OrganizationsTree{},
			}
			node.Children = append(node.Children, childNode)

			// 递归构建子树
			if err := buildDescendantsTree(childNode); err != nil {
				return err
			}
		}
		return nil
	}

	// 开始递归构建
	if err := buildDescendantsTree(tree); err != nil {
		return nil, err
	}

	return tree, nil
}

// FindActiveDescendantsById 获取指定组织及其子组织
func (m *customOrganizationsModel) FindActiveDescendantsById(ctx context.Context, id int64) (*OrganizationsTree, error) {
	// 首先根据id查询活跃组织
	org, err := m.FindActiveById(ctx, id)
	if err != nil {
		return nil, err
	}

	// 创建组织树根节点
	tree := &OrganizationsTree{
		Organizations: org,
		Children:      []*OrganizationsTree{},
	}

	// 递归构建子树
	var buildDescendantsTree func(*OrganizationsTree) error
	buildDescendantsTree = func(node *OrganizationsTree) error {
		// 根据当前节点ID查询活跃子组织
		children, err := m.FindActiveByParentId(ctx, node.Id)
		if err != nil {
			return err
		}

		// 为每个子组织创建树节点
		for _, child := range children {
			childNode := &OrganizationsTree{
				Organizations: child,
				Children:      []*OrganizationsTree{},
			}
			node.Children = append(node.Children, childNode)

			// 递归构建子树
			if err := buildDescendantsTree(childNode); err != nil {
				return err
			}
		}
		return nil
	}

	// 开始递归构建
	if err := buildDescendantsTree(tree); err != nil {
		return nil, err
	}

	return tree, nil
}

// SoftDelete 软删除组织
func (m *customOrganizationsModel) SoftDelete(ctx context.Context, id int64) error {
	one, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}
	one.DeletedAt.Valid = true
	one.DeletedAt.Time = time.Now()
	err = m.Update(ctx, one)
	return err
}

// Restore 恢复已删除组织
func (m *customOrganizationsModel) Restore(ctx context.Context, id int64) error {
	orgOrganizationsIdKey := fmt.Sprintf("%s%v", cacheOrgOrganizationsIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set deleted_at = NULL where id = $1", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, orgOrganizationsIdKey)
	return err
}

// Disable 禁用组织
func (m *customOrganizationsModel) Disable(ctx context.Context, id int64) error {
	orgOrganizationsIdKey := fmt.Sprintf("%s%v", cacheOrgOrganizationsIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set disabled_at = NOW() where id = $1 and deleted_at IS NULL and disabled_at IS NULL", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, orgOrganizationsIdKey)
	return err
}

// Enable 启用组织
func (m *customOrganizationsModel) Enable(ctx context.Context, id int64) error {
	orgOrganizationsIdKey := fmt.Sprintf("%s%v", cacheOrgOrganizationsIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set disabled_at = NULL where id = $1 and deleted_at IS NULL", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, orgOrganizationsIdKey)
	return err
}

// BatchSoftDelete 批量软删除
func (m *customOrganizationsModel) BatchSoftDelete(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	// 构建占位符
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	// 清除相关缓存
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf("%s%v", cacheOrgOrganizationsIdPrefix, id)
	}

	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set deleted_at = NOW() where id IN (%s) and deleted_at IS NULL",
			m.table, strings.Join(placeholders, ","))
		return conn.ExecCtx(ctx, query, args...)
	}, keys...)
	return err
}

// BatchDisable 批量禁用
func (m *customOrganizationsModel) BatchDisable(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	// 构建占位符
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	// 清除相关缓存
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = fmt.Sprintf("%s%v", cacheOrgOrganizationsIdPrefix, id)
	}

	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set disabled_at = NOW() where id IN (%s) and deleted_at IS NULL and disabled_at IS NULL",
			m.table, strings.Join(placeholders, ","))
		return conn.ExecCtx(ctx, query, args...)
	}, keys...)
	return err
}

// Insert 重写Insert方法，使用PostgreSQL的RETURNING子句获取插入后的ID
func (m *customOrganizationsModel) Insert(ctx context.Context, data *Organizations) (sql.Result, error) {
	// 使用QueryRowCtx来处理RETURNING子句，获取插入后的ID
	var insertedID int64
	query := fmt.Sprintf("insert into %s (%s) values ($1, $2, $3, $4) RETURNING id", m.table, organizationsRowsExpectAutoSet)
	err := m.QueryRowNoCacheCtx(ctx, &insertedID, query, data.ParentId, data.Name, data.DisabledAt, data.DeletedAt)
	if err != nil {
		return nil, err
	}

	// 设置插入后的ID到data对象中
	data.Id = insertedID

	// 清除相关缓存
	orgOrganizationsIdKey := fmt.Sprintf("%s%v", cacheOrgOrganizationsIdPrefix, insertedID)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		// 这里返回一个模拟的Result，包含正确的LastInsertId
		return &customResult{insertedID: insertedID}, nil
	}, orgOrganizationsIdKey)

	return &customResult{insertedID: insertedID}, err
}

func (m *customOrganizationsModel) FindOne(ctx context.Context, id int64) (*Organizations, error) {
	orgOrganizationsIdKey := fmt.Sprintf("%s%v", cacheOrgOrganizationsIdPrefix, id)
	var resp Organizations
	err := m.QueryRowCtx(ctx, &resp, orgOrganizationsIdKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		// 排除软删除的
		query := fmt.Sprintf("select %s from %s where id = $1 and deleted_at IS NULL limit 1", organizationsRows, m.table)
		return conn.QueryRowCtx(ctx, v, query, id)
	})
	switch {
	case err == nil:
		return &resp, nil
	case errors.Is(err, sqlc.ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
