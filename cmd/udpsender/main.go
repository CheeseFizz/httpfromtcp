package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	localaddress := "localhost:42070"
	remoteaddress := "localhost:42069"

	laddr, err := net.ResolveUDPAddr("udp", localaddress)
	if err != nil {
		fmt.Println(err)
		return
	}
	raddr, err := net.ResolveUDPAddr("udp", remoteaddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	fmt.Printf("Writing to %s\n", raddr.String())

	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		in, err := r.ReadString(byte('\n'))
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = conn.Write([]byte(in))
		if err != nil {
			log.Println(err)
		}
	}
}
