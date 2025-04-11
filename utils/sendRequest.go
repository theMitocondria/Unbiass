package utils

import(
	"net/http"
	"io"
    "fmt"
	"github.com/theMitocondria/Unbiass/inits" 
)

func SendRequest( Type string , Link string ,Body io.Reader) (*http.Response, error) {
    req, err := http.NewRequest(Type, Link , Body)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := inits.HTTPClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to send request: %v", err)
    }

    return resp, nil
}