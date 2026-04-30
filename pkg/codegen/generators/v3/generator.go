package generatorv3

import (
	"fmt"

	asyncapi "github.com/dimonoff/asyncapi-codegen/pkg/asyncapi/v3"
	"github.com/dimonoff/asyncapi-codegen/pkg/codegen/generators"
	"github.com/dimonoff/asyncapi-codegen/pkg/codegen/options"
)

// Generator is the structure that contains information to generate the code from
// the specification.
type Generator struct {
	Options       options.Options
	Specification asyncapi.Specification
	ModulePath    string
	ModuleVersion string
}

// Generate generates the source code from the specification.
func (g Generator) Generate() (string, error) {
	content, err := g.generateImports(g.Options)
	if err != nil {
		return "", err
	}

	for remainingParts, part := true, ""; remainingParts; part = "" {
		switch {
		case g.Options.Generate.Application:
			part, err = g.generateApp()
			g.Options.Generate.Application = false
		case g.Options.Generate.User:
			part, err = g.generateUser()
			g.Options.Generate.User = false
		case g.Options.Generate.Types:
			part, err = g.generateTypes()
			g.Options.Generate.Types = false
		default:
			remainingParts = false
		}

		if err != nil {
			return "", err
		}

		content += part
	}

	return content, nil
}

// PartTypes is the multi-file output key for "core" type definitions
// (controller, options, error, channel parameters, channel-path constants).
const PartTypes = "types"

// PartMessages is the multi-file output key for the message section
// (channel-scoped message comments + message structs declared under
// `components.messages`).
const PartMessages = "messages"

// PartSchemas is the multi-file output key for the schema section
// (Go types generated from `components.schemas`).
const PartSchemas = "schemas"

// PartApp is the multi-file output key for application-side code
// (the App controller and its subscriber interface).
const PartApp = "app"

// PartUser is the multi-file output key for user-side code
// (the User controller and its subscriber interface).
const PartUser = "user"

// GenerateImports renders the standard imports/package header used by every
// generated file. Each file emitted in directory mode starts with this block;
// `goimports` then prunes unused imports per file.
func (g Generator) GenerateImports() (string, error) {
	return g.generateImports(g.Options)
}

// GenerateParts returns the body of each enabled generation category, keyed
// by PartTypes / PartMessages / PartSchemas / PartApp / PartUser. The
// returned strings DO NOT include the package/import header — callers (e.g.
// the multi-file writer in pkg/codegen) are expected to prepend the result
// of GenerateImports.
//
// When `--generate types` is enabled, the type output is split into three
// per-section files instead of being written as one monolithic file:
//
//   - PartTypes    → core (controller, options, channel params, paths)
//   - PartMessages → message structs and broker codecs
//   - PartSchemas  → schema definitions from components.schemas
//
// This keeps the legacy single-file mode unchanged (it still uses the
// monolithic `Generate()` method) while giving directory-mode users a more
// navigable layout.
func (g Generator) GenerateParts() (map[string]string, error) {
	parts := map[string]string{}

	if g.Options.Generate.Types {
		tg := TypesGenerator{Specification: g.Specification}

		core, err := tg.GenerateCore()
		if err != nil {
			return nil, err
		}
		parts[PartTypes] = core

		msgs, err := tg.GenerateMessages()
		if err != nil {
			return nil, err
		}
		parts[PartMessages] = msgs

		schemas, err := tg.GenerateSchemas()
		if err != nil {
			return nil, err
		}
		parts[PartSchemas] = schemas
	}

	if g.Options.Generate.Application {
		body, err := g.generateApp()
		if err != nil {
			return nil, err
		}
		parts[PartApp] = body
	}

	if g.Options.Generate.User {
		body, err := g.generateUser()
		if err != nil {
			return nil, err
		}
		parts[PartUser] = body
	}

	return parts, nil
}

func (g Generator) generateImports(opts options.Options) (string, error) {
	imps, err := g.Specification.CustomImports()
	if err != nil {
		return "", fmt.Errorf("failed to generate custom imports: %w", err)
	}

	return ImportsGenerator{
		PackageName:   opts.PackageName,
		ModuleVersion: g.ModuleVersion,
		ModuleName:    g.ModulePath,
		CustomImports: imps,
	}.Generate()
}

func (g Generator) generateTypes() (string, error) {
	return TypesGenerator{Specification: g.Specification}.Generate()
}

func (g Generator) generateApp() (string, error) {
	var content string

	// Generate application listener
	listener, err := NewSubscriberGenerator(
		generators.SideIsApplication,
		g.Specification,
	).Generate()
	if err != nil {
		return "", err
	}
	content += listener

	// Generate application controller
	controller, err := NewControllerGenerator(
		generators.SideIsApplication,
		g.Specification,
	).Generate()
	if err != nil {
		return "", err
	}
	content += controller

	return content, nil
}

func (g Generator) generateUser() (string, error) {
	var content string

	// Generate user listener
	listener, err := NewSubscriberGenerator(
		generators.SideIsUser,
		g.Specification,
	).Generate()
	if err != nil {
		return "", err
	}
	content += listener
	// Generate user controller
	controller, err := NewControllerGenerator(
		generators.SideIsUser,
		g.Specification,
	).Generate()
	if err != nil {
		return "", err
	}
	content += controller

	return content, nil
}
