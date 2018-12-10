package util

import "testing"

func InferCommand(input string, options []string) (guess string) {
	scores := make(map[string]struct {
		string
		int
	}, len(options))

	for _, opt := range options {
		scores[opt] = struct {
			string
			int
		}{string: opt, int: 0}
	}

	const MaxLookahead = 4

	for _, char := range input {
		for name, data := range scores {
			for i := 0; i < MaxLookahead+1 && i < len(data.string); i++ {
				if char == int32(data.string[i]) {
					scores[name] = struct {
						string
						int
					}{string: data.string[i+1:], int: data.int + MaxLookahead - i}
					break
				}
			}
		}
	}

	var score int

	for o, s := range scores {
		if s.int > score {
			guess, score = o, s.int
		}
	}

	return
}

func TestInference(*testing.T) {

}
