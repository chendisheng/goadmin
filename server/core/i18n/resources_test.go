package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadResourceRootScansNestedLocalesDirectories(t *testing.T) {
	registry := NewRegistry(LanguageZHCN)
	root := t.TempDir()

	writeLocaleTestFile(t, filepath.Join(root, "core", "i18n", "locales", "zh-CN", "common.yaml"), `
translation:
  common.ok: 成功
  auth.login.title: 登录
  nested:
    value: 嵌套值
`)
	writeLocaleTestFile(t, filepath.Join(root, "plugin", "builtin", "example", "locales", "en-US", "plugin.json"), `{
  "translation": {
    "plugin.example.title": "Example Plugin",
    "plugin.example.description": "Plugin description"
  }
}`)

	if err := registry.LoadResourceRoot(root); err != nil {
		t.Fatalf("LoadResourceRoot returned error: %v", err)
	}

	if got, ok := registry.TranslateLanguage(LanguageZHCN, "common.ok"); !ok || got != "成功" {
		t.Fatalf("TranslateLanguage(zh-CN, common.ok) = %q, %v; want 成功, true", got, ok)
	}
	if got, ok := registry.TranslateLanguage(LanguageZHCN, "auth.login.title"); !ok || got != "登录" {
		t.Fatalf("TranslateLanguage(zh-CN, auth.login.title) = %q, %v; want 登录, true", got, ok)
	}
	if got, ok := registry.TranslateLanguage(LanguageZHCN, "nested.value"); !ok || got != "嵌套值" {
		t.Fatalf("TranslateLanguage(zh-CN, nested.value) = %q, %v; want 嵌套值, true", got, ok)
	}
	if got, ok := registry.TranslateLanguage(LanguageENUS, "plugin.example.title"); !ok || got != "Example Plugin" {
		t.Fatalf("TranslateLanguage(en-US, plugin.example.title) = %q, %v; want Example Plugin, true", got, ok)
	}
	if got, ok := registry.TranslateLanguage(LanguageENUS, "plugin.example.description"); !ok || got != "Plugin description" {
		t.Fatalf("TranslateLanguage(en-US, plugin.example.description) = %q, %v; want Plugin description, true", got, ok)
	}
}

func TestLoadRepositoryModuleLocales(t *testing.T) {
	registry := NewRegistry(LanguageZHCN)
	root := filepath.Join("..", "..")

	if err := registry.LoadResourceRoot(root); err != nil {
		t.Fatalf("LoadResourceRoot returned error: %v", err)
	}

	for _, tc := range []struct {
		language string
		key      string
		want     string
	}{
		{LanguageZHCN, "user.username_required", "用户名不能为空"},
		{LanguageENUS, "user.username_required", "username is required"},
		{LanguageZHCN, "role.code_required", "角色编码不能为空"},
		{LanguageENUS, "role.code_required", "role code is required"},
		{LanguageZHCN, "menu.path_required", "菜单路径不能为空"},
		{LanguageENUS, "menu.path_required", "menu path is required"},
		{LanguageZHCN, "dictionary.item.label_required", "字典项标签不能为空"},
		{LanguageENUS, "dictionary.item.label_required", "dictionary item label is required"},
		{LanguageZHCN, "upload.extension_not_allowed", "上传文件扩展名不允许"},
		{LanguageENUS, "upload.extension_not_allowed", "upload extension is not allowed"},
		{LanguageZHCN, "casbin.summary.not_configured", "授权模块未配置"},
		{LanguageENUS, "casbin.summary.not_configured", "authorization module is not configured"},
		{LanguageZHCN, "book.not_found", "图书不存在"},
		{LanguageENUS, "book.not_found", "book not found"},
		{LanguageZHCN, "casbin_model.operation_failed", "授权模型操作失败"},
		{LanguageENUS, "casbin_model.operation_failed", "casbin_model operation failed"},
		{LanguageZHCN, "casbin_rule.service_not_configured", "授权策略服务未配置"},
		{LanguageENUS, "casbin_rule.service_not_configured", "casbin_rule service is not configured"},
	} {
		if got, ok := registry.TranslateLanguage(tc.language, tc.key); !ok || got != tc.want {
			t.Fatalf("TranslateLanguage(%q, %q) = %q, %v; want %q, true", tc.language, tc.key, got, ok, tc.want)
		}
	}
}

func writeLocaleTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
