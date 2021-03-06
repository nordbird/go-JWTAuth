package controllers

import (
	"encoding/json"
	"io"
	"medods-auth/models"
	"net/http"
)

func Respond(w http.ResponseWriter, data models.WebResponse) {
	w.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func Hello(w http.ResponseWriter, r *http.Request) {
	var request models.WebRequest
	var response models.WebResponse

	response.Message = "Hello!"

	routes := map[string]string{}

	routes["route /signin"] = "creating tokens"
	routes["route /refresh"] = "refresh tokens"
	routes["route /delete-one-refresh"] = "deleting one refresh token"
	routes["route /delete-all-refresh"] = "deleting all refresh token"

	request.GUID = "some_GUID"
	request.AccessToken = "this_is_access_token"
	request.RefreshToken = "this_is_refresh_token_in_base64"
	response.Payload = map[string]interface{}{"routes": routes, "data format": request}

	jsonStr, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		Respond(w, models.WebResponse{Message: err.Error()})
		return
	}

	_, err = io.WriteString(w, string(jsonStr))
	if err != nil {
		Respond(w, models.WebResponse{Message: err.Error()})
		return
	}
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	var request models.WebRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		Respond(w, models.WebResponse{Message: err.Error()})
		return
	}

	request.AccessToken, request.RefreshToken, err = AddTokens(request.GUID)
	if err != nil {
		Respond(w, models.WebResponse{Message: err.Error()})
		return
	}

	var response models.WebResponse
	response.Payload = map[string]interface{}{"tokens": request}

	Respond(w, response)
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	var request models.WebRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		Respond(w, models.WebResponse{Message: err.Error()})
		return
	}

	refreshTokenValid, message := models.IsValidToken(request.GUID, request.RefreshToken)
	if !refreshTokenValid {
		Respond(w, models.WebResponse{Message: message})
		return
	}

	tokenPair, err := models.CreateDBTokenPair(request.AccessToken, request.RefreshToken)
	if err != nil {
		Respond(w, models.WebResponse{Message: err.Error()})
		return
	}

	err = DeleteDBTokenPair(request.GUID, tokenPair)
	if err != nil {
		Respond(w, models.WebResponse{Message: err.Error()})
		return
	}

	request.AccessToken, request.RefreshToken, err = AddTokens(request.GUID)
	if err != nil {
		Respond(w, models.WebResponse{Message: err.Error()})
		return
	}

	var response models.WebResponse
	response.Payload = map[string]interface{}{"tokens": request}

	Respond(w, response)
}

func DeleteOneRefreshToken(w http.ResponseWriter, r *http.Request) {
	var request models.WebRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		Respond(w, models.WebResponse{Message: err.Error()})
		return
	}

	accessTokenValid, message := models.IsValidToken(request.GUID, request.AccessToken)
	if !accessTokenValid {
		Respond(w, models.WebResponse{Message: message})
		return
	}

	tokenPair, err := models.CreateDBTokenPair(request.AccessToken, request.RefreshToken)
	if err != nil {
		Respond(w, models.WebResponse{Message: err.Error()})
		return
	}

	err = DeleteDBTokenPair(request.GUID, tokenPair)
	if err != nil {
		Respond(w, models.WebResponse{Message: err.Error()})
		return
	}

	Respond(w, models.WebResponse{Message: "Ok"})
}

func DeleteAllRefreshToken(w http.ResponseWriter, r *http.Request) {
	var request models.WebRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		Respond(w, models.WebResponse{Message: err.Error()})
		return
	}

	accessTokenValid, message := models.IsValidToken(request.GUID, request.AccessToken)
	if !accessTokenValid {
		Respond(w, models.WebResponse{Message: message})
		return
	}

	err = DeleteAllDBTokenPair(request.GUID)
	if err != nil {
		Respond(w, models.WebResponse{Message: err.Error()})
		return
	}

	Respond(w, models.WebResponse{Message: "Ok"})
}

func AddTokens(guid string) (string, string, error) {
	accessTokenString, refreshTokenString, err := models.CreateWebTokenPair(guid)
	if err != nil {
		return "", "", err
	}

	tokenPair, err := models.CreateDBTokenPair(accessTokenString, refreshTokenString)
	if err != nil {
		return "", "", err
	}

	err = AddDBTokenPair(guid, tokenPair)
	if err != nil {
		return "", "", err
	}
	return accessTokenString, refreshTokenString, err
}
