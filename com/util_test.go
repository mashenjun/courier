package com

import (
	"fmt"
	"gopkg.in/go-playground/validator.v9"
	"log"
	"testing"
)

func Test_Validation(t *testing.T) {
	type Address struct {
		Street string `validate:"required"`
		City   string `validate:"required"`
		Planet string `validate:"required"`
		Phone  string `validate:"required"`
	}

	type User struct {
		FirstName      string     `validate:"required"`
		LastName       string     `validate:"required"`
		Age            uint8      `validate:"gte=0,lte=130"`
		Email          string     `validate:"required,email"`
		FavouriteColor string     `validate:"hexcolor|rgb|rgba"`
		Addresses      []*Address `validate:"required,dive,required"` // a person can have a home and cottage...
	}

	validate := validator.New()

	//addr := Address{
	//	Street: "Eavesdown Docks",
	//	Planet: "Persphone",
	//	Phone:  "none",
	//}
	addrMap := map[string]string{
		"street": "Eavesdown",
		"planet": "Persphone",
		"phone": "none",
	}
	log.Printf("%+v\n", addrMap)
	err := validate.Struct(addrMap)
	if err != nil {
		if vErr, ok := err.(validator.ValidationErrors); ok {
			for _, v := range vErr {
				fmt.Printf("%+v\n",v.Field())
			}
		}
		t.Fatalf("could not validate struct: %+v", err)
	}
}