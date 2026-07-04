package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func main() {
	baseURL := "https://pay-sandbox.sepay.vn/v1/checkout/init"
	params := url.Values{}
	params.Add("currency", "VND")
	params.Add("merchant", "SP-TEST-TX45579A")
	params.Add("operation", "PURCHASE")
	params.Add("order_amount", "100000")
	params.Add("order_description", "Thanh toan don hang INV_001")
	params.Add("order_invoice_number", "INV_001")
	params.Add("signature", "2G8i7E91E4XdJ0jQihVsMO2b5UDuYlyYPhOhfB88luM=")

	fullURL := baseURL + "?" + params.Encode()
	fmt.Println("GET URL:", fullURL)

	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	
	fmt.Println("Status:", resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Response (first 100 chars):", string(body[:100]))
}
