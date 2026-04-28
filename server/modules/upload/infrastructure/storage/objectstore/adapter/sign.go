package adapter

import (
	"net/url"
	"strings"

	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"
)

func buildSignedURL(base string, opts storagecontract.SignedURLOptions) string {
	values := url.Values{}
	if method := strings.TrimSpace(opts.Method); method != "" {
		values.Set("method", strings.ToUpper(method))
	}
	if opts.Expires > 0 {
		values.Set("expires", opts.Expires.String())
	}
	if contentType := strings.TrimSpace(opts.ResponseContentType); contentType != "" {
		values.Set("response_content_type", contentType)
	}
	if disposition := strings.TrimSpace(opts.ResponseContentDisposition); disposition != "" {
		values.Set("response_content_disposition", disposition)
	}
	if len(values) == 0 {
		return base
	}
	return base + "?" + values.Encode()
}
