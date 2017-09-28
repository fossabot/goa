package design

import (
	"net/url"
	"regexp"
	"sort"
	"strings"

	"goa.design/goa/design"
	"goa.design/goa/eval"
)

var (
	// Root holds the root expression built on process initialization.
	Root = &RootExpr{Design: design.Root}

	// WildcardRegex is the regular expression used to capture path
	// parameters.
	WildcardRegex = regexp.MustCompile(`/{\*?([a-zA-Z0-9_]+)}`)

	// ErrorResult is the built-in result type for error responses.
	ErrorResult = design.ErrorResult
)

const (
	// DefaultView is the name of the default view.
	DefaultView = "default"
)

const (
	// Boolean is the type for a JSON boolean.
	Boolean = design.Boolean

	// Int is the type for a signed integer.
	Int = design.Int

	// Int32 is the type for a signed 32-bit integer.
	Int32 = design.Int32

	// Int64 is the type for a signed 64-bit integer.
	Int64 = design.Int64

	// UInt is the type for a signed integer.
	UInt = design.UInt

	// UInt32 is the type for an unsigned 32-bit integer.
	UInt32 = design.UInt32

	// UInt64 is the type for an unsigned 64-bit integer.
	UInt64 = design.UInt64

	// Float32 is the type for a 32-bit floating number.
	Float32 = design.Float32

	// Float64 is the type for a 64-bit floating number.
	Float64 = design.Float64

	// String is the type for a JSON string.
	String = design.String

	// Bytes is the type for binary data.
	Bytes = design.Bytes

	// Any is the type for an arbitrary JSON value (interface{} in Go).
	Any = design.Any
)

// Empty represents empty values.
var Empty = design.Empty

type (
	// ParamHolder is the interface implemented by the design data structures
	// that represent HTTP constructs with path and query string parameters.
	ParamHolder interface {
		eval.Expression
		// Params returns the attribute holding the parameters. It takes
		// care of initializing the underlying struct field so that it
		// never returns nil.
		Params() *design.AttributeExpr
	}

	// HeaderHolder is the interface implemented by the design data
	// structures that represent HTTP constructs with HTTP headers.
	HeaderHolder interface {
		eval.Expression
		// Headers returns the attribute holding the headers. It takes
		// care of initializing the underlying struct field so that it
		// never returns nil.
		Headers() *design.AttributeExpr
	}

	// RootExpr is the data structure built by the top level HTTP DSL.
	RootExpr struct {
		// Design is the transport agnostic root expression.
		Design *design.RootExpr
		// Path is the common request path prefix to all the service
		// HTTP endpoints.
		Path string
		// Consumes lists the mime types supported by the API
		// controllers.
		Consumes []string
		// Produces lists the mime types generated by the API
		// controllers.
		Produces []string
		// HTTPServices contains the services created by the DSL.
		HTTPServices []*ServiceExpr
		// HTTPErrors lists the error HTTP responses.
		HTTPErrors []*ErrorExpr
		// Metadata is a set of key/value pairs with semantic that is
		// specific to each generator.
		Metadata design.MetadataExpr
		// params defines common request parameters to all the service
		// HTTP endpoints. The keys may use the "attribute:param" syntax
		// where "attribute" is the name of the attribute and "param"
		// the name of the HTTP parameter.
		params *design.AttributeExpr
		// headers defines common headers to all the service HTTP
		// endpoints. The keys may use the "attribute:header" syntax
		// where "attribute" is the name of the attribute and "header"
		// the name of the HTTP header.
		headers *design.AttributeExpr
	}
)

// Schemes returns the list of HTTP schemes used by the API servers.
func (r *RootExpr) Schemes() []string {
	if r.Design == nil {
		return nil
	}
	schemes := make(map[string]bool)
	for _, s := range r.Design.API.Servers {
		if u, err := url.Parse(s.URL); err != nil {
			schemes[u.Scheme] = true
		}
	}
	if len(schemes) == 0 {
		return nil
	}
	ss := make([]string, len(schemes))
	i := 0
	for s := range schemes {
		ss[i] = s
		i++
	}
	sort.Strings(ss)
	return ss
}

// Service returns the service with the given name if any.
func (r *RootExpr) Service(name string) *ServiceExpr {
	for _, res := range r.HTTPServices {
		if res.Name() == name {
			return res
		}
	}
	return nil
}

// ServiceFor creates a new or returns the existing service definition for the
// given service.
func (r *RootExpr) ServiceFor(s *design.ServiceExpr) *ServiceExpr {
	if res := r.Service(s.Name); res != nil {
		return res
	}
	res := &ServiceExpr{
		ServiceExpr: s,
	}
	r.HTTPServices = append(r.HTTPServices, res)
	return res
}

// Headers initializes and returns the attribute holding the API headers.
func (r *RootExpr) Headers() *design.AttributeExpr {
	if r.headers == nil {
		r.headers = &design.AttributeExpr{Type: &design.Object{}}
	}
	return r.headers
}

// MappedHeaders computes the mapped attribute expression from Headers.
func (r *RootExpr) MappedHeaders() *design.MappedAttributeExpr {
	return design.NewMappedAttributeExpr(r.headers)
}

// Params initializes and returns the attribute holding the API parameters.
func (r *RootExpr) Params() *design.AttributeExpr {
	if r.params == nil {
		r.params = &design.AttributeExpr{Type: &design.Object{}}
	}
	return r.params
}

// MappedParams computes the mapped attribute expression from Params.
func (r *RootExpr) MappedParams() *design.MappedAttributeExpr {
	return design.NewMappedAttributeExpr(r.params)
}

// EvalName is the expression name used by the evaluation engine to display
// error messages.
func (r *RootExpr) EvalName() string {
	return "API HTTP"
}

// WalkSets iterates through the service to finalize and validate them.
func (r *RootExpr) WalkSets(walk eval.SetWalker) {
	var (
		services  eval.ExpressionSet
		endpoints eval.ExpressionSet
		servers   eval.ExpressionSet
	)
	{
		services = make(eval.ExpressionSet, len(r.HTTPServices))
		sort.SliceStable(r.HTTPServices, func(i, j int) bool {
			if r.HTTPServices[j].ParentName == r.HTTPServices[i].Name() {
				return true
			}
			return false
		})
		for i, svc := range r.HTTPServices {
			services[i] = svc
			for _, e := range svc.HTTPEndpoints {
				endpoints = append(endpoints, e)
			}
			for _, s := range svc.FileServers {
				servers = append(servers, s)
			}
		}
	}
	walk(services)
	walk(endpoints)
	walk(servers)
}

// DependsOn is a no-op as the DSL runs when loaded.
func (r *RootExpr) DependsOn() []eval.Root { return nil }

// Packages returns the Go import path to this and the dsl packages.
func (r *RootExpr) Packages() []string {
	return []string{
		"goa.design/goa/http/design",
		"goa.design/goa/http/dsl",
	}
}

// ExtractWildcards returns the names of the wildcards that appear in path.
func ExtractWildcards(path string) []string {
	matches := WildcardRegex.FindAllStringSubmatch(path, -1)
	wcs := make([]string, len(matches))
	for i, m := range matches {
		wcs[i] = m[1]
	}
	return wcs
}

// NameMap returns the attribute and HTTP element name encoded in the given
// string. The encoding uses a simple "attribute:element" notation which allows
// to map header or body field names to underlying attributes. The second
// element of the encoding is optional in which case both the element and
// attribute have the same name.
func NameMap(encoded string) (string, string) {
	elems := strings.Split(encoded, ":")
	attName := elems[0]
	name := attName
	if len(elems) > 1 {
		name = elems[1]
	}
	return attName, name
}
