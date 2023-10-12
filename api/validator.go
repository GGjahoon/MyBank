package api

import (
	"github.com/GGjahoon/MySimpleBank/util"
	"github.com/go-playground/validator/v10"
)

// declare a validator function: validCurrency
// the function takes a fieldLevel interface as input,and return true when validate success
// this validator function will be Registered in server
var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	//fieldLevel.Field() to get the value of the filed,it's a reflection value
	//.Interface() to get its value as an empty interface. .(string) to convert this value to a string
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return util.IsSupportedCurrency(currency)
	}
	return false
}
