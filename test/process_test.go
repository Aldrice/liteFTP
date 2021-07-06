package main

import (
	"github.com/Aldrice/liteFTP/utils"
	"net"
	"testing"
)

func TestFormatAddr(t *testing.T) {
	type args struct {
		addr *net.TCPAddr
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "local addr",
			args: struct{ addr *net.TCPAddr }{
				addr: &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8121},
			},
			want: "127,0,0,1,31,185",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.FormatAddr(tt.args.addr); got != tt.want {
				t.Errorf("FormatAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}
