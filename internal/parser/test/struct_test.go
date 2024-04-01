package test

import (
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang-cz/gospeak"
	"github.com/golang-cz/gospeak/internal/parser"
	"github.com/webrpc/webrpc/schema"
	"golang.org/x/tools/go/packages"
)

func parseTestStructCode(t *testing.T, inputFields string) *schema.Type {
	t.Helper()

	srcCode := fmt.Sprintf(`package test

	import (
		"context"
		"time"

		"github.com/golang-cz/gospeak/internal/parser/test/uuid"
		"github.com/golang-cz/gospeak/internal/parser/test/empty"
	)

	type TestStruct struct {
		%s
	}
	
	//go:webrpc json -out=/dev/null
	type TestAPI interface{
		Test(ctx context.Context) (tst *TestStruct, err error)
	}

	type Struct = empty.Struct // type alias

	type Number int // should be rendered as a number in JSON

	type Locale int // implements MarshalText(), should be rendered as a string in JSON

	// MarshalText implements encoding.TextMarshaler.
	func (locale Locale) MarshalText() ([]byte, error) {
		return []byte{}, nil
	}

	// UnmarshalText implements encoding.TextUnmarshaler.
	func (locale *Locale) UnmarshalText(data []byte) error {
		return nil
	}

	type Embedded struct {
		Number Number
	}

	// Ensure all the imports are used.
	var _ time.Time
	var _ uuid.UUID
	var _ Number
	var _ Locale
	`, inputFields)

	p, err := testParser(srcCode)
	if err != nil {
		t.Fatal(fmt.Errorf("error creating test parser: %w", err))
	}

	if err := parseTestStruct(p); err != nil {
		t.Fatal(fmt.Errorf("error parsing: %q: %w", inputFields, err))
	}

	for _, t := range p.Schema.Types {
		if t.Name != "TestStruct" {
			continue
		}

		return t
	}

	t.Fatal("couldn't find TestStruct type")
	return nil
}

func testParser(srcCode string) (*parser.Parser, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting working directory: %w", err)
	}

	pkg1 := filepath.Join(wd, "proto.go")
	pkg2 := filepath.Join(wd, "uuid/uuid.go")
	pkg3 := filepath.Join(wd, "empty/empty.go")

	cfg := &packages.Config{
		Dir:  wd,
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Overlay: map[string][]byte{
			pkg1: []byte(srcCode),
			pkg2: []byte(`
				package uuid

				type UUID [16]byte

				// MarshalText implements encoding.TextMarshaler.
				func (uuid UUID) MarshalText() ([]byte, error) {
					return []byte{}, nil
				}

				// UnmarshalText implements encoding.TextUnmarshaler.
				func (uuid *UUID) UnmarshalText(data []byte) error {
					return nil
				}
			`),
			pkg3: []byte(`
				package empty

				type Struct struct{}
			`),
		},
	}

	pkgs, err := packages.Load(cfg, "file="+pkg1, "file="+pkg2, "file="+pkg3)
	if err != nil {
		return nil, fmt.Errorf("error loading Go packages: %v\n%s", err, prefixLinesWithLineNumber(srcCode))
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("%v\n%s", spew.Sdump(pkg.Errors), prefixLinesWithLineNumber(srcCode))
		}
	}

	if len(pkgs) != 3 {
		return nil, fmt.Errorf("expected 2 Go packages, got %v\n%s", len(pkgs), spew.Sdump(pkgs))
	}

	pkg := pkgs[0]

	_, _ = gospeak.CollectInterfaces(pkg)

	p := parser.New(pkg)
	p.Schema.SchemaName = "TestAPI"
	p.Schema.SchemaVersion = "v0.0.1"

	return p, nil
}

func parseTestStruct(p *parser.Parser) error {
	scope := p.Pkg.Types.Scope()

	obj := scope.Lookup("TestStruct")
	if obj == nil {
		return fmt.Errorf("type TestStruct not defined")
	}

	testStruct, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		return fmt.Errorf("type TestStruct is %T", obj.Type().Underlying())
	}

	_, err := p.ParseStruct("TestStruct", testStruct)
	if err != nil {
		return fmt.Errorf("failed to parse struct TestStruct: %w", err)
	}

	return nil
}
