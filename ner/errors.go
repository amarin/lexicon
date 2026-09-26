package ner

import "errors"

// ErrUnknownProfile: Doc.Profile is not in Config.Profiles.
var ErrUnknownProfile = errors.New("ner: unknown profile")
