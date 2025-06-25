package tools

import (
	"fmt"
	"io"
	"net/http"

	"github.com/firebase/genkit/go/ai"
)

type CurlToolInput struct {
	URL string
}

// CurlTool returns a tool function that dowloads content of a given file.
func CurlTool() func(ctx *ai.ToolContext, input CurlToolInput) ([]byte, error) {
	return func(ctx *ai.ToolContext, input CurlToolInput) ([]byte, error) {
		fmt.Printf("... Fetching log file %s\n", input.URL)
		resp, err := http.Get(input.URL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, nil
		}

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return content, nil

	}
}
