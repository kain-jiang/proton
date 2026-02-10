package trait

import (
	"context"
	"reflect"
)

type LangSetter interface {
	SetLang(string)
	LangKey() string
}

func ConvertLangs(ctx context.Context, s LangReader, objs any, lang, zone string) *Error {
	v := reflect.ValueOf(objs)

	// 确保是切片
	if v.Kind() != reflect.Slice {
		return &Error{
			Internal: ECNULL,
			Detail:   "input not a LangSetter, check code",
		}
	}

	// 遍历切片
	for j := 0; j < v.Len(); j++ {
		o := v.Index(j).Interface()
		i := o.(LangSetter)
		alias, err := s.GetAppLang(ctx, lang, i.LangKey(), zone)
		if err != nil {
			return err
		}
		i.SetLang(alias)
	}
	return nil
}
