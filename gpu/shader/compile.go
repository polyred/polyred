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
	"sort"
	"strings"
)

// BindingKind is the resource type of a kernel parameter.
type BindingKind int

const (
	StorageBuffer BindingKind = iota
	UniformBuffer
)

// Binding describes one kernel parameter's GPU binding.
type Binding struct {
	Index int
	Name  string
	Kind  BindingKind
}

// Kernel is a compiled compute kernel.
type Kernel struct {
	Name     string
	Bindings []Binding
	MSL      string
}

// builtins maps allowed Go call targets to their MSL spelling.
var builtins = map[string]string{
	"sqrt": "sqrt", "abs": "abs", "min": "min", "max": "max",
	"floor": "floor", "ceil": "ceil", "sin": "sin", "cos": "cos",
	// type conversions
	"float32": "float", "float": "float", "uint": "uint", "int": "int",
}

// goToMSLType maps a Go scalar type name to its MSL spelling.
func goToMSLType(name string) (string, bool) {
	switch name {
	case "float32":
		return "float", true
	case "uint", "uint32":
		return "uint", true
	case "int", "int32":
		return "int", true
	}
	return "", false
}

// Compile parses src and compiles every kernel function it finds, returning them
// keyed by function name. Struct types referenced as uniform parameters are
// emitted into each kernel's MSL.
func Compile(src string) (map[string]*Kernel, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "kernel.go", src, 0)
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

func compileKernel(fn *ast.FuncDecl, structs map[string]*ast.StructType) (*Kernel, error) {
	params := flattenParams(fn.Type.Params)
	if len(params) == 0 {
		return nil, fmt.Errorf("kernel needs a thread-id parameter")
	}

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

	gid := params[0]
	gidType, ok := identType(gid.typ)
	if !ok || (gidType != "int" && gidType != "uint" && gidType != "int32" && gidType != "uint32") {
		return nil, fmt.Errorf("first parameter %q must be the int/uint thread id", gid.name)
	}
	c.env[gid.name] = "uint"

	var bindings []Binding
	var usedStructs []string
	bufIndex := 0
	var sig []string
	for _, p := range params[1:] {
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
		case *ast.Ident: // struct-by-value uniform
			if _, ok := structs[t.Name]; !ok {
				return nil, fmt.Errorf("parameter %q: unsupported type %q", p.name, t.Name)
			}
			sig = append(sig, fmt.Sprintf("constant %s& %s [[buffer(%d)]]", t.Name, p.name, bufIndex))
			bindings = append(bindings, Binding{Index: bufIndex, Name: p.name, Kind: UniformBuffer})
			c.env[p.name] = t.Name
			usedStructs = append(usedStructs, t.Name)
			bufIndex++
		default:
			return nil, fmt.Errorf("parameter %q: unsupported parameter type", p.name)
		}
	}
	// Thread id parameter goes last in MSL.
	sig = append(sig, fmt.Sprintf("uint %s [[thread_position_in_grid]]", gid.name))

	var body strings.Builder
	bc := &compiler{structs: structs, env: c.env, written: c.written, buf: body}
	if err := bc.stmts(fn.Body.List, 1); err != nil {
		return nil, err
	}

	var msl strings.Builder
	msl.WriteString("#include <metal_stdlib>\nusing namespace metal;\n\n")
	for _, name := range usedStructs {
		emitStruct(&msl, name, structs[name])
	}
	fmt.Fprintf(&msl, "kernel void %s(%s) {\n%s}\n", fn.Name.Name, strings.Join(sig, ",\n    "), bc.buf.String())

	return &Kernel{Name: fn.Name.Name, Bindings: bindings, MSL: msl.String()}, nil
}

func emitStruct(w *strings.Builder, name string, st *ast.StructType) {
	fmt.Fprintf(w, "struct %s {\n", name)
	for _, f := range st.Fields.List {
		ft, _ := identType(f.Type)
		mt, ok := goToMSLType(ft)
		if !ok {
			mt = ft
		}
		for _, n := range f.Names {
			fmt.Fprintf(w, "    %s %s;\n", mt, n.Name)
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
		c.buf.WriteString("return;\n")
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
	c.buf.WriteString(fmt.Sprintf("if (%s) {\n", cond))
	if err := c.stmts(st.Body.List, depth+1); err != nil {
		return err
	}
	c.indent(depth)
	c.buf.WriteString("}\n")
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
		return base + "." + ex.Sel.Name, nil
	case *ast.CallExpr:
		return c.call(ex)
	case *ast.UnaryExpr:
		v, err := c.expr(ex.X)
		if err != nil {
			return "", err
		}
		return ex.Op.String() + v, nil
	}
	return "", fmt.Errorf("unsupported expression %T", e)
}

func (c *compiler) call(ex *ast.CallExpr) (string, error) {
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
		if id, ok := ex.Fun.(*ast.Ident); ok {
			if mt, ok := goToMSLType(id.Name); ok {
				return mt
			}
			if id.Name == "uint" || id.Name == "int" || id.Name == "float32" {
				m, _ := goToMSLType(id.Name)
				return m
			}
			// builtin like sqrt returns float
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
		if lt == "float" || rt == "float" {
			return "float"
		}
		if lt == "uint" || rt == "uint" {
			return "uint"
		}
		return lt
	case *ast.ParenExpr:
		return c.inferType(ex.X)
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
