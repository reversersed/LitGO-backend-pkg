package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	shared_pb "github.com/reversersed/LitGO-proto/gen/go/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
)

type ValidationErrors validator.ValidationErrors
type Validator struct {
	*validator.Validate
}

func New() *Validator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	_ = v.RegisterValidation("primitiveid", validate_PrimitiveId)
	_ = v.RegisterValidation("lowercase", validate_LowercaseCharacter)
	_ = v.RegisterValidation("uppercase", validate_UppercaseCharacter)
	_ = v.RegisterValidation("digitrequired", validate_AtLeastOneDigit)
	_ = v.RegisterValidation("specialsymbol", validate_SpecialSymbol)
	_ = v.RegisterValidation("onlyenglish", validate_OnlyEnglish)
	_ = v.RegisterValidation("eqfield", validate_FieldsEqual)
	return &Validator{v}
}
func (v *Validator) StructValidation(data any) error {
	result := v.Validate.Struct(data)

	if result == nil {
		return nil
	}
	if er, ok := result.(*validator.InvalidValidationError); ok {
		return status.Error(codes.Internal, er.Error())
	}
	details := make([]protoadapt.MessageV1, 0)
	errors, ok := result.(validator.ValidationErrors)
	if !ok {
		return status.Error(codes.Internal, "wrong errors format provided")
	}
	for _, i := range errors {
		actual := fmt.Sprintf("%v", i.Value())
		if strings.Contains(strings.ToLower(i.Field()), "password") {
			actual = ""
		}
		details = append(details, &shared_pb.ErrorDetail{
			Field:       i.Field(),
			Struct:      i.StructNamespace(),
			Tag:         i.Tag(),
			TagValue:    i.Param(),
			Description: errorToStringByTag(i),
			Actualvalue: actual,
		})
	}
	stat, err := status.New(codes.InvalidArgument, "validation failed, see the details").WithDetails(details...)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return stat.Err()
}
func errorToStringByTag(err validator.FieldError) string {
	mapListErrors := map[string]string{
		"required":             "%s: field is required",
		"oneof":                "%s: field can only be: %s",
		"min":                  "%s: must be at least %s characters length",
		"max":                  "%s: can't be more than %s characters length",
		"lte":                  "%s: must be less or equal than %s",
		"gte":                  "%s: must be greater or equal than %s",
		"lt":                   "%s: must be less than %s",
		"gt":                   "%s: must be greater than %s",
		"email":                "%s: must be a valid email",
		"jwt":                  "%s: must be a JWT token",
		"lowercase":            "%s: must contain at least one lowercharacter",
		"uppercase":            "%s: must contain at least one uppercharacter",
		"digitrequired":        "%s: must contain at least one digit",
		"specialsymbol":        "%s: must contain at least one special symbol",
		"onlyenglish":          "%s: must contain only latin characters",
		"primitiveid":          "%s: must be a primitive id type",
		"eqfield":              "%s: field must be equal to %s field's value",
		"required_without_all": "%s: at least one field must be present",
	}
	format, ok := mapListErrors[err.Tag()]

	if !ok {
		return err.Error()
	} else {
		if strings.Count(format, "%s") == 2 {
			return fmt.Sprintf(format, err.Field(), err.Param())
		} else {
			return fmt.Sprintf(format, err.Field())
		}
	}
}
func validate_FieldsEqual(fl validator.FieldLevel) bool {
	return fl.Field().String() == fl.Parent().FieldByName(fl.Param()).String()
}
func validate_PrimitiveId(field validator.FieldLevel) bool {
	var obj primitive.ObjectID
	if slice, ok := field.Field().Interface().([]string); ok {
		if len(slice) == 0 {
			return true
		}
		for _, v := range slice {
			_, err := primitive.ObjectIDFromHex(v)
			if err != nil {
				return false
			}
		}
		return true
	}
	_, err := primitive.ObjectIDFromHex(field.Field().String())
	return (err == nil) || (field.Field().Kind() == reflect.TypeOf(obj).Kind()) || len(field.Field().String()) == 0
}
func validate_OnlyEnglish(field validator.FieldLevel) bool {
	mathed, err := regexp.MatchString(`^[a-zA-Z]+$`, field.Field().String())
	if err != nil {
		return false
	}
	if !mathed && len(field.Field().String()) > 0 {
		return false
	}
	return true
}
func validate_LowercaseCharacter(field validator.FieldLevel) bool {
	mathed, err := regexp.MatchString("[a-z]+", field.Field().String())
	if err != nil {
		return false
	}
	if !mathed {
		return false
	}
	return true
}
func validate_UppercaseCharacter(field validator.FieldLevel) bool {
	mathed, err := regexp.MatchString("[A-Z]+", field.Field().String())
	if err != nil {
		return false
	}
	if !mathed {
		return false
	}
	return true
}
func validate_AtLeastOneDigit(field validator.FieldLevel) bool {
	mathed, err := regexp.MatchString("[0-9]+", field.Field().String())
	if err != nil {
		return false
	}
	if !mathed {
		return false
	}
	return true
}
func validate_SpecialSymbol(field validator.FieldLevel) bool {
	mathed, err := regexp.MatchString("[!@#\\$%\\^&*()_\\+-.,]+", field.Field().String())
	if err != nil {
		return false
	}
	if !mathed {
		return false
	}
	return true
}
