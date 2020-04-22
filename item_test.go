package main

import (
	"reflect"
	"testing"
	"time"
)

func TestMarshalTag(t *testing.T) {
	type args struct {
		tagString string
	}
	tests := []struct {
		name string
		args args
		want []*Tag
	}{
		{
			name: "normal",
			args: args{
				tagString: "Go:1.12:1.13:1.14,Python:3.7:3.8",
			},
			want: []*Tag{
				{Name: "Go", Versions: []string{"1.12", "1.13", "1.14"}},
				{Name: "Python", Versions: []string{"3.7", "3.8"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MarshalTag(tt.args.tagString); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnmarshalTag(t *testing.T) {
	type args struct {
		Tags []*Tag
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{
				Tags: []*Tag{
					{Name: "Go", Versions: []string{"1.12", "1.13", "1.14"}},
					{Name: "Python", Versions: []string{"3.7", "3.8"}},
				},
			},
			want: "Go:1.12:1.13:1.14,Python:3.7:3.8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnmarshalTag(tt.args.Tags); got != tt.want {
				t.Errorf("UnmarshalTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateFormat(t *testing.T) {
	got := time.Date(2020, 4, 22, 20, 43, 00, 0, time.UTC)
	want := "20200422"
	if DateFormat(got) != want {
		t.Errorf("DateFormat() = %v, want %v", got, want)
	}
}
