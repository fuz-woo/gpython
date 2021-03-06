// Copyright 2018 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Implement builtin functions eval and exec

package vm

import (
	"strings"

	"github.com/go-python/gpython/py"
)

func builtinEvalOrExec(self py.Object, args py.Tuple, kwargs, currentLocals, currentGlobals, builtins py.Dict, mode string) (py.Object, error) {
	var (
		cmd     py.Object
		globals py.Object = py.None
		locals  py.Object = py.None
	)
	err := py.UnpackTuple(args, kwargs, mode, 1, 3, &cmd, &globals, &locals)
	if err != nil {
		return nil, err
	}
	if globals == py.None {
		globals = currentGlobals
		if locals == py.None {
			locals = currentLocals
		}
	} else if locals == py.None {
		locals = globals
	}
	// FIXME this can be a mapping too
	globalsDict, err := py.DictCheck(globals)
	if err != nil {
		return nil, py.ExceptionNewf(py.TypeError, "globals must be a dict")
	}
	localsDict, err := py.DictCheck(locals)
	if err != nil {
		return nil, py.ExceptionNewf(py.TypeError, "locals must be a dict")
	}

	// Set __builtins__ if not set
	if _, ok := globalsDict[py.String("__builtins__")]; !ok {
		globalsDict[py.String("__builtins__")] = builtins
	}

	var codeStr string
	var code *py.Code
	switch x := cmd.(type) {
	case *py.Code:
		code = x
	case py.String:
		codeStr = string(x)
	case py.Bytes:
		codeStr = string(x)
	default:
		return nil, py.ExceptionNewf(py.TypeError, "%s() arg 1 must be a string, bytes or code object", mode)

	}
	if code == nil {
		codeStr = strings.TrimLeft(codeStr, " \t")
		obj, err := py.Compile(codeStr, "<string>", mode, 0, true)
		if err != nil {
			return nil, err
		}
		code = obj.(*py.Code)
	}
	if code.GetNumFree() > 0 {
		return nil, py.ExceptionNewf(py.TypeError, "code passed to %s() may not contain free variables", mode)
	}
	return EvalCode(code, globalsDict, localsDict)
}

func builtinEval(self py.Object, args py.Tuple, kwargs, currentLocals, currentGlobals, builtins py.Dict) (py.Object, error) {
	return builtinEvalOrExec(self, args, kwargs, currentLocals, currentGlobals, builtins, "eval")
}

func builtinExec(self py.Object, args py.Tuple, kwargs, currentLocals, currentGlobals, builtins py.Dict) (py.Object, error) {
	_, err := builtinEvalOrExec(self, args, kwargs, currentLocals, currentGlobals, builtins, "exec")
	if err != nil {
		return nil, err
	}
	return py.None, nil
}
