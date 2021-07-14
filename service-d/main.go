package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Permission struct {
	Claims       map[string]string `json:"claims,omitempty"`
	ResourceID   string            `json:"rsid,omitempty"`
	ResourceName string            `json:"rsname,omitempty"`
	Scopes       []string          `json:"scopes,omitempty"`
}

func main() {
	var (
		host string
		port string
	)

	flag.StringVar(&host, "h", "0.0.0.0", "Listening host")
	flag.StringVar(&port, "p", "8000", "Listening port")
	flag.Parse()

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(host+":"+port, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	accessToken, err := getBearerToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	keyCloakTokenEndpoint := "http://keycloak:8080/auth/realms/kafka-authz/protocol/openid-connect/token"
	permissionsArray, err := getPermissions(accessToken, "service-d", keyCloakTokenEndpoint)
	permissionsMap, err := getPermissionsMap(*permissionsArray)
	fmt.Println(permissionsMap)

	if strings.Contains(permissionsMap[r.URL.Path], r.Method) {
		fmt.Println("TRUE")
	} else {
		fmt.Println("FALSE")
	}

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Write([]byte("This is an listener server.\n"))
}

func getBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization header didn't provide")
	}
	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return "", fmt.Errorf("Authorization header provided but it looks like non bearer type")
	}
	return splitToken[1], nil
}

func getPermissions(assessToken string, audience string, keyCloakTokenEndpoint string) (*[]Permission, error) {
	payload := strings.NewReader("audience=" + audience + "&grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Auma-ticket&response_mode=permissions")

	client := &http.Client{}
	req, err := http.NewRequest("POST", keyCloakTokenEndpoint, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+assessToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println(string(body))
	var permissions []Permission

	json.Unmarshal(body, &permissions)

	//TODO: invalid token or not permissions
	if res.StatusCode >= 300 {
		return &permissions, fmt.Errorf("Token invalid")
	}
	return &permissions, nil
}

func getPermissionsMap(permissions []Permission) (map[string]string, error) {
	result := make(map[string]string)
	for _, permission := range permissions {
		result[permission.ResourceName] = strings.Join(permission.Scopes, ",")
	}
	return result, nil
}
