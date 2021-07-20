package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type permission struct {
	Claims       map[string]string `json:"claims,omitempty"`
	ResourceID   string            `json:"rsid,omitempty"`
	ResourceName string            `json:"rsname,omitempty"`
	Scopes       []string          `json:"scopes,omitempty"`
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	keyCloakTokenEndpoint := "http://keycloak:8080/auth/realms/meetup/protocol/openid-connect/token"

	// Get JWT from request
	accessToken, err := getBearerToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	// Get permissions from AuthServer (KeyCloak)
	permissions, err := getPermissions(accessToken, "service-d", keyCloakTokenEndpoint)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	// Log incoming HTTP request
	logRequest(r)

	// Check permissions
	if strings.Contains(permissions[r.URL.Path], r.Method) {
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte("ServiceD. Current time is " + (time.Now().Format("2006.01.02 15:04:05"))))
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func getBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization header didn't provide")
	}
	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return "", fmt.Errorf("Authorization header provided but it's not bearer type")
	}
	return splitToken[1], nil
}

func getPermissions(assessToken string, audience string, keyCloakTokenEndpoint string) (map[string]string, error) {
	client := &http.Client{}

	payload := strings.NewReader("audience=" + audience + "&grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Auma-ticket&response_mode=permissions")

	req, err := http.NewRequest("POST", keyCloakTokenEndpoint, payload)

	if err != nil {
		return nil, fmt.Errorf("failed to construct request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+assessToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	permissionsMap := make(map[string]string)

	if res.StatusCode == 401 {
		return nil, fmt.Errorf("invalid token")
	} else if res.StatusCode == 403 {
		return permissionsMap, nil
		// return nil, fmt.Errorf("Does not have any permission")
	}

	var permissions []permission
	json.Unmarshal(body, &permissions)

	for _, permission := range permissions {
		permissionsMap[permission.ResourceName] = strings.Join(permission.Scopes, ",")
	}

	return permissionsMap, nil
}

func logRequest(r *http.Request) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("---New HTTP request received---\n%s", requestDump)
}
