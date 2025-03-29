package main

import validation "github.com/go-ozzo/ozzo-validation"

var (
	passwordLength = validation.Length(8, 72)
)
