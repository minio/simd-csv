package simdcsv

import (
	"fmt"
	"strings"
	"testing"
)

func exampleToString(input string) string {
	out := strings.ReplaceAll(input, `\n`, " \n")
	return strings.ReplaceAll(out[7:], `  `, "")
}

func TestAmbiguityWithFSM(t *testing.T) {

	// A chunk is AMBIGUOUS if and only if the remaining valid starting
	// states are all either unquoted states or quoted state

	const ambiguous = `
       l  i  c  e  ,  "  ,  "  ,  1  6 \n  B  o  b  ,  "  ,  "  ,  1  7
    R  U  U  U  U  F  Q  Q  E  F  U  U  R  U  U  U  F  Q  Q  E  F  U  U
    F  U  U  U  U  F  Q  Q  E  F  U  U  R  U  U  U  F  Q  Q  E  F  U  U
    U  U  U  U  U  F  Q  Q  E  F  U  U  R  U  U  U  F  Q  Q  E  F  U  U
    Q  Q  Q  Q  Q  Q  E  F  Q  Q  Q  Q  Q  Q  Q  Q  Q  E  F  Q  Q  Q  Q
    E  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X`

	lines := strings.Split(ambiguous, "\n")
	csv := exampleToString(lines[1])

	initialStates := []int32{'R', 'F', 'U', 'Q', 'E'}

	endStates := make(map[uint8]bool)
	for i, state := range initialStates {
		out := augmentedFSM(csv, state)
		// fmt.Println(out)
		if out != lines[i+2] {
			t.Errorf("TestAmbiguityWithFSM mismatch: got %s, want %s", out, lines[i+2])
		}

		if out[len(out)-1] != 'X' {
			endStates[out[len(out)-1]] = true
		}
	}

	// Except for E, all other starting states successfully pass through the chunk.
	// Since the remaining starting states R, F, U, and Q fall into two categories, the
	// chunk is ambiguous.

	isAmbiguous := len(endStates) >= 2
	if !isAmbiguous {
		t.Errorf("TestAmbiguityWithFSM mismatch: got %v, want true", isAmbiguous)
	}
}

func TestUnambiguityWithFSM(t *testing.T) {

	const unambiguous = `
       l  i  c  e  ,  " \n  "  ,  1  6 \n  B  o  b  ,  "  M  "  ,  1  7
    R  U  U  U  U  F  Q  Q  E  F  U  U  R  U  U  U  F  Q  Q  E  F  U  U
    F  U  U  U  U  F  Q  Q  E  F  U  U  R  U  U  U  F  Q  Q  E  F  U  U
    U  U  U  U  U  F  Q  Q  E  F  U  U  R  U  U  U  F  Q  Q  E  F  U  U
    Q  Q  Q  Q  Q  Q  E  R  Q  Q  Q  Q  Q  Q  Q  Q  Q  E  X  X  X  X  X
    E  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X`

	lines := strings.Split(unambiguous, "\n")
	csv := exampleToString(lines[1])

	endStates := make(map[uint8]bool)

	initialStates := []int32{'R', 'F', 'U', 'Q', 'E'}

	for i, state := range initialStates {
		out := augmentedFSM(csv, state)
		//fmt.Println(out)
		if out != lines[i+2] {
			t.Errorf("TestUnambiguityWithFSM mismatch: got %s, want %s", out, lines[i+2])
		}

		if out[len(out)-1] != 'X' {
			endStates[out[len(out)-1]] = true
		}
	}

	// The chunk has an invalid state Q, because other are not
	// allowed after the state E, which is transited from the starting state Q
	// after reading the string `lice,"\n",16\nBob,"`. Thus, all valid starting
	// states are unquoted, and the example chunk is therefore unambiguous.

	isAmbiguous := len(endStates) >= 2
	if isAmbiguous {
		t.Errorf("TestUnambiguityWithFSM mismatch: got %v, want false", isAmbiguous)
	}
}

func TestAmbiquityWithPatterns(t *testing.T) {

	// q-o pattern
	//                    | quote  other
	// -------------------|-------------
	// R (Record start)   |   Q      Q
	// F (Field start)    |   Q      Q
	// U (Unquoted field) |   -      -
	// Q (Quoted field)   |   E      -
	// E (quoted End)     |   Q      Q

	// o-q pattern
	//                    | other  quote
	// -------------------|-------------
	// R (Record start)   |   U      -
	// F (Field start)    |   U      -
	// U (Unquoted field) |   U      -
	// Q (Quoted field)   |   Q      E
	// E (quoted End)     |   -      -

	// Both q-o and o-q patterns have a crucial property: for all
	// possible input states, the FSM transits into the same output state,
	// after reading an input string following the pattern

	// The chunk is ambiguous if and only if it contains neither
	// q-o pattern strings nor o-q pattern strings

	const ambiguous = `
       l  i  c  e  ,  "  ,  "  ,  1  6 \n  B  o  b  ,  "  ,  "  ,  1  7
   q-o .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
   o-q .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .`

	lines := strings.Split(ambiguous, "\n")
	csv := exampleToString(lines[1])

	hasQo := detectQoPattern([]byte(csv))
	hasOq := detectOqPattern([]byte(csv))
	isAmbiguous := hasQo == false && hasOq == false
	if !isAmbiguous {
		t.Errorf("TestAmbiquityWithPatterns mismatch: got %v, want true", isAmbiguous)
	}
}

func TestUnambiquityWithPatterns(t *testing.T) {

		const unambiguous = `
       l  i  c  e  ,  " \n  "  ,  1  6 \n  B  o  b  ,  "  M  "  ,  1  7
   q-o .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  X  .  .  .  .  .
   o-q .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  X  .  .  .  .`

	lines := strings.Split(unambiguous, "\n")
	csv := exampleToString(lines[1])

	hasQo := detectQoPattern([]byte(csv))
	hasOq := detectOqPattern([]byte(csv))

	isUnambiguous := hasQo || hasOq
	if !isUnambiguous {
		t.Errorf("TestUnambiquityWithPatterns mismatch: got %v, want false", isUnambiguous)
	}
}

func augmentedFSM(input string, state int32) string {
	out := fmt.Sprintf("    %c", state)
	for _, r := range input {
		state = augmentedFsmRune(r, state)
		out += fmt.Sprintf("  %c", state)
	}
	return out
}

func TestSyntaxErrorHandling(t *testing.T) {

	// transition table
	//                    | quote comma newline other
	// -------------------|--------------------------
	// R (Record start)   |   Q     F      R      U
	// F (Field start)    |   Q     F      R      U
	// U (Unquoted field) |   X     F      R      U
	// Q (Quoted field)   |   E     Q      Q      Q
	// E (quoted End)     |   Q     F      R      X
	// X (Error)          |   X     X      X      X

	const chunk1 = `
       A  l  i  c  e  ,  "  F  "  ,  "  H  i \n  " \n  B  o  b  ,  "  M  "  ,  "  H
    R  U  U  U  U  U  F  Q  Q  E  F  Q  Q  Q  Q  E  R  U  U  U  F  Q  Q  E  F  Q  Q`

	lines := strings.Split(chunk1, "\n")
	csv := exampleToString(lines[1])
	out := augmentedFSM(csv, 'R')
	if out != lines[2] {
		t.Errorf("TestSyntaxErrorHandling mismatch: got %s, want %s", out, lines[2])
	}

	const chunk2_WithError = `
       e  l  l  o \n  " \n  C  h  r  i  s  ,  M  "  ,  "  b  y  e  " \n  D  a  v  e
    Q  Q  Q  Q  Q  Q  E  R  U  U  U  U  U  F  U  X  X  X  X  X  X  X  X  X  X  X  X`

	lines = strings.Split(chunk2_WithError, "\n")
	csv = exampleToString(lines[1])
	out = augmentedFSM(csv, 'Q')
	if out != lines[2] {
		t.Errorf("TestSyntaxErrorHandling mismatch: got %s, want %s", out, lines[2])
	}

	const chunk3 = `
       ,  "  M  "  ,  "  M  o  r  n  i  n  g  ! \n  " \n
    X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X`

	lines = strings.Split(chunk3, "\n")
	csv = exampleToString(lines[1])
	out = augmentedFSM(csv, 'X')
	if out != lines[2] {
		t.Errorf("TestSyntaxErrorHandling mismatch: got %s, want %s", out, lines[2])
	}
}
