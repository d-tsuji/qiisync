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
			if got := marshalTag(tt.args.tagString); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("marshalTag() = %v, want %v", got, tt.want)
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
			if got := unmarshalTag(tt.args.Tags); got != tt.want {
				t.Errorf("unmarshalTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateFormat(t *testing.T) {
	got := time.Date(2020, 4, 22, 20, 43, 00, 0, time.UTC)
	want := "20200422"
	if dateFormat(got) != want {
		t.Errorf("dateFormat() = %v, want %v", got, want)
	}
}
