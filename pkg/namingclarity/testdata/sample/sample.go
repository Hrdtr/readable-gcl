package sample

import (
	"context"
	"math/big"
	"testing"
)

// Single-letter argument should be flagged.
func calculate(x int) int { // want `single-letter argument prohibited: x`
	return x * 2
}

// Multi-letter argument is allowed.
func addNumbers(firstNumber int, secondNumber int) int {
	return firstNumber + secondNumber
}

// Prohibited Hungarian notation.
func demoHungarianNotation() {
	var countInt int // want `meaningless representation suffix prohibited: countInt`
	_ = countInt

	var activeBool bool // want `meaningless representation suffix prohibited: activeBool`
	_ = activeBool

	var userBigInt *big.Int // allowed: math/big.Int
	_ = userBigInt
}

// Bad initialisms.
func demoInitialisms() {
	var userId string // want `disallowed initialism casing "Id" in userId, use "ID" instead`
	_ = userId

	var userID string // allowed
	_ = userID
}

// Disallowed abbreviations.
func demoAbbreviations() {
	var userCfg string // want `disallowed identifier term "cfg" prohibited: userCfg`
	_ = userCfg

	var userConfiguration string // allowed
	_ = userConfiguration
}

// Special type rules.
func demoSpecialTypes(ctx context.Context) {
	var customContext context.Context // want `"customContext" should be named "ctx" or suffixed with "Ctx"`
	_ = customContext

	var requestCtx context.Context // allowed: ends with Ctx
	_ = requestCtx
}

type BaseEntity struct {
	Timestamp int
}

// Unexported struct fields.
type Person struct {
	BaseEntity
	ID   string // Exported struct field allowed
	URL  string // Exported struct field allowed
	pInt int    // want `meaningless representation suffix prohibited: pInt`
}

func helperForTests(t *testing.T) { // allowed: t is testing.T
	_ = t
}

func (p *Person) String() string { // want `single-letter receiver prohibited: p`
	return p.ID
}

func (person *Person) GetID() string { // allowed: multi-letter receiver
	return person.ID
}

func (_ *Person) BlankReceiver() {} // allowed: blank receiver

func (_ Person) ValueBlankReceiver() {} // allowed: blank receiver

func helperAnonymous(_ int, _ string) {} // allowed: blank arguments

func tupleReturn() (int, string) {
	return 1, "test"
}

func demoRangeAndAssignments() {
	items := []string{"hello"}
	for idxInt, valStr := range items { // want `meaningless representation suffix prohibited: idxInt` `meaningless representation suffix prohibited: valStr`
		_ = idxInt
		_ = valStr
	}

	resultInt, titleStr := tupleReturn() // want `meaningless representation suffix prohibited: resultInt` `meaningless representation suffix prohibited: titleStr`
	_ = resultInt
	_ = titleStr

	dict := map[string]int{"k": 1}
	entryInt, ok := dict["k"] // want `meaningless representation suffix prohibited: entryInt`
	_ = entryInt
	_ = ok

	itemOne, itemTwo := dict["k"] // allowed multi LHS assignment without tuple
	_ = itemOne
	_ = itemTwo

	var inferredValue = 100 // inferred var
	_ = inferredValue

	var personInstance Person
	personInstance.ID = "assigned" // selector assignment

	var errResult error // want `"errResult" should be named "err", suffixed with "Err", or prefixed with "Err"`
	_ = errResult

	var ErrNotFound error // allowed sentinel
	_ = ErrNotFound
}

