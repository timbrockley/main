/*

Copyright 2023-2024, Tim Brockley. All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.

*/

package conv

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/mtraver/base91"
)

//------------------------------------------------------------
//############################################################
//------------------------------------------------------------

//------------------------------------------------------------
// Base64_encode
//------------------------------------------------------------

func Base64_encode(dataString string) string {
	//------------------------------------------------------------
	if dataString == "" {
		return dataString
	}
	//------------------------------------------------------------
	return base64.StdEncoding.EncodeToString([]byte(dataString))
	//------------------------------------------------------------
}

//------------------------------------------------------------
// Base64_decode
//------------------------------------------------------------

func Base64_decode(dataString string) (string, error) {
	//------------------------------------------------------------
	var err error
	var dataBytes []byte
	//------------------------------------------------------------
	if dataString == "" {
		return dataString, err
	}
	//------------------------------------------------------------
	dataBytes, err = base64.StdEncoding.DecodeString(dataString)
	if err != nil {
		dataBytes = []byte{}
	}
	//------------------------------------------------------------
	return string(dataBytes), err
	//------------------------------------------------------------
}

//------------------------------------------------------------
// Base64url_encode
//------------------------------------------------------------

func Base64url_encode(dataString string) string {
	//------------------------------------------------------------
	if dataString == "" {
		return dataString
	}
	//------------------------------------------------------------
	dataString = Base64_encode(dataString)
	//------------------------------------------------------------
	replacer := strings.NewReplacer(
		"+", "-",
		"/", "_",
		"=", "",
	)
	//------------------------------------------------------------
	return replacer.Replace(dataString)
	//------------------------------------------------------------
}

//------------------------------------------------------------
// Base64url_decode
//------------------------------------------------------------

func Base64url_decode(dataString string) (string, error) {
	//------------------------------------------------------------
	var err error
	//------------------------------------------------------------
	if dataString == "" {
		return dataString, err
	}
	//------------------------------------------------------------
	replacer := strings.NewReplacer(
		"-", "+",
		"_", "/",
	)
	//------------------------------------------------------------
	dataString = replacer.Replace(dataString)
	//------------------------------------------------------------
	switch len(dataString) % 4 { // Pad with trailing '='s
	case 2:
		dataString += "==" // 2 pad chars
	case 3:
		dataString += "=" // 1 pad char
	}
	//------------------------------------------------------------
	return Base64_decode(dataString)
	//------------------------------------------------------------
}

//------------------------------------------------------------
// Base64_Base64url
//------------------------------------------------------------

func Base64_Base64url(dataString string) string {
	//------------------------------------------------------------
	if dataString == "" {
		return dataString
	}
	//------------------------------------------------------------
	replacer := strings.NewReplacer(
		"+", "-",
		"/", "_",
		"=", "",
	)
	//------------------------------------------------------------
	return replacer.Replace(dataString)
	//------------------------------------------------------------
}

//------------------------------------------------------------
// Base64url_Base64
//------------------------------------------------------------

func Base64url_Base64(dataString string) string {
	//------------------------------------------------------------
	if dataString == "" {
		return dataString
	}
	//------------------------------------------------------------
	replacer := strings.NewReplacer(
		"-", "+",
		"_", "/",
	)
	//------------------------------------------------------------
	dataString = replacer.Replace(dataString)
	//------------------------------------------------------------
	switch len(dataString) % 4 { // Pad with trailing '='s
	case 2:
		dataString += "==" // 2 pad chars
	case 3:
		dataString += "=" // 1 pad char
	}
	//------------------------------------------------------------
	return dataString
	//------------------------------------------------------------
}

//------------------------------------------------------------
//############################################################
//------------------------------------------------------------

//------------------------------------------------------------
// Base91_encode
//------------------------------------------------------------

func Base91_encode(dataString string, escapeBool bool) string {
	//------------------------------------------------------------
	if dataString == "" {
		return dataString
	}
	//------------------------------------------------------------
	dataString = base91.StdEncoding.EncodeToString([]byte(dataString))
	//------------------------------------------------------------
	if escapeBool {
		//------------------------------------------------------------
		replacer := strings.NewReplacer(
			"\x22", "-q",
			"\x24", "-d",
			"\x60", "-g",
		)
		//------------------------------------------------------------
		dataString = replacer.Replace(dataString)
		//------------------------------------------------------------
	}
	//------------------------------------------------------------
	return dataString
	//------------------------------------------------------------
}

//------------------------------------------------------------
// Base91_decode
//------------------------------------------------------------

func Base91_decode(dataString string, unescapeBool bool) (string, error) {
	//------------------------------------------------------------
	var err error
	var dataBytes []byte
	//------------------------------------------------------------
	if dataString == "" {
		return dataString, err
	}
	//------------------------------------------------------------
	if unescapeBool {
		//------------------------------------------------------------
		replacer := strings.NewReplacer(
			"-g", "\x60",
			"-d", "\x24",
			"-q", "\x22",
		)
		//------------------------------------------------------------
		dataString = replacer.Replace(dataString)
		//------------------------------------------------------------
	}
	//------------------------------------------------------------
	dataBytes, err = base91.StdEncoding.DecodeString(dataString)
	if err != nil {
		dataBytes = []byte{}
	}
	//------------------------------------------------------------
	return string(dataBytes), err
	//------------------------------------------------------------
}

//------------------------------------------------------------
//############################################################
//------------------------------------------------------------

//------------------------------------------------------------
// JSON_Marshal => json encodes input into "bytes" without escaping html characters
//------------------------------------------------------------

func JSON_Marshal(input interface{}) ([]byte, error) {
	//------------------------------------------------------------
	var err error
	var encodeBuffer bytes.Buffer
	//------------------------------------------------------------
	encoder := json.NewEncoder(&encodeBuffer)
	encoder.SetEscapeHTML(false)
	//------------------------------------------------------------
	err = encoder.Encode(input)
	if err != nil {
		return nil, err
	}
	//------------------------------------------------------------
	return bytes.TrimRight(encodeBuffer.Bytes(), "\n"), err
	//------------------------------------------------------------
}

//------------------------------------------------------------
// JSON_MarshalIndent => json encodes input into "bytes" without escaping html characters
//------------------------------------------------------------

func JSON_MarshalIndent(input interface{}, prefix string, indent string) ([]byte, error) {
	//------------------------------------------------------------
	var err error
	var encodeBuffer bytes.Buffer
	var indentBuffer bytes.Buffer
	//------------------------------------------------------------
	encoder := json.NewEncoder(&encodeBuffer)
	encoder.SetEscapeHTML(false)
	//------------------------------------------------------------
	err = encoder.Encode(input)
	if err != nil {
		return nil, err
	}
	//------------------------------------------------------------
	err = json.Indent(&indentBuffer, bytes.TrimRight(encodeBuffer.Bytes(), "\n"), prefix, indent)
	if err != nil {
		return nil, err
	}
	//------------------------------------------------------------
	return indentBuffer.Bytes(), err
	//------------------------------------------------------------
}

//------------------------------------------------------------
// JSON_encode
//------------------------------------------------------------

func JSON_encode(jsonInterface interface{}) (string, error) {
	//------------------------------------------------------------
	var err error
	var jsonBytes []byte
	//------------------------------------------------------------
	jsonBytes, err = JSON_Marshal(jsonInterface)
	if err != nil {
		jsonBytes = []byte{}
	}
	//------------------------------------------------------------
	return string(jsonBytes), err
	//------------------------------------------------------------
}

//------------------------------------------------------------
// JSON_decode
//------------------------------------------------------------

func JSON_decode(jsonString string) (interface{}, error) {
	//------------------------------------------------------------
	var err error
	var jsonInterface interface{}
	//------------------------------------------------------------
	err = json.Unmarshal([]byte(jsonString), &jsonInterface)
	if err != nil {
		jsonInterface = nil
	}
	//------------------------------------------------------------
	return jsonInterface, err
	//------------------------------------------------------------

}

//------------------------------------------------------------
//############################################################
//------------------------------------------------------------
