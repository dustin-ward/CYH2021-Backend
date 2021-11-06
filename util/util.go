package util

import "log"

func ErrHandle(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
