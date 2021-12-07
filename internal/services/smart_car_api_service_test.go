package services

import (
	"reflect"
	"testing"
)

func Test_parseSmartCarYears(t *testing.T) {
	singleYear := "2019"
	rangeYear := "2012-2015"
	plusYears := "2019+" // figure out way to stub out time so we always get same. https://stackoverflow.com/questions/18970265/is-there-an-easy-way-to-stub-out-time-now-globally-during-test
	garbage := "bobby"

	tests := []struct {
		name     string
		yearsPtr *string
		want     []int
		wantErr  bool
	}{
		{
			name:     "nil return",
			yearsPtr: nil,
			want:     nil,
			wantErr:  false,
		},
		{
			name:     "parse single year",
			yearsPtr: &singleYear,
			want:     []int{ 2019 },
			wantErr:  false,
		},
		{
			name:     "parse range year",
			yearsPtr: &rangeYear,
			want:     []int{2012, 2013, 2014, 2015},
			wantErr:  false,
		},
		{
			name:     "parse year plus",
			yearsPtr: &plusYears,
			want:     []int{2019, 2020, 2021},
			wantErr:  false,
		},
		{
			name:     "error on garbage",
			yearsPtr: &garbage,
			want:     nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSmartCarYears(tt.yearsPtr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSmartCarYears() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSmartCarYears() got = %v, want %v", got, tt.want)
			}
		})
	}
}
