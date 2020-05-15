// Copyright 2018 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Dict and StringDict type
//
// The idea is that most dicts just have strings for keys so we use
// the simpler StringDict and promote it into a Dict when necessary

package py

import "bytes"

const dictDoc = `dict() -> new empty dictionary
dict(mapping) -> new dictionary initialized from a mapping object's
    (key, value) pairs
dict(iterable) -> new dictionary initialized as if via:
    d = {}
    for k, v in iterable:
        d[k] = v
dict(**kwargs) -> new dictionary initialized with the name=value pairs
    in the keyword argument list.  For example:  dict(one=1, two=2)`

var (
	StringDictType = NewType("dict", dictDoc)
	DictType       = NewType("dict", dictDoc)
	expectingDict  = ExceptionNewf(TypeError, "a dict is required")
)

func init() {
	DictType.Dict[String("items")] = MustNewMethod("items", func(self Object, args Tuple) (Object, error) {
		err := UnpackTuple(args, nil, "items", 0, 0)
		if err != nil {
			return nil, err
		}
		sMap := self.(Dict)
		o := make([]Object, 0, len(sMap))
		for k, v := range sMap {
			o = append(o, Tuple{k, v})
		}
		return NewIterator(o), nil
	}, 0, "items() -> list of D's (key, value) pairs, as 2-tuples")

	DictType.Dict[String("get")] = MustNewMethod("get", func(self Object, args Tuple) (Object, error) {
		var length = len(args)
		switch {
		case length == 0:
			return nil, ExceptionNewf(TypeError, "%s expected at least 1 arguments, got %d", "items()", length)
		case length > 2:
			return nil, ExceptionNewf(TypeError, "%s expected at most 2 arguments, got %d", "items()", length)
		}
		sMap := self.(Dict)
		if str, ok := args[0].(String); ok {
			if res, ok := sMap[str]; ok {
				return res, nil
			}

			switch length {
			case 2:
				return args[1], nil
			default:
				return None, nil
			}
		}
		return nil, ExceptionNewf(KeyError, "%v", args[0])
	}, 0, "gets(key, default) -> If there is a val corresponding to key, return val, otherwise default")
}

// String to object dictionary
//
// Used for variables etc where the keys can only be strings
type Dict map[Object]Object

// Type of this StringDict object
func (o Dict) Type() *Type {
	return DictType
}

// Make a new dictionary
func NewDict() Dict {
	return make(Dict)
}

// Make a new dictionary with reservation for n entries
func NewDictSized(n int) Dict {
	return make(Dict, n)
}

// Checks that obj is exactly a dictionary and returns an error if not
func DictCheckExact(obj Object) (Dict, error) {
	dict, ok := obj.(Dict)
	if !ok {
		return nil, expectingDict
	}
	return dict, nil
}

// Checks that obj is exactly a dictionary and returns an error if not
func DictCheck(obj Object) (Dict, error) {
	// FIXME should be checking subclasses
	return DictCheckExact(obj)
}

// Copy a dictionary
func (d Dict) Copy() Dict {
	e := make(Dict, len(d))
	for k, v := range d {
		e[k] = v
	}
	return e
}

func (a Dict) M__str__() (Object, error) {
	return a.M__repr__()
}

func (a Dict) M__repr__() (Object, error) {
	var out bytes.Buffer
	out.WriteRune('{')
	spacer := false
	for key, value := range a {
		if spacer {
			out.WriteString(", ")
		}
		keyStr, err := ReprAsString(key)
		if err != nil {
			return nil, err
		}
		valueStr, err := ReprAsString(value)
		if err != nil {
			return nil, err
		}
		out.WriteString(keyStr)
		out.WriteString(": ")
		out.WriteString(valueStr)
		spacer = true
	}
	out.WriteRune('}')
	return String(out.String()), nil
}

// Returns a list of keys from the dict
func (d Dict) M__iter__() (Object, error) {
	o := make([]Object, 0, len(d))
	for k := range d {
		o = append(o, k)
	}
	return NewIterator(o), nil
}

func (d Dict) M__getitem__(key Object) (Object, error) {
	str, ok := key.(String)
	if ok {
		res, ok := d[str]
		if ok {
			return res, nil
		}
	}
	return nil, ExceptionNewf(KeyError, "%v", key)
}

func (d Dict) M__setitem__(key, value Object) (Object, error) {
	//str, ok := key.(String)
	//if !ok {
	//	return nil, ExceptionNewf(KeyError, "FIXME can only have string keys!: %v", key)
	//}
	d[key] = value
	return None, nil
}

func (a Dict) M__eq__(other Object) (Object, error) {
	b, ok := other.(Dict)
	if !ok {
		return NotImplemented, nil
	}
	if len(a) != len(b) {
		return False, nil
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return False, nil
		}
		res, err := Eq(av, bv)
		if err != nil {
			return nil, err
		}
		if res == False {
			return False, nil
		}
	}
	return True, nil
}

func (a Dict) M__ne__(other Object) (Object, error) {
	res, err := a.M__eq__(other)
	if err != nil {
		return nil, err
	}
	if res == NotImplemented {
		return res, nil
	}
	if res == True {
		return False, nil
	}
	return True, nil
}

func (a Dict) M__contains__(other Object) (Object, error) {
	key, ok := other.(String)
	if !ok {
		return nil, ExceptionNewf(KeyError, "FIXME can only have string keys!: %v", key)
	}

	if _, ok := a[key]; ok {
		return True, nil
	}
	return False, nil
}
