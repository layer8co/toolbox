// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/koonix/x/file"
	"github.com/koonix/x/must"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

var (
	destPath = flag.String("write-to", "", "file to write the imports to")
	rootPath = flag.String("root", "", "starting path to operate from")
)

const cobraCommandType = "*github.com/spf13/cobra.Command"

var cobraMethods = map[string]bool{
	"AddCommand": true,
	"AddGroup":   true,
}

func main() {
	err := app()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func app() (retErr error) {

	flag.Parse()

	// =====

	if *destPath == "" {
		*destPath = os.Getenv("GOFILE")
	}
	if *destPath == "" {
		return fmt.Errorf("provide the destination file either through $GOFILE or --write-to")
	}
	*destPath = filepath.FromSlash(*destPath)

	// =====

	if *rootPath == "" {
		root, err := getRoot()
		if err != nil {
			return fmt.Errorf("could not get root path: %w", err)
		}
		*rootPath = root
	}
	*rootPath = filepath.FromSlash(*rootPath)

	// =====

	pkgPatterns := flag.Args()
	if len(pkgPatterns) == 0 {
		pkgPatterns = []string{"./..."}
	}

	// =====

	cfg := &packages.Config{
		Mode: packages.LoadSyntax,
		Dir:  *rootPath,
	}

	pkgs, err := packages.Load(cfg, pkgPatterns...)
	if err != nil {
		return fmt.Errorf("could not load packages %v: %w", pkgPatterns, err)
	}

	types := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.SelectorExpr)(nil),
	}

	pkgsToImport := make([]*packages.Package, 0)

	for _, pkg := range pkgs {
		skip := false
		insp := inspector.New(pkg.Syntax)
		insp.Nodes(types, func(node ast.Node, _ bool) (proceed bool) {
			if skip {
				return false
			}
			switch node := node.(type) {
			case *ast.FuncDecl:
				if node.Name.Name == "init" {
					return true
				}
			case *ast.SelectorExpr:
				if !cobraMethods[node.Sel.Name] {
					return false
				}
				if getType(leftExpr(node), pkg) != cobraCommandType {
					return false
				}
				if !isPackage(leftmostExpr(node), pkg) {
					return false
				}
				pkgsToImport = append(pkgsToImport, pkg)
				skip = true
			}
			return false
		})
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, *destPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("could not parse file %q: %w", *destPath, err)
	}

	for _, x := range astutil.Imports(fset, f) {
		for _, y := range x {
			path := must.Get(strconv.Unquote(y.Path.Value))
			astutil.DeleteNamedImport(fset, f, "_", path)
		}
	}

	for _, pkg := range pkgsToImport {
		astutil.AddNamedImport(fset, f, "_", pkg.PkgPath)
	}

	dest, err := file.OpenAtomicBufio(*destPath, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer dest.CloseOnSuccess(&retErr)

	err = format.Node(dest, fset, f)
	if err != nil {
		return fmt.Errorf("could not format file %q: %w", *destPath, err)
	}

	return nil
}

func leftExpr(sel *ast.SelectorExpr) ast.Node {
	left, isSel := sel.X.(*ast.SelectorExpr)
	if !isSel {
		return sel.X
	}
	return left.Sel
}

func leftmostExpr(sel *ast.SelectorExpr) ast.Node {
	for {
		left, isSel := sel.X.(*ast.SelectorExpr)
		if !isSel {
			return sel.X
		}
		sel = left
	}
}

func getType(node ast.Node, pkg *packages.Package) string {
	ident, ok := node.(*ast.Ident)
	if !ok {
		return ""
	}
	obj := pkg.TypesInfo.Uses[ident]
	if obj == nil {
		return ""
	}
	t := obj.Type()
	if t == nil {
		return ""
	}
	return t.String()
}

func isPackage(node ast.Node, pkg *packages.Package) bool {
	ident, ok := node.(*ast.Ident)
	if !ok {
		return false
	}
	obj := pkg.TypesInfo.Uses[ident]
	if obj == nil {
		return false
	}
	_, ok = obj.(*types.PkgName)
	return ok
}

func getRoot() (string, error) {

	args := []string{"go", "list", "-f", "{{.Root}}"}

	output, err1 := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err1 == nil {
		before, _ := strings.CutSuffix(string(output), "\n")
		return before, nil
	}

	_, err2 := os.Stat("go.mod")
	if err2 == nil {
		return ".", nil
	}

	return "", fmt.Errorf(
		"could not run %q: %w\n"+
			"could not stat go.mod: %w\n",
		strings.Join(args, " "), err1,
		err2,
	)
}
