package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const BASE_URL = "https://api.jquants.com/v1"

type HolidayDivision int

/*
非営業日 0
営業日 1
東証半日立会日 2
非営業日(祝日取引あり) 3
*/

const (
	Holiday        HolidayDivision = 0
	BusinessDay    HolidayDivision = 1
	HalfDay        HolidayDivision = 2
	TradingHoliday HolidayDivision = 3
)

type TradingCalendarParam struct {
	HolidayDivision *HolidayDivision
	From            *time.Time
	To              *time.Time
}

type Client interface {
	TradingCalendar(param TradingCalendarParam) ([]DateHolidayDivision, error)
}

type clientImpl struct {
	idToken string
}

func New(mailAddress, password string) (Client, error) {
	clt := &clientImpl{}
	refreshToken, err := clt.getRefreshToken(mailAddress, password)
	if err != nil {
		return nil, nil
	}
	idToken, err := clt.getIDToken(refreshToken)
	if err != nil {
		return nil, nil
	}
	clt.idToken = idToken
	return clt, nil
}

type refreshTokenRequest struct {
	MailAddress string `json:"mailaddress"`
	Password    string `json:"password"`
}

type refreshTokenResponse struct {
	RefreshToken string `json:"refreshToken"`
}

func (c *clientImpl) getRefreshToken(mailAddress, password string) (string, error) {
	path := "/token/auth_user"
	url := BASE_URL + path
	reqBytes, err := json.Marshal(refreshTokenRequest{
		MailAddress: mailAddress,
		Password:    password,
	})
	if err != nil {
		return "", fmt.Errorf("getRefreshToken json.Marshal failed: %w", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBytes))
	if err != nil {
		return "", fmt.Errorf("getRefreshToken http.Post failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("getRefreshToken got non 200 HTTP status code %d: %s", resp.StatusCode, resp.Status)
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("getRefreshToken io.ReadAll failed: %w", err)
	}
	result := refreshTokenResponse{}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", fmt.Errorf("getRefreshToken json.Unmarshal failed: %w", err)
	}
	return result.RefreshToken, nil
}

type idTokenResponse struct {
	IDToken string `json:"idToken"`
}

func (c *clientImpl) getIDToken(refreshToken string) (string, error) {
	path := "/token/auth_refresh"
	endpoint := BASE_URL + path
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("getIDToken url.Parse failed: %w", err)
	}
	q := u.Query()
	q.Set("refreshtoken", refreshToken)
	u.RawQuery = q.Encode()
	resp, err := http.Post(u.String(), "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("getIDToken http.Post failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("getIDToken got non 200 HTTP status code %d: %s", resp.StatusCode, resp.Status)
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("getIDToken io.ReadAll failed: %w", err)
	}
	result := idTokenResponse{}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", fmt.Errorf("getIDToken json.Unmarshal failed: %w", err)
	}
	return result.IDToken, nil
}

type tradingCalendarResponse struct {
	TradingCalendar []struct {
		Date            string `json:"date"`
		HolidayDivision string `json:"holidaydivision"`
	} `json:"trading_calendar"`
}

type DateHolidayDivision struct {
	Date            time.Time
	HolidayDivision HolidayDivision
}

func (c *clientImpl) TradingCalendar(param TradingCalendarParam) ([]DateHolidayDivision, error) {
	path := "/markets/trading_calendar"
	endpoint := BASE_URL + path
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("TradingCalendar url.Parse failed: %w", err)
	}
	q := u.Query()
	if param.HolidayDivision != nil {
		q.Set("holidaydivision", strconv.Itoa(int(*param.HolidayDivision)))
	}
	if param.From != nil && param.To != nil {
		q.Set("from", param.From.Format("2006-01-02"))
		q.Set("to", param.To.Format("2006-01-02"))
	}
	u.RawQuery = q.Encode()
	fmt.Printf("URL: %s\n", u.String())
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("TradingCalendar http.NewRequest failed: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.idToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("TradingCalendar http.Do failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TradingCalendar got non 200 HTTP status code %d: %s", resp.StatusCode, resp.Status)
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("TradingCalendar io.ReadAll failed: %w", err)
	}
	result := tradingCalendarResponse{}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("TradingCalendar json.Unmarshal failed: %w", err)
	}
	dates := []DateHolidayDivision{}
	for _, x := range result.TradingCalendar {
		date, err := time.Parse("2006-01-02", x.Date)
		if err != nil {
			return nil, fmt.Errorf("TradingCalendar time.Parse failed %s: %w", x.Date, err)
		}
		division, err := strconv.Atoi(x.HolidayDivision)
		if err != nil {
			return nil, fmt.Errorf("TradingCalendar strconv.Atoi failed %s: %w", x.HolidayDivision, err)
		}
		dates = append(dates, DateHolidayDivision{
			Date:            date,
			HolidayDivision: HolidayDivision(division),
		})
	}
	return dates, nil
}
