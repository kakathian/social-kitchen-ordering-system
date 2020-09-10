package kitchen

import (
	"testing"
)

func TestStart(t *testing.T) {
	type args struct {
		noOfOrdersToRead int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Start(tt.args.noOfOrdersToRead)
		})
	}
}

func Test_internalProcess(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			internalProcess()
		})
	}
}
