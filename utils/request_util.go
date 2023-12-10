package utils

import (
	"io"
	"net/http"
)

func ReadContent(response *http.Response) []byte {
	defer response.Body.Close()

	all, err := io.ReadAll(response.Body)
	if err != nil {
		Error(err)
		return nil
	}
	return all
}
