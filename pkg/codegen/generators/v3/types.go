package generatorv3

import (
	"bytes"

	asyncapi "github.com/dimonoff/asyncapi-codegen/pkg/asyncapi/v3"
)

// TypesGenerator is a code generator for types that will generate all schemas
// contained in an asyncapi specification to golang structures code.
type TypesGenerator struct {
	asyncapi.Specification
}

// Generate will create a new types code generator.
//
// This is the legacy entry-point used by the single-file output mode: it
// renders the original `types.tmpl` template, which concatenates the core
// types (controller, options, channel parameters, channel paths) together
// with all message structs and all schema definitions in one go. The output
// is unchanged so existing `//go:generate ... -o ./asyncapi.gen.go`
// directives keep producing byte-identical fixtures.
func (tg TypesGenerator) Generate() (string, error) {
	tmplt, err := loadTemplate(
		typesTemplatePath,
		schemaDefinitionTemplatePath,
		schemaNameTemplatePath,
		messageTemplatePath,

		marshalingAdditionalPropertiesTemplatePath,
		marshalingTimeTemplatePath,
	)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err := tmplt.Execute(buf, tg); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GenerateCore renders only the "core" type definitions: AsyncAPI version
// constant, the internal controller struct and its options, the
// MessageWithCorrelationID interface, the Error type, per-channel
// `*Parameters` structs and the channel-path constants.
//
// Used by the multi-file output mode to populate `types.gen.go`.
func (tg TypesGenerator) GenerateCore() (string, error) {
	return tg.renderSingle(typesCoreTemplatePath)
}

// GenerateMessages renders only the message section: the explanatory
// comment block for channel-scoped `$ref` messages plus every concrete
// message struct (with its `New…`, `brokerMessageTo…` and `toBrokerMessage`
// helpers) declared under `components.messages`.
//
// Used by the multi-file output mode to populate `messages.gen.go`.
func (tg TypesGenerator) GenerateMessages() (string, error) {
	return tg.renderSingle(typesMessagesTemplatePath)
}

// GenerateSchemas renders only the schema section: every Go type generated
// from `components.schemas` definitions.
//
// Used by the multi-file output mode to populate `schemas.gen.go`.
func (tg TypesGenerator) GenerateSchemas() (string, error) {
	return tg.renderSingle(typesSchemasTemplatePath)
}

// renderSingle is the shared rendering helper for the per-section
// templates. It loads the requested top-level template together with every
// helper template referenced from message / schema templates so partial
// rendering produces complete output.
func (tg TypesGenerator) renderSingle(top string) (string, error) {
	tmplt, err := loadTemplate(
		top,
		schemaDefinitionTemplatePath,
		schemaNameTemplatePath,
		messageTemplatePath,

		marshalingAdditionalPropertiesTemplatePath,
		marshalingTimeTemplatePath,
	)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err := tmplt.Execute(buf, tg); err != nil {
		return "", err
	}
	return buf.String(), nil
}
