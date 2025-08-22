package appfeatures

import (
	"fmt"
	"strings"
)

// Features describes the features of an application.
type Features uint

// The different kinds of individual features.
const (
	FeatureNone       Features = 0 // The zero value for this type.
	FeatureConnection          = 1 << iota
	FeaturePairing
	FeatureSendFile
	FeatureReceiveFile
	FeatureNetwork
	FeatureMediaPlayer
)

// FeatureMap holds a list of descriptions for each feature.
var FeatureMap = map[Features]string{
	FeatureConnection:  "Bluetooth Connection",
	FeaturePairing:     "Bluetooth Pairing",
	FeatureSendFile:    "OBEX Send Files",
	FeatureReceiveFile: "OBEX Receive Files",
	FeatureNetwork:     "PANU/DUN Network Connection",
	FeatureMediaPlayer: "Media Player",
}

// Add adds the provided features to the existing features.
func (c *Features) Add(features ...Features) {
	for _, f := range features {
		*c |= f
	}
}

// Remove removes the provided features from the existing features.
func (c *Features) Remove(features ...Features) {
	for _, f := range features {
		*c &= f
	}
}

// AbsentFeatures returns a list of features that are not present in the existing features.
func (c *Features) AbsentFeatures() []Features {
	s := make([]Features, 0, len(FeatureMap))

	for feature := range FeatureMap {
		if *c&feature == 0 {
			s = append(s, feature)
		}
	}

	return s
}

// String converts a set of features to a comma-separated string of
// their respective descriptions.
func (c *Features) String() string {
	s := make([]string, 0, len(FeatureMap))

	for feature, title := range FeatureMap {
		if *c&feature != 0 {
			s = append(s, title)
		}
	}

	return strings.Join(s, ", ")
}

// Slice returns a slice of individual app features.
func (c *Features) Slice() []Features {
	s := make([]Features, 0, len(FeatureMap))

	for feature := range FeatureMap {
		if *c&feature != 0 {
			s = append(s, feature)
		}
	}

	return s
}

// FeatureSet holds all supported features and feature related errors.
type FeatureSet struct {
	Supported Features
	Errors    Errors
}

// NewFeatureSet returns a new set (of features).
func NewFeatureSet(features Features, errors Errors) *FeatureSet {
	return &FeatureSet{
		Supported: features,
		Errors:    errors,
	}
}

// MergedFeatureSet merges and returns all available features in a feature set.
func MergedFeatureSet() *FeatureSet {
	var features Features

	for c := range FeatureMap {
		features |= c
	}

	return &FeatureSet{Supported: features}
}

// HasAny returns if the feature feature sets contains any of the provided features.
func (c *FeatureSet) HasAny(compare ...Features) bool {
	var compared int

	for _, toCompare := range compare {
		if compared > 0 {
			break
		}

		if c.Supported&toCompare != 0 {
			compared++
		}
	}

	return compared > 0
}

// Has returns if the feature set has all of the provided features.
func (c *FeatureSet) Has(compare ...Features) bool {
	var compared int

	for _, toCompare := range compare {
		if c.Supported&toCompare != 0 {
			compared++
		}
	}

	return compared > 0 && compared == len(compare)
}

// Error describes an error which occurred while attempting
// to enable support for the specified feature.
type Error struct {
	Feature       Features
	FeatureErrors error
}

// Errors holds a list of feature based errors.
type Errors struct {
	errors map[Features]Error
}

// NewError returns a feature-based Error.
func NewError(c Features, err error) *Error {
	return &Error{
		Feature:       c,
		FeatureErrors: err,
	}
}

// Append appends a single feature error to the feature error list.
func (c *Errors) Append(e *Error) {
	if c.errors == nil {
		c.errors = make(map[Features]Error)
	}

	c.errors[e.Feature] = *e
}

// Exists checks and returns all feature based errors.
func (c *Errors) Exists() (map[Features]Error, bool) {
	return c.errors, c.errors != nil
}

// Error returns a text representation of the feature error.
func (c *Error) Error() string {
	return fmt.Sprintf(
		"Capabilities '%s' cannot be activated: %s",
		c.Feature.String(), c.FeatureErrors,
	)
}
