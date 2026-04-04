package generate

const moduleTemplate = `package {{.PackageName}}

const Name = "{{.Name}}"
const ManifestPath = "modules/{{.EntityLower}}/manifest.yaml"

type Module struct {
	Name         string
	ManifestPath string
}

func NewModule() Module {
	return Module{Name: Name, ManifestPath: ManifestPath}
}
`

const manifestTemplate = `name: {{.Name}}
version: v1
kind: {{.Kind}}
module: {{.Module}}
dependencies:
  - core/auth/bootstrap
  - core/auth/casbin
{{.ManifestRoutes}}{{.ManifestMenus}}{{.ManifestPermissions}}capabilities:
  - basic-crud
  - policy-generated
  - frontend-generated
`

const configTemplate = `name: {{.Name}}
version: v1
kind: {{.Kind}}
module: {{.Module}}
settings:
  environment: generated
  enabled: true
`

const modelTemplate = `package model

import "time"

type {{.Entity}} struct {
{{.ModelFields}}	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

func (m {{.Entity}}) Clone() {{.Entity}} {
	clone := m
{{.CloneBlock}}	return clone
}
`

const repositoryTemplate = `package repository

import (
	"context"
	"errors"

	"goadmin/modules/{{.EntityLower}}/domain/model"
)

var ErrNotFound = errors.New("{{.EntityLower}} not found")

type Repository interface {
	List(ctx context.Context, keyword string, page int, pageSize int) ([]model.{{.Entity}}, int64, error)
	Get(ctx context.Context, id string) (*model.{{.Entity}}, error)
	Create(ctx context.Context, item *model.{{.Entity}}) (*model.{{.Entity}}, error)
	Update(ctx context.Context, item *model.{{.Entity}}) (*model.{{.Entity}}, error)
	Delete(ctx context.Context, id string) error
}
`

const commandTemplate = `package command

{{if .HasInputTime}}import "time"
{{end}}

type Create{{.Entity}} struct {
{{.CommandFields}}}

type Update{{.Entity}} struct {
{{.CommandFields}}}
`

const queryTemplate = `package query

type List{{.EntityPlural}} struct {
	Keyword  string
	Page     int
	PageSize int
}
`

const serviceTemplate = `package service

import (
	"context"
	"fmt"

	"goadmin/modules/{{.EntityLower}}/application/command"
	"goadmin/modules/{{.EntityLower}}/application/query"
	"goadmin/modules/{{.EntityLower}}/domain/model"
	"goadmin/modules/{{.EntityLower}}/domain/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("{{.EntityLower}} repository is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) List(ctx context.Context, q query.List{{.EntityPlural}}) ([]model.{{.Entity}}, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, fmt.Errorf("{{.EntityLower}} service is not configured")
	}
	return s.repo.List(ctx, q.Keyword, q.Page, q.PageSize)
}

func (s *Service) Get(ctx context.Context, id string) (*model.{{.Entity}}, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("{{.EntityLower}} service is not configured")
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input command.Create{{.Entity}}) (*model.{{.Entity}}, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("{{.EntityLower}} service is not configured")
	}
	item := &model.{{.Entity}}{}
{{.CreateAssignments}}	return s.repo.Create(ctx, item)
}

func (s *Service) Update(ctx context.Context, id string, input command.Update{{.Entity}}) (*model.{{.Entity}}, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("{{.EntityLower}} service is not configured")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	cloned := item.Clone()
	item = &cloned
{{.UpdateAssignments}}	return s.repo.Update(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("{{.EntityLower}} service is not configured")
	}
	return s.repo.Delete(ctx, id)
}
`

const gormRepositoryTemplate = `package repo

import (
	"context"
	"fmt"
	{{if .NeedsStringsImport}}"strings"{{end}}
{{if .NeedsPrimaryIDGeneration}}
	"time"
{{end}}

	"goadmin/modules/{{.EntityLower}}/domain/model"

	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("{{.EntityLower}} gorm repository requires db")
	}
	return &GormRepository{db: db}, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("{{.EntityLower}} migrate requires db")
	}
	return db.AutoMigrate(&model.{{.Entity}}{})
}

func (r *GormRepository) List(ctx context.Context, keyword string, page int, pageSize int) ([]model.{{.Entity}}, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("{{.EntityLower}} gorm repository is not configured")
	}
	base := r.db.WithContext(ctx).Model(&model.{{.Entity}}{})
	{{.SearchFilterBlock}}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, pageSize = normalizePage(page, pageSize)
	var items []model.{{.Entity}}
	if err := base.Order("updated_at DESC, created_at DESC, id ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *GormRepository) Get(ctx context.Context, id string) (*model.{{.Entity}}, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("{{.EntityLower}} gorm repository is not configured")
	}
	var item model.{{.Entity}}
	if err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Create(ctx context.Context, item *model.{{.Entity}}) (*model.{{.Entity}}, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("{{.EntityLower}} gorm repository is not configured")
	}
	if item == nil {
		return nil, fmt.Errorf("{{.EntityLower}} item is nil")
	}
{{if .NeedsPrimaryIDGeneration}}
	if strings.TrimSpace(item.Id) == "" {
		item.Id = nextRecordID("{{.EntityLower}}")
	}
{{end}}
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func (r *GormRepository) Update(ctx context.Context, item *model.{{.Entity}}) (*model.{{.Entity}}, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("{{.EntityLower}} gorm repository is not configured")
	}
	if item == nil {
		return nil, fmt.Errorf("{{.EntityLower}} item is nil")
	}
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func (r *GormRepository) Delete(ctx context.Context, id string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("{{.EntityLower}} gorm repository is not configured")
	}
	if err := r.db.WithContext(ctx).Delete(&model.{{.Entity}}{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}
{{if .NeedsPrimaryIDGeneration}}

func nextRecordID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
}
{{end}}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return page, pageSize
}
`

const requestTemplate = `package request

{{if .HasInputTime}}import "time"
{{end}}

type ListRequest struct {
{{.ListRequestFields}}}

type CreateRequest struct {
{{.RequestFields}}}

type UpdateRequest struct {
{{.RequestFields}}}
`

const responseTemplate = `package response

{{if .HasInputTime}}import "time"
{{end}}

type Item struct {
{{.ResponseFields}}}

type List struct {
	Total int64  ` + "`json:\"total\"`" + `
	Items []Item ` + "`json:\"items\"`" + `
}
`

const handlerTemplate = `package handler

import (
	"net/http"

	coreerrors "goadmin/core/errors"
	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	"goadmin/modules/{{.EntityLower}}/application/command"
	"goadmin/modules/{{.EntityLower}}/application/query"
	{{.EntityLower}}service "goadmin/modules/{{.EntityLower}}/application/service"
	"goadmin/modules/{{.EntityLower}}/domain/model"
	{{.EntityLower}}req "goadmin/modules/{{.EntityLower}}/transport/http/request"
	{{.EntityLower}}resp "goadmin/modules/{{.EntityLower}}/transport/http/response"
	"go.uber.org/zap"
)

type Handler struct {
	service *{{.EntityLower}}service.Service
	logger  *zap.Logger
}

func New(service *{{.EntityLower}}service.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c coretransport.Context) {
	var req {{.EntityLower}}req.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, body := response.Failure(coreerrors.Wrap(err, coreerrors.CodeBadRequest, "invalid list request"), requestID(c))
		c.JSON(status, body)
		return
	}
	items, total, err := h.service.List(c.RequestContext(), query.List{{.EntityPlural}}{
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success({{.EntityLower}}resp.List{
		Total: total,
		Items: mapItems(items),
	}, requestID(c)))
}

func (h *Handler) Get(c coretransport.Context) {
	item, err := h.service.Get(c.RequestContext(), c.Param("id"))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) Create(c coretransport.Context) {
	var req {{.EntityLower}}req.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(coreerrors.Wrap(err, coreerrors.CodeBadRequest, "invalid create request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Create(c.RequestContext(), command.Create{{.Entity}}(req))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) Update(c coretransport.Context) {
	var req {{.EntityLower}}req.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(coreerrors.Wrap(err, coreerrors.CodeBadRequest, "invalid update request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Update(c.RequestContext(), c.Param("id"), command.Update{{.Entity}}(req))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) Delete(c coretransport.Context) {
	if err := h.service.Delete(c.RequestContext(), c.Param("id")); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(map[string]any{"deleted": true}, requestID(c)))
}

func requestID(c coretransport.Context) string {
	if value, exists := c.Get(coremiddleware.RequestIDContextKey); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

func mapItem(item model.{{.Entity}}) {{.EntityLower}}resp.Item {
	return {{.EntityLower}}resp.Item{
{{.ResponseAssignments}}	}
}

func mapItems(items []model.{{.Entity}}) []{{.EntityLower}}resp.Item {
	result := make([]{{.EntityLower}}resp.Item, 0, len(items))
	for _, item := range items {
		result = append(result, mapItem(item))
	}
	return result
}
`

const routerTemplate = `package http

import (
	coretransport "goadmin/core/transport"
	{{.EntityLower}}service "goadmin/modules/{{.EntityLower}}/application/service"
	"goadmin/modules/{{.EntityLower}}/transport/http/handler"
	"go.uber.org/zap"
)

type Dependencies struct {
	Service *{{.EntityLower}}service.Service
	Logger  *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/{{.EntityPlural}}")
	root.GET("", h.List)
	root.GET("/:id", h.Get)
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
}
`

const frontendApiTemplate = `import http from './http';

const basePath = '/{{.EntityPlural}}'

export function list{{.EntityPlural}}(params = {}) {
  return http.get(basePath, { params });
}

export function get{{.Entity}}(id) {
  return http.get(basePath + '/' + id);
}

export function create{{.Entity}}(data) {
  return http.post(basePath, data);
}

export function update{{.Entity}}(id, data) {
  return http.put(basePath + '/' + id, data);
}

export function delete{{.Entity}}(id) {
  return http.delete(basePath + '/' + id);
}
`

const frontendRouterTemplate = `const route = {
  path: '/{{.EntityPlural}}',
  name: '{{.Entity}}',
  component: () => import('@/views/{{.EntityLower}}/index.vue'),
  meta: {
    title: '{{.Entity}}',
    icon: 'menu',
  },
}

export default route
`

const frontendViewTemplate = `<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { create{{.Entity}}, delete{{.Entity}}, list{{.EntityPlural}}, update{{.Entity}} } from '@/api/{{.EntityLower}}';
import { formatDateTime } from '@/utils/admin';

type {{.Entity}}Item = {
  id: string;
{{- range .DisplayFields}}
  {{.JSONName}}?: {{.TSValueType}};
{{- end}}
  created_at?: string;
  updated_at?: string;
};

type {{.Entity}}FormState = {
{{- range .FormFields}}
  {{.JSONName}}: {{.TSFormValueType}};
{{- end}}
};

const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const rows = ref<{{.Entity}}Item[]>([]);
const total = ref(0);
const editingId = ref('');

const query = reactive({
  keyword: '',
  page: 1,
  page_size: 10,
});

const defaultForm = (): {{.Entity}}FormState => ({
{{- range .FormFields}}
  {{.JSONName}}: {{.FormDefaultValue}},
{{- end}}
});

const form = reactive<{{.Entity}}FormState>(defaultForm());

function resetForm() {
  Object.assign(form, defaultForm());
}

async function loadItems() {
  tableLoading.value = true;
  try {
    const response = await list{{.EntityPlural}}({ ...query });
    rows.value = response.items ?? [];
    total.value = response.total ?? 0;
  } finally {
    tableLoading.value = false;
  }
}

function openCreate() {
  editingId.value = '';
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: {{.Entity}}Item) {
  editingId.value = row.id;
  Object.assign(form, {
{{- range .FormFields}}
    {{.JSONName}}: {{.FrontendEditExpression}},
{{- end}}
  });
  dialogVisible.value = true;
}

async function submitForm() {
  dialogLoading.value = true;
  try {
    const payload: {{.Entity}}FormState = {
{{- range .FormFields}}
      {{.JSONName}}: {{.FrontendSubmitExpression}},
{{- end}}
    };

    if (editingId.value) {
      await update{{.Entity}}(editingId.value, payload);
      ElMessage.success('{{.Entity}} 已更新');
    } else {
      await create{{.Entity}}(payload);
      ElMessage.success('{{.Entity}} 已创建');
    }

    dialogVisible.value = false;
    await loadItems();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存失败');
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: {{.Entity}}Item) {
  await ElMessageBox.confirm('确认删除 {{.Entity}} ' + row.id + ' 吗？', '删除{{.Entity}}', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消',
  });
  await delete{{.Entity}}(row.id);
  ElMessage.success('{{.Entity}} 已删除');
  await loadItems();
}

function handleSearch() {
  query.page = 1;
  void loadItems();
}

function handleReset() {
  query.keyword = '';
  query.page = 1;
  void loadItems();
}

function handlePageChange(page: number) {
  query.page = page;
  void loadItems();
}

function handleSizeChange(pageSize: number) {
  query.page_size = pageSize;
  query.page = 1;
  void loadItems();
}

onMounted(() => {
  void loadItems();
});
</script>

<template>
  <div class="admin-page">
    <AdminTable
      title="{{.Entity}}管理"
      description="由 goadmin-cli 生成的 CRUD 页面，可直接用于列表、编辑和删除。"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadItems">刷新</el-button>
        <el-button v-permission="'{{.EntityLower}}:create'" type="primary" @click="openCreate">新增{{.Entity}}</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item label="关键字">
            <el-input v-model="query.keyword" clearable placeholder="搜索{{.Entity}}数据" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="id" label="ID" min-width="160" />
{{- range .DisplayFields}}
        <el-table-column
          prop="{{.JSONName}}"
          label="{{.DisplayLabel}}"
          min-width="{{if eq .FrontendControl "datetime"}}180{{else if eq .FrontendControl "number"}}120{{else if eq .FrontendControl "textarea"}}220{{else}}140{{end}}"
{{- if or (eq .FrontendControl "input") (eq .FrontendControl "textarea")}}
          show-overflow-tooltip
{{- end}}
        >
          <template #default="{ row }">
            {{.FrontendDisplayExpression}}
          </template>
        </el-table-column>
{{- end}}
        <el-table-column label="创建时间" min-width="180">
          <template #default="{ row }">
            {{"{{"}} formatDateTime(row.created_at) {{"}}"}}
          </template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="180">
          <template #default="{ row }">
            {{"{{"}} formatDateTime(row.updated_at) {{"}}"}}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'{{.EntityLower}}:update'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'{{.EntityLower}}:delete'" link type="danger" @click="removeRow(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
        <div class="admin-pagination">
          <el-pagination
            background
            layout="total, sizes, prev, pager, next, jumper"
            :total="total"
            :current-page="query.page"
            :page-size="query.page_size"
            :page-sizes="[10, 20, 50, 100]"
            @current-change="handlePageChange"
            @size-change="handleSizeChange"
          />
        </div>
      </template>
    </AdminTable>

    <AdminFormDialog
      v-model="dialogVisible"
      :title="editingId ? '编辑{{.Entity}}' : '新增{{.Entity}}'"
      :loading="dialogLoading"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
{{- range .FormFields}}
        <el-form-item label="{{.DisplayLabel}}">
{{- if eq .FrontendControl "datetime"}}
          <el-date-picker
            v-model="form.{{.JSONName}}"
            type="datetime"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            placeholder="请选择{{.DisplayLabel}}"
            style="width: 100%"
          />
{{- else if eq .FrontendControl "switch"}}
          <el-switch v-model="form.{{.JSONName}}" inline-prompt active-text="是" inactive-text="否" />
{{- else if eq .FrontendControl "number"}}
          <el-input-number v-model="form.{{.JSONName}}" :controls="false" style="width: 100%" />
{{- else if eq .FrontendControl "textarea"}}
          <el-input v-model="form.{{.JSONName}}" type="textarea" :rows="4" :placeholder="'请输入{{.DisplayLabel}}'" />
{{- else}}
          <el-input v-model="form.{{.JSONName}}" :placeholder="'请输入{{.DisplayLabel}}'" />
{{- end}}
        </el-form-item>
{{- end}}
      </el-form>
    </AdminFormDialog>
  </div>
</template>
`

const pluginTemplate = `package {{.PackageName}}

import (
	"fmt"
	"net/http"

	coretransport "goadmin/core/transport"
	pluginiface "goadmin/plugin/interface"
)

type Plugin struct{}

func New() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "{{.EntityLower}}"
}

func (p *Plugin) Register(ctx *pluginiface.Context, registrar pluginiface.Registrar) error {
	if registrar == nil {
		return fmt.Errorf("plugin registrar is required")
	}
	if err := registrar.AddRoute(pluginiface.Route{
		Name:   "{{.EntityLower}}Ping",
		Method: http.MethodGet,
		Path:   "{{.RoutePrefix}}/ping",
		Access: pluginiface.AccessPublic,
		Handler: func(c coretransport.Context) {
			c.JSON(http.StatusOK, map[string]any{
				"message": "pong from {{.EntityLower}} plugin",
				"plugin":  "{{.EntityLower}}",
			})
		},
	}); err != nil {
		return err
	}

	if err := registrar.AddMenu(pluginiface.Menu{
		Plugin:     "{{.EntityLower}}",
		ID:         "{{.EntityLower}}-root",
		Name:       "{{.Title}}",
		Path:       "/plugin/{{.EntityLower}}",
		Component:  "Layout",
		Icon:       "plug",
		Sort:       100,
		Permission: "plugin:{{.EntityLower}}:view",
		Type:       pluginiface.MenuTypeDirectory,
		Visible:    true,
		Enabled:    true,
		Redirect:   "/plugin/{{.EntityLower}}/home",
	}); err != nil {
		return err
	}

	if err := registrar.AddMenu(pluginiface.Menu{
		Plugin:     "{{.EntityLower}}",
		ID:         "{{.EntityLower}}-home",
		ParentID:   "{{.EntityLower}}-root",
		Name:       "Home",
		Path:       "/plugin/{{.EntityLower}}/home",
		Component:  "{{.ViewPath}}",
		Icon:       "sparkles",
		Sort:       1,
		Permission: "plugin:{{.EntityLower}}:view",
		Type:       pluginiface.MenuTypeMenu,
		Visible:    true,
		Enabled:    true,
	}); err != nil {
		return err
	}

	return registrar.AddPermission(pluginiface.Permission{
		Plugin:      "{{.EntityLower}}",
		Object:      "plugin:{{.EntityLower}}",
		Action:      "view",
		Description: "View {{.Title}} plugin",
	})
}
`

const pluginFrontendTemplate = `export const plugin{{.Title}} = {
  name: '{{.EntityLower}}',
  routePrefix: '{{.RoutePrefix}}',
  viewPath: '{{.ViewPath}}',
}
`

const pluginViewTemplate = `<template>
  <div class="page-container">
    <h1>{{.Title}} Plugin</h1>
    <p>This plugin page is generated by goadmin-cli.</p>
  </div>
</template>

<script setup>
</script>
`

const pageViewTemplate = `<template>
  <div class="page-container">
    <h1>{{.Title}}</h1>
    <p>This page is generated by goadmin-cli.</p>
  </div>
</template>

<script setup>
</script>
`

const pageRouterTemplate = `const route = {
  path: '{{.RoutePath}}',
  name: '{{.RouteName}}',
  component: () => import('@/views/{{.Component}}.vue'),
  meta: {
    title: '{{.Title}}',
    permission: '{{.Permission}}',
  },
}

export default route
`
