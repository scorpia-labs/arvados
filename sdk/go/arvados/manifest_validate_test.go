// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package arvados

import (
	check "gopkg.in/check.v1"
	"strings"
)

var _ = check.Suite(&ManifestValidateSuite{})

type ManifestValidateSuite struct{}

func (s *ManifestValidateSuite) TestValidateManifest(c *check.C) {
	tests := []struct {
		manifest string
		valid    bool
		errMatch string
	}{
		// from test_keep_manifest.rb
		{manifest: "", valid: true},
		{manifest: " ", valid: false, errMatch: "Invalid manifest: does not end with newline"},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e 0:0:abc.txt\n", valid: false, errMatch: "missing or invalid locator"},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:abc.txt\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e a41d8cd98f00b204e9800998ecf8427e+0 0:0:abc.txt\n", valid: false, errMatch: "missing or invalid locator"},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo/bar.txt\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:.foo.txt\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:.foo\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:...\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:.../.foo./.../bar\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo/...\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo/.../bar\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo/.bar/baz.txt\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo/bar./baz.txt\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 000000000000000000000000000000:0777:foo.txt\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:0:0\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:\\040\n", valid: true},
		{manifest: ". 00000000000000000000000000000000+0 0:0:0\n", valid: true},
		{manifest: ". 00000000000000000000000000000000+0 0:0:d41d8cd98f00b204e9800998ecf8427e+0+Ad41d8cd98f00b204e9800998ecf8427e00000000@ffffffff\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0+Ad41d8cd98f00b204e9800998ecf8427e00000000@ffffffff 0:0:empty.txt\n", valid: true},
		{manifest: "./empty_dir d41d8cd98f00b204e9800998ecf8427e+0 0:0:.\n", valid: true},

		{manifest: ". d41d8cd98f00b204e9800998ecf8427e 0:0:abc.txt", valid: false, errMatch: "Invalid manifest: does not end with newline"},
		{manifest: "abc d41d8cd98f00b204e9800998ecf8427e 0:0:abc.txt\n", valid: false, errMatch: "stream name \"abc\""},
		{manifest: "abc/./foo d41d8cd98f00b204e9800998ecf8427e 0:0:abc.txt\n", valid: false, errMatch: "stream name \"abc/./foo\""},
		{manifest: "./abc/../foo d41d8cd98f00b204e9800998ecf8427e 0:0:abc.txt\n", valid: false, errMatch: "stream name \"./abc/../foo\""},
		{manifest: "./abc/. d41d8cd98f00b204e9800998ecf8427e 0:0:abc.txt\n", valid: false, errMatch: "stream name \"./abc/.\""},
		{manifest: "./abc/.. d41d8cd98f00b204e9800998ecf8427e 0:0:abc.txt\n", valid: false, errMatch: "stream name \"./abc/..\""},
		{manifest: "./abc/./foo d41d8cd98f00b204e9800998ecf8427e 0:0:abc.txt\n", valid: false, errMatch: "stream name \"./abc/./foo\""},
		{manifest: ". 8cf8463b34caa8ac871a52d5dd7ad1ef+1 0:1:.\n", valid: false, errMatch: "invalid file token \"0:1:.\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:..\n", valid: false, errMatch: "invalid file token \"0:0:..\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:./abc.txt\n", valid: false, errMatch: "invalid file token \"0:0:./abc.txt\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:../abc.txt\n", valid: false, errMatch: "invalid file token \"0:0:../abc.txt\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:abc.txt/.\n", valid: false, errMatch: "invalid file token \"0:0:abc.txt/.\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:abc.txt/..\n", valid: false, errMatch: "invalid file token \"0:0:abc.txt/..\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo/./bar\n", valid: false, errMatch: "invalid file token \"0:0:foo/./bar\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo/../bar\n", valid: false, errMatch: "invalid file token \"0:0:foo/../bar\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:abc.txt\n./dir1 a41d8cd98f00b204e9800998ecf8427e+0 abc.txt\n", valid: false, errMatch: "invalid file token \"abc.txt\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:abc.txt\n./dir1 a41d8cd98f00b204e9800998ecf8427e+0 0:abc.txt\n", valid: false, errMatch: "invalid file token \"0:abc.txt\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:abc.txt\n./dir1 a41d8cd98f00b204e9800998ecf8427e+0 0:0:abc.txt xyz.txt\n", valid: false, errMatch: "invalid file token \"xyz.txt\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt d41d8cd98f00b204e9800998ecf8427e+0\n", valid: false, errMatch: "invalid file token \"d41d8cd98f00b204e9800998ecf8427e+0\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:\n", valid: false, errMatch: "invalid file token \"0:0:\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0\n", valid: false, errMatch: "no file tokens"},
		{manifest: ". 0:0:foo.txt d41d8cd98f00b204e9800998ecf8427e+0\n", valid: false, errMatch: "missing or invalid locator \"0:0:foo.txt\""},
		{manifest: ". 0:0:foo.txt\n", valid: false, errMatch: "missing or invalid locator \"0:0:foo.txt\""},
		{manifest: ".\n", valid: false, errMatch: "missing or invalid locator"},
		{manifest: ".", valid: false, errMatch: "does not end with newline"},
		{manifest: ". \n", valid: false, errMatch: "trailing space"},
		{manifest: ".  \n", valid: false, errMatch: "trailing space"},
		{manifest: " . d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\n", valid: false, errMatch: "missing stream name"},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt \n", valid: false, errMatch: "trailing space"},

		// whitespace
		{manifest: "\v. d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\n", valid: false, errMatch: "stream name \"\\v."},
		{manifest: "./foo\vbar d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\n", valid: false, errMatch: "stream name \"./foo\\vbar"},
		{manifest: "\t. d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\n", valid: false, errMatch: "stream name \"\\t"},
		{manifest: ".\td41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\n", valid: false, errMatch: "stream name \".\\t"},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\t\n", valid: false, errMatch: "invalid file token \"0:0:foo.txt\\t\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0\t 0:0:foo.txt\n", valid: false, errMatch: "missing or invalid locator \"d41d8cd98f00b204e9800998ecf8427e+0\\t\""},

		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0  0:0:foo.txt\n", valid: false, errMatch: "invalid file token \"\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\n \n", valid: false, errMatch: "trailing space"},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\n\n", valid: false, errMatch: "missing stream name"},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\n ", valid: false, errMatch: "does not end with newline"},
		{manifest: "\n. d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\n", valid: false, errMatch: "missing stream name"},

		// empty file and stream name components
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:/foo.txt\n", valid: false, errMatch: "invalid file token \"0:0:/foo.txt\""},
		{manifest: "./ d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\n", valid: false, errMatch: "stream name \"./\""},
		{manifest: ".//foo d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\n", valid: false, errMatch: "stream name \".//foo\""},
		{manifest: "./foo/ d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\n", valid: false, errMatch: "stream name \"./foo/\""},
		{manifest: "./foo//bar d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo.txt\n", valid: false, errMatch: "stream name \"./foo//bar\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo//bar.txt\n", valid: false, errMatch: "invalid file token \"0:0:foo//bar.txt\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo/\n", valid: false, errMatch: "invalid file token \"0:0:foo/\""},

		// escaped chars
		{manifest: "./empty_dir d41d8cd98f00b204e9800998ecf8427e+0 0:0:\\056\n", valid: true},
		{manifest: "./empty_dir d41d8cd98f00b204e9800998ecf8427e+0 0:0:\\056\\056\n", valid: false, errMatch: "invalid file token \"0:0:\\\\056\\\\056\""},
		{manifest: "./empty_dir d41d8cd98f00b204e9800998ecf8427e+0 0:0:\\056\\056\\057foo\n", valid: false, errMatch: "invalid file token \"0:0:\\\\056\\\\056\\\\057foo\""},
		{manifest: "./empty_dir d41d8cd98f00b204e9800998ecf8427e+0 0\\0720\\072foo\n", valid: false, errMatch: "invalid file token \"0\\\\0720\\\\072foo\""},
		{manifest: "./empty_dir d41d8cd98f00b204e9800998ecf8427e+0 \\060:\\060:foo\n", valid: false, errMatch: "invalid file token \"\\\\060:\\\\060:foo\""},

		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\\057bar\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:\\072\n", valid: true},
		{manifest: ".\\057Data d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: true},
		{manifest: "\\056\\057Data d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: true},
		{manifest: "./\\134444 d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: true},

		{manifest: "./\\\\444 d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: false, errMatch: "stream name \"./\\\\\\\\444\""},
		{manifest: "./\\011foo d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: true},
		{manifest: "./\\011/.. d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: false, errMatch: "stream name \"./\\\\011/..\""},
		{manifest: ".\\056\\057 d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: false, errMatch: "stream name \".\\\\056\\\\057\""},
		{manifest: ".\\057\\056 d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: false, errMatch: "stream name \".\\\\057\\\\056\""},
		{manifest: ".\\057Data d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\\444\n", valid: false, errMatch: ">8-bit encoded chars not allowed on file token \"0:0:foo\\\\444\""},
		{manifest: "./\\444 d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: false, errMatch: ">8-bit encoded chars not allowed on stream token \"./\\\\444\""},

		{manifest: "./\tfoo d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: false, errMatch: "stream name \"./\\tfoo\""},
		{manifest: "./foo\\ d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: false, errMatch: "stream name \"./foo\\\\\""},
		{manifest: "./foo\r d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: false, errMatch: "stream name"},
		{manifest: "./foo\\444 d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: false, errMatch: ">8-bit encoded chars not allowed on stream token \"./foo\\\\444\""},
		{manifest: "./foo\\888 d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: false, errMatch: "stream name \"./foo\\\\888\""},

		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\\\n", valid: false, errMatch: "invalid file token \"0:0:foo\\\\\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\r\n", valid: false, errMatch: "invalid file token"},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\\444\n", valid: false, errMatch: ">8-bit encoded chars not allowed on file token \"0:0:foo\\\\444\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\\888\n", valid: false, errMatch: "invalid file token \"0:0:foo\\\\888\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\\057/bar\n", valid: false, errMatch: "invalid file token \"0:0:foo\\\\057/bar\""},
		{manifest: ".\\057/Data d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: false, errMatch: "stream name \".\\\\057/Data\""},
		{manifest: "./Data\\040Folder d41d8cd98f00b204e9800998ecf8427e+0 0:0:foo\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:\\057foo/bar\n", valid: false, errMatch: "invalid file token \"0:0:\\\\057foo/bar\""},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 0:0:\\134057foo/bar\n", valid: true},
		{manifest: ". d41d8cd98f00b204e9800998ecf8427e+0 \\040:\\040:foo.txt\n", valid: false, errMatch: "invalid file token \"\\\\040:\\\\040:foo.txt\""},
	}

	for i, tc := range tests {
		err := ValidateManifest(tc.manifest)
		if tc.valid {
			if err != nil {
				c.Check(err, check.IsNil, check.Commentf("Test %d: expected valid, got error: %v, manifest: %q", i, err, tc.manifest))
			}
		} else {
			if err == nil {
				c.Check(err, check.NotNil, check.Commentf("Test %d: expected invalid, got no error, manifest: %q", i, tc.manifest))
			} else if !strings.Contains(err.Error(), tc.errMatch) {
				c.Check(strings.Contains(err.Error(), tc.errMatch), check.Equals, true, check.Commentf("Test %d: error %q does not match %q, manifest: %q", i, err.Error(), tc.errMatch, tc.manifest))
			}
		}
	}
}
