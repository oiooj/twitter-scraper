package twitterscraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const bearerToken string = "AAAAAAAAAAAAAAAAAAAAAFR3JwEAAAAAEquyLtEZhuvH7jVLCzrba2Wearo%3DhIcSZNPXNWb3ZOrf2CcexWQbQgNPzhp7MInFZNsjihbzY9KRFe"

// RequestAPI get JSON from frontend API and decodes it
func (s *Scraper) RequestAPI(req *http.Request, target interface{}) error {
	if s.guestToken == "" || s.guestCreatedAt.Before(time.Now().Add(-time.Hour*3)) {
		err := s.GetGuestToken()
		if err != nil {
			return err
		}
	}

	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("X-Guest-Token", s.guestToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// private profiles return forbidden, but also data
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusForbidden {
		content, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("response status %s: %s", resp.Status, content)
	}

	if resp.Header.Get("X-Rate-Limit-Remaining") == "0" {
		s.guestToken = ""
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// GetGuestToken from Twitter API
func (s *Scraper) GetGuestToken() error {
	req, err := http.NewRequest("POST", "https://api.twitter.com/1.1/guest/activate.json", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		content, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("response status %s: %s", resp.Status, content)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var jsn map[string]interface{}
	if err := json.Unmarshal(body, &jsn); err != nil {
		return err
	}
	var ok bool
	if s.guestToken, ok = jsn["guest_token"].(string); !ok {
		return fmt.Errorf("guest_token not found")
	}
	s.guestCreatedAt = time.Now()

	return nil
}
