package arvados

import (
	"strings"
	"testing"
)

func TestManifestScannerValid(t *testing.T) {
	manifest := ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:a 0:0:b 0:33:output.txt\n" +
		"./c d41d8cd98f00b204e9800998ecf8427e+0 0:0:d\n"

	scanner := NewManifestScanner(manifest)

	if !scanner.Scan() {
		t.Fatalf("expected first stream, got error: %v", scanner.Err())
	}
	s := scanner.Stream()
	if s.StreamName != "." {
		t.Errorf("expected stream name ., got %q", s.StreamName)
	}
	if len(s.Locators) != 1 || s.Locators[0].Digest != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Errorf("bad locators: %v", s.Locators)
	}
	if len(s.FileTokens) != 3 {
		t.Errorf("bad file tokens: %v", s.FileTokens)
	} else if s.FileTokens[2].Name != "output.txt" {
		t.Errorf("expected output.txt, got %q", s.FileTokens[2].Name)
	}

	if !scanner.Scan() {
		t.Fatalf("expected second stream, got error: %v", scanner.Err())
	}
	s = scanner.Stream()
	if s.StreamName != "./c" {
		t.Errorf("expected stream name ./c, got %q", s.StreamName)
	}
	if len(s.FileTokens) != 1 || s.FileTokens[0].Name != "d" {
		t.Errorf("bad file tokens: %v", s.FileTokens)
	}

	if scanner.Scan() {
		t.Errorf("expected end of manifest, got stream %v", scanner.Stream())
	}
	if scanner.Err() != nil {
		t.Errorf("expected no error at EOF, got: %v", scanner.Err())
	}
}

func TestManifestScannerInvalid(t *testing.T) {
	testCases := []struct {
		name     string
		manifest string
		errSub   string
	}{
		{
			name:     "no trailing newline",
			manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:a",
			errSub:   "line 1: no trailing newline",
		},
		{
			name:     "no locators",
			manifest: ". 0:0:a\n",
			errSub:   "line 1: bad file segment \"0:0:a\": appears before locators",
		},
		{
			name:     "no file tokens",
			manifest: ". d41d8cd98f00b204e9800998ecf8427e+0\n",
			errSub:   "line 1: no file segments",
		},
		{
			name:     "file token before locator",
			manifest: ". 0:0:a d41d8cd98f00b204e9800998ecf8427e+0\n",
			errSub:   "line 1: bad file segment \"0:0:a\": appears before locators",
		},
		{
			name:     "locator after file token",
			manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:a d41d8cd98f00b204e9800998ecf8427e+0\n",
			errSub:   "line 1: bad locator \"d41d8cd98f00b204e9800998ecf8427e+0\": appears after file segments",
		},
		{
			name:     "invalid stream name - not starting with dot",
			manifest: "a d41d8cd98f00b204e9800998ecf8427e+0 0:0:a\n",
			errSub:   "line 1: stream name must start with '.'",
		},
		{
			name:     "invalid stream name - ends with slash",
			manifest: "./a/ d41d8cd98f00b204e9800998ecf8427e+0 0:0:a\n",
			errSub:   "line 1: stream name must not end with '/'",
		},
		{
			name:     "invalid stream name - contains dot component",
			manifest: "././a d41d8cd98f00b204e9800998ecf8427e+0 0:0:a\n",
			errSub:   "line 1: invalid path component '.' or '..'",
		},
		{
			name:     "invalid file name - contains dot component",
			manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:./a\n",
			errSub:   "line 1: bad file segment \"0:0:./a\": filename component cannot be '.' or '..'",
		},
		{
			name:     "invalid file name - starts with slash",
			manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:/a\n",
			errSub:   "line 1: bad file segment \"0:0:/a\": filename must not start or end with '/'",
		},
		{
			name:     "invalid file name - ends with slash",
			manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:a/\n",
			errSub:   "line 1: bad file segment \"0:0:a/\": filename must not start or end with '/'",
		},
		{
			name:     "invalid locator",
			manifest: ". d41d8cd98f00b204e9800998ecf8427e 0:0:a\n",
			errSub:   "line 1: bad locator \"d41d8cd98f00b204e9800998ecf8427e\": invalid format",
		},
		{
			name:     "invalid unprintable character",
			manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:a\t",
			errSub:   "line 1: invalid character",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scanner := NewManifestScanner(tc.manifest)
			for scanner.Scan() {
			}
			err := scanner.Err()
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.errSub)
			}
			if !strings.Contains(err.Error(), tc.errSub) {
				t.Fatalf("expected error containing %q, got %q", tc.errSub, err.Error())
			}
		})
	}
}
