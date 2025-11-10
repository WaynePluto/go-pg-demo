package pkgs

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/samber/mo"
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

// 绑定并返回路径参数。
func BindUri[T any](c *gin.Context) mo.Result[*T] {
	var req T
	if err := c.ShouldBindUri(&req); err != nil {
		return mo.Err[*T](NewApiError(http.StatusBadRequest, err.Error()))
	}
	return mo.Ok(&req)
}

func BindQuery[T any](c *gin.Context) mo.Result[*T] {
	var req T
	if err := c.ShouldBindQuery(&req); err != nil {
		return mo.Err[*T](NewApiError(http.StatusBadRequest, err.Error()))
	}
	return mo.Ok(&req)
}

// 绑定并返回请求体的JSON数据。
func BindJSON[T any](c *gin.Context) mo.Result[*T] {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		return mo.Err[*T](NewApiError(http.StatusBadRequest, err.Error()))
	}
	return mo.Ok(&req)
}

// 同时绑定路径参数和请求体的JSON数据。
func BindUriAndJSON[T any](c *gin.Context) mo.Result[*T] {
	var uriReq T
	if err := c.ShouldBindUri(&uriReq); err != nil {
		return mo.Err[*T](NewApiError(http.StatusBadRequest, err.Error()))
	}
	// 打印req
	fmt.Printf("uri请求参数: %+v\n", uriReq)
	var jsonReq T
	if err := c.ShouldBindJSON(&jsonReq); err != nil {
		return mo.Err[*T](NewApiError(http.StatusBadRequest, err.Error()))
	}
	fmt.Printf("json请求参数: %+v\n", jsonReq)
	// 合并两个结构体的字段值，jsonReq的值覆盖uriReq的值
	uriVal := reflect.ValueOf(&uriReq).Elem()
	jsonVal := reflect.ValueOf(&jsonReq).Elem()
	for i := 0; i < uriVal.NumField(); i++ {
		field := uriVal.Field(i)
		jsonField := jsonVal.Field(i)
		// 仅当jsonField不为零值时才覆盖uriField
		if !jsonField.IsZero() {
			field.Set(jsonField)
		}
	}
	return mo.Ok(&uriReq)
}

// 验证给定的请求结构体。
func ValidateV2[T any](v *RequestValidator) func(req *T) mo.Result[*T] {
	return func(req *T) mo.Result[*T] {
		if err := v.validate.Struct(req); err != nil {
			errMsg := err.Error()
			// 获取第一个验证错误
			if validationErrors, ok := err.(validator.ValidationErrors); ok && len(validationErrors) > 0 {
				// 尝试获取自定义错误消息
				field, _ := reflect.TypeOf(req).Elem().FieldByName(validationErrors[0].Field())
				message := field.Tag.Get("message")
				if message == "" {
					// 如果没有自定义消息，则使用翻译后的错误
					errMsg = validationErrors[0].Translate(v.trans)
				} else {
					// 使用自定义错误消息
					errMsg = message
				}
			}
			return mo.Err[*T](NewApiError(http.StatusBadRequest, errMsg))
		}
		return mo.Ok(req)
	}
}
