// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.
// Modified from https://github.com/g3n/engine/blob/master/loader/obj/obj.go

// Package obj is used to parse the Wavefront OBJ file format (*.obj), including
// associated materials (*.mtl). Not all features of the OBJ format are
// supported. Basic format info: http://paulbourke.net/dataformats/obj/, and
// http://paulbourke.net/dataformats/mtl/
package obj

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"poly.red/color"
	"poly.red/math"
)

// File contains all decoded data from the obj and mtl files
type File struct {
	Objs      []Object             // decoded objs
	Matlib    string               // name of the material lib
	Materials map[string]*Material // maps material name to obj
	Vertices  []float32            // vertices positions array
	Normals   []float32            // vertices normals
	Uvs       []float32            // vertices texture coordinates
	Warnings  []string             // warning messages
	MtlDir    string               // Directory of material file

	// Intermediate parsing status
	line          uint      // current line number
	objCurrent    *Object   // current obj
	matCurrent    *Material // current material
	smoothCurrent bool      // current smooth state
}

type Object struct {
	Name      string
	Faces     []Face
	Materials []string
}

type Face struct {
	Vertices []int  // Indices to the face vertices
	Uvs      []int  // Indices to the face UV coordinates
	Normals  []int  // Indices to the face normals
	Material string // Material name
	Smooth   bool   // Smooth face
}

// Material contains all information about an obj material
type Material struct {
	Name       string     // Material name
	Illum      int        // Illumination model
	Opacity    float32    // Opacity factor
	Refraction float32    // Refraction factor
	Shininess  float32    // Shininess (specular exponent)
	Ambient    color.RGBA // Ambient color reflectivity
	Diffuse    color.RGBA // Diffuse color reflectivity
	Specular   color.RGBA // Specular color reflectivity
	Emissive   color.RGBA // Emissive color
	MapKd      string     // Texture file linked to diffuse color
}

// Local constants
const (
	blanks   = "\r\n\t "
	invINDEX = math.MaxUint32
	objType  = "obj"
	mtlType  = "mtl"
)

func Load(objpath string) (*File, error) {
	fobj, err := os.Open(objpath)
	if err != nil {
		return nil, err
	}
	defer fobj.Close()

	f := &File{
		Objs:      make([]Object, 0),
		Warnings:  make([]string, 0),
		Materials: make(map[string]*Material),
		Vertices:  make([]float32, 0),
		Normals:   make([]float32, 0),
		Uvs:       make([]float32, 0),
		line:      1,
	}

	err = f.parse(fobj, f.parseObjLine)
	if err != nil {
		return nil, err
	}

	if f.Matlib != "" {
		objdir := filepath.Dir(fobj.Name())
		mtllibPath := filepath.Join(objdir, f.Matlib)
		f.MtlDir = objdir
		mtlf, errMTL := os.Open(mtllibPath)
		defer mtlf.Close()
		if errMTL != nil {
			panic(errMTL)
		}
		if errMTL == nil {
			err = f.parse(mtlf, f.parseMtlLine) // will set err to nil if successful
		}
	}

	// If the mtllib line fails try <obj_filename>.mtl in the same directory.
	// process is basically identical to the above code block.
	if err != nil {
		objdir := strings.TrimSuffix(fobj.Name(), ".obj")
		mtlpath := objdir + ".mtl"
		f.MtlDir = objdir
		mtlf, errMTL := os.Open(mtlpath)
		defer mtlf.Close()
		if errMTL == nil {
			err = f.parse(mtlf, f.parseMtlLine) // will set err to nil if successful
			if err == nil {
				// log a warning
				msg := fmt.Sprintf("using material file %s", mtlpath)
				f.appendWarn(mtlType, msg)
			}
		}
	}

	if err != nil {
		fmt.Println("Using default material")
		for key := range f.Materials {
			f.Materials[key] = nil
		}
		// TODO: could be an error of some custom type. But people
		// tend to ignore errors and pass them up the call stack instead
		// of handling them... so all this work would probably be wasted.
		f.appendWarn(mtlType, "unable to parse a material file for obj. using default material instead.")
	}

	return f, nil
}

func (f *File) parse(reader io.Reader, parseLine func(string) error) error {
	bufin := bufio.NewReader(reader)
	f.line = 1
	for {
		line, err := bufin.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		line = strings.Trim(line, blanks)
		perr := parseLine(line)
		if perr != nil {
			return perr
		}
		if err == io.EOF {
			break
		}
		f.line++
	}
	return nil
}

func (f *File) parseObjLine(line string) error {
	fields := strings.Fields(line)
	if len(fields) == 0 { // empty line
		return nil
	}
	ltype := fields[0]
	if strings.HasPrefix(ltype, "#") { // comments
		return nil
	}

	switch ltype {
	case "mtllib":
		return f.parseMatlib(fields[1:])
	case "o":
		return f.parseObj(fields[1:])
	case "g":
		return f.parseObj(fields[1:])
	case "v":
		return f.parseVertex(fields[1:])
	case "vn":
		return f.parseNormal(fields[1:])
	case "vt":
		return f.parseUV(fields[1:])
	case "f":
		return f.parseFace(fields[1:])
	case "usemtl":
		return f.parseUsemtl(fields[1:])
	case "s":
		return f.parseSmooth(fields[1:])
	default:
		f.appendWarn(objType, "unsupported field: "+ltype)
	}
	return nil
}

func (f *File) parseMatlib(fields []string) error {
	// mtllib <name>
	if len(fields) < 1 {
		return f.formatError("Material library (mtllib) with no fields")
	}
	f.Matlib = fields[0]
	return nil
}

func (f *File) parseObj(fields []string) error {
	// o <name>
	if len(fields) < 1 {
		return f.formatError("obj line (o) with no fields")
	}

	// create obj information
	var ob Object
	ob.Name = fields[0]
	ob.Faces = make([]Face, 0)
	ob.Materials = make([]string, 0)

	f.Objs = append(f.Objs, ob)
	f.objCurrent = &f.Objs[len(f.Objs)-1]
	return nil
}

func (f *File) parseVertex(fields []string) error {
	// v <x> <y> <z> [w]
	if len(fields) < 3 {
		return f.formatError(fmt.Sprintf("vertex has less than 3 vertices in 'v' line: %v", fields))
	}
	w := float32(1)
	if len(fields) == 4 {
		v, err := strconv.ParseFloat(fields[3], 32)
		if err != nil {
			return err
		}
		w = 1 / float32(v)
	}
	for _, fi := range fields[:3] {
		val, err := strconv.ParseFloat(fi, 32)
		if err != nil {
			return err
		}
		if w == 1 {
			f.Vertices = append(f.Vertices, float32(val))
		} else {
			f.Vertices = append(f.Vertices, float32(val)*w)
		}
	}
	return nil
}

func (f *File) parseNormal(fields []string) error {
	// vn <x> <y> <z>
	if len(fields) < 3 {
		return f.formatError(fmt.Sprintf("normal has less than 3 vertices in 'vn' line: %v", fields))
	}
	for _, fi := range fields[:3] {
		val, err := strconv.ParseFloat(fi, 32)
		if err != nil {
			return err
		}
		f.Normals = append(f.Normals, float32(val))
	}
	return nil
}

func (f *File) parseUV(fields []string) error {
	// vt <u> <v> <w>
	if len(fields) < 2 {
		return f.formatError("Less than 2 texture coords. in 'vt' line")
	}
	for _, fi := range fields[:2] {
		val, err := strconv.ParseFloat(fi, 32)
		if err != nil {
			return err
		}
		f.Uvs = append(f.Uvs, float32(val))
	}
	return nil
}

func (f *File) parseFace(fields []string) error {
	// f v1[/vt1][/vn1] v2[/vt2][/vn2] v3[/vt3][/vn3] ...
	if f.objCurrent == nil {
		// If a face line is encountered before a group (g) or obj (o),
		// create a new "default" obj. This 'handles' the case when
		// a g or o line is not specified (allowed in OBJ format)
		f.parseObj([]string{fmt.Sprintf("unnamed%d", f.line)})
	}

	// If current obj has no material, appends last material if defined
	if len(f.objCurrent.Materials) == 0 && f.matCurrent != nil {
		f.objCurrent.Materials = append(f.objCurrent.Materials, f.matCurrent.Name)
	}

	if len(fields) < 3 {
		return f.formatError("face line with less 3 fields")
	}
	var face Face
	face.Vertices = make([]int, len(fields))
	face.Uvs = make([]int, len(fields))
	face.Normals = make([]int, len(fields))
	if f.matCurrent != nil {
		face.Material = f.matCurrent.Name
	} else {
		// No avaliable material found. Use a polyred_default material.
		face.Material = "polyred_default"
	}
	face.Smooth = f.smoothCurrent

	for pos, fi := range fields {

		// Separate the current field in its components: v vt vn
		vfields := strings.Split(fi, "/")
		if len(vfields) < 1 {
			return f.formatError("face field with no parts")
		}

		// Get the index of this vertex position (must always exist)
		val, err := strconv.ParseInt(vfields[0], 10, 32)
		if err != nil {
			return err
		}

		// Positive index is an absolute vertex index
		if val > 0 {
			face.Vertices[pos] = int(val - 1)
			// Negative vertex index is relative to the last parsed vertex
		} else if val < 0 {
			current := (len(f.Vertices) / 3) - 1
			face.Vertices[pos] = current + int(val) + 1
			// Vertex index could never be 0
		} else {
			return f.formatError("face vertex index value equal to 0")
		}

		// Get the index of this vertex UV coordinate (optional)
		if len(vfields) > 1 && len(vfields[1]) > 0 {
			val, err := strconv.ParseInt(vfields[1], 10, 32)
			if err != nil {
				return err
			}

			// Positive index is an absolute UV index
			if val > 0 {
				face.Uvs[pos] = int(val - 1)
				// Negative vertex index is relative to the last parsed uv
			} else if val < 0 {
				current := (len(f.Uvs) / 2) - 1
				face.Uvs[pos] = current + int(val) + 1
				// UV index could never be 0
			} else {
				return f.formatError("face uv index value equal to 0")
			}
		} else {
			face.Uvs[pos] = invINDEX
		}

		// Get the index of this vertex normal (optional)
		if len(vfields) >= 3 {
			val, err = strconv.ParseInt(vfields[2], 10, 32)
			if err != nil {
				return err
			}

			// Positive index is an absolute normal index
			if val > 0 {
				face.Normals[pos] = int(val - 1)
				// Negative vertex index is relative to the last parsed normal
			} else if val < 0 {
				current := (len(f.Normals) / 3) - 1
				face.Normals[pos] = current + int(val) + 1
				// Normal index could never be 0
			} else {
				return f.formatError("face normal index value equal to 0")
			}
		} else {
			face.Normals[pos] = invINDEX
		}
	}

	f.objCurrent.Faces = append(f.objCurrent.Faces, face)
	return nil
}

func (f *File) parseUsemtl(fields []string) error {
	// usemtl <name>
	if len(fields) < 1 {
		return f.formatError("Usemtl with no fields")
	}

	// See similar nil test in parseface()
	if f.objCurrent == nil {
		f.parseObj([]string{fmt.Sprintf("unnamed%d", f.line)})
	}

	// Checks if this material has already been parsed
	name := fields[0]
	mat := f.Materials[name]
	// Creates material descriptor
	if mat == nil {
		mat = &Material{Name: name}
		f.Materials[name] = mat
	}
	f.objCurrent.Materials = append(f.objCurrent.Materials, name)
	// Set this as the current material
	f.matCurrent = mat
	return nil
}

// parseSmooth parses a "s" decription line:
func (f *File) parseSmooth(fields []string) error {
	// s <0|1>
	if len(fields) < 1 {
		return f.formatError("'s' with no fields")
	}
	if fields[0] == "0" || fields[0] == "off" {
		f.smoothCurrent = false
		return nil
	}
	f.smoothCurrent = true
	return nil
}

func (f *File) parseMtlLine(line string) error {
	fields := strings.Fields(line)
	if len(fields) == 0 { // empty line
		return nil
	}
	ltype := fields[0]
	if strings.HasPrefix(ltype, "#") { // comments
		return nil
	}

	switch ltype {
	case "newmtl":
		return f.parseNewmtl(fields[1:])
	case "d":
		return f.parseDissolve(fields[1:])
	case "Ka":
		return f.parseKa(fields[1:])
	case "Kd":
		return f.parseKd(fields[1:])
	case "Ke":
		return f.parseKe(fields[1:])
	case "Ks":
		return f.parseKs(fields[1:])
	case "Ni":
		return f.parseNi(fields[1:])
	case "Ns":
		return f.parseNs(fields[1:])
	case "illum":
		return f.parseIllum(fields[1:])
	case "map_Kd":
		return f.parseMapKd(fields[1:])
	default:
		f.appendWarn(mtlType, "field not supported: "+ltype)
	}
	return nil
}

func (f *File) parseNewmtl(fields []string) error {
	// newmtl <mat_name>
	if len(fields) < 1 {
		return f.formatError("newmtl with no fields")
	}
	// Checks if material has already been seen
	name := fields[0]
	mat := f.Materials[name]
	// Creates material descriptor
	if mat == nil {
		mat = &Material{Name: name}
		f.Materials[name] = mat
	}
	f.matCurrent = mat
	return nil
}

func (f *File) parseDissolve(fields []string) error {
	// d <factor>
	if len(fields) < 1 {
		return f.formatError("'d' with no fields")
	}
	val, err := strconv.ParseFloat(fields[0], 32)
	if err != nil {
		return f.formatError("'d' parse float error")
	}
	f.matCurrent.Opacity = float32(val)
	return nil
}

func (f *File) parseKa(fields []string) error {
	// Ka r g b
	if len(fields) < 3 {
		return f.formatError("'Ka' with less than 3 fields")
	}
	var colors [3]float32
	for pos, f := range fields[:3] {
		val, err := strconv.ParseFloat(f, 32)
		if err != nil {
			return err
		}
		colors[pos] = float32(val)
	}
	f.matCurrent.Ambient = color.FromValue(colors[0], colors[1], colors[2], 1)
	return nil
}

func (f *File) parseKd(fields []string) error {
	// Kd r g b
	if len(fields) < 3 {
		return f.formatError("'Kd' with less than 3 fields")
	}
	var colors [3]float32
	for pos, f := range fields[:3] {
		val, err := strconv.ParseFloat(f, 32)
		if err != nil {
			return err
		}
		colors[pos] = float32(val)
	}
	f.matCurrent.Diffuse = color.FromValue(colors[0], colors[1], colors[2], 1)
	return nil
}

func (f *File) parseKe(fields []string) error {
	// Ke r g b
	if len(fields) < 3 {
		return f.formatError("'Ke' with less than 3 fields")
	}
	var colors [3]float32
	for pos, f := range fields[:3] {
		val, err := strconv.ParseFloat(f, 32)
		if err != nil {
			return err
		}
		colors[pos] = float32(val)
	}
	f.matCurrent.Emissive = color.FromValue(colors[0], colors[1], colors[2], 1)
	return nil
}

func (f *File) parseKs(fields []string) error {
	// Ks r g b
	if len(fields) < 3 {
		return f.formatError("'Ks' with less than 3 fields")
	}
	var colors [3]float32
	for pos, f := range fields[:3] {
		val, err := strconv.ParseFloat(f, 32)
		if err != nil {
			return err
		}
		colors[pos] = float32(val)
	}
	f.matCurrent.Specular = color.FromValue(colors[0], colors[1], colors[2], 1)
	return nil
}

func (f *File) parseNi(fields []string) error {
	// Ni <optical_density>
	if len(fields) < 1 {
		return f.formatError("'Ni' with no fields")
	}
	val, err := strconv.ParseFloat(fields[0], 32)
	if err != nil {
		return f.formatError("'d' parse float error")
	}
	f.matCurrent.Refraction = float32(val)
	return nil
}

func (f *File) parseNs(fields []string) error {
	// Ns <specular_exponent>
	if len(fields) < 1 {
		return f.formatError("'Ns' with no fields")
	}
	val, err := strconv.ParseFloat(fields[0], 32)
	if err != nil {
		return f.formatError("'d' parse float error")
	}
	f.matCurrent.Shininess = float32(val)
	return nil
}

func (f *File) parseIllum(fields []string) error {
	// illum <ilum_#>
	if len(fields) < 1 {
		return f.formatError("'illum' with no fields")
	}
	val, err := strconv.ParseUint(fields[0], 10, 32)
	if err != nil {
		return f.formatError("'d' parse int error")
	}
	f.matCurrent.Illum = int(val)
	return nil
}

func (f *File) parseMapKd(fields []string) error {
	// map_Kd [-options] <filename>
	if len(fields) < 1 {
		return f.formatError("No fields")
	}
	f.matCurrent.MapKd = fields[0]
	return nil
}

func (f *File) formatError(msg string) error {
	return fmt.Errorf("%s in line:%d", msg, f.line)
}

func (f *File) appendWarn(ftype string, msg string) {
	wline := fmt.Sprintf("%s(%d): %s", ftype, f.line, msg)
	f.Warnings = append(f.Warnings, wline)
}
