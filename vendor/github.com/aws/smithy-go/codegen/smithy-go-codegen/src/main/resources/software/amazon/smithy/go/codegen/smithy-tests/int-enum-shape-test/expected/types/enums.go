// Code generated by smithy-go-codegen DO NOT EDIT.


package types

type Number = int32

// Enum values for Number
const (
	NumberAce Number = 1
	NumberTwo Number = 2
	NumberThree Number = 3
	NumberFour Number = 4
	NumberFive Number = 5
	NumberSix Number = 6
	NumberSeven Number = 7
	NumberEight Number = 8
	NumberNine Number = 9
	NumberTen Number = 10
	NumberJack Number = 11
	NumberQueen Number = 12
	NumberKing Number = 13
)

type Suit string

// Enum values for Suit
const (
	SuitDiamond Suit = "DIAMOND"
	SuitClub Suit = "CLUB"
	SuitHeart Suit = "HEART"
	SuitSpade Suit = "SPADE"
)

// Values returns all known values for Suit. Note that this can be expanded in the
// future, and so it is only as up to date as the client.
//
// The ordering of this slice is not guaranteed to be stable across updates.
func (Suit) Values() []Suit {
	return []Suit{
		"DIAMOND",
		"CLUB",
		"HEART",
		"SPADE",
	}
}