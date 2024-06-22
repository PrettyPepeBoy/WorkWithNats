package validation

import (
	"TestTaskNats/internal/models"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

var checkers = map[string]func(val reflect.Value) bool{
	"email": func(val reflect.Value) bool {

		var mailsSuffixes = [3]string{"@mail.ru", "@yandex.ru", "@gmail.com"}

		if val.Kind() != reflect.String {
			return false
		}
		str := val.String()
		if len(str) < 4 {
			return false
		}
		if len(str) > 50 {
			return false
		}

		for _, elem := range mailsSuffixes {
			if strings.HasPrefix(str, elem) {
				return true
			}
		}

		return false
	},

	"password": func(val reflect.Value) bool {
		if val.Len() > 20 {
			return false
		}

		if val.Len() < 8 {
			return false
		}

		//todo what password can not have?
		return true
	},

	"is_word": func(val reflect.Value) bool {
		if len(val.String()) > 20 {
			return false
		}
		if len(val.String()) == 0 {
			return false
		}
		for elem := range val.String() {
			if !((elem >= 'a' && elem <= 'z') || (elem >= 'A' && elem <= 'Z')) {
				return false
			}
		}
		return true
	},

	"is_number": func(val reflect.Value) bool {
		if val.Int() < 0 {
			return false
		}
		if val.Int() > 2<<30 {
			return false
		}
		return true
	},
}

//рефлексия применяется только к объектам о которых мы не знаем

func ValidProductBody(product models.ProductBody) bool {
	val, tp := reflect.ValueOf(product), reflect.TypeOf(product)
	for i := 0; i < val.NumField(); i++ {
		str := tp.Field(i).Tag.Get("validate")
		tags := strings.Split(str, ",")
		for _, elem := range tags {
			fn, ok := checkers[elem]
			if !ok {
				logrus.Error("not existing tag ", elem)
				return false
			}

			if !fn(val.Field(i).Elem()) {
				return false
			}
		}
	}
	return true
}
