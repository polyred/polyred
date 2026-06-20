// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package shader is the Go→shader compiler for the GPU abstraction. It parses a
// restricted subset of Go compute kernels and emits backend shading-language
// source (MSL in this phase) plus the matching binding layout, so kernels are
// authored in Go instead of hand-written per backend.
//
// See specs/foundations/gpu-phase2-goshader.md and docs/gpu-abstraction.md §6b.
//
// Supported subset (compute, this phase): a kernel is a top-level func whose
// first parameter is the thread id (an int/uint, conventionally named gid) and
// whose remaining parameters are []float32 storage buffers or a struct-by-value
// uniform. Bodies may use arithmetic, indexing, short/var declarations, for/if,
// type conversions (uint/int/float32) and a whitelist of math builtins.
package shader

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"sort"
	"strings"
)

// BindingKind is the resource type of a kernel parameter.
type BindingKind int

const (
	StorageBuffer BindingKind = iota
	UniformBuffer
	SampledTexture
	SamplerBinding
)

// Binding describes one kernel parameter's GPU binding.
type Binding struct {
	Index int
	Name  string
	Kind  BindingKind
}

// Kernel is a compiled kernel (compute, vertex, or fragment).
type Kernel struct {
	Name     string
	Stage    Stage
	Bindings []Binding
	MSL      string
}

// builtins maps allowed Go call targets to their MSL spelling.
var builtins = map[string]string{
	"sqrt": "sqrt", "abs": "abs", "min": "min", "max": "max",
	"floor": "floor", "ceil": "ceil", "sin": "sin", "cos": "cos",
	"pow": "pow", "clamp": "clamp", "mix": "mix", "exp": "exp", "log": "log",
	"dot": "dot", "normalize": "normalize", "length": "length",
	"cross": "cross", "reflect": "reflect",
	// type conversions
	"float32": "float", "float": "float", "uint": "uint", "int": "int",
}

// goToMSLType maps a Go scalar/vector type name to its MSL spelling.
func goToMSLType(name string) (string, bool) {
	switch name {
	case "float32":
		return "float", true
	case "uint", "uint32":
		return "uint", true
	case "int", "int32":
		return "int", true
	case "Vec2":
		return "float2", true
	case "Vec3":
		return "float3", true
	case "Vec4":
		return "float4", true
	case "Mat4":
		return "float4x4", true
	}
	return "", false
}

// Stage is the pipeline stage a kernel targets.
type Stage int

const (
	StageCompute Stage = iota
	StageVertex
	StageFragment
)

// stageOf reads a //gpu:vertex / //gpu:fragment directive from a func's doc
// comment; absent a directive the kernel is compute.
func stageOf(doc *ast.CommentGroup) Stage {
	if doc == nil {
		return StageCompute
	}
	for _, c := range doc.List {
		switch strings.TrimSpace(c.Text) {
		case "//gpu:vertex":
			return StageVertex
		case "//gpu:fragment":
			return StageFragment
		}
	}
	return StageCompute
}

// Compile parses src and compiles every kernel function it finds, returning them
// keyed by function name. Struct types referenced as uniform parameters are
// emitted into each kernel's MSL.
func Compile(src string) (map[string]*Kernel, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "kernel.go", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("shader: parse: %w", err)
	}

	structs := map[string]*ast.StructType{}
	var funcs []*ast.FuncDecl
	for _, d := range file.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			for _, s := range decl.Specs {
				if ts, ok := s.(*ast.TypeSpec); ok {
					if st, ok := ts.Type.(*ast.StructType); ok {
						structs[ts.Name.Name] = st
					}
				}
			}
		case *ast.FuncDecl:
			if decl.Recv == nil && decl.Body != nil {
				funcs = append(funcs, decl)
			}
		}
	}

	out := map[string]*Kernel{}
	for _, fn := range funcs {
		k, err := compileKernel(fn, structs)
		if err != nil {
			return nil, fmt.Errorf("shader: kernel %s: %w", fn.Name.Name, err)
		}
		out[k.Name] = k
	}
	return out, nil
}

type compiler struct {
	structs map[string]*ast.StructType
	env     map[string]string // var name -> MSL type
	written map[string]bool   // buffer params written to (=> non-const)
	buf     strings.Builder
}

func isIntType(t string) bool {
	return t == "int" || t == "uint" || t == "int32" || t == "uint32"
}

func isVecType(t string) bool {
	return t == "float2" || t == "float3" || t == "float4"
}

// isSwizzle reports whether s is a vector component selector (1-4 of x/y/z/w).
func isSwizzle(s string) bool {
	if s == "" || len(s) > 4 {
		return false
	}
	for _, c := range s {
		if c != 'x' && c != 'y' && c != 'z' && c != 'w' {
			return false
		}
	}
	return true
}

func compileKernel(fn *ast.FuncDecl, structs map[string]*ast.StructType) (*Kernel, error) {
	stage := stageOf(fn.Doc)
	params := flattenParams(fn.Type.Params)

	c := &compiler{structs: structs, env: map[string]string{}, written: map[string]bool{}}

	// First pass: detect which buffer params are written (appear on the LHS of
	// an index assignment), so reads stay const.
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, lhs := range as.Lhs {
			if ix, ok := lhs.(*ast.IndexExpr); ok {
				if id, ok := ix.X.(*ast.Ident); ok {
					c.written[id.Name] = true
				}
			}
		}
		return true
	})

	// compute and vertex kernels take a leading id parameter
	// (thread_position_in_grid / vertex_id); fragment kernels do not.
	hasID := stage == StageCompute || stage == StageVertex
	bufParams := params
	var idName string
	if hasID {
		if len(params) == 0 {
			return nil, fmt.Errorf("kernel needs a leading id parameter")
		}
		gid := params[0]
		gidType, ok := identType(gid.typ)
		if !ok || !isIntType(gidType) {
			return nil, fmt.Errorf("first parameter %q must be the int/uint id", gid.name)
		}
		c.env[gid.name] = "uint"
		idName = gid.name
		bufParams = params[1:]
	}

	var bindings []Binding
	var usedStructs []string
	bufIndex := 0
	texIndex := 0
	samplerIndex := 0
	stageInUsed := false
	var sig []string
	for _, p := range bufParams {
		switch t := p.typ.(type) {
		case *ast.ArrayType: // []float32 storage buffer
			if t.Len != nil {
				return nil, fmt.Errorf("parameter %q: only slices ([]float32) are supported as buffers", p.name)
			}
			elt, ok := identType(t.Elt)
			if !ok {
				return nil, fmt.Errorf("parameter %q: unsupported slice element", p.name)
			}
			mt, ok := goToMSLType(elt)
			if !ok {
				return nil, fmt.Errorf("parameter %q: unsupported slice element %q", p.name, elt)
			}
			qual := "device "
			if !c.written[p.name] {
				qual += "const "
			}
			sig = append(sig, fmt.Sprintf("%s%s* %s [[buffer(%d)]]", qual, mt, p.name, bufIndex))
			bindings = append(bindings, Binding{Index: bufIndex, Name: p.name, Kind: StorageBuffer})
			c.env[p.name] = mt + "*"
			bufIndex++
		case *ast.Ident:
			// Texture / sampler params use separate MSL index spaces.
			switch t.Name {
			case "Texture2D":
				sig = append(sig, fmt.Sprintf("texture2d<float> %s [[texture(%d)]]", p.name, texIndex))
				bindings = append(bindings, Binding{Index: texIndex, Name: p.name, Kind: SampledTexture})
				c.env[p.name] = "texture2d"
				texIndex++
				continue
			case "Sampler":
				sig = append(sig, fmt.Sprintf("sampler %s [[sampler(%d)]]", p.name, samplerIndex))
				bindings = append(bindings, Binding{Index: samplerIndex, Name: p.name, Kind: SamplerBinding})
				c.env[p.name] = "sampler"
				samplerIndex++
				continue
			}
			// struct param
			if _, ok := structs[t.Name]; !ok {
				return nil, fmt.Errorf("parameter %q: unsupported type %q", p.name, t.Name)
			}
			c.env[p.name] = t.Name
			usedStructs = append(usedStructs, t.Name)
			if stage == StageFragment && !stageInUsed {
				// the first struct param of a fragment is the interpolated
				// vertex output (varyings) delivered via stage_in
				sig = append(sig, fmt.Sprintf("%s %s [[stage_in]]", t.Name, p.name))
				stageInUsed = true
			} else {
				// struct-by-value uniform
				sig = append(sig, fmt.Sprintf("constant %s& %s [[buffer(%d)]]", t.Name, p.name, bufIndex))
				bindings = append(bindings, Binding{Index: bufIndex, Name: p.name, Kind: UniformBuffer})
				bufIndex++
			}
		default:
			return nil, fmt.Errorf("parameter %q: unsupported parameter type", p.name)
		}
	}
	// The id parameter goes last in MSL with the stage-appropriate attribute.
	if hasID {
		attr := "[[thread_position_in_grid]]"
		if stage == StageVertex {
			attr = "[[vertex_id]]"
		}
		sig = append(sig, fmt.Sprintf("uint %s %s", idName, attr))
	}

	var body strings.Builder
	bc := &compiler{structs: structs, env: c.env, written: c.written, buf: body}
	if err := bc.stmts(fn.Body.List, 1); err != nil {
		return nil, err
	}

	// Function keyword and return type per stage.
	kw, ret := "kernel", "void"
	switch stage {
	case StageVertex, StageFragment:
		if stage == StageVertex {
			kw = "vertex"
		} else {
			kw = "fragment"
		}
		if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
			return nil, fmt.Errorf("%s kernel must return exactly one value", kw)
		}
		rt, _ := identType(fn.Type.Results.List[0].Type)
		if mt, ok := goToMSLType(rt); ok {
			// built-in vector return (e.g. fragment float4)
			ret = mt
		} else if _, isStruct := structs[rt]; isStruct {
			// vertex output struct (varyings + [[position]])
			ret = rt
			usedStructs = append(usedStructs, rt)
		} else {
			return nil, fmt.Errorf("unsupported return type %q", rt)
		}
	}

	var msl strings.Builder
	msl.WriteString("#include <metal_stdlib>\nusing namespace metal;\n\n")
	for _, name := range usedStructs {
		emitStruct(&msl, name, structs[name])
	}
	fmt.Fprintf(&msl, "%s %s %s(%s) {\n%s}\n", kw, ret, fn.Name.Name, strings.Join(sig, ",\n    "), bc.buf.String())

	return &Kernel{Name: fn.Name.Name, Stage: stage, Bindings: bindings, MSL: msl.String()}, nil
}

func emitStruct(w *strings.Builder, name string, st *ast.StructType) {
	fmt.Fprintf(w, "struct %s {\n", name)
	for _, f := range st.Fields.List {
		ft, _ := identType(f.Type)
		mt, ok := goToMSLType(ft)
		if !ok {
			mt = ft
		}
		// A `gpu:"position"` tag marks the clip-space position output.
		attr := ""
		if f.Tag != nil {
			tag := reflect.StructTag(strings.Trim(f.Tag.Value, "`"))
			switch tag.Get("gpu") {
			case "position":
				attr = " [[position]]"
			}
		}
		for _, n := range f.Names {
			fmt.Fprintf(w, "    %s %s%s;\n", mt, n.Name, attr)
		}
	}
	w.WriteString("};\n\n")
}

type param struct {
	name string
	typ  ast.Expr
}

func flattenParams(fl *ast.FieldList) []param {
	var ps []param
	if fl == nil {
		return ps
	}
	for _, f := range fl.List {
		for _, n := range f.Names {
			ps = append(ps, param{name: n.Name, typ: f.Type})
		}
	}
	return ps
}

func identType(e ast.Expr) (string, bool) {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name, true
	case *ast.StarExpr:
		return identType(t.X)
	}
	return "", false
}

func (c *compiler) indent(n int) { c.buf.WriteString(strings.Repeat("    ", n)) }

func (c *compiler) stmts(list []ast.Stmt, depth int) error {
	for _, s := range list {
		if err := c.stmt(s, depth); err != nil {
			return err
		}
	}
	return nil
}

func (c *compiler) stmt(s ast.Stmt, depth int) error {
	switch st := s.(type) {
	case *ast.AssignStmt:
		return c.assign(st, depth)
	case *ast.DeclStmt:
		return c.declStmt(st, depth)
	case *ast.ForStmt:
		return c.forStmt(st, depth)
	case *ast.IfStmt:
		return c.ifStmt(st, depth)
	case *ast.IncDecStmt:
		c.indent(depth)
		e, err := c.expr(st.X)
		if err != nil {
			return err
		}
		c.buf.WriteString(e + st.Tok.String() + ";\n")
		return nil
	case *ast.BlockStmt:
		return c.stmts(st.List, depth)
	case *ast.ReturnStmt:
		c.indent(depth)
		if len(st.Results) == 0 {
			c.buf.WriteString("return;\n")
			return nil
		}
		v, err := c.expr(st.Results[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(&c.buf, "return %s;\n", v)
		return nil
	default:
		return fmt.Errorf("unsupported statement %T", s)
	}
}

func (c *compiler) assign(st *ast.AssignStmt, depth int) error {
	if len(st.Lhs) != 1 || len(st.Rhs) != 1 {
		return fmt.Errorf("only single assignments are supported")
	}
	rhs, err := c.expr(st.Rhs[0])
	if err != nil {
		return err
	}
	lhs, err := c.expr(st.Lhs[0])
	if err != nil {
		return err
	}
	c.indent(depth)
	if st.Tok == token.DEFINE {
		// infer the declared type of a new variable from the RHS
		typ := c.inferType(st.Rhs[0])
		if id, ok := st.Lhs[0].(*ast.Ident); ok {
			c.env[id.Name] = typ
		}
		c.buf.WriteString(fmt.Sprintf("%s %s = %s;\n", typ, lhs, rhs))
		return nil
	}
	c.buf.WriteString(fmt.Sprintf("%s %s %s;\n", lhs, st.Tok.String(), rhs))
	return nil
}

func (c *compiler) declStmt(st *ast.DeclStmt, depth int) error {
	gd, ok := st.Decl.(*ast.GenDecl)
	if !ok || gd.Tok != token.VAR {
		return fmt.Errorf("unsupported declaration")
	}
	for _, spec := range gd.Specs {
		vs := spec.(*ast.ValueSpec)
		mt := "float"
		if vs.Type != nil {
			if gt, ok := identType(vs.Type); ok {
				if m, ok := goToMSLType(gt); ok {
					mt = m
				}
			}
		}
		for i, name := range vs.Names {
			c.env[name.Name] = mt
			c.indent(depth)
			if i < len(vs.Values) {
				v, err := c.expr(vs.Values[i])
				if err != nil {
					return err
				}
				c.buf.WriteString(fmt.Sprintf("%s %s = %s;\n", mt, name.Name, v))
			} else {
				c.buf.WriteString(fmt.Sprintf("%s %s = 0;\n", mt, name.Name))
			}
		}
	}
	return nil
}

func (c *compiler) forStmt(st *ast.ForStmt, depth int) error {
	var init, cond, post string
	if as, ok := st.Init.(*ast.AssignStmt); ok && as.Tok == token.DEFINE {
		id := as.Lhs[0].(*ast.Ident)
		typ := c.inferType(as.Rhs[0])
		c.env[id.Name] = typ
		v, err := c.expr(as.Rhs[0])
		if err != nil {
			return err
		}
		init = fmt.Sprintf("%s %s = %s", typ, id.Name, v)
	}
	if st.Cond != nil {
		v, err := c.expr(st.Cond)
		if err != nil {
			return err
		}
		cond = v
	}
	if ix, ok := st.Post.(*ast.IncDecStmt); ok {
		v, err := c.expr(ix.X)
		if err != nil {
			return err
		}
		post = v + ix.Tok.String()
	}
	c.indent(depth)
	c.buf.WriteString(fmt.Sprintf("for (%s; %s; %s) {\n", init, cond, post))
	if err := c.stmts(st.Body.List, depth+1); err != nil {
		return err
	}
	c.indent(depth)
	c.buf.WriteString("}\n")
	return nil
}

func (c *compiler) ifStmt(st *ast.IfStmt, depth int) error {
	cond, err := c.expr(st.Cond)
	if err != nil {
		return err
	}
	c.indent(depth)
	fmt.Fprintf(&c.buf, "if (%s) {\n", cond)
	if err := c.stmts(st.Body.List, depth+1); err != nil {
		return err
	}
	c.indent(depth)
	switch e := st.Else.(type) {
	case nil:
		c.buf.WriteString("}\n")
	case *ast.BlockStmt:
		c.buf.WriteString("} else {\n")
		if err := c.stmts(e.List, depth+1); err != nil {
			return err
		}
		c.indent(depth)
		c.buf.WriteString("}\n")
	case *ast.IfStmt: // else if
		c.buf.WriteString("} else ")
		return c.ifStmt(e, depth)
	default:
		return fmt.Errorf("unsupported else clause")
	}
	return nil
}

func (c *compiler) expr(e ast.Expr) (string, error) {
	switch ex := e.(type) {
	case *ast.Ident:
		return ex.Name, nil
	case *ast.BasicLit:
		return ex.Value, nil
	case *ast.ParenExpr:
		v, err := c.expr(ex.X)
		if err != nil {
			return "", err
		}
		return "(" + v + ")", nil
	case *ast.BinaryExpr:
		l, err := c.expr(ex.X)
		if err != nil {
			return "", err
		}
		r, err := c.expr(ex.Y)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s %s)", l, ex.Op.String(), r), nil
	case *ast.IndexExpr:
		base, err := c.expr(ex.X)
		if err != nil {
			return "", err
		}
		idx, err := c.expr(ex.Index)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s[%s]", base, idx), nil
	case *ast.SelectorExpr:
		base, err := c.expr(ex.X)
		if err != nil {
			return "", err
		}
		// Vector component access / swizzle: on a float2/3/4, X/Y/Z/W (and
		// multi-letter swizzles) map to the lowercase MSL components.
		if bt := c.inferType(ex.X); bt == "float2" || bt == "float3" || bt == "float4" {
			if sw := strings.ToLower(ex.Sel.Name); isSwizzle(sw) {
				return base + "." + sw, nil
			}
		}
		return base + "." + ex.Sel.Name, nil
	case *ast.CallExpr:
		return c.call(ex)
	case *ast.CompositeLit:
		return c.compositeLit(ex)
	case *ast.UnaryExpr:
		v, err := c.expr(ex.X)
		if err != nil {
			return "", err
		}
		return ex.Op.String() + v, nil
	}
	return "", fmt.Errorf("unsupported expression %T", e)
}

// compositeLit translates Vec4{...} -> float4(...) and a user-struct literal
// VOut{...} -> VOut{ordered fields}. Keyed and positional forms are supported.
func (c *compiler) compositeLit(ex *ast.CompositeLit) (string, error) {
	tname, ok := identType(ex.Type)
	if !ok {
		return "", fmt.Errorf("unsupported composite literal")
	}

	// Built-in vectors (Vec4 -> float4) take priority even if the source also
	// declares them as a struct for Go's benefit.
	if mt, ok := goToMSLType(tname); ok {
		var elems []string
		for _, e := range ex.Elts {
			v, err := c.expr(e)
			if err != nil {
				return "", err
			}
			elems = append(elems, v)
		}
		return fmt.Sprintf("%s(%s)", mt, strings.Join(elems, ", ")), nil
	}

	if st, isStruct := c.structs[tname]; isStruct {
		var fieldNames []string
		for _, f := range st.Fields.List {
			for _, n := range f.Names {
				fieldNames = append(fieldNames, n.Name)
			}
		}
		keyed := map[string]string{}
		var positional []string
		isKeyed := false
		for _, e := range ex.Elts {
			if kv, ok := e.(*ast.KeyValueExpr); ok {
				isKeyed = true
				v, err := c.expr(kv.Value)
				if err != nil {
					return "", err
				}
				keyed[kv.Key.(*ast.Ident).Name] = v
			} else {
				v, err := c.expr(e)
				if err != nil {
					return "", err
				}
				positional = append(positional, v)
			}
		}
		var ordered []string
		if isKeyed {
			for _, fn := range fieldNames {
				if v, ok := keyed[fn]; ok {
					ordered = append(ordered, v)
				} else {
					ordered = append(ordered, "0")
				}
			}
		} else {
			ordered = positional
		}
		return fmt.Sprintf("%s{%s}", tname, strings.Join(ordered, ", ")), nil
	}

	mt, ok := goToMSLType(tname)
	if !ok {
		return "", fmt.Errorf("unsupported composite type %q", tname)
	}
	var elems []string
	for _, e := range ex.Elts {
		v, err := c.expr(e)
		if err != nil {
			return "", err
		}
		elems = append(elems, v)
	}
	return fmt.Sprintf("%s(%s)", mt, strings.Join(elems, ", ")), nil
}

func (c *compiler) call(ex *ast.CallExpr) (string, error) {
	// Method call: texture sampling, e.g. tex.Sample(samp, uv) -> tex.sample(...)
	if sel, ok := ex.Fun.(*ast.SelectorExpr); ok {
		if sel.Sel.Name != "Sample" {
			return "", fmt.Errorf("unsupported method %q (only Texture2D.Sample)", sel.Sel.Name)
		}
		base, err := c.expr(sel.X)
		if err != nil {
			return "", err
		}
		var args []string
		for _, a := range ex.Args {
			v, err := c.expr(a)
			if err != nil {
				return "", err
			}
			args = append(args, v)
		}
		return fmt.Sprintf("%s.sample(%s)", base, strings.Join(args, ", ")), nil
	}

	id, ok := ex.Fun.(*ast.Ident)
	if !ok {
		return "", fmt.Errorf("unsupported call target")
	}
	msl, ok := builtins[id.Name]
	if !ok {
		return "", fmt.Errorf("call to %q is not in the builtin/conversion whitelist", id.Name)
	}
	var args []string
	for _, a := range ex.Args {
		v, err := c.expr(a)
		if err != nil {
			return "", err
		}
		args = append(args, v)
	}
	return fmt.Sprintf("%s(%s)", msl, strings.Join(args, ", ")), nil
}

// inferType returns the MSL type of an expression for declaration purposes.
func (c *compiler) inferType(e ast.Expr) string {
	switch ex := e.(type) {
	case *ast.BasicLit:
		if ex.Kind == token.INT {
			return "int"
		}
		return "float"
	case *ast.Ident:
		if t, ok := c.env[ex.Name]; ok {
			return strings.TrimSuffix(t, "*")
		}
	case *ast.CallExpr:
		if sel, ok := ex.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Sample" {
			return "float4" // Texture2D.Sample returns a float4
		}
		if id, ok := ex.Fun.(*ast.Ident); ok {
			if mt, ok := goToMSLType(id.Name); ok {
				return mt
			}
			// vector-preserving builtins return their argument's type
			switch id.Name {
			case "normalize", "cross", "reflect", "min", "max", "clamp", "abs":
				if len(ex.Args) > 0 {
					return c.inferType(ex.Args[0])
				}
			}
			// builtins like sqrt/dot/length return float (scalar)
			return "float"
		}
	case *ast.IndexExpr:
		if id, ok := ex.X.(*ast.Ident); ok {
			if t, ok := c.env[id.Name]; ok {
				return strings.TrimSuffix(t, "*")
			}
		}
	case *ast.BinaryExpr:
		lt := c.inferType(ex.X)
		rt := c.inferType(ex.Y)
		// a vector operand makes the result that vector type (vec op scalar)
		if isVecType(lt) {
			return lt
		}
		if isVecType(rt) {
			return rt
		}
		if lt == "float" || rt == "float" {
			return "float"
		}
		if lt == "uint" || rt == "uint" {
			return "uint"
		}
		return lt
	case *ast.ParenExpr:
		return c.inferType(ex.X)
	case *ast.CompositeLit:
		if tn, ok := identType(ex.Type); ok {
			if mt, ok := goToMSLType(tn); ok {
				return mt
			}
			return tn // user struct
		}
	case *ast.SelectorExpr:
		// struct field: look up field type
		if base, ok := ex.X.(*ast.Ident); ok {
			if sname, ok := c.env[base.Name]; ok {
				if st, ok := c.structs[sname]; ok {
					for _, f := range st.Fields.List {
						for _, n := range f.Names {
							if n.Name == ex.Sel.Name {
								ft, _ := identType(f.Type)
								if m, ok := goToMSLType(ft); ok {
									return m
								}
							}
						}
					}
				}
			}
		}
	}
	return "float"
}

// SortedBindings returns a kernel's bindings ordered by index (stable output).
func (k *Kernel) SortedBindings() []Binding {
	b := append([]Binding(nil), k.Bindings...)
	sort.Slice(b, func(i, j int) bool { return b[i].Index < b[j].Index })
	return b
}
