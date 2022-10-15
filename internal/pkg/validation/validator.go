package validation

import (
	"reflect"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var trans ut.Translator

func InitValidator() error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		chinese := zh.New()
		uni := ut.New(chinese)
		trans, _ = uni.GetTranslator("zh")

		// 注册一个函数，获取struct tag里自定义的label作为字段名
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := fld.Tag.Get("label")
			return name
		})

		registerCustomValidator(v, trans)

		return zhTranslations.RegisterDefaultTranslations(v, trans)
	}

	return nil
}

// registerCustomValidator 注册自定义验证器
func registerCustomValidator(v *validator.Validate, trans ut.Translator) {
	validatorPhone(v, trans)
	validatorIds(v, trans)
}

func Translate(err error) string {
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, err := range errs {
			return err.Translate(trans)
		}
	}

	return err.Error()
}

func init() {
	_ = InitValidator()
}
