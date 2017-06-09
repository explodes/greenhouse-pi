package stats

import (
	"reflect"
	"runtime"
	"strings"
)

func functionName(i interface{}) string {
	qname := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	parts := strings.Split(qname, "/")
	qname = parts[len(parts)-1]
	parts = strings.Split(qname, ".")
	return strings.Join(parts[1:], ".")
}

func testFunctionName(i interface{}) string {
	name := functionName(i)
	parts := strings.Split(name, "_")
	if len(parts) < 2 {
		panic("Test name must be in <function>_<Condition> format")
	}
	return strings.Join(parts[1:], "_")
}
