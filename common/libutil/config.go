package libutil

import (
	"bytes"
	"go_tcp/common/logging"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
)

func TRLogger(logFile, logLevel, name, suffix string, daemon bool) error {
	logFile, _ = filepath.Abs(logFile)
	if err := os.MkdirAll(filepath.Dir(logFile), os.ModeDir|os.ModePerm); err != nil {
		return err
	}
	handler, err := logging.NewTimeRotationHandler(logFile, suffix, nil)
	if err != nil {
		return err
	}
	handler.SetLevelString(logLevel)
	handler.SetFormat(func(name, timeString string, rd *logging.Record) string {
		return "[" + timeString + "] " + name + " " + rd.Level.String() + " " + rd.Message + "\n"
	})
	logging.AddHandler(name, handler)
	if daemon {
		logging.DisableStdout()
	}
	return nil
}

func ParseJSON(filename string, v interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var lines [][]byte
	buf := bytes.NewBuffer(data)
	for {
		line, err := buf.ReadBytes('\n')
		line = bytes.Trim(line, " \t\r\n")
		if len(line) > 0 && !bytes.HasPrefix(line, []byte("//")) {
			lines = append(lines, line)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	data = bytes.Join(lines, []byte{})
	return json.Unmarshal(data, v)
}

func flagName(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func marshalStructField(name string, field reflect.StructField, value reflect.Value) {
	switch field.Type.Kind() {
	case reflect.Struct:
		marshalStruct(flagName(name, field.Name), value)
	case reflect.String:
		flag.StringVar(value.Addr().Interface().(*string), flagName(name, field.Name), value.String(), string(field.Tag))
	case reflect.Bool:
		flag.BoolVar(value.Addr().Interface().(*bool), flagName(name, field.Name), value.Bool(), string(field.Tag))
	case reflect.Int:
		flag.IntVar(value.Addr().Interface().(*int), flagName(name, field.Name), int(value.Int()), string(field.Tag))
	case reflect.Int64:
		flag.Int64Var(value.Addr().Interface().(*int64), flagName(name, field.Name), value.Int(), string(field.Tag))
	case reflect.Float64:
		flag.Float64Var(value.Addr().Interface().(*float64), flagName(name, field.Name), value.Float(), string(field.Tag))
	}
}

func marshalStruct(name string, st reflect.Value) {
	for i := 0; i < st.NumField(); i++ {
		marshalStructField(name, st.Type().Field(i), st.Field(i))
	}
}

func MarshalToFlag(v interface{}) {
	marshalStruct("", reflect.ValueOf(v).Elem())
}
