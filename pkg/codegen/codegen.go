package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/dimonoff/asyncapi-codegen/pkg/asyncapi"
	"github.com/dimonoff/asyncapi-codegen/pkg/asyncapi/parser"
	asyncapiv2 "github.com/dimonoff/asyncapi-codegen/pkg/asyncapi/v2"
	asyncapiv3 "github.com/dimonoff/asyncapi-codegen/pkg/asyncapi/v3"
	generatorv2 "github.com/dimonoff/asyncapi-codegen/pkg/codegen/generators/v2"
	templatesv2 "github.com/dimonoff/asyncapi-codegen/pkg/codegen/generators/v2/templates"
	generatorv3 "github.com/dimonoff/asyncapi-codegen/pkg/codegen/generators/v3"
	templatesv3 "github.com/dimonoff/asyncapi-codegen/pkg/codegen/generators/v3/templates"
	"github.com/dimonoff/asyncapi-codegen/pkg/codegen/options"
	"github.com/dimonoff/asyncapi-codegen/pkg/utils/template"
	"golang.org/x/tools/imports"
)

// CodeGen is the main structure for the code generation.
type CodeGen struct {
	Specification asyncapi.Specification
	modulePath    string
	moduleVersion string
}

// FromFile returns a code generator from a Specification file path.
func FromFile(path string, dependencies ...string) (CodeGen, error) {
	// Get Specification from file
	spec, err := parser.FromFile(parser.FromFileParams{
		Path: path,
	})
	if err != nil {
		return CodeGen{}, err
	}

	// Get dependencies
	for _, path := range dependencies {
		dep, err := parser.FromFile(parser.FromFileParams{
			Path:         path,
			MajorVersion: spec.MajorVersion(),
		})
		if err != nil {
			return CodeGen{}, err
		}

		if err := spec.AddDependency(path, dep); err != nil {
			return CodeGen{}, err
		}
	}

	return New(spec)
}

// New creates a new code generation structure that can be used to generate code.
func New(spec asyncapi.Specification) (CodeGen, error) {
	modulePath, moduleVersion := modulePathVersion()

	return CodeGen{
		Specification: spec,
		modulePath:    modulePath,
		moduleVersion: moduleVersion,
	}, nil
}

func modulePathVersion() (path, version string) {
	path = "unknown module path"
	version = "unknown version"
	if bi, ok := debug.ReadBuildInfo(); ok {
		if bi.Main.Path != "" {
			path = bi.Main.Path
		}
		if bi.Main.Version != "" {
			version = bi.Main.Version
		}
	}

	return path, version
}

// Generate generates code from the code generation structure, that have already
// processed the AsyncAPI file when creating it.
//
// Output behaviour
//
//   - If `opt.OutputPath` ends in `.go`, the generator runs in legacy
//     single-file mode: every enabled `--generate` category is concatenated
//     into one Go file written verbatim at that path.
//   - Otherwise `opt.OutputPath` is treated as a directory (created if it
//     does not exist) and the generator emits one file per category:
//     `types.gen.go`, `app.gen.go`, `user.gen.go`. Each file gets its own
//     package/import header and is formatted independently with goimports,
//     so unused imports are pruned per file.
func (cg CodeGen) Generate(opt options.Options) error {
	if err := template.SetConvertKeyFn(opt.ConvertKeys); err != nil {
		return err
	}

	if err := template.SetNamifyFn(opt.NamingScheme); err != nil {
		return err
	}

	if opt.IgnoreStringFormat {
		template.DisableDateOrTimeGeneration()
	}
	if opt.ForcePointers {
		templatesv2.ForcePointerOnFields()
		templatesv3.ForcePointerOnFields()
	}

	// Process Specification
	if err := cg.Specification.Process(); err != nil {
		return err
	}

	// Single-file legacy mode (kept so existing `//go:generate` directives
	// targeting `*.gen.go` continue to work unchanged).
	if strings.HasSuffix(opt.OutputPath, ".go") {
		return cg.generateSingleFile(opt)
	}

	return cg.generateMultiFile(opt)
}

func (cg CodeGen) generateSingleFile(opt options.Options) error {
	content, err := cg.generateContent(opt)
	if err != nil {
		return err
	}

	fileContent, err := formatGoSource(content, opt.DisableFormatting)
	if err != nil {
		return err
	}

	return os.WriteFile(opt.OutputPath, fileContent, 0o644)
}

func (cg CodeGen) generateMultiFile(opt options.Options) error {
	// Ensure the output directory exists.
	if err := os.MkdirAll(opt.OutputPath, 0o755); err != nil {
		return fmt.Errorf("create output directory %q: %w", opt.OutputPath, err)
	}

	header, parts, err := cg.generateParts(opt)
	if err != nil {
		return err
	}

	// Sort keys to make the output order deterministic.
	names := make([]string, 0, len(parts))
	for name := range parts {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		body := parts[name]
		if strings.TrimSpace(body) == "" {
			continue
		}

		fileContent, err := formatGoSource(header+body, opt.DisableFormatting)
		if err != nil {
			return fmt.Errorf("format %s: %w", name, err)
		}

		dst := filepath.Join(opt.OutputPath, name+".gen.go")
		if err := os.WriteFile(dst, fileContent, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dst, err)
		}
	}

	return nil
}

func formatGoSource(content string, disableFormatting bool) ([]byte, error) {
	if disableFormatting {
		return []byte(content), nil
	}
	return imports.Process("", []byte(content), &imports.Options{
		TabWidth:  8,
		TabIndent: true,
		Comments:  true,
		Fragment:  true,
	})
}

// generateParts returns the package/imports header (shared by every file) and
// a map of category-name → body for the per-file output mode.
func (cg CodeGen) generateParts(opt options.Options) (header string, parts map[string]string, err error) {
	switch v := cg.Specification.MajorVersion(); v {
	case 2:
		spec, err := asyncapiv2.FromUnknownVersion(cg.Specification)
		if err != nil {
			return "", nil, err
		}
		gen := generatorv2.Generator{
			Specification: *spec,
			Options:       opt,
			ModulePath:    cg.modulePath,
			ModuleVersion: cg.moduleVersion,
		}
		header, err = gen.GenerateImports()
		if err != nil {
			return "", nil, err
		}
		parts, err = gen.GenerateParts()
		return header, parts, err
	case 3:
		spec, err := asyncapiv3.FromUnknownVersion(cg.Specification)
		if err != nil {
			return "", nil, err
		}
		gen := generatorv3.Generator{
			Specification: *spec,
			Options:       opt,
			ModulePath:    cg.modulePath,
			ModuleVersion: cg.moduleVersion,
		}
		header, err = gen.GenerateImports()
		if err != nil {
			return "", nil, err
		}
		parts, err = gen.GenerateParts()
		return header, parts, err
	default:
		return "", nil, fmt.Errorf("unsupported major version (%q)", v)
	}
}

func (cg CodeGen) generateContent(opt options.Options) (string, error) {
	version := cg.Specification.MajorVersion()
	switch version {
	case 2:
		spec, err := asyncapiv2.FromUnknownVersion(cg.Specification)
		if err != nil {
			return "", err
		}

		return generatorv2.Generator{
			Specification: *spec,
			Options:       opt,
			ModulePath:    cg.modulePath,
			ModuleVersion: cg.moduleVersion,
		}.Generate()
	case 3:
		spec, err := asyncapiv3.FromUnknownVersion(cg.Specification)
		if err != nil {
			return "", err
		}

		return generatorv3.Generator{
			Specification: *spec,
			Options:       opt,
			ModulePath:    cg.modulePath,
			ModuleVersion: cg.moduleVersion,
		}.Generate()
	default:
		return "", fmt.Errorf("unsupported major version (%q)", version)
	}
}
