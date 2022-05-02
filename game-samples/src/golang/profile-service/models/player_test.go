package models

import "testing"

func TestValidEmails(t *testing.T) {
	var tests = []string{"good@gmail", "good@gmail.com", "good@somedomain"}

	for _, email := range tests {
		res := validateEmail(email)

		t.Logf("Testing '%s'", email)

		if res == false {
			t.Error(
				"For", email,
				"expected", false,
				"got", res,
			)
		}
	}

}

func TestInvalidEmails(t *testing.T) {

	var tests = []string{"bademail"}

	for _, email := range tests {
		res := validateEmail(email)

		if res == true {
			t.Error(
				"For", email,
				"expected", true,
				"got", res,
			)
		}
	}

}

func TestHashPassword(t *testing.T) {
	var tests = []string{"mypass", "somepass", "som1pass"}

	for _, pass := range tests {
		hash, err := hashPassword(pass)

		if err != nil {
			t.Error(
				"For", pass,
				"expected", "no error",
				"got", err,
			)
		}

		err = validatePassword(pass, hash)

		if err != nil {
			t.Error(
				"For", pass,
				"expected", "no error",
				"got", err,
			)
		}
	}

}
