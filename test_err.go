package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

func main() {
	resp, _ := http.Get("https://pay-sandbox.sepay.vn/v1/checkout?signature=s9R%2Fp4INHpC040imdSd8apLMTcrTknur%2F3OvOC3iEHM%3D&merchant=SP-TEST-TX45579A&operation=PURCHASE&order_invoice_number=TEST_LINK_001&order_amount=150000&currency=VND&order_description=Test+thanh+toan+link")
	b, _ := ioutil.ReadAll(resp.Body)
	html := string(b)
	
	re := regexp.MustCompile(`(?s)displayErrorCard\((.*?)\)`)
	matches := re.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		fmt.Println("Found displayErrorCard call:", match[1])
	}
	
	// Also print any script tag content that contains 'displayErrorCard' (but not its definition)
	// Or just look for standard error messages
	if len(matches) <= 1 {
	    // Maybe the error is in the HTML directly?
	    re2 := regexp.MustCompile(`(?s)<div class="message">(.*?)</div>`)
	    matches2 := re2.FindAllStringSubmatch(html, -1)
	    for _, match := range matches2 {
		    fmt.Println("Message div:", match[1])
	    }
	}
}
