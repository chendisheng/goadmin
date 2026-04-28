package i18n

import (
	"context"
	"testing"
)

func TestNormalizeLanguage(t *testing.T) {
	tests := map[string]string{
		" zh_cn ": "zh-CN",
		"en-GB":   "en-US",
		"":        "",
	}

	for input, want := range tests {
		if got := NormalizeLanguage(input); got != want {
			t.Fatalf("NormalizeLanguage(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	if got := ParseAcceptLanguage("fr-FR, en-US;q=0.0, zh-CN;q=1.0"); got != LanguageZHCN {
		t.Fatalf("ParseAcceptLanguage() = %q, want %q", got, LanguageZHCN)
	}

	if got := ParseAcceptLanguage("en;q=1.0, zh;q=0.5"); got != LanguageENUS {
		t.Fatalf("ParseAcceptLanguage() = %q, want %q", got, LanguageENUS)
	}

	if got := ResolveLanguage("fr-FR", LanguageENUS); got != LanguageENUS {
		t.Fatalf("ResolveLanguage() = %q, want %q", got, LanguageENUS)
	}
}

func TestRequestLanguageBinding(t *testing.T) {
	requestID := "req-1"
	BindRequestLanguage(requestID, "en-GB")
	t.Cleanup(func() { ClearRequestLanguage(requestID) })

	if got := RequestLanguage(requestID); got != LanguageENUS {
		t.Fatalf("RequestLanguage() = %q, want %q", got, LanguageENUS)
	}

	ctx := ContextWithLanguage(context.Background(), "zh_hant")
	if got, ok := LanguageFromContext(ctx); !ok || got != LanguageZHCN {
		t.Fatalf("LanguageFromContext() = %q, %v; want %q, true", got, ok, LanguageZHCN)
	}
}

func TestConfigureRuntimeLanguages(t *testing.T) {
	originalDefault := DefaultLanguage()
	originalSupported := SupportedLanguages()
	t.Cleanup(func() {
		Configure(originalDefault, originalSupported)
	})

	Configure(LanguageENUS, []string{LanguageENUS})

	if got := DefaultLanguage(); got != LanguageENUS {
		t.Fatalf("DefaultLanguage() = %q, want %q", got, LanguageENUS)
	}

	if got := SupportedLanguages(); len(got) != 1 || got[0] != LanguageENUS {
		t.Fatalf("SupportedLanguages() = %#v, want [%q]", got, LanguageENUS)
	}

	if got := ResolveLanguage(LanguageZHCN); got != LanguageENUS {
		t.Fatalf("ResolveLanguage() = %q, want %q", got, LanguageENUS)
	}
}
