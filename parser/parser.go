package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	ajaxEndpoint = `https://www.rusprofile.ru/ajax.php?action=search`
	pagePrefix   = `https://www.rusprofile.ru`
)

var (
	kppRegex = regexp.MustCompile(`<span.+clip_kpp.+>(\d{9})<\/span>`)
)

type ServiceError struct{
	Code int
	Message string
}

func (e ServiceError) Error() string {
	return fmt.Sprintf("parser: rusprofile.ru error: %s (code=%d)", e.Message, e.Code)
}

type ajaxComp struct{
	INN string `json:"inn"`
	Name string `json:"name"`
	Link string `json:"link"`
	CEOName string `json:"ceo_name"`
}

type ajaxIP struct{
	INN string `json:"inn"`
	Link string `json:"link"`
	Name string `json:"name"`
}

type ajaxResponse struct{
	Success bool `json:"success"`
	Code int `json:"code"`
	Message string `json:"message"`

	Ul []ajaxComp `json:"ul"`
	Ip []ajaxIP `json:"ip"`
}

type CompanyInfo struct{
	KPP string
	Title string
	DirectorName string
}

type HTTPClient interface{
	Do(r *http.Request) (*http.Response, error)
}

var cacheKey = func() float64 {
	return rand.Float64()
}

func Query(ctx context.Context, client HTTPClient, inn string) (*CompanyInfo, error) {
	queryURL := fmt.Sprintf("%s&query=%s&cacheKey=%v", ajaxEndpoint, inn, cacheKey())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("parser: Query: %v", err)
	}

	// Attempt to evade rusprofile.ru bot detection.
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "ru,en-US;q=0.7,en;q=0.3")
	req.Header.Set("Cookie", "screen_for_ad")
	req.Header.Set("DNT", "1")
	req.Header.Set("Referer", "https://www.rusprofile.ru/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; rv:91.0) Gecko/20100101 Firefox/91.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("parser: Query: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("parser: Query: HTTP %v", resp.StatusCode)
	}
	var respBody ajaxResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("parser: Query: HTTP %v (JSON parse err: %v)", resp.StatusCode, err)
	}
	if !respBody.Success {
		return nil, ServiceError{Code: respBody.Code, Message: respBody.Message}
	}

	// rusprofile.ru does prefix search and may use other identifiers
	// so find the company we were looking for.
	var (
		foundComp *ajaxComp
	)
	for _, comp := range respBody.Ul {
		// !~~ and ~~! is formatting, added to some INNs for some reason.
		compINN := strings.TrimPrefix(comp.INN, "!~~")
		compINN = strings.TrimSuffix(compINN, "~~!")
		if inn == compINN {
			foundComp = &comp
			break
		}
	}
	if foundComp == nil {
		for _, comp := range respBody.Ip {
			// !~~ and ~~! is formatting, added to some INNs for some reason.
			compINN := strings.TrimPrefix(comp.INN, "!~~")
			compINN = strings.TrimSuffix(compINN, "~~!")
			if inn == compINN {
				return &CompanyInfo{
					Title: "ИП " +comp.Name,
					DirectorName: comp.Name,
				}, nil
			}
		}
	}
	if foundComp == nil {
		return nil, nil
	}

	res := CompanyInfo{
		Title: foundComp.Name,
		DirectorName: foundComp.CEOName,
	}
	// KPP is not present in ajax.php so we have to parse it from the page.
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, pagePrefix+foundComp.Link, nil)
	if err != nil {
		return &res, fmt.Errorf("parser: Query: KPP get: %v", err)
	}

	// Attempt to evade rusprofile.ru bot detection.
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru,en-US;q=0.7,en;q=0.3")
	req.Header.Set("Cookie", "screen_for_ad=desktop")
	req.Header.Set("DNT", "1")
	req.Header.Set("Referer", "https://www.rusprofile.ru/")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; rv:91.0) Gecko/20100101 Firefox/91.0")
	time.Sleep(300*time.Millisecond)

	resp, err = client.Do(req)
	if err != nil {
		return &res, fmt.Errorf("parser: Query: KPP get: %v", err)
	}
	if resp.StatusCode / 100 != 2 {
		return nil, fmt.Errorf("parser: Query KPP: HTTP %v", resp.StatusCode)
	}
	defer resp.Body.Close()
	bodyBlob, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &res, fmt.Errorf("parser: Query: KPP get: %v", err)
	}
	kppMatch := kppRegex.FindSubmatch(bodyBlob)
	if kppMatch != nil {
		res.KPP = string(kppMatch[1])
	}
	return &res, nil
}
