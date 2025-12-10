package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gnaps-api/models"
	"gnaps-api/workers"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
)

// SmsEncoding represents the SMS encoding type
type SmsEncoding string

const (
	EncodingGSM7Bit   SmsEncoding = "gsm_7bit"
	EncodingGSM7BitEx SmsEncoding = "gsm_7bit_ex"
	EncodingUTF16     SmsEncoding = "utf16"
)

// GSM 7-bit character set
const gsm7bitChars = "@£$¥èéùìòÇ\nØø\rÅåΔ_ΦΓΛΩΠΨΣΘΞÆæßÉ !\"#¤%&'()*+,-./0123456789:;<=>?¡ABCDEFGHIJKLMNOPQRSTUVWXYZÄÖÑÜ§¿abcdefghijklmnopqrstuvwxyzäöñüà"
const gsm7bitExChars = "^{}\\[~]|€"

var (
	gsm7bitRegExp   *regexp.Regexp
	gsm7bitExRegExp *regexp.Regexp
)

func init() {
	// Build regex patterns for GSM character detection
	gsm7bitRegExp = regexp.MustCompile(`^[` + regexp.QuoteMeta(gsm7bitChars) + `]*$`)
	gsm7bitExRegExp = regexp.MustCompile(`^[` + regexp.QuoteMeta(gsm7bitChars+gsm7bitExChars) + `]*$`)
}

// SmsCounter counts SMS segments based on encoding
type SmsCounter struct {
	text string
}

// NewSmsCounter creates a new SMS counter
func NewSmsCounter(text string) *SmsCounter {
	return &SmsCounter{text: text}
}

// MessageLength returns the max characters per single SMS for each encoding
func (s *SmsCounter) MessageLength() map[SmsEncoding]int {
	return map[SmsEncoding]int{
		EncodingGSM7Bit:   160,
		EncodingGSM7BitEx: 160,
		EncodingUTF16:     70,
	}
}

// MultiMessageLength returns the max characters per SMS in a multi-part message
func (s *SmsCounter) MultiMessageLength() map[SmsEncoding]int {
	return map[SmsEncoding]int{
		EncodingGSM7Bit:   153,
		EncodingGSM7BitEx: 153,
		EncodingUTF16:     67,
	}
}

// DetectEncoding detects the encoding type for the text
func (s *SmsCounter) DetectEncoding() SmsEncoding {
	if gsm7bitRegExp.MatchString(s.text) {
		return EncodingGSM7Bit
	}
	if gsm7bitExRegExp.MatchString(s.text) {
		return EncodingGSM7BitEx
	}
	return EncodingUTF16
}

// CountGSM7BitEx counts extended GSM characters (which take 2 bytes)
func (s *SmsCounter) CountGSM7BitEx() int {
	count := 0
	for _, char := range s.text {
		if strings.ContainsRune(gsm7bitExChars, char) {
			count++
		}
	}
	return count
}

// Count returns the number of SMS segments needed for the message
func (s *SmsCounter) Count() int {
	encoding := s.DetectEncoding()
	length := utf8.RuneCountInString(s.text)

	// Extended GSM characters take 2 bytes
	if encoding == EncodingGSM7BitEx {
		length += s.CountGSM7BitEx()
	}

	perMessage := s.MessageLength()[encoding]
	if length > perMessage {
		perMessage = s.MultiMessageLength()[encoding]
	}

	if perMessage == 0 {
		return 1
	}

	count := (length + perMessage - 1) / perMessage // Ceiling division
	if count == 0 {
		count = 1
	}
	return count
}

// SmsService handles SMS operations
type SmsService struct {
	db        *gorm.DB
	smsWorker *workers.SmsWorker
	apiURL    string
	username  string
	password  string
}

// NewSmsService creates a new SMS service
func NewSmsService(db *gorm.DB) *SmsService {
	apiURL := os.Getenv("SMS_API_URL")
	if apiURL == "" {
		apiURL = "https://deywuro.com/api/sms"
	}

	return &SmsService{
		db:       db,
		apiURL:   apiURL,
		username: os.Getenv("SMS_USERNAME"),
		password: os.Getenv("SMS_PASSWORD"),
	}
}

// SetWorker sets the SMS worker for background processing
func (s *SmsService) SetWorker(worker *workers.SmsWorker) {
	s.smsWorker = worker
}

// DeywuroRequest represents the request to deywuro.com API
type DeywuroRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Message     string `json:"message"`
}

// DeywuroResponse represents the response from deywuro.com API
type DeywuroResponse struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	MessageID   string `json:"message_id,omitempty"`
	CreditsLeft int    `json:"credits_left,omitempty"`
}

// RefineRecipient cleans and formats a phone number for Ghana
func (s *SmsService) RefineRecipient(recipient string) string {
	// Remove all non-numeric characters except +
	re := regexp.MustCompile(`[^0-9+]`)
	recipient = re.ReplaceAllString(recipient, "")

	// Remove + prefix if present
	recipient = strings.TrimPrefix(recipient, "+")

	if len(recipient) > 8 {
		// Get last 9 digits and prefix with 233 (Ghana)
		if len(recipient) >= 9 {
			recipient = "233" + recipient[len(recipient)-9:]
		}
	} else {
		recipient = ""
	}

	return recipient
}

// GetPackage retrieves the message package for an owner
func (s *SmsService) GetPackage(ownerType string, ownerID int64) (*models.MessagePackage, error) {
	var pkg models.MessagePackage
	err := s.db.Where("owner_type = ? AND owner_id = ?",
		ownerType, ownerID).First(&pkg).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

// GetUsedUnits calculates the total SMS units used by an owner
func (s *SmsService) GetUsedUnits(ownerType string, ownerID int64) float64 {
	var total float64
	s.db.Model(&models.MessageLog{}).
		Where("owner_type = ? AND owner_id = ? AND msg_type = ?", ownerType, ownerID, "SMS").
		Select("COALESCE(SUM(units), 0)").
		Scan(&total)
	return total
}

// GetAvailableUnits returns available SMS units for an owner
func (s *SmsService) GetAvailableUnits(ownerType string, ownerID int64) (int, error) {
	pkg, err := s.GetPackage(ownerType, ownerID)
	if err != nil {
		// If no package exists, return 0 available units (not an error)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	units := 0
	if pkg.Units != nil {
		units = *pkg.Units
	}

	used := s.GetUsedUnits(ownerType, ownerID)
	return units - int(used), nil
}

// WasSentToday checks if the same message was sent to the same recipient today
func (s *SmsService) WasSentToday(recipient, message string) bool {
	var count int64
	today := time.Now().Format("2006-01-02")

	// Check if last 9 digits match
	recipientSuffix := recipient
	if len(recipient) >= 9 {
		recipientSuffix = recipient[len(recipient)-9:]
	}

	s.db.Model(&models.MessageLog{}).
		Where("DATE(created_at) = ? AND recipient LIKE ? AND message = ? AND msg_type = ?",
			today, "%"+recipientSuffix+"%", message, "SMS").
		Count(&count)

	return count > 0
}

// SendSMS sends an SMS to a single recipient
func (s *SmsService) SendSMS(payload workers.SmsSendPayload) error {
	// Check for empty message
	if strings.TrimSpace(payload.Message) == "" {
		return s.logSmsError(payload.Recipient, payload.Message, payload.OwnerType, payload.OwnerID,
			"Message cannot be empty")
	}

	// Refine recipient number
	recipient := s.RefineRecipient(payload.Recipient)
	if recipient == "" || len(recipient) <= 8 {
		return s.logSmsError(payload.Recipient, payload.Message, payload.OwnerType, payload.OwnerID,
			"Invalid phone number")
	}

	// Get SMS package
	pkg, err := s.GetPackage(payload.OwnerType, payload.OwnerID)
	if err != nil {
		return s.logSmsError(payload.Recipient, payload.Message, payload.OwnerType, payload.OwnerID,
			"SMS module must be set up. Contact support.")
	}

	// Check sender name
	if pkg.Sendername == nil || *pkg.Sendername == "" {
		return s.logSmsError(payload.Recipient, payload.Message, payload.OwnerType, payload.OwnerID,
			"SMS sender name must be present and approved")
	}

	// Calculate SMS units needed
	smsCount := NewSmsCounter(payload.Message).Count()

	// Check available units (unless free)
	if !payload.Free {
		availableUnits, _ := s.GetAvailableUnits(payload.OwnerType, payload.OwnerID)
		if availableUnits < smsCount {
			return s.logSmsError(payload.Recipient, payload.Message, payload.OwnerType, payload.OwnerID,
				"SMS package is not available (insufficient units)")
		}
	}

	// Check if already sent today (prevent duplicates)
	if s.WasSentToday(recipient, payload.Message) {
		log.Printf("SMS already sent to %s today, skipping", recipient)
		return nil
	}

	// Send via API
	return s.sendViaAPI(recipient, payload.Message, *pkg.Sendername, smsCount, payload.OwnerType, payload.OwnerID, payload.Free)
}

// sendViaAPI sends the SMS via the deywuro.com API
func (s *SmsService) sendViaAPI(recipient, message, senderName string, smsCount int, ownerType string, ownerID int64, free bool) error {
	// Create message log entry
	msgType := "SMS"
	msgLog := &models.MessageLog{
		MsgType:   &msgType,
		Recipient: &recipient,
		Message:   &message,
		OwnerType: &ownerType,
		OwnerId:   &ownerID,
	}

	// Prepare request
	reqBody := DeywuroRequest{
		Username:    s.username,
		Password:    s.password,
		Source:      senderName,
		Destination: recipient,
		Message:     message,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to marshal request: %v", err)
		msgLog.GatewayResponse = &errMsg
		s.db.Create(msgLog)
		return err
	}

	// Make HTTP request
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create request: %v", err)
		msgLog.GatewayResponse = &errMsg
		s.db.Create(msgLog)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("Request failed: %v", err)
		msgLog.GatewayResponse = &errMsg
		s.db.Create(msgLog)
		return err
	}
	defer resp.Body.Close()

	// Parse response
	var gatewayResp DeywuroResponse
	var sendError error
	units := float64(0)

	if err := json.NewDecoder(resp.Body).Decode(&gatewayResp); err != nil {
		// If not JSON, store raw response
		errMsg := "Non-JSON response received"
		msgLog.GatewayResponse = &errMsg
		sendError = errors.New(errMsg)
	} else {
		respJSON, _ := json.Marshal(gatewayResp)
		respStr := string(respJSON)
		msgLog.GatewayResponse = &respStr

		// Check if successful (code 0 typically means success)
		if gatewayResp.Code == 0 {
			units = float64(smsCount)
		} else {
			// API returned an error code
			sendError = fmt.Errorf("SMS gateway error: %s (code: %d)", gatewayResp.Message, gatewayResp.Code)
		}
	}

	msgLog.Units = &units

	// Save log
	if err := s.db.Create(msgLog).Error; err != nil {
		log.Printf("Failed to save message log: %v", err)
	}

	return sendError
}

// logSmsError logs an SMS error
func (s *SmsService) logSmsError(recipient, message, ownerType string, ownerID int64, errorMsg string) error {
	msgType := "SMS"
	units := float64(0)
	msgLog := &models.MessageLog{
		MsgType:         &msgType,
		Recipient:       &recipient,
		Message:         &message,
		OwnerType:       &ownerType,
		OwnerId:         &ownerID,
		GatewayResponse: &errorMsg,
		Units:           &units,
	}
	s.db.Create(msgLog)
	return errors.New(errorMsg)
}

// SendBulkSMS sends SMS to multiple recipients
func (s *SmsService) SendBulkSMS(payload workers.SmsBulkSendPayload) error {
	var lastErr error

	for _, recipient := range payload.Recipients {
		// Split by comma, newline, or forward slash (like Ruby code)
		parts := regexp.MustCompile(`[,/\n]`).Split(recipient, -1)

		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			singlePayload := workers.SmsSendPayload{
				Message:    payload.Message,
				Recipient:  part,
				OwnerType:  payload.OwnerType,
				OwnerID:    payload.OwnerID,
				SenderName: payload.SenderName,
				Free:       payload.Free,
			}

			if err := s.SendSMS(singlePayload); err != nil {
				log.Printf("Failed to send SMS to %s: %v", part, err)
				lastErr = err
			}
		}
	}

	return lastErr
}

// EnqueueSMS enqueues an SMS for background sending
func (s *SmsService) EnqueueSMS(message, recipient, ownerType string, ownerID int64, senderName string, free bool) error {
	if s.smsWorker == nil {
		// Send synchronously if worker not available
		return s.SendSMS(workers.SmsSendPayload{
			Message:    message,
			Recipient:  recipient,
			OwnerType:  ownerType,
			OwnerID:    ownerID,
			SenderName: senderName,
			Free:       free,
		})
	}

	return s.smsWorker.EnqueueSmsSend(workers.SmsSendPayload{
		Message:    message,
		Recipient:  recipient,
		OwnerType:  ownerType,
		OwnerID:    ownerID,
		SenderName: senderName,
		Free:       free,
	})
}

// EnqueueBulkSMS enqueues a bulk SMS for background sending
func (s *SmsService) EnqueueBulkSMS(message string, recipients []string, ownerType string, ownerID int64, senderName string, free bool) error {
	if s.smsWorker == nil {
		// Send synchronously if worker not available
		return s.SendBulkSMS(workers.SmsBulkSendPayload{
			Message:    message,
			Recipients: recipients,
			OwnerType:  ownerType,
			OwnerID:    ownerID,
			SenderName: senderName,
			Free:       free,
		})
	}

	return s.smsWorker.EnqueueSmsBulkSend(workers.SmsBulkSendPayload{
		Message:    message,
		Recipients: recipients,
		OwnerType:  ownerType,
		OwnerID:    ownerID,
		SenderName: senderName,
		Free:       free,
	})
}
