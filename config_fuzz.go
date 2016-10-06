// +build gofuzz

package incognitomail

import (
  "bytes"
)

// FuzzConfig is the entry point for fuzzing stuff related to config parsing.
func FuzzConfig(data []byte) int {
  reader := bytes.NewReader(data)

  err := ReadConfigFromReader(reader)
  if err != nil {
    return 0
  }

  return 1
}