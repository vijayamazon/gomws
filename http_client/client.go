package mwsHttpClient

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// http client wrapper to handle request to mws
type MwsHttpClient struct {
	// The host of the end point
	Host       string
	// The path to the operation
	Path       string
	// The query parameters send to the server
	Parameters NormalizedParameters
	// Whether or not the parameters are signed
	signed     bool
}

// calculateStringToSignV2 Calculate the signature to sign the query for signature version 2
func (client *MwsHttpClient) calculateStringToSignV2() string {
	var stringToSign bytes.Buffer

	client.Parameters.Set("Timestamp", now())

	stringToSign.WriteString("POST\n")
	stringToSign.WriteString(client.Host)
	stringToSign.WriteString("\n")
	stringToSign.WriteString(client.Path)
	stringToSign.WriteString("\n")
	stringToSign.WriteString(client.Parameters.Encode())

	return stringToSign.String()
}

// signature generate the signature by the parameters and the secretKey using HmacSHA256
func (client *MwsHttpClient) signature(params NormalizedParameters, secretKey string) string {
	stringToSign := client.calculateStringToSignV2()
	signature2 := SignV2(stringToSign, secretKey)
	return signature2
}

// SignQuery generate the signature and add the signature to the http parameters
func (client *MwsHttpClient) SignQuery(secretKey string) {
	signature := client.signature(client.Parameters, secretKey)
	client.Parameters.Set("Signature", signature)
	client.signed = true
}

// AugmentParameters add new parameters to http's query and indicate the query is not signed
func (client *MwsHttpClient) AugmentParameters(params map[string]string) {
	for k, v := range params {
		client.Parameters.Set(k, v)
	}

	client.signed = false
}

func (client *MwsHttpClient) EndPoint() string, error {
	return "https://" + client.Host + client.Path
}

// Request send the http request to mws server
// If the query is indicated un signed, an error will return
func (client *MwsHttpClient) Request() string, error {
	if !client.signed {
		return "", fmt.Errorf("Query is not signed")
	}

	encodedParams := client.Parameters.Encode()
	req, err := http.NewRequest(
		"POST",
		client.EndPoint(),
		bytes.NewBufferString(encodedParams),
	)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(encodedParams)))

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil

	// roleBuffer := new(bytes.Buffer)
	// roleBuffer.ReadFrom(roleResponse.Body)

	// credentials := Credentials{}

	// err = json.Unmarshal(roleBuffer.Bytes(), &credentials)

	// return json.NewDecoder(r.Body).Decode(target)
}

const (
	Iso8061Format = time.RFC3339 // "2006-01-02T15:04:05Z07:00"
)

// Current timestamp in iso8061 format
func now() string {
	return time.Now().UTC().Format(Iso8061Format)
}
