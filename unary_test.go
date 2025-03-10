package main

import (
	"github.com/bhavyagada/xeneinterpreter/runtime"
	"testing"
)

func Test_NegativeInt(t *testing.T) {
	tests := map[string]runtime.Value{
		"var_a = -3":                    -3,
		"var_a = 3; return -var_a":      -3,
		"var_a = [3]; return -var_a[0]": -3,
		"5*-3":     -15,
		"-3*5":     -15,
		"4-7":      -3,
		"-(-3)":    3,
		"-[3][0]":  -3,
		"-abs(-3)": -3,
		"(-1)":     -1,
		"1-(-1)":   2,
		"1 - -1":   2,
	}
	for code, val := range tests {
		value, err := execString(code, runtime.DefaultTimeout)
		if err != nil {
			t.Errorf("%v\nCode: %v", err.Error(), code)
		} else if !runtime.Equals(value, val) {
			t.Errorf("Failed got %v expected %v\nTest:%v", value, val, code)
		}
	}
}

func Test_NegativeIntFail(t *testing.T) {
	tests := []string{
		"-[1]", "1---1", "-true", "1--1",
	}
	for _, code := range tests {
		if val, err := execString(code, runtime.DefaultTimeout); err == nil {
			t.Errorf("Failed to fail! \nResult: %v\nCode:%v", val, code)
		}
	}
}

func Test_NegativeBool(t *testing.T) {
	tests := map[string]runtime.Value{
		"!true": false,
		"var_a = true; return !var_a":      false,
		"var_a = [true]; return !var_a[0]": false,
		"true == !false":                   true,
		"!(1 < 2)":                         false,
		"!(!true)":                         true,
	}
	for code, val := range tests {
		value, err := execString(code, runtime.DefaultTimeout)
		if err != nil {
			t.Errorf("%v\nCode: %v", err.Error(), code)
		} else if !runtime.Equals(value, val) {
			t.Errorf("Failed got %v expected %v\nTest:%v", value, val, code)
		}
	}
}

func Test_NegativeBoolFail(t *testing.T) {
	tests := []string{
		"![true]", "!!true", "!3 < 4",
	}
	for _, code := range tests {
		if val, err := execString(code, runtime.DefaultTimeout); err == nil {
			t.Errorf("Failed to fail! \nResult: %v\nCode:%v", val, code)
		}
	}
}

func Test_Increments(t *testing.T) {
	tests := map[string]runtime.Value{
		"var_a = 2; var_a++":                            2,
		"var_a = 2; var_a++;var_a":                      3,
		"var_a = [1]; var_a[0]++; var_a[0]":             2,
		"var_a = [1]; var_a[0]--; var_a[0]":             0,
		"var_a--":                                       0,
		"var_a--;var_a":                                 -1,
		"new_list(1).map(function var_a -> var_a++)[0]": 0,
		"var_a = 3; var_a-- + 3":                        6,
		"var_a---1":                                     -1,
		"var_a---1;var_a":                               -1,
	}
	for code, val := range tests {
		value, err := execString(code, runtime.DefaultTimeout)
		if err != nil {
			t.Errorf("%v\nCode: %v", err.Error(), code)
		} else if !runtime.Equals(value, val) {
			t.Errorf("Failed got %v expected %v\nTest:%v", value, val, code)
		}
	}
}

func Test_IncrementsFail(t *testing.T) {
	tests := []string{
		"[1]++", "[1].length--",
	}
	for _, code := range tests {
		if val, err := execString(code, runtime.DefaultTimeout); err == nil {
			t.Errorf("Failed to fail! \nResult: %v\nCode:%v", val, code)
		}
	}
}
