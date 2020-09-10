package pkg

import "testing"

func TestCalculateMaxAge(t *testing.T) {
	type args struct {
		shelfLife int32
		decayRate float32
		factor    int8
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "CalculateMaxAge_Factor_1",
			args: args{shelfLife: 18, decayRate: 0.5, factor: 1},
			want: 12,
		},
		{
			name: "CalculateMaxAge_Factor_2",
			args: args{shelfLife: 18, decayRate: 0.5, factor: 2},
			want: 9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateMaxAge(tt.args.shelfLife, tt.args.decayRate, tt.args.factor); got != tt.want {
				t.Errorf("CalculateMaxAge() = %v, want %v", got, tt.want)
			}
		})
	}
}
