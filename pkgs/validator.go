package pkgs

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// 提供可重用的请求验证逻辑。
type RequestValidator struct {
	validate *validator.Validate
	trans    ut.Translator
}

// 创建一个新的 RequestValidator。
func NewRequestValidator() *RequestValidator {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("label")
		if name == "" {
			return fld.Name
		}
		return name
	})
	chinese := zh.New()
	uni := ut.New(chinese, chinese)
	trans, _ := uni.GetTranslator("zh")
	zh_translations.RegisterDefaultTranslations(validate, trans)
	return &RequestValidator{
		validate: validate,
		trans:    trans,
	}
}

// 验证给定的请求结构体。
func (v *RequestValidator) Validate(c *gin.Context, req interface{}) error {
	if err := v.validate.Struct(req); err != nil {
		// 获取第一个验证错误
		if validationErrors, ok := err.(validator.ValidationErrors); ok && len(validationErrors) > 0 {
			// 尝试获取自定义错误消息
			field, _ := reflect.TypeOf(req).Elem().FieldByName(validationErrors[0].Field())
			message := field.Tag.Get("message")
			if message == "" {
				// 如果没有自定义消息，则使用翻译后的错误
				Error(c, http.StatusBadRequest, validationErrors[0].Translate(v.trans))
			} else {
				// 使用自定义错误消息
				Error(c, http.StatusBadRequest, message)
			}
		} else {
			Error(c, http.StatusBadRequest, err.Error())
		}
		return err
	}
	return nil
}
