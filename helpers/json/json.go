package json

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func DecodeReader(reader io.Reader, item interface{}) error {
	dec := json.NewDecoder(reader)
	return dec.Decode(item)
}

func DecodeFile(file string, item interface{}) error {
	reader, err := os.Open(file)
	if err != nil {
		return err
	}
	defer reader.Close()
	return DecodeReader(reader, item)
}

func EncodeWriter(writer io.Writer, item interface{}) error {
	enc := json.NewEncoder(writer)
	return enc.Encode(item)
}

func EncodeFile(file string, item interface{}) error {
	writer, err := os.Create(file)
	if err != nil {
		return err
	}
	defer writer.Close()

	return EncodeWriter(writer, item)
}

func Stringify(o interface{}) string {
	res, err := json.MarshalIndent(o, "", "  ")
	if err == nil {
		return fmt.Sprintf("%s", res)
	} else {
		return fmt.Sprintf("%s", err)
	}
}
