package copier

import "github.com/jinzhu/copier"

func Copy(dst any, src any, options ...option) error {
	if err := copier.CopyWithOption(dst, src, copyOption(options...)); err != nil {
		return err
	}
	return nil
}
