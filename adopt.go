package main

func adoptEntry(e *Entry) {
	if len(e.Level) == 0 {
		if len(e.Priority) == 3 {
			e.Level = priorityToLevel(e.Priority[1])
		}
	}

	if len(e.Time) == 0 {
		if len(e.SourceTimestamp) > 3 {
			e.Time = e.SourceTimestamp
		} else if len(e.RealtimeTimestamp) > 3 {
			e.Time = e.RealtimeTimestamp
		}
	}
}

func priorityToLevel(pr byte) []byte {
	switch pr {
	case '7':
		return []byte(`"debug"`)
	case '6', '5':
		return []byte(`"info"`)
	case '4':
		return []byte(`"warn"`)
	case '3', '2', '1':
		return []byte(`"error"`)
	default:
		return []byte(`"unknown"`)
	}
}
