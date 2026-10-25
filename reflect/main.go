package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func Validate(v any) error {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return fmt.Errorf("передан nil")
		}
		val = val.Elem()
		typ = val.Type()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("ожидалась структура, получен %s", val.Kind())
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}
		if err := checkField(field.Name, val.Field(i), tag); err != nil {
			return err
		}
	}
	return nil
}

func checkField(name string, value reflect.Value, tag string) error {
	rules := strings.Split(tag, ";")
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		parts := strings.SplitN(rule, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("поле %s: некорректное правило %q", name, rule)
		}
		key, raw := parts[0], parts[1]
		switch key {
		case "min", "max":
			if err := checkBound(name, value, key, raw); err != nil {
				return err
			}
		case "regexp":
			if err := checkRegexp(name, value, raw); err != nil {
				return err
			}
		default:
			return fmt.Errorf("поле %s: неизвестное правило %q", name, key)
		}
	}
	return nil
}

func checkBound(name string, value reflect.Value, op, raw string) error {
	if value.Kind() == reflect.String {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("поле %s: некорректный %s=%s", name, op, raw)
		}
		// Длина в рунах: кириллица и иероглифы считаются как один символ.
		length := len([]rune(value.String()))
		if op == "min" && length < limit {
			return fmt.Errorf("поле %s: длина %d меньше min=%d", name, length, limit)
		}
		if op == "max" && length > limit {
			return fmt.Errorf("поле %s: длина %d больше max=%d", name, length, limit)
		}
		return nil
	}

	number, ok := asNumber(value)
	if !ok {
		return fmt.Errorf("поле %s: правило %s применимо к строке или числу", name, op)
	}
	limit, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fmt.Errorf("поле %s: некорректный %s=%s", name, op, raw)
	}
	if op == "min" && number < limit {
		return fmt.Errorf("поле %s: значение %v меньше min=%s", name, value.Interface(), raw)
	}
	if op == "max" && number > limit {
		return fmt.Errorf("поле %s: значение %v больше max=%s", name, value.Interface(), raw)
	}
	return nil
}

func checkRegexp(name string, value reflect.Value, pattern string) error {
	if value.Kind() != reflect.String {
		return fmt.Errorf("поле %s: regexp применим только к строке", name)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("поле %s: некорректный regexp: %w", name, err)
	}
	if !re.MatchString(value.String()) {
		return fmt.Errorf("поле %s: значение %q не соответствует regexp", name, value.String())
	}
	return nil
}

func asNumber(value reflect.Value) (float64, bool) {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(value.Uint()), true
	case reflect.Float32, reflect.Float64:
		return value.Float(), true
	default:
		return 0, false
	}
}

type User struct {
	Name  string `validate:"min=3"`
	Age   int    `validate:"min=18;max=65"`
	Email string `validate:"regexp=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
}

func main() {
	if err := Validate(User{Name: "Ив", Age: 18, Email: "test@example.com"}); err != nil {
		fmt.Println("Validation error:", err)
	}

	if err := Validate(User{Name: "Иван", Age: 70, Email: "test@example.com"}); err != nil {
		fmt.Println("Validation error:", err)
	}

	if err := Validate(User{Name: "Иван", Age: 35, Email: "invalid email"}); err != nil {
		fmt.Println("Validation error:", err)
	}

	if err := Validate(User{Name: "Иван", Age: 35, Email: "test@example.com"}); err != nil {
		fmt.Println("Validation error:", err)
	}
}
