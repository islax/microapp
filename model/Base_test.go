package model

import "testing"

func TestValidateString(t *testing.T) {
	type args struct {
		value      string
		constraint StringType
		regex      []string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"Valid Email", args{value: "abc@def.gh", constraint: Email}, true, false},
		{"Invalid Email", args{value: "abcdef", constraint: Email}, false, false},
		{"Valid Alphanumeric (ANC)", args{value: "abc3434def", constraint: AlphaNumeric}, true, false},
		{"Invalid Alphanumeric (ANC)", args{value: "abc343-4def", constraint: AlphaNumeric}, false, false},
		{"Valid Alphanumeric with hyphen (ANH)", args{value: "abc343-4def", constraint: AlphaNumericAndHyphen}, true, false},
		{"Invalid Alphanumeric with hyphen (ANH)", args{value: "abc343-4def", constraint: AlphaNumericAndHyphen}, true, false},
		{"RegEx constraint with valid regex and valid value", args{value: "abcZ", constraint: RegEx, regex: []string{"^[A-Za-z]+$"}}, true, false},
		{"RegEx constraint with valid regex and invalid value", args{value: "abc343", constraint: RegEx, regex: []string{"^[A-Za-z]+$"}}, false, false},
		{"RegEx constraint with invalid regex", args{value: "abc343-4def", constraint: RegEx, regex: []string{"())"}}, false, true},
		{"RegEx constraint with no regex", args{value: "abc343-4def", constraint: RegEx}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateString(tt.args.value, tt.args.constraint, tt.args.regex...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateString() = %v, want %v", got, tt.want)
			}
		})
	}
}
