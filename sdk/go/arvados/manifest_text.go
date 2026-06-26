package arvados

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Keep locator: 32 hex chars + size + optional hints
	locatorRegex = regexp.MustCompile(`^([0-9a-f]{32})\+([0-9]+)(\+[A-Z][-A-Za-z0-9@_]*)*$`)
)

// ManifestStream represents a single stream in a manifest, parsed and validated.
type ManifestStream struct {
	StreamName string
	Locators   []BlockLocator
	FileTokens []FileToken
}

// BlockLocator represents a parsed and validated Keep locator.
type BlockLocator struct {
	Digest string
	Size   int
	Hints  []string
	Text   string
}

// FileToken represents a parsed and validated file segment token.
type FileToken struct {
	Position int64
	Length   int64
	Name     string
}

// ManifestScanner parses and validates an Arvados manifest string.
// It iterates over streams, providing one ManifestStream at a time.
type ManifestScanner struct {
	text   string
	pos    int
	lineno int
	err    error
	stream ManifestStream
}

// NewManifestScanner creates a new ManifestScanner for the given manifest text.
func NewManifestScanner(text string) *ManifestScanner {
	return &ManifestScanner{
		text:   text,
		pos:    0,
		lineno: 0,
	}
}

// Scan advances the scanner to the next stream. It returns false when the end
// of the manifest is reached or an error occurs.
func (s *ManifestScanner) Scan() bool {
	if s.err != nil {
		return false
	}
	if s.pos >= len(s.text) {
		return false
	}

	s.lineno++

	// Find the end of the current line
	lineEnd := strings.IndexByte(s.text[s.pos:], '\n')

	// Check for invalid characters in the line
	checkLen := lineEnd
	if lineEnd == -1 {
		checkLen = len(s.text) - s.pos
	}
	for i := 0; i < checkLen; i++ {
		c := s.text[s.pos+i]
		if c < 32 || c > 126 {
			s.err = fmt.Errorf("line %d: invalid character %q", s.lineno, c)
			return false
		}
	}

	if lineEnd == -1 {
		if s.text[s.pos:] == "" {
			return false
		}
		s.err = fmt.Errorf("line %d: no trailing newline", s.lineno)
		return false
	}

	line := s.text[s.pos : s.pos+lineEnd]
	s.pos += lineEnd + 1 // +1 to skip the newline

	if len(line) == 0 {
		s.err = fmt.Errorf("line %d: empty stream", s.lineno)
		return false
	}

	tokens := strings.Split(line, " ")

	if len(tokens) == 0 || tokens[0] == "" {
		s.err = fmt.Errorf("line %d: no stream name", s.lineno)
		return false
	}

	s.stream = ManifestStream{
		StreamName: tokens[0],
	}

	// Validate stream name
	if err := s.validateStreamName(s.stream.StreamName); err != nil {
		s.err = fmt.Errorf("line %d: %v", s.lineno, err)
		return false
	}

	s.stream.StreamName = manifestUnescape(s.stream.StreamName)

	var anyFileTokens bool
	var anyLocators bool

	for i := 1; i < len(tokens); i++ {
		token := tokens[i]
		if token == "" {
			s.err = fmt.Errorf("line %d: invalid empty token", s.lineno)
			return false
		}

		if strings.ContainsRune(token, ':') {
			// File segment
			if !anyLocators {
				s.err = fmt.Errorf("line %d: bad file segment %q: appears before locators", s.lineno, token)
				return false
			}
			anyFileTokens = true

			ft, err := s.parseFileToken(token)
			if err != nil {
				s.err = fmt.Errorf("line %d: bad file segment %q: %v", s.lineno, token, err)
				return false
			}
			s.stream.FileTokens = append(s.stream.FileTokens, ft)
		} else {
			// Keep locator
			if anyFileTokens {
				s.err = fmt.Errorf("line %d: bad locator %q: appears after file segments", s.lineno, token)
				return false
			}
			anyLocators = true

			loc, err := s.parseLocator(token)
			if err != nil {
				s.err = fmt.Errorf("line %d: bad locator %q: %v", s.lineno, token, err)
				return false
			}
			s.stream.Locators = append(s.stream.Locators, loc)
		}
	}

	if !anyLocators {
		s.err = fmt.Errorf("line %d: no locators", s.lineno)
		return false
	}
	if !anyFileTokens {
		s.err = fmt.Errorf("line %d: no file segments", s.lineno)
		return false
	}

	return true
}

// Stream returns the current stream parsed by the most recent call to Scan.
func (s *ManifestScanner) Stream() ManifestStream {
	return s.stream
}

// Err returns the first error encountered by the scanner.
func (s *ManifestScanner) Err() error {
	return s.err
}

func (s *ManifestScanner) validateStreamName(name string) error {
	unescaped := manifestUnescape(name)
	if unescaped == "" || unescaped[0] != '.' {
		return errors.New("stream name must start with '.'")
	}
	if unescaped == "." {
		return nil
	}
	if strings.HasSuffix(unescaped, "/") {
		return errors.New("stream name must not end with '/'")
	}

	parts := strings.Split(unescaped, "/")
	if parts[0] != "." {
		return errors.New("first path component must be '.'")
	}

	for i := 1; i < len(parts); i++ {
		if parts[i] == "" {
			return errors.New("empty path component")
		}
		if parts[i] == "." || parts[i] == ".." {
			return errors.New("invalid path component '.' or '..'")
		}
	}
	return nil
}

func (s *ManifestScanner) parseLocator(token string) (BlockLocator, error) {
	if !locatorRegex.MatchString(token) {
		return BlockLocator{}, errors.New("invalid format")
	}

	parts := strings.Split(token, "+")
	if len(parts) < 2 {
		return BlockLocator{}, errors.New("missing size hint")
	}

	size, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil || size < 0 {
		return BlockLocator{}, errors.New("invalid size hint")
	}

	loc := BlockLocator{
		Digest: parts[0],
		Size:   int(size),
		Text:   token,
	}

	if len(parts) > 2 {
		loc.Hints = parts[2:]
	}

	return loc, nil
}

func (s *ManifestScanner) parseFileToken(token string) (FileToken, error) {
	parts := strings.SplitN(token, ":", 3)
	if len(parts) != 3 {
		return FileToken{}, errors.New("must have 3 parts separated by ':'")
	}

	pos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || pos < 0 {
		return FileToken{}, errors.New("invalid position")
	}

	length, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || length < 0 {
		return FileToken{}, errors.New("invalid size")
	}

	unescapedName := manifestUnescape(parts[2])

	if strings.HasPrefix(unescapedName, "/") || strings.HasSuffix(unescapedName, "/") {
		return FileToken{}, errors.New("filename must not start or end with '/'")
	}
	if strings.Contains(unescapedName, "//") {
		return FileToken{}, errors.New("filename must not contain '//'")
	}

	nameParts := strings.Split(unescapedName, "/")
	for _, p := range nameParts {
		if p == "." || p == ".." {
			return FileToken{}, errors.New("filename component cannot be '.' or '..'")
		}
	}

	return FileToken{
		Position: pos,
		Length:   length,
		Name:     unescapedName,
	}, nil
}
