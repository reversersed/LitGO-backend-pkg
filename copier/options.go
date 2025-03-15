package copier

import (
	"fmt"

	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type option func(*copier.Option)

func copyOption(opt ...option) copier.Option {
	var option copier.Option

	for _, v := range opt {
		v(&option)
	}
	return option
}

var (
	WithPrimitiveToStringConverter = func(c *copier.Option) {
		c.Converters = append(c.Converters, copier.TypeConverter{SrcType: primitive.ObjectID{}, DstType: string(""), Fn: func(src any) (dst any, err error) {
			s, ok := src.(primitive.ObjectID)
			if !ok {
				return nil, fmt.Errorf("unable to convert %v to primitive object id", src)
			}
			return s.Hex(), nil
		}})
		c.Converters = append(c.Converters, copier.TypeConverter{SrcType: []primitive.ObjectID{}, DstType: []string{}, Fn: func(src any) (dst any, err error) {
			s, ok := src.([]primitive.ObjectID)
			if !ok {
				return nil, fmt.Errorf("unable to convert %v to primitive array object id", src)
			}
			result := make([]string, 0)
			for _, i := range s {
				result = append(result, i.Hex())
			}
			return result, nil
		}})

		c.Converters = append(c.Converters, copier.TypeConverter{DstType: primitive.ObjectID{}, SrcType: string(""), Fn: func(src any) (dst any, err error) {
			s, ok := src.(string)
			if !ok {
				return nil, fmt.Errorf("unable to convert %v to string", src)
			}
			i, err := primitive.ObjectIDFromHex(s)
			if err != nil {
				return nil, err
			}
			return i, nil
		}})
		c.Converters = append(c.Converters, copier.TypeConverter{DstType: []primitive.ObjectID{}, SrcType: []string{}, Fn: func(src any) (dst any, err error) {
			s, ok := src.([]string)
			if !ok {
				return nil, fmt.Errorf("unable to convert %v to string array", src)
			}
			result := make([]primitive.ObjectID, 0)
			for _, i := range s {
				d, err := primitive.ObjectIDFromHex(i)
				if err != nil {
					return nil, err
				}
				result = append(result, d)
			}
			return result, nil
		}})
	}
	WithIgnoreEmptyFields = func(c *copier.Option) {
		c.IgnoreEmpty = true
	}
)
