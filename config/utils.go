package config

import "bytes"

func ignoreComments(bs []byte) []byte {
	length := len(bs)
	if length == 0 {
		return bs
	}

	b := &bytes.Buffer{}

	// state      example        description
	//   0           -          Initial state
	//
	//                          Encountered '/' in state 0,
	//   1      int a = b; /    which means you may encounter a comment,
	//                          then enter state 1
	//
	//                          Encountered '/' in state 1,
	//   4      int a = b; //   which means entering the comment part,
	//                          then enter state 4
	//
	//                          Encountered '"' in state 0,
	//   7      char s[] = "    which means entering the string constant,
	//                          then enter state 7
	state := 0

	for i := 0; i < length; i++ {
		switch state {
		case 0:
			switch bs[i] {
			case '/':
				state = 1
			case '"':
				state = 7
				b.WriteByte(bs[i])
			case ' ':
			case '\r':
			case '\n':
			default:
				b.WriteByte(bs[i])
			}

		case 1:
			switch bs[i] {
			case '/':
				state = 4
			default:
				state = 0

				b.WriteByte(bs[i-1])
				b.WriteByte(bs[i])
			}

		case 4:
			if bs[i] == '\n' {
				state = 0
			}
			if bs[i] == '\r' {
				state = 0

				if bs[i+1] == '\n' {
					i++
				}
			}

		case 7:
			switch bs[i] {
			case '"':
				state = 0
				b.WriteByte(bs[i])
			default:
				b.WriteByte(bs[i])
			}

		default:
		}
	}

	return b.Bytes()
}
