package model

import "testing"

func TestValidateString(t *testing.T) {
	type args struct {
		value      string
		constraint string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"Valid Email", args{value: "abc@def.gh", constraint: "EML"}, true, false},
		{"Invalid Email", args{value: "abcdef", constraint: "EML"}, false, false},
		{"Valid Alphanumeric (ANC)", args{value: "abc3434def", constraint: "ANC"}, true, false},
		{"Invalid Alphanumeric (ANC)", args{value: "abc343-4def", constraint: "ANC"}, false, false},
		{"Valid Alphanumeric with hyphen (ANH)", args{value: "abc343-4def", constraint: "ANH"}, true, false},
		{"Invalid Alphanumeric with hyphen (ANH)", args{value: "abc343-4def", constraint: "ANH"}, true, false},
		{"Valid Regex and valid value", args{value: "abcZ", constraint: "^[A-Za-z]+$"}, true, false},
		{"Valid Regex and invalid value", args{value: "abc343", constraint: "^[A-Za-z]+$"}, false, false},
		{"Invalid Regex", args{value: "abc343-4def", constraint: "())"}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateString(tt.args.value, tt.args.constraint)
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
