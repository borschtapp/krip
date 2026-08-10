package utils

import (
	"math"
	"regexp"
	"strings"
)

// Unit is a recognized mass or energy unit found alongside a number.
//
// Only SI/scientific abbreviations are recognized (g, mg, kcal, kJ, г, мг, ...), not
// translated unit words, since these are used verbatim across most languages' nutrition
// labels. Scripts/words not listed here are left unrecognized and fall back to a bare number.
type Unit string

const (
	UnitMicrogram   Unit = "mcg"
	UnitMilligram   Unit = "mg"
	UnitGram        Unit = "g"
	UnitKilogram    Unit = "kg"
	UnitKilocalorie Unit = "kcal"
	UnitKilojoule   Unit = "kj"
)

// kilojoulesPerKilocalorie is 1 kcal in kJ.
const kilojoulesPerKilocalorie = 4.184

// Unit token must be followed by a non-letter/non-digit or end of string. \b isn't used
// since Go's regexp treats it as ASCII-only, which would reject the Cyrillic units.
var numberUnitRegex = regexp.MustCompile(`(?i)(\d+(?:[.,]\d+)?)\s*(kilocalories?|kilojoules?|micrograms?|milligrams?|kilograms?|calories?|kcal|kj|mcg|µg|mg|kg|cal|grams?|g|мг|г)(?:[^\p{L}\p{N}]|$)`)

var unitAliases = map[string]Unit{
	"mcg": UnitMicrogram, "µg": UnitMicrogram, "microgram": UnitMicrogram, "micrograms": UnitMicrogram,
	"mg": UnitMilligram, "milligram": UnitMilligram, "milligrams": UnitMilligram, "мг": UnitMilligram,
	"g": UnitGram, "gram": UnitGram, "grams": UnitGram, "г": UnitGram,
	"kg": UnitKilogram, "kilogram": UnitKilogram, "kilograms": UnitKilogram,
	// "calorie(s)" colloquially means kilocalorie on nutrition labels, not the physical small calorie.
	"cal": UnitKilocalorie, "calorie": UnitKilocalorie, "calories": UnitKilocalorie,
	"kcal": UnitKilocalorie, "kilocalorie": UnitKilocalorie, "kilocalories": UnitKilocalorie,
	"kj": UnitKilojoule, "kilojoule": UnitKilojoule, "kilojoules": UnitKilojoule,
}

// massToGrams is the multiplier to convert a mass unit's magnitude to grams.
var massToGrams = map[Unit]float64{
	UnitMicrogram: 1e-6,
	UnitMilligram: 1e-3,
	UnitGram:      1,
	UnitKilogram:  1000,
}

// maxPlausibleMilligrams bounds the result of a mass conversion into milligrams.
// Sodium and cholesterol - the only fields ever converted to UnitMilligram - are always small, mg-scale quantities.
var maxPlausibleMilligrams = 10000.0

// roundUnit rounds a unit conversion result to 6 decimal places. This absorbs the binary floating-point noise.
func roundUnit(f float64) float64 {
	return math.Round(f*1e6) / 1e6
}

// FindNumberWithUnit extracts the first number in str and its unit, if recognized.
// unit is "" when none is found, meaning the caller should assume its own default.
func FindNumberWithUnit(str string) (value *float64, unit Unit) {
	if m := numberUnitRegex.FindStringSubmatch(str); m != nil {
		if n, err := ParseFloat(m[1]); err == nil {
			return &n, unitAliases[strings.ToLower(m[2])]
		}
	}
	return FindNumber(str), ""
}

// ParseMass parses str as a mass and converts it to target (e.g. UnitGram, UnitMilligram).
// Returns nil if str has a recognized unit that isn't a mass (e.g. "12 kcal") - wrong
// scale entirely, so there's no sensible number to fall back on.
func ParseMass(str string, target Unit) *float64 {
	value, unit := FindNumberWithUnit(str)
	if value == nil {
		return nil
	}
	if unit == "" {
		return value // trust the caller's documented default unit
	}
	fromFactor, ok := massToGrams[unit]
	if !ok {
		return nil
	}
	result := roundUnit(*value * fromFactor / massToGrams[target])
	if target == UnitMilligram && result > maxPlausibleMilligrams {
		return value // source unit is likely mislabeled; trust the raw number in target's unit instead
	}
	return &result
}

// ParseKilocalories parses str as an energy value and converts it to kilocalories.
// Returns nil if str has a recognized unit that isn't energy (e.g. "12 g").
func ParseKilocalories(str string) *float64 {
	value, unit := FindNumberWithUnit(str)
	if value == nil {
		return nil
	}
	switch unit {
	case "", UnitKilocalorie:
		return value
	case UnitKilojoule:
		kcal := roundUnit(*value / kilojoulesPerKilocalorie)
		return &kcal
	default:
		return nil
	}
}
