package retry

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestDoFunc(t *testing.T) {
	var try = 0
	_ = DoFunc(5, 0, func() error {
		if try < 5 {
			try++
			return errors.New("Try is not five")
		}
		return nil
	})
	if try != 5 {
		t.Error("Retry failed, expected try = 5")
	}

}

func TestDoFunc_Nil(t *testing.T) {
	var try = 5
	_ = DoFunc(1, 0, func() error {
		try--
		return nil
	})

	if try != 4 {
		t.Error("Failed to stop retry, expected try = 4")
	}
}

func TestDo(t *testing.T) {
	var notFunc int

	sum := func(nums ...int) (int, error) {
		var result int
		for _, n := range nums {
			result = result + n
		}
		return result, nil
	}

	div := func(a, b float64) (float64, error) {
		if b == 0 {
			return 0, errors.New("Can not divide by zero")
		}
		return a / b, nil
	}

	voidFunc := func() {

	}

	noErrorFunc := func() bool {
		fmt.Println("I'll executed only once as I don't return any error interface")
		return false
	}

	testcases := []struct {
		Tag           string
		Func          interface{}
		Args          []interface{}
		Result        interface{}
		ExpectedError bool
	}{
		{
			Tag:    "Add 1 to 4 and expected result 10",
			Func:   sum,
			Args:   []interface{}{1, 2, 3, 4},
			Result: 10,
		},
		{
			Tag:    "Add 1 to 5 and expected result 15",
			Func:   sum,
			Args:   []interface{}{1, 2, 3, 4, 5},
			Result: 15,
		},
		{
			Tag:    "Div 9.0/3.0 and expected result 3.0",
			Func:   div,
			Args:   []interface{}{9.0, 3.0},
			Result: 3.0,
		},
		{
			Tag:           "Div 9.0/0.0 and expected result 0 with error",
			Func:          div,
			Args:          []interface{}{9.0, 0.0},
			Result:        0.0,
			ExpectedError: true,
		},
		{
			Tag:           "As div is not a variadic func, if args mismatch we expect error",
			Func:          div,
			Args:          []interface{}{12.0, 3.0, 4.0},
			ExpectedError: true,
		},
		{
			Tag:           "As 'notFunc' is not a function we expect an error from retry package",
			Func:          notFunc,
			ExpectedError: true,
		},
		{
			Tag:           "As 'voidFunc' is does not return anything we expect an error from retry package",
			Func:          voidFunc,
			ExpectedError: true,
		},
		{
			Tag:    "As 'noErrorFunc' is does not return error we silently try to execute the func only once",
			Func:   noErrorFunc,
			Result: false,
		},
	}

	for _, tc := range testcases {

		out, err := Do(2, 1*time.Millisecond, tc.Func, tc.Args...)
		if err != nil && !tc.ExpectedError {
			t.Error(tc.Tag, err)
		}

		if !tc.ExpectedError && out != nil {
			if out[0].Interface() != tc.Result {
				t.Errorf("Failed: %s \nExpected: %v \nGot: %v", tc.Tag, tc.Result, out[0].Interface())
			}
		}
	}
}
