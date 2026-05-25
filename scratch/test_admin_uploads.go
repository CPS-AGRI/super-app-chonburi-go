package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type LoginResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

func main() {
	// 1. Login to get token
	loginURL := "http://127.0.0.1:8080/api/v1/auth/login"
	loginData := map[string]string{
		"email":    "testadmin@example.com",
		"password": "password",
	}
	jsonData, err := json.Marshal(loginData)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Login connection failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Login failed with status %d: %s\n", resp.StatusCode, string(body))
		return
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		panic(err)
	}
	token := loginResp.Data.Token
	fmt.Printf("Logged in successfully. Token length: %d\n", len(token))

	// 2. Upload KTB Reconciliation file
	ktbPath := `C:\Users\phnjk\.gemini\antigravity-ide\brain\6d9cd73e-a39d-4db6-9b99-1e51cf80f4bf\scratch\ktb_reconciliation_sample.txt`
	ktbURL := "http://127.0.0.1:8080/api/v1/admin/tax-new/reconciliation/upload"
	fmt.Println("Uploading KTB reconciliation file...")
	uploadMultipartFile(ktbURL, ktbPath, token)

	// 3. Upload e-LAAS Daily Summary file
	elaasPath := `C:\Users\phnjk\.gemini\antigravity-ide\brain\6d9cd73e-a39d-4db6-9b99-1e51cf80f4bf\scratch\elaas_daily_summary_sample.csv`
	elaasURL := "http://127.0.0.1:8080/api/v1/admin/tax-new/elaas/upload"
	fmt.Println("Uploading e-LAAS daily summary file...")
	uploadMultipartFile(elaasURL, elaasPath, token)
}

func uploadMultipartFile(targetURL, filePath, token string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Failed to open file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		fmt.Printf("Failed to create form file: %v\n", err)
		return
	}

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Printf("Failed to copy file content: %v\n", err)
		return
	}
	writer.Close()

	req, err := http.NewRequest("POST", targetURL, body)
	if err != nil {
		fmt.Printf("Failed to create HTTP request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Upload request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\nResponse: %s\n\n", resp.StatusCode, string(respBody))
}
