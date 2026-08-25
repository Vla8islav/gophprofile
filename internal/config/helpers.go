package config

import (
	"reflect"

	"go.uber.org/zap"
)

func logSetFlags[T OptionsServer](options *T, logger *zap.Logger) {
	if options == nil {
		return
	}
	fields := make([]zap.Field, 0)

	t := reflect.TypeOf(options).Elem()
	v := reflect.ValueOf(options).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		flagName, ok := field.Tag.Lookup("command_arg")
		if !ok {
			continue
		}
		fieldValue := v.Field(i)
		beenSet := fieldValue.FieldByName("BeenSet")
		value := fieldValue.FieldByName("Value")
		if !beenSet.Bool() {
			continue
		}
		fields = append(fields, zap.String("-"+flagName, value.String()))
	}

	if len(fields) == 0 {
		logger.Info("no command-line flags were set")
		return
	}
	logger.Info("command line options", fields...)
}

func logSetEnv[T OptionsServer](options *T, logger *zap.Logger) {
	if options == nil {
		return
	}
	fields := make([]zap.Field, 0)

	// get type and value
	t := reflect.TypeOf(options).Elem()
	v := reflect.ValueOf(options).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		envName, ok := field.Tag.Lookup("env")
		if !ok {
			continue
		}
		fieldValue := v.Field(i)
		beenSet := fieldValue.FieldByName("BeenSet")
		value := fieldValue.FieldByName("Value")
		if !beenSet.Bool() {
			continue
		}
		fields = append(fields, zap.String(envName, value.String()))
	}

	if len(fields) == 0 {
		logger.Info("no environment variables were set")
		return
	}
	logger.Info("environment variables", fields...)
}

func logConfigOptions[T OptionsServer](options *T, logger *zap.Logger) {
	if options == nil {
		return
	}
	fields := make([]zap.Field, 0)

	t := reflect.TypeOf(options).Elem()
	v := reflect.ValueOf(options).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonFieldName, ok := field.Tag.Lookup("json")
		if !ok {
			continue
		}
		if jsonFieldName == "-" {
			continue
		}
		fieldValue := v.Field(i)
		beenSet := fieldValue.FieldByName("BeenSet")
		value := fieldValue.FieldByName("Value")
		if !beenSet.Bool() {
			continue
		}
		fields = append(fields, zap.String("-"+jsonFieldName, value.String()))
	}
	if len(fields) == 0 {
		logger.Info("no config file options were set")
		return
	}
	logger.Info("config file options", fields...)
}

func mergeOptions[T OptionsServer](mergeInto *T, newValues T) {
	t := reflect.TypeOf(mergeInto).Elem()
	vInto := reflect.ValueOf(mergeInto).Elem()
	vNew := reflect.ValueOf(newValues)
	for i := 0; i < t.NumField(); i++ {
		fieldValueInto := vInto.Field(i)
		intoBeenSet := fieldValueInto.FieldByName("BeenSet")
		intoValue := fieldValueInto.FieldByName("Value")

		fieldValueNew := vNew.Field(i)
		newBeenSet := fieldValueNew.FieldByName("BeenSet")
		newValue := fieldValueNew.FieldByName("Value")

		if newBeenSet.Bool() {
			intoValue.Set(newValue)
			intoBeenSet.Set(newBeenSet)
		}
	}
}
