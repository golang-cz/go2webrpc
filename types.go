package gospeak

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/webrpc/webrpc/schema"
)

func (p *parser) parseType(typ types.Type) (*schema.VarType, error) {
	return p.parseNamedType("", typ)
}

func (p *parser) parseNamedType(typeName string, typ types.Type) (varType *schema.VarType, err error) {
	// Return a parsedType from cache, if exists.
	if parsedType, ok := p.parsedTypes[typ]; ok {
		return parsedType, nil
	}

	// Otherwise, create a new parsedType record and warm up the cache up-front.
	// Claim the cache key and fill in the value later in defer(). Meanwhile, any
	// following recursive call(s) to this function (ie. on recursive types like
	// self-referencing structs, linked lists, graphs, circular dependencies etc.)
	// will return early with the same pointer.
	//
	// Note: We're parsing sequentially, no need for sync.Map.
	cacheDoNotReturn := &schema.VarType{
		Expr: typeName,
	}
	p.parsedTypes[typ] = cacheDoNotReturn

	defer func() {
		if varType != nil {
			*cacheDoNotReturn = *varType // Update the cache value via pointer dereference.
			varType = cacheDoNotReturn
		}
	}()

	switch v := typ.(type) {
	case *types.Named:
		pkg := v.Obj().Pkg()
		underlying := v.Underlying()

		//typeName := p.goTypeName(typ)
		typeName := p.goTypeName(v)

		if pkg != nil {
			if pkg.Path() == "time" && typeName == "Time" {
				return &schema.VarType{
					Expr: "timestamp",
					Type: schema.T_Timestamp,
				}, nil
			}

			underlying := v.Underlying()
			underlyingTypeName := p.goTypeName(underlying)

			if pkg.Path() == "github.com/golang-cz/gospeak" && strings.HasPrefix(typeName, "Enum[") && strings.HasSuffix(typeName, "]") {

				name := v.Obj().Id()

				enumElemType, err := p.parseNamedType(underlyingTypeName, underlying)
				if err != nil {
					return nil, fmt.Errorf("parsing gospeak.Enum underlying type: %w", err)
				}

				enumValues := []*schema.TypeField{}
				for _, file := range p.pkg.Syntax {
					for _, decl := range file.Decls {
						if typeDeclaration, ok := decl.(*ast.GenDecl); ok && typeDeclaration.Tok == token.TYPE {
							for _, spec := range typeDeclaration.Specs {
								if typeSpec, ok := spec.(*ast.TypeSpec); ok {
									//typeName := typeSpec.Name.Name
									if indent, ok := typeSpec.Type.(*ast.Ident); ok {
										typeName := indent.Name
										if typeName == "Enum" {
											doc := typeDeclaration.Doc
											if doc != nil {
												for i, comment := range doc.List {
													commentValue, _ := strings.CutPrefix(comment.Text, "//")
													name, value, found := strings.Cut(commentValue, "=") // approved = 0
													if !found {                                          // approved
														value = fmt.Sprintf("%v", i)
														name = commentValue
													}
													enumValues = append(enumValues, &schema.TypeField{
														Name: strings.TrimSpace(name),
														TypeExtra: schema.TypeExtra{
															Value: strings.TrimSpace(value),
														},
													})
												}
											}
										}
									}

								}
							}
						}
					}
				}

				enumType := &schema.Type{
					Kind:   schema.TypeKind_Enum,
					Name:   name,
					Type:   enumElemType,
					Fields: enumValues, // webrpc TODO: should be Enums
				}

				p.schema.Types = append(p.schema.Types, enumType)
				return &schema.VarType{
					Expr: name,
					Type: schema.T_Struct, // webrpc TODO: should be schema.T_Enum
					Struct: &schema.VarStructType{ // webrpc TODO: should be EnumType{}
						Name: name,
						Type: enumType,
					},
				}, nil
			}
		}

		// If the type implements encoding.TextMarshaler, it's a string.
		if isTextMarshaler(v, pkg) {
			return &schema.VarType{
				Expr: "string",
				Type: schema.T_String,
			}, nil
		}

		switch u := underlying.(type) {

		case *types.Pointer:
			// Named pointer. Webrpc can't handle that.
			// Example:
			//   type NamedPtr *Obj

			// Go for the underlying element type name (ie. `Obj`).
			return p.parseNamedType(p.goTypeName(underlying), u.Underlying())

		case *types.Slice, *types.Array:
			// Named slice/array. Webrpc can't handle that.
			// Example:
			//  type NamedSlice []int
			//  type NamedSlice []Obj

			// If the named type is a slice/array and implements json.Marshaler,
			// we assume it's []any.
			if isJsonMarshaller(v, pkg) {
				return &schema.VarType{
					Expr: "[]any",
					Type: schema.T_List,
					List: &schema.VarListType{
						Elem: &schema.VarType{
							Expr: "any",
							Type: schema.T_Any,
						},
					},
				}, nil
			}

			var elem types.Type
			//                  = u.Elem().Underlying()
			// NOTE: Calling the above assignment fails to build with this error:
			//       "u.Elem undefined (type types.Type has no field or method Elem)"
			//       even though both *types.Slice and *types.Array have the .Elem() method.
			switch underlyingElem := u.(type) {
			case *types.Slice:
				elem = underlyingElem.Elem().Underlying()
			case *types.Array:
				elem = underlyingElem.Elem().Underlying()
			}

			// If the named type is a slice/array and its underlying element
			// type is basic type (ie. `int`), we go for it directly.
			if basic, ok := elem.(*types.Basic); ok {
				basicType, err := p.parseBasic(basic)
				if err != nil {
					return nil, fmt.Errorf("failed to parse []namedBasicType: %w", err)
				}
				return &schema.VarType{
					Expr: fmt.Sprintf("[]%v", basicType.String()),
					Type: schema.T_List,
					List: &schema.VarListType{
						Elem: basicType,
					},
				}, nil
			}

			// Otherwise, go for the underlying element type name (ie. `Obj`).
			return p.parseNamedType(p.goTypeName(underlying), u.Underlying())

		default:
			if isJsonMarshaller(v, pkg) {
				return &schema.VarType{
					Expr: "any",
					Type: schema.T_Any,
				}, nil
			}

			return p.parseNamedType(typeName, underlying)
		}

	case *types.Basic:
		return p.parseBasic(v)

	case *types.Struct:
		return p.parseStruct(typeName, v)

	case *types.Slice:
		return p.parseSlice(typeName, v)

	case *types.Interface:
		return p.parseAny(typeName, v)

	case *types.Map:
		return p.parseMap(typeName, v)

	case *types.Pointer:
		if typeName == "" {
			return p.parseNamedType(p.goTypeName(v), v.Elem())
		}
		return p.parseNamedType(typeName, v.Elem())

	default:
		return nil, fmt.Errorf("unsupported argument type %T", typ)
	}
}

func findFirstLetter(s string) int {
	for i, char := range s {
		if unicode.IsLetter(char) {
			return i
		}
	}
	return 0
}

func (p *parser) goTypeName(typ types.Type) string {
	name := typ.String() // []*github.com/golang-cz/gospeak/pkg.Typ

	firstLetter := findFirstLetter(name)
	prefix := name[:firstLetter] // []*
	name = name[firstLetter:]    // github.com/golang-cz/gospeak/pkg.Typ

	name = strings.TrimPrefix(name, p.schemaPkgName+".")       // Typ (ignore root pkg)
	name = strings.TrimPrefix(name, "command-line-arguments.") // Typ (ignore "command-line-arguments" pkg autogenerated by Go tool chain)
	name = filepath.Base(name)                                 // pkg.Typ

	if name == "invalid type" {
		name = "invalidType"
	}

	return prefix + name // []*pkg.Typ
}

func (p *parser) goTypeImport(typ types.Type) string {
	name := typ.String() // []*github.com/golang-cz/gospeak/pkg.Typ

	firstLetter := findFirstLetter(name)
	name = name[firstLetter:] // github.com/golang-cz/gospeak/pkg.Typ

	lastDot := strings.LastIndex(name, ".")
	if lastDot <= 0 {
		return ""
	}

	name = name[:lastDot] // github.com/golang-cz/gospeak/pkg
	switch name {
	case p.schemaPkgName, "command-line-arguments", "time":
		return ""
	}

	return name
}

func (p *parser) parseBasic(typ *types.Basic) (*schema.VarType, error) {
	var varType schema.VarType
	if err := schema.ParseVarTypeExpr(p.schema, typ.Name(), &varType); err != nil {
		return nil, fmt.Errorf("failed to parse basic type: %v: %w", typ.Name(), err)
	}

	return &varType, nil
}

func (p *parser) parseStruct(typeName string, structTyp *types.Struct) (*schema.VarType, error) {
	structType := &schema.Type{
		Kind: "struct",
		Name: typeName,
	}

	for i := 0; i < structTyp.NumFields(); i++ {
		structField := structTyp.Field(i)
		if !structField.Exported() {
			continue
		}
		structTags := structTyp.Tag(i)

		if structField.Embedded() || strings.Contains(structTags, `json:",inline"`) {
			varType, err := p.parseNamedType("", structField.Type())
			if err != nil {
				return nil, fmt.Errorf("parsing var %v: %w", structField.Name(), err)
			}

			if varType.Type == schema.T_Struct {
				for _, embeddedField := range varType.Struct.Type.Fields {
					structType.Fields = appendOrOverrideExistingField(structType.Fields, embeddedField)
				}
			}
			continue
		}

		field, err := p.parseStructField(typeName, structField, structTags)
		if err != nil {
			return nil, fmt.Errorf("parsing struct field %v: %w", i, err)
		}
		if field != nil {
			structType.Fields = appendOrOverrideExistingField(structType.Fields, field)
		}
	}

	p.schema.Types = append(p.schema.Types, structType)

	return &schema.VarType{
		Expr: typeName,
		Type: schema.T_Struct,
		Struct: &schema.VarStructType{
			Name: typeName,
			Type: structType,
		},
	}, nil
}

// parses single Go struct field
// if the field is embedded, ie. `json:",inline"`, recursively parse
func (p *parser) parseStructField(structTypeName string, field *types.Var, structTags string) (*schema.TypeField, error) {
	optional := false

	fieldName := field.Name()
	jsonFieldName := fieldName
	goFieldName := fieldName

	fieldType := field.Type()
	goFieldType := p.goTypeName(fieldType)
	goFieldImport := p.goTypeImport(fieldType)

	jsonTag, ok := getJsonTag(structTags)
	if ok {
		if jsonTag.Name == "-" { // struct field ignored by `json:"-"` struct tag
			return nil, nil
		}

		if jsonTag.Name != "" {
			jsonFieldName = jsonTag.Name
		}

		optional = jsonTag.Omitempty

		if jsonTag.IsString { // struct field forced to be string by `json:",string"`
			structField := &schema.TypeField{
				Name: jsonFieldName,
				Type: &schema.VarType{
					Expr: "string",
					Type: schema.T_String,
				},
				TypeExtra: schema.TypeExtra{
					Meta: []schema.TypeFieldMeta{
						{"go.field.name": goFieldName},
						{"go.field.type": goFieldType},
					},
					Optional: optional,
				},
			}
			if goFieldImport != "" {
				structField.TypeExtra.Meta = append(structField.TypeExtra.Meta,
					schema.TypeFieldMeta{"go.type.import": goFieldImport},
				)
			}
			structField.TypeExtra.Meta = append(structField.TypeExtra.Meta,
				schema.TypeFieldMeta{"go.tag.json": jsonTag.Value},
			)

			return structField, nil
		}
	}

	if _, ok := field.Type().Underlying().(*types.Pointer); ok {
		optional = true
	}

	if _, ok := field.Type().Underlying().(*types.Struct); ok {
		// Anonymous struct fields.
		// Example:
		//   type Something struct {
		// 	   AnonymousField struct { // no explicit struct type name
		//       Name string
		//     }
		//   }
		structTypeName = structTypeName + "Anonymous" + field.Name()
	}

	varType, err := p.parseNamedType(structTypeName, fieldType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse var %v: %w", field.Name(), err)
	}

	structField := &schema.TypeField{
		Name: jsonFieldName,
		Type: varType,
		TypeExtra: schema.TypeExtra{
			Meta: []schema.TypeFieldMeta{
				{"go.field.name": goFieldName},
				{"go.field.type": goFieldType},
			},
			Optional: optional,
		},
	}
	if goFieldImport != "" {
		structField.TypeExtra.Meta = append(structField.TypeExtra.Meta,
			schema.TypeFieldMeta{"go.type.import": goFieldImport},
		)
	}
	if jsonTag.Value != "" {
		structField.TypeExtra.Meta = append(structField.TypeExtra.Meta, schema.TypeFieldMeta{"go.tag.json": jsonTag.Value})
	}

	return structField, nil
}

func (p *parser) parseSlice(typeName string, sliceTyp *types.Slice) (*schema.VarType, error) {
	elem, err := p.parseNamedType(typeName, sliceTyp.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to parse slice type: %w", err)
	}

	varType := &schema.VarType{
		Expr: fmt.Sprintf("[]%v", elem.String()),
		Type: schema.T_List,
		List: &schema.VarListType{
			Elem: elem,
		},
	}

	return varType, nil
}

func (p *parser) parseAny(typeName string, iface *types.Interface) (*schema.VarType, error) {
	varType := &schema.VarType{
		Expr: "any",
		Type: schema.T_Any,
	}

	return varType, nil
}

func (p *parser) parseMap(typeName string, m *types.Map) (*schema.VarType, error) {
	key, err := p.parseNamedType(typeName, m.Key())
	if err != nil {
		return nil, fmt.Errorf("failed to parse map key type: %w", err)
	}

	value, err := p.parseNamedType(typeName, m.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to parse map value type: %w", err)
	}

	varType := &schema.VarType{
		Expr: fmt.Sprintf("map<%v,%v>", key, value),
		Type: schema.T_Map,
		Map: &schema.VarMapType{
			Key:   key.Type,
			Value: value,
		},
	}

	return varType, nil
}

var textMarshalerRegex = regexp.MustCompile(`^func \((.+)\)\.MarshalText\(\) \((.+ )?\[\]byte, ([a-z]+ )?error\)$`)
var textUnmarshalerRegex = regexp.MustCompile(`^func \((.+)\)\.UnmarshalText\((.+ )?\[\]byte\) \(?(.+ )?error\)?$`)

// Returns true if the given type implements encoding.TextMarshaler/TextUnmarshaler interfaces.
func isTextMarshaler(typ types.Type, pkg *types.Package) bool {
	marshalTextMethod, _, _ := types.LookupFieldOrMethod(typ, true, pkg, "MarshalText")
	if marshalTextMethod == nil || !textMarshalerRegex.MatchString(marshalTextMethod.String()) {
		return false
	}

	unmarshalTextMethod, _, _ := types.LookupFieldOrMethod(typ, true, pkg, "UnmarshalText")
	if unmarshalTextMethod == nil || !textUnmarshalerRegex.MatchString(unmarshalTextMethod.String()) {
		return false
	}

	return true
}

var jsonMarshalerRegex = regexp.MustCompile(`^func \((.+)\)\.MarshalJSON\(\) \((.+ )?\[\]byte, ([a-z]+ )?error\)$`)
var jsonUnmarshalerRegex = regexp.MustCompile(`^func \((.+)\)\.UnmarshalJSON\((.+ )?\[\]byte\) \(?(.+ )?error\)?$`)

// Returns true if the given type implements json.Marshaler/Unmarshaler interfaces.
func isJsonMarshaller(typ types.Type, pkg *types.Package) bool {
	marshalJsonMethod, _, _ := types.LookupFieldOrMethod(typ, true, pkg, "MarshalJSON")
	if marshalJsonMethod == nil || !jsonMarshalerRegex.MatchString(marshalJsonMethod.String()) {
		return false
	}

	unmarshalJsonMethod, _, _ := types.LookupFieldOrMethod(typ, true, pkg, "UnmarshalJSON")
	if unmarshalJsonMethod == nil || !jsonUnmarshalerRegex.MatchString(unmarshalJsonMethod.String()) {
		return false
	}

	return true
}

// Appends message field to the given slice, while also removing any previously defined field of the same name.
// This lets us overwrite embedded fields, exactly how Go does it behind the scenes in the JSON marshaller.
func appendOrOverrideExistingField(slice []*schema.TypeField, newItem *schema.TypeField) []*schema.TypeField {
	// Let's try to find an existing item of the same name and delete it.
	for i, item := range slice {
		if item.Name == newItem.Name {
			// Delete item.
			copy(slice[i:], slice[i+1:])
			slice = slice[:len(slice)-1]
		}
	}
	// And then append the new item at the end of the slice.
	return append(slice, newItem)
}

// This regex will return the following three submatches,
// given `db:"id,omitempty,pk" json:"id,string"` struct tag:
//
//	[0]: json:"id,string"
//	[1]: id
//	[2]: ,string
var jsonTagRegex, _ = regexp.Compile(`\s?json:\"([^,\"]*)(,[^\"]*)?\"`)

type jsonTag struct {
	Name      string
	Value     string
	IsString  bool
	Omitempty bool
}

func getJsonTag(structTags string) (jsonTag, bool) {
	if !strings.Contains(structTags, `json:"`) {
		return jsonTag{}, false
	}

	submatches := jsonTagRegex.FindStringSubmatch(structTags)

	// Submatches from the jsonTagRegex:
	// [0]: json:"deleted_by,omitempty,string"
	// [1]: deleted_by
	// [2]: ,omitempty,string
	if len(submatches) != 3 {
		return jsonTag{}, false
	}

	jsonTag := jsonTag{
		Name:      submatches[1],
		Value:     submatches[1] + submatches[2],
		IsString:  strings.Contains(submatches[2], ",string"),
		Omitempty: strings.Contains(submatches[2], ",omitempty"),
	}

	return jsonTag, true
}
