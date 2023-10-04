package patcheskibabcom

import (
	"fmt"
	"io"
	"net/http"
)

const (
	patchURL = "https://patches.kibab.com/patches/dn.php5?id=%d"
)

// PatchByID fetches a patch from patches.kibab.com by ID.
func PatchByID(id int) (string, error) {
	url := fmt.Sprintf(patchURL, id)
	// Send an HTTP GET request to the URL
	response, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error fetching the patch #%d: %w", id, err)
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("error reading the patch #%d body: %w", id, err)
	}

	return string(body), nil
}
