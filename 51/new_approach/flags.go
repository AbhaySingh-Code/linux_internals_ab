package main

import "fmt"

//working with bitmask and bitwise operators.

func main() {
	var flags byte = 0x02    //SYN -> A single byte contains 8 bits. The hexadecimal value 0x02 translates to 00000010 in binary.
	isSYN := flags&0x02 != 0 //It is a logical AND condition if flags & 0x02 != 0, as flags is 0x02 anding it with 0x02 (1 & 1 = 1) (1 & 0 = 0 )
	//( 0 & 1 = 0) (0 & 0 = 1)
	isACK := flags&0x10 != 0
	fmt.Println("SYN:", isSYN, "ACK:", isACK)

	// now trying to set multiple flags
	/*
		0 0 0 0 0 0 1 0   (0x02 - SYN switch is ON)
		| 0 0 0 1 0 0 0 0   (0x10 - ACK switch is ON)
		-----------------
		0 0 0 1 0 0 1 0   (The Result: Both switches are now ON)
	*/
	flags = 0x02 | 0x10         //SYN + ACK
	fmt.Printf("%08b\n", flags) // Print the number in binary format, padded with zeros so it is 8 digits long.
}
