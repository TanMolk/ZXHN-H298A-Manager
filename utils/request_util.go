package utils

import (
	"io"
	"net/http"
)

func ReadContent(response *http.Response) []byte {
	if response != nil && response.Body != nil {

		defer response.Body.Close()

		all, err := io.ReadAll(response.Body)
		if err != nil {
			Error(err)
			return nil
		}
		return all

	}
	return nil
}
