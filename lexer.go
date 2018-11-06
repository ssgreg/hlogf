package main

func fetchValue(data []byte) ([]byte, int, bool) {
	var r []byte

	i := 0
	length, ok := handleSpaces(data[i:])
	if !ok {
		return nil, 0, false
	}
	i += length

	switch data[i] {
	case '"':
		length, ok = handleString(data[i+1:])
		if !ok {
			return nil, 0, false
		}
		r = data[i : i+length+2]
		i += length + 2
	case '{', '[':
		length, ok = handleNonBasicValue(data[i+1:], data[i])
		if !ok {
			return nil, 0, false
		}
		r = data[i : i+length+2]
		i += length + 2
	default:
		length, ok = handleNonStringValue(data[i:])
		if !ok {
			return nil, 0, false
		}
		r = data[i : i+length]
		i += length
	}

	length, ok = handleSpaces(data[i:])
	if !ok {
		return nil, 0, false
	}
	i += length

	return r, i, true
}

func fetchKey(data []byte) ([]byte, int, bool) {
	var r []byte

	i := 0
	length, ok := handleSpaces(data[i:])
	if !ok {
		return nil, 0, false
	}
	i += length

	if data[i] != '"' {
		return nil, 0, false
	}
	length, ok = handleString(data[i+1:])
	if !ok {
		return nil, 0, false
	}
	r = data[i+1 : i+length+1]
	i += length + 2

	length, ok = handleSpaces(data[i:])
	if !ok {
		return nil, 0, false
	}
	i += length

	return r, i, true
}

func handleString(data []byte) (int, bool) {
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case '"':
			return i, true
		case '\\':
			i++
		}
	}

	return 0, false
}

func handleNonStringValue(data []byte) (int, bool) {
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case ' ', '\t', '\r', '\n', '{', '}', '[', ']', ',':
			return i, true
		}
	}

	return len(data), true
}

func handleSpaces(data []byte) (int, bool) {
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case ' ', '\t', '\r', '\n':
		default:
			return i, true
		}
	}

	return 0, true
}

func handleNonBasicValue(data []byte, delim byte) (int, bool) {
	switch delim {
	case '{':
		delim = '}'
	case '[':
		delim = ']'
	}

	for i := 0; i < len(data); i++ {
		switch data[i] {
		case delim:
			return i, true
		case '"':
			i++
			length, ok := handleString(data[i:])
			if !ok {
				return 0, false
			}
			i += length
		case '{', '[':
			i++
			length, ok := handleNonBasicValue(data[i:], data[i-1])
			if !ok {
				return 0, false
			}
			i += length
		}
	}

	return 0, false
}
