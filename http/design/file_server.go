package design

import (
	"fmt"
	"path"
	"strings"

	"goa.design/goa/design"
)

type (
	// FileServerExpr defines an endpoint that servers static assets.
	FileServerExpr struct {
		// Service is the parent service.
		Service *ServiceExpr
		// Description for docs
		Description string
		// Docs points to the service external documentation
		Docs *design.DocsExpr
		// FilePath is the file path to the static asset(s)
		FilePath string
		// RequestPath is the HTTP path that servers the assets.
		RequestPath string
		// Metadata is a list of key/value pairs
		Metadata design.MetadataExpr
	}
)

// EvalName returns the generic definition name used in error messages.
func (f *FileServerExpr) EvalName() string {
	suffix := fmt.Sprintf("file server %s", f.FilePath)
	var prefix string
	if f.Service != nil {
		prefix = f.Service.EvalName() + " "
	}
	return prefix + suffix
}

// Finalize normalizes the request path.
func (f *FileServerExpr) Finalize() {
	f.RequestPath = path.Join(Root.Path, f.Service.Path, f.RequestPath)
	// Make sure request path starts with a "/" so codegen can rely on it.
	if !strings.HasPrefix(f.RequestPath, "/") {
		f.RequestPath = "/" + f.RequestPath
	}
}

// IsDir returns true if the file server serves a directory, false otherwise.
func (f *FileServerExpr) IsDir() bool {
	return WildcardRegex.MatchString(f.RequestPath)
}
