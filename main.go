package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	baseNestHost       = "video.nest.com"
	loginEndpoint      = "/api/dropcam/share.login"
	loginContentType   = "application/x-www-form-urlencoded"
	getCamerasEndpoint = "/api/dropcam/cameras.get_by_public_token"
	getImageEndpoint   = "/get_image"
	sessionCookieName  = "website_2"
)

type loginResponse struct {
	Status            int                 `json:"status"`
	Items             []loginResponseItem `json:"items"`
	StatusDescription string              `json:"status_description"`
	StatusDetail      string              `json:"status_detail"`
}

type loginResponseItem struct {
	SessionToken string `json:"session_token"`
}

type getCamerasResponse struct {
	Status            int                     `json:"status"`
	Items             []getCamerasResponeItem `json:"items"`
	StatusDescription string                  `json:"status_description"`
	StatusDetail      string                  `json:"status_detail"`
}

type getCamerasResponeItem struct {
	Name                   string `json:"name"`
	UUID                   string `json:"uuid"`
	NexusAPINestDomainHost string `json:"nexus_api_nest_domain_host"`
}

var (
	httpCli *http.Client = &http.Client{Timeout: time.Second * 30}
	debug   bool         = false
	quiet   bool         = false
)

func main() {
	var cameraToken string
	flag.StringVar(&cameraToken, "token", "", "the camera's token")

	var password string
	flag.StringVar(&password, "password", "", "the camera's password, if link is password-protected")

	var outFile string
	flag.StringVar(&outFile, "out", "out.png", "the output file, must end in .png or .jpeg")

	var imageWidth int
	flag.IntVar(&imageWidth, "width", 512, "the image width in pixels")

	// uses global
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.BoolVar(&quiet, "quiet", false, "disable all logging")

	flag.Parse()

	if cameraToken == "" {
		logError("missing required flag -token", "", true)
	}

	if !strings.HasSuffix(outFile, ".png") && !strings.HasSuffix(outFile, ".jpeg") {
		logError("-out must end in .png or .jpeg", "", true)
	}

	loginInfo, err := login(password, cameraToken)
	if err != nil {
		logError("error creating login session", err.Error(), true)
	}
	sessionToken := loginInfo.Items[0].SessionToken

	cameras, err := getCameras(sessionToken, cameraToken)
	if err != nil {
		logError("error getting camera(s)", err.Error(), true)
	}

	// Use the first camera returned
	// [TODO] Figure out when multiple cameras can be returned by single live video
	camera := cameras.Items[0]
	logInfo(fmt.Sprintf("using camera name=%s, uuid=%s", camera.Name, camera.UUID))

	img, err := getImage(sessionToken, camera.NexusAPINestDomainHost, camera.UUID, imageWidth)
	if err != nil {
		logError("error getting image", err.Error(), true)
	}

	err = saveImage(img, outFile)
	if err != nil {
		logError("error saving image", err.Error(), true)
	}

	logInfo("done")
}

func login(password, cameraToken string) (*loginResponse, error) {
	logInfo("creating new login session")

	formData := url.Values{
		"token": {cameraToken},
	}

	if password != "" {
		formData["password"] = []string{password}
	}

	url := fmt.Sprintf("https://%s%s", baseNestHost, loginEndpoint)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	// Set some minimally required headers
	req.Header.Set("Content-Type", loginContentType)
	req.Header.Set("Referer", fmt.Sprintf("https://%s/live/%s", baseNestHost, cameraToken))

	logDebug(fmt.Sprintf("making POST request to %s", url))
	resp, err := httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status returned by server: %s", resp.Status)
	}

	respRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	logDebug(fmt.Sprintf("got JSON login response from server: %s", respRaw))

	var loginInfo loginResponse
	err = json.Unmarshal(respRaw, &loginInfo)
	if err != nil {
		return nil, err
	}

	if loginInfo.StatusDescription != "ok" || len(loginInfo.Items) != 1 {
		return nil, fmt.Errorf("got unexpected json response from server: %+v", loginInfo)
	}

	logInfo("successfully created login session")
	return &loginInfo, nil
}

func getCameras(sessionToken, cameraToken string) (*getCamerasResponse, error) {
	logInfo("getting camera(s)")

	url := fmt.Sprintf("https://%s%s?token=%s&_=%d", baseNestHost, getCamerasEndpoint, cameraToken, time.Now().Unix())
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionToken})

	logDebug(fmt.Sprintf("making GET request to %s", url))
	resp, err := httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status returned by server: %s", resp.Status)
	}

	respRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	logDebug(fmt.Sprintf("got JSON login response from server: %s", respRaw))

	var camerasResponse getCamerasResponse
	err = json.Unmarshal(respRaw, &camerasResponse)
	if err != nil {
		return nil, err
	}

	if camerasResponse.StatusDescription != "ok" || len(camerasResponse.Items) < 1 {
		return nil, fmt.Errorf("got unexpected json response from server: %+v", camerasResponse)
	}

	logInfo("successfully created login session")
	return &camerasResponse, nil
}

func getImage(sessionToken string, apiURL string, cameraUUID string, imageWidth int) (image.Image, error) {
	logInfo("getting image from camera")

	url := fmt.Sprintf("https://%s%s?uuid=%s&width=%d", apiURL, getImageEndpoint, cameraUUID, imageWidth)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionToken})

	logDebug(fmt.Sprintf("making GET request to %s", url))
	resp, err := httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status returned by server: %s", resp.Status)
	}

	logDebug("decoding image response from server")
	return jpeg.Decode(resp.Body)
}

func saveImage(img image.Image, outFile string) error {
	logInfo(fmt.Sprintf("saving image to %s", outFile))

	imgOut, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer imgOut.Close()

	switch {
	case strings.HasSuffix(outFile, ".png"):
		if err := png.Encode(imgOut, img); err != nil {
			return err
		}

	case strings.HasSuffix(outFile, ".jpeg"):
		if err := jpeg.Encode(imgOut, img, nil); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unrecognized image type for file %s", outFile)
	}

	return nil
}

func logInfo(msg string) {
	if !quiet {
		fmt.Printf("[INFO]: %s\n", msg)
	}
}

func logDebug(msg string) {
	if debug && !quiet {
		fmt.Printf("[DEBUG]: %s\n", msg)
	}
}

func logError(msg string, debugMsg string, fatal bool) {
	// If we have additional debug context with original message, include it
	if debug && debugMsg != "" && !quiet {
		fmt.Printf("[ERROR]: %s - %s\n", msg, debugMsg)
	} else if !quiet {
		fmt.Printf("[ERROR]: %s\n", msg)
	}

	if fatal {
		os.Exit(1)
	}
}
