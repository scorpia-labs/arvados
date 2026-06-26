// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package arvados

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	streamTokenRegexp    = regexp.MustCompile(`^([^\x00-\x20\\]|\\[0-3][0-7][0-7])+$`)
	streamNameRegexp     = regexp.MustCompile(`^(\.)(\/[^\/]+)*$`)
	fileTokenRegexp      = regexp.MustCompile(`^[0-9]+:[0-9]+:([^\x00-\x20\\]|\\[0-3][0-7][0-7])+$`)
	fileNameRegexp       = regexp.MustCompile(`^[0-9]+:[0-9]+:([^\/]+(\/[^\/]+)*)$`)
	emptyDirTokenRegexp  = regexp.MustCompile(`^0:0:\.$`)
	locatorRegexp        = regexp.MustCompile(`^[0-9a-fA-F]{32}\+[0-9]+(\+[A-Z][A-Za-z0-9@_-]*)*$`)
	non8BitEncodedRegexp = regexp.MustCompile(`(^|[^\\])\\[4-7][0-7][0-7]`)
	unescapeRegexp       = regexp.MustCompile(`\\(\\|[0-7]{3})`)
)

// unescape unescapes backslash sequences like \ooo (octal) and \\ in manifest tokens.
func unescape(s string) string {
	return unescapeRegexp.ReplaceAllStringFunc(s, func(match string) string {
		if match == `\\` {
			return `\`
		}
		val, _ := strconv.ParseInt(match[1:], 8, 32)
		return string([]byte{byte(val)})
	})
}

// hasDotOrDotDotPath returns true if the given slash-separated path contains "." or ".." components.
// For stream names, the first component is allowed to be "." according to manifest spec, so we
// usually strip the leading "." before checking, or check for "/./" and "/../".
func hasDotOrDotDotPath(s string) bool {
	parts := strings.Split(s, "/")
	for _, p := range parts {
		if p == "." || p == ".." {
			return true
		}
	}
	return false
}

// ValidateManifest validates the given manifest text according to the Arvados specification.
// It returns nil if valid, or an error detailing the validation failure.
func ValidateManifest(manifest string) error {
	if manifest == "" {
		return nil
	}
	if !strings.HasSuffix(manifest, "\n") {
		return errors.New("Invalid manifest: does not end with newline")
	}

	lines := strings.Split(manifest, "\n")
	// The last element is empty because of the trailing newline
	lines = lines[:len(lines)-1]

	for i, line := range lines {
		lineCount := i + 1

		if strings.HasSuffix(line, " ") {
			return fmt.Errorf("Manifest invalid for stream %d: trailing space", lineCount)
		}

		words := strings.Split(line, " ")
		if len(words) == 0 || words[0] == "" {
			return fmt.Errorf("Manifest invalid for stream %d: missing stream name", lineCount)
		}

		// 1. Validate stream name (first token)
		word := words[0]
		if non8BitEncodedRegexp.MatchString(word) {
			return fmt.Errorf("Manifest invalid for stream %d: >8-bit encoded chars not allowed on stream token %q", lineCount, word)
		}

		unescapedWord := unescape(word)
		validStreamName := streamTokenRegexp.MatchString(word) &&
			streamNameRegexp.MatchString(unescapedWord) &&
			!strings.Contains(unescapedWord, "/./") &&
			!strings.Contains(unescapedWord, "/../") &&
			!strings.HasSuffix(unescapedWord, "/.") &&
			!strings.HasSuffix(unescapedWord, "/..")
		if !validStreamName {
			return fmt.Errorf("Manifest invalid for stream %d: missing or invalid stream name %q", lineCount, word)
		}

		words = words[1:]

		// 2. Validate Locators
		locatorCount := 0
		for len(words) > 0 && locatorRegexp.MatchString(words[0]) {
			locatorCount++
			words = words[1:]
		}

		if locatorCount == 0 {
			badToken := ""
			if len(words) > 0 {
				badToken = fmt.Sprintf(" %q", words[0])
			}
			return fmt.Errorf("Manifest invalid for stream %d: missing or invalid locator%s", lineCount, badToken)
		}

		// 3. Validate File Tokens
		fileTokenCount := 0
		for len(words) > 0 {
			word = words[0]
			if non8BitEncodedRegexp.MatchString(word) {
				return fmt.Errorf("Manifest invalid for stream %d: >8-bit encoded chars not allowed on file token %q", lineCount, word)
			}

			unescapedFileWord := unescape(word)

			isEmptyDir := emptyDirTokenRegexp.MatchString(unescapedFileWord)

			isValidFile := false
			if fileTokenRegexp.MatchString(word) && fileNameRegexp.MatchString(unescapedFileWord) {
				// check for . or ..
				matches := fileNameRegexp.FindStringSubmatch(unescapedFileWord)
				if len(matches) > 1 {
					filenamePart := matches[1]
					if !hasDotOrDotDotPath(filenamePart) {
						isValidFile = true
					}
				}
			}

			if isEmptyDir || isValidFile {
				fileTokenCount++
				words = words[1:]
			} else {
				break
			}
		}

		if len(words) > 0 {
			return fmt.Errorf("Manifest invalid for stream %d: invalid file token %q", lineCount, words[0])
		} else if fileTokenCount == 0 {
			return fmt.Errorf("Manifest invalid for stream %d: no file tokens", lineCount)
		}
	}

	return nil
}
