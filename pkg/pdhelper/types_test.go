package pdhelper

import (
	"reflect"
	"testing"
)

func TestPdSchedulerShow_FetchStoreIds(t *testing.T) {
	tests := []struct {
		name string
		it   PdSchedulerShow
		want []uint
	}{
		{
			name: "normal",
			it: []string{
				"evict-leader-scheduler-2",
				"evict-leader-scheduler-3",
				"evict-leader-scheduler-4",
			},
			want: []uint{
				2,
				3,
				4,
			},
		},
		{
			name: "no expected data",
			it: []string{
				"unexpected-2",
				"foo-bar-3",
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.it.FetchStoreIds(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchStoreIds() = %v, want %v", got, tt.want)
			}
		})
	}
}
