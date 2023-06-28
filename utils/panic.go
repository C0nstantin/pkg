package utils

// func panic if err input error  panic if err not nil
func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
