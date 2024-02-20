package nodebuilder

import (
	"bytes"
	"github.com/BurntSushi/toml"
	"io"
)

// TODO(@Wondertan): We should have a description for each field written into w,
// 	so users can instantly understand purpose of each field. Ideally, we should have a utility
// program to parse comments 	from actual sources(*.go files) and generate docs from comments.

// Hint: use 'ast' package.
// Encode encodes a given Config into w.
func (cfg *Config) Encode(w io.Writer) error {
	return toml.NewEncoder(w).Encode(cfg)
}

// Decode decodes a Config from a given reader r.
func (cfg *Config) Decode(r io.Reader) error {
	// Read the content of the io.Reader into a byte slice
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// Print the body
	//fmt.Println("Got configuration from IndexedDB", string(body))

	// Create a new reader from the byte slice
	newReader := bytes.NewReader(body)

	// Use the new reader for the decoder
	_, err = toml.NewDecoder(newReader).Decode(cfg)
	return err
}
