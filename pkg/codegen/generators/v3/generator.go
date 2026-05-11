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
		case g.Options.Generate.Publisher:
			part, err = g.generatePublisher()
			g.Options.Generate.Publisher = false
		case g.Options.Generate.Subscriber:
			part, err = g.generateSubscriber()
			g.Options.Generate.Subscriber = false
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

// PartMessages is the multi-file output key for the message section.
const PartMessages = "messages"

// PartSchemas is the multi-file output key for the schema section.
const PartSchemas = "schemas"

// PartPublisher is the multi-file output key for publisher-side code
// (the Publisher controller and its handler interface).
const PartPublisher = "publisher"

// PartSubscriber is the multi-file output key for subscriber-side code
// (the Subscriber controller and its handler interface).
const PartSubscriber = "subscriber"

// GenerateImports renders the standard imports/package header used by every
// generated file.
func (g Generator) GenerateImports() (string, error) {
	return g.generateImports(g.Options)
}

// GenerateParts returns the body of each enabled generation category, keyed
// by PartTypes / PartMessages / PartSchemas / PartPublisher / PartSubscriber.
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

	if g.Options.Generate.Publisher {
		body, err := g.generatePublisher()
		if err != nil {
			return nil, err
		}
		parts[PartPublisher] = body
	}

	if g.Options.Generate.Subscriber {
		body, err := g.generateSubscriber()
		if err != nil {
			return nil, err
		}
		parts[PartSubscriber] = body
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

func (g Generator) generatePublisher() (string, error) {
	var content string

	// Generate publisher handler interface
	listener, err := NewSubscriberGenerator(
		generators.SideIsPublisher,
		g.Specification,
	).Generate()
	if err != nil {
		return "", err
	}
	content += listener

	// Generate publisher controller
	controller, err := NewControllerGenerator(
		generators.SideIsPublisher,
		g.Specification,
	).Generate()
	if err != nil {
		return "", err
	}
	content += controller

	return content, nil
}

func (g Generator) generateSubscriber() (string, error) {
	var content string

	// Generate subscriber handler interface
	listener, err := NewSubscriberGenerator(
		generators.SideIsSubscriber,
		g.Specification,
	).Generate()
	if err != nil {
		return "", err
	}
	content += listener

	// Generate subscriber controller
	controller, err := NewControllerGenerator(
		generators.SideIsSubscriber,
		g.Specification,
	).Generate()
	if err != nil {
		return "", err
	}
	content += controller

	return content, nil
}
