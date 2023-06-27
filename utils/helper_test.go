package utils

import (
	"testing"
)

func TestHighest(t *testing.T) {
	type args struct {
		scoreList []float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "test",
			args: args{scoreList: []float64{1, 2, 3, 4, 5}},
			want: 5,
		},
		{
			name: "test2",
			args: args{scoreList: []float64{9, 20, 3, 14, 5}},
			want: 20,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Highest(tt.args.scoreList); got != tt.want {
				t.Errorf("Highest() = %v, want %v", got, tt.want)
			}
		})
	}
}
