package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"io"
	"net/http"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type MomoPaymentService struct {
	paymentRepo      *repositories.MomoPaymentRepository
	registrationRepo *repositories.RegistrationRepository
	eventRepo        *repositories.EventRepository
	db               *gorm.DB
}

// GatewayCredentials holds the parsed gateway parameters
type GatewayCredentials struct {
	AccountID   string `json:"accountId"`
	SecretKey   string `json:"skey"`
	CallbackURL string `json:"callbackUrl"`
}

func NewMomoPaymentService(paymentRepo *repositories.MomoPaymentRepository, registrationRepo *repositories.RegistrationRepository, eventRepo *repositories.EventRepository, db *gorm.DB) *MomoPaymentService {
	return &MomoPaymentService{
		paymentRepo:      paymentRepo,
		registrationRepo: registrationRepo,
		eventRepo:        eventRepo,
		db:               db,
	}
}

// GetGatewayCredentials fetches Hubtel credentials from payment_gateways table
func (s *MomoPaymentService) GetGatewayCredentials() (*GatewayCredentials, error) {
	var gateway models.PaymentGateway
	gatewayName := "FidelityPay"

	err := s.db.Where("name = ? AND (is_deleted IS NULL OR is_deleted = ?)", gatewayName, false).First(&gateway).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("payment gateway '%s' not found", gatewayName)
		}
		return nil, fmt.Errorf("failed to fetch payment gateway: %w", err)
	}

	if gateway.GatewayParameters == nil {
		return nil, fmt.Errorf("gateway parameters not configured for '%s'", gatewayName)
	}

	var credentials GatewayCredentials
	if err := json.Unmarshal(*gateway.GatewayParameters, &credentials); err != nil {
		return nil, fmt.Errorf("failed to parse gateway parameters: %w", err)
	}

	if credentials.AccountID == "" || credentials.SecretKey == "" {
		return nil, fmt.Errorf("incomplete gateway credentials: accountId and skey are required")
	}

	return &credentials, nil
}

// MomoPaymentRequest represents the request to initiate a MoMo payment
type MomoPaymentRequest struct {
	Amount            float64 `json:"amount"`
	PhoneNumber       string  `json:"phone_number"`
	Network           string  `json:"network"` // MTN, TELECEL, AIRTELTIGO
	FeeName           string  `json:"fee_name"`
	PayeeID           int64   `json:"payee_id"`
	PayeeType         string  `json:"payee_type"` // e.g., "EventRegistration", "EventPayment"
	UserID            *int64  `json:"user_id,omitempty"`
	SchoolID          int64   `json:"school_id"`
	SchoolName        string  `json:"school_name"`
	NumberOfAttendees int     `json:"number_of_attendees,omitempty"` // For event payments
	EventCode         string  `json:"event_code,omitempty"`          // For event payments
}

// MomoPaymentResponse represents the response from MoMo payment initiation
type MomoPaymentResponse struct {
	Error                bool   `json:"error"`
	Message              string `json:"message"`
	PaymentTransactionID uint   `json:"payment_transaction_id"`
}

// MomoStatusResponse represents the response from checking payment status
type MomoStatusResponse struct {
	Status        string `json:"status"` // pending, successful, failed
	BankStatus    string `json:"bank_status,omitempty"`
	TransStatus   string `json:"trans_status,omitempty"`
	Message       string `json:"message"`
	TransactionID string `json:"transaction_id,omitempty"`
	RawBody       []byte `json:"-"` // Raw response body for saving to gateway_response (not exposed in JSON)
}

// HubtelPaymentRequest represents the request to Hubtel API
type HubtelPaymentRequest struct {
	CustomerMsisdn     string  `json:"CustomerMsisdn"`
	Channel            string  `json:"Channel"`
	Amount             float64 `json:"Amount"`
	PrimaryCallbackUrl string  `json:"PrimaryCallbackUrl"`
	Description        string  `json:"Description"`
	ClientReference    string  `json:"ClientReference"`
}

// HubtelPaymentResponse represents the response from Hubtel API
type HubtelPaymentResponse struct {
	ResponseCode string `json:"ResponseCode"`
	Data         struct {
		TransactionId   string  `json:"TransactionId"`
		ClientReference string  `json:"ClientReference"`
		Amount          float64 `json:"Amount"`
		Charges         float64 `json:"Charges"`
		Description     string  `json:"Description"`
	} `json:"Data"`
}

// InitiatePayment creates a new MoMo payment record (equivalent to Ruby's momo_pay)
func (s *MomoPaymentService) InitiatePayment(req MomoPaymentRequest) (*MomoPaymentResponse, error) {
	// Check if there's already a pending payment for this payee
	existingPayment, _ := s.paymentRepo.FindPendingByPayeeAndType(req.SchoolID, req.PayeeID, req.PayeeType)
	if existingPayment != nil {
		return &MomoPaymentResponse{
			Error:                false,
			Message:              "Payment already initiated",
			PaymentTransactionID: existingPayment.ID,
		}, nil
	}

	// Generate transaction ID like Ruby: S#{school_id}T#{timestamp}-#{member_no}
	transactionID := fmt.Sprintf("S%dT%d-EVT%d", req.SchoolID, time.Now().Unix(), req.PayeeID)

	// Create payment details JSON - include registration data for EventPayment type
	paymentDetails := map[string]interface{}{
		"school_name": "GNAPS",
		"aggregator":  "FidelityPay",
		"paid_by":     req.SchoolName,
	}

	// For EventPayment type, store registration data to create registration after payment success
	if req.PayeeType == "EventPayment" {
		paymentDetails["event_code"] = req.EventCode
		paymentDetails["school_id"] = req.SchoolID
		paymentDetails["number_of_attendees"] = req.NumberOfAttendees
		paymentDetails["payment_phone"] = req.PhoneNumber
		paymentDetails["payment_method"] = req.Network
	}

	detailsJSON, _ := json.Marshal(paymentDetails)
	jsonDetails := datatypes.JSON(detailsJSON)

	// Create the payment record with "created" status
	// Background job will process it and change to "pending" when sent to Hubtel
	status := "created"
	retries := 0
	payment := &models.MomoPayment{
		Amount:            &req.Amount,
		Status:            &status,
		PayeeId:           &req.PayeeID,
		PayeeType:         &req.PayeeType,
		UserId:            req.UserID,
		SchoolId:          &req.SchoolID,
		FeeName:           &req.FeeName,
		MomoNetwork:       &req.Network,
		MomoNumber:        &req.PhoneNumber,
		TransactionDate:   time.Now(),
		MomoTransactionId: &transactionID,
		PaymentDetails:    &jsonDetails,
		Retries:           &retries,
	}

	// Set owner fields from payee entity
	// For EventPayment type, get ownership from Event by code
	if req.PayeeType == "EventPayment" && req.EventCode != "" {
		ownerInfo := repositories.GetOwnerFromEvent(s.db, req.EventCode)
		if ownerInfo != nil {
			payment.SetOwner(ownerInfo.OwnerType, ownerInfo.OwnerID)
		}
	} else {
		// For other types (SchoolBillPayment, EventRegistration), get ownership from payee
		ownerInfo := repositories.GetOwnerFromPayee(s.db, req.PayeeType, req.PayeeID)
		if ownerInfo != nil {
			payment.SetOwner(ownerInfo.OwnerType, ownerInfo.OwnerID)
		}
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return &MomoPaymentResponse{
			Error:   true,
			Message: "Payment request initiation failed, please contact admin",
		}, err
	}

	return &MomoPaymentResponse{
		Error:                false,
		Message:              "Transaction initiated successfully",
		PaymentTransactionID: payment.ID,
	}, nil
}

// ProcessPaymentWithHubtel processes payment with Hubtel API (equivalent to Ruby's EpayPaymentRequest job)
func (s *MomoPaymentService) ProcessPaymentWithHubtel(paymentID uint) error {
	payment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		return fmt.Errorf("failed to find payment: %w", err)
	}
	if payment == nil {
		return fmt.Errorf("payment not found: id=%d", paymentID)
	}

	// Skip if payment is not in "created" status (already processed)
	if payment.Status != nil && *payment.Status != "created" {
		fmt.Printf("Payment %d already processed (status: %s), skipping\n", paymentID, *payment.Status)
		return nil
	}

	// Get Hubtel credentials from payment_gateways table
	credentials, err := s.GetGatewayCredentials()
	if err != nil {
		// For development/testing without Hubtel credentials
		fmt.Printf("Hubtel credentials not configured: %v. Simulating payment for ID: %d\n", err, paymentID)
		// Update status to pending even in dev mode
		s.paymentRepo.UpdateStatus(paymentID, "pending", "", nil)
		return nil
	}

	// Build Hubtel API URL
	baseURL := fmt.Sprintf("https://rmp.hubtel.com/merchantaccount/merchants/%s/receive/mobilemoney", credentials.AccountID)

	// Format phone number for Hubtel API
	// Hubtel expects format: 233XXXXXXXXX (country code + 9 digits)
	phoneNumber := *payment.MomoNumber
	// Remove any spaces, dashes, or other characters
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")

	// If starts with 0, remove it and add 233
	if strings.HasPrefix(phoneNumber, "0") {
		phoneNumber = "233" + phoneNumber[1:]
	} else if !strings.HasPrefix(phoneNumber, "233") {
		// If doesn't start with 233, add it (take last 9 digits)
		if len(phoneNumber) >= 9 {
			phoneNumber = "233" + phoneNumber[len(phoneNumber)-9:]
		}
	}

	// Map network to Hubtel channel
	channel := s.mapNetworkToHubtelChannel(*payment.MomoNetwork)

	// Build callback URL with reference
	ref := *payment.MomoTransactionId
	callbackURL := credentials.CallbackURL
	if callbackURL == "" {
		callbackURL = "https://sch-cmp.adesua360.com/api/epayments" // Default callback URL
	}
	fullCallbackURL := fmt.Sprintf("%s?action_type=verify_fidelity_pay&ref=%s", callbackURL, ref)

	// Prepare payment request - round Amount to 2 decimal places like Ruby does
	roundedAmount := float64(int(*payment.Amount*100)) / 100
	hubtelReq := HubtelPaymentRequest{
		CustomerMsisdn:     phoneNumber,
		Channel:            channel,
		Amount:             roundedAmount,
		PrimaryCallbackUrl: fullCallbackURL,
		Description:        fmt.Sprintf("Payment for %s", *payment.FeeName),
		ClientReference:    ref,
	}

	jsonPayload, err := json.Marshal(hubtelReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Log request for debugging
	fmt.Printf("Hubtel Payment Request URL: %s\n", baseURL)
	fmt.Printf("Hubtel Payment Request Body: %s\n", string(jsonPayload))
	fmt.Printf("Hubtel AccountID: %s\n", credentials.AccountID)
	fmt.Printf("Hubtel SecretKey (first 20 chars): %.20s...\n", credentials.SecretKey)

	// Make HTTP request to Hubtel
	httpReq, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("No-Cache", "true")
	httpReq.Header.Set("Authorization", "Basic "+credentials.SecretKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		// Save error to gateway_response
		s.saveGatewayResponse(payment.ID, map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("hubtel request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Log response for debugging
	fmt.Printf("Hubtel Payment Response (HTTP %d): %s\n", resp.StatusCode, string(body))

	// Save raw gateway response body directly
	s.saveRawGatewayResponse(payment.ID, body)

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("hubtel returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var hubtelResp HubtelPaymentResponse
	if err := json.Unmarshal(body, &hubtelResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Update payment with order ID if available
	if hubtelResp.Data.TransactionId != "" {
		// Update payment details with order ID
		var details map[string]interface{}
		if payment.PaymentDetails != nil {
			json.Unmarshal(*payment.PaymentDetails, &details)
		} else {
			details = make(map[string]interface{})
		}
		details["orderID"] = hubtelResp.Data.TransactionId
		detailsJSON, _ := json.Marshal(details)
		jsonDetails := datatypes.JSON(detailsJSON)
		s.paymentRepo.Update(payment.ID, map[string]interface{}{
			"payment_details": jsonDetails,
		})
	}

	// Update status from "created" to "pending" - payment has been sent to Hubtel
	s.paymentRepo.UpdateStatus(payment.ID, "pending", "", nil)
	fmt.Printf("Payment %d sent to Hubtel, status changed to pending\n", payment.ID)

	return nil
}

// ProcessCreatedPayments fetches payments with "created" status and processes them with Hubtel
// This is called by the background ticker every 3 seconds
func (s *MomoPaymentService) ProcessCreatedPayments() error {
	// Get all payments with "created" status
	createdPayments, err := s.paymentRepo.GetCreatedPayments()
	if err != nil {
		return fmt.Errorf("failed to get created payments: %w", err)
	}

	if len(createdPayments) == 0 {
		return nil
	}

	fmt.Printf("Found %d created payments to process with Hubtel\n", len(createdPayments))

	// Process each created payment
	for _, payment := range createdPayments {
		fmt.Printf("Processing created payment %d...\n", payment.ID)
		if err := s.ProcessPaymentWithHubtel(payment.ID); err != nil {
			fmt.Printf("Error processing payment %d: %v\n", payment.ID, err)
			// Continue processing other payments even if one fails
			continue
		}
	}

	return nil
}

// mapNetworkToHubtelChannel maps network names to Hubtel channel codes
func (s *MomoPaymentService) mapNetworkToHubtelChannel(network string) string {
	switch network {
	case "MTN":
		return "mtn-gh"
	case "TELECEL", "VODAFONE":
		return "vodafone-gh"
	case "AIRTELTIGO":
		return "tigo-gh"
	default:
		return "mtn-gh"
	}
}

// saveGatewayResponse saves the gateway response to the payment record
func (s *MomoPaymentService) saveGatewayResponse(paymentID uint, response interface{}) {
	responseJSON, _ := json.Marshal(response)
	jsonResponse := datatypes.JSON(responseJSON)
	s.paymentRepo.Update(paymentID, map[string]interface{}{
		"gateway_response": jsonResponse,
	})
}

// saveRawGatewayResponse saves the raw gateway response body directly without formatting
func (s *MomoPaymentService) saveRawGatewayResponse(paymentID uint, rawBody []byte) {
	jsonResponse := datatypes.JSON(rawBody)
	s.paymentRepo.Update(paymentID, map[string]interface{}{
		"gateway_response": jsonResponse,
	})
}

// CheckPaymentStatus checks the status of a payment
func (s *MomoPaymentService) CheckPaymentStatus(paymentID uint) (*MomoStatusResponse, error) {
	payment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		return nil, errors.New("payment not found")
	}

	if payment == nil {
		return nil, errors.New("payment not found")
	}

	// If already completed or failed, return the status
	if payment.Status != nil && (*payment.Status == "successful" || *payment.Status == "success" || *payment.Status == "failed") {
		bankStatus := ""
		if payment.BankStatus != nil {
			bankStatus = *payment.BankStatus
		}
		transStatus := ""
		if payment.TransStatus != nil {
			transStatus = *payment.TransStatus
		}
		transactionID := ""
		if payment.MomoTransactionId != nil {
			transactionID = *payment.MomoTransactionId
		}

		// Normalize status
		status := *payment.Status
		if status == "success" {
			status = "successful"
		}

		return &MomoStatusResponse{
			Status:        status,
			BankStatus:    bankStatus,
			TransStatus:   transStatus,
			Message:       s.getStatusMessage(status),
			TransactionID: transactionID,
		}, nil
	}

	// Get gateway credentials
	credentials, credErr := s.GetGatewayCredentials()
	if credErr != nil || credentials == nil {
		// For development: simulate status progression after retries when gateway not configured
		s.paymentRepo.IncrementRetries(payment.ID)

		// After 3 retries (about 15 seconds with 5-second polling), mark as successful
		if payment.Retries != nil && *payment.Retries >= 3 {
			successStatus := "successful"
			approvedBank := "APPROVED"
			s.paymentRepo.UpdateStatus(payment.ID, successStatus, approvedBank, nil)

			// Handle payment success based on payee type
			if payment.PayeeType != nil {
				if *payment.PayeeType == "EventRegistration" {
					// Legacy flow: update existing registration
					s.registrationRepo.UpdatePaymentStatus(uint(*payment.PayeeId), "paid", fmt.Sprintf("MOMO_%d", payment.ID))
				} else if *payment.PayeeType == "EventPayment" {
					// New flow: create registration after payment success
					if err := s.createEventRegistrationFromPayment(payment); err != nil {
						fmt.Printf("Error creating event registration for payment %d: %v\n", payment.ID, err)
					}
				} else if *payment.PayeeType == "SchoolBillPayment" {
					// School bill payment: update school bill balance
					if err := s.processSchoolBillPaymentSuccess(payment); err != nil {
						fmt.Printf("Error processing school bill payment %d: %v\n", payment.ID, err)
					}
				}
			}

			// Create finance transaction for successful payment
			if err := s.createFinanceTransaction(payment); err != nil {
				fmt.Printf("Error creating finance transaction for payment %d: %v\n", payment.ID, err)
			}

			return &MomoStatusResponse{
				Status:     "successful",
				BankStatus: "APPROVED",
				Message:    "Payment completed successfully",
			}, nil
		}

		return &MomoStatusResponse{
			Status:  "pending",
			Message: "Payment is being processed. Please approve on your phone.",
		}, nil
	}

	// Query Hubtel for actual payment status
	if payment.MomoTransactionId == nil {
		return &MomoStatusResponse{
			Status:  "pending",
			Message: "Payment is being processed.",
		}, nil
	}

	hubtelStatus, err := s.queryHubtelTransactionStatus(credentials, *payment.MomoTransactionId)
	if err != nil {
		fmt.Printf("Error checking Hubtel status for payment %d: %v\n", payment.ID, err)
		// Save error to gateway_response
		s.saveGatewayResponse(payment.ID, map[string]interface{}{"error": err.Error()})
		// Increment retries on error
		s.paymentRepo.IncrementRetries(payment.ID)

		// If retries exceed 69, mark as failed (like Ruby code)
		if payment.Retries != nil && *payment.Retries > 69 {
			failedStatus := "failed"
			s.paymentRepo.UpdateStatus(payment.ID, failedStatus, "TIMEOUT", nil)
			return &MomoStatusResponse{
				Status:  "failed",
				Message: "Payment verification timeout. Please try again.",
			}, nil
		}

		return &MomoStatusResponse{
			Status:  "pending",
			Message: "Payment is being processed. Please approve on your phone.",
		}, nil
	}

	// Save Hubtel response to gateway_response (like Ruby: self.gateway_response = response_body)
	if hubtelStatus.RawBody != nil {
		s.saveRawGatewayResponse(payment.ID, hubtelStatus.RawBody)
	}

	// Update payment based on Hubtel response
	s.paymentRepo.IncrementRetries(payment.ID)

	if hubtelStatus.Status == "successful" || hubtelStatus.Status == "success" {
		s.paymentRepo.UpdateStatus(payment.ID, "successful", "APPROVED", &hubtelStatus.TransStatus)

		// Handle payment success based on payee type
		if payment.PayeeType != nil {
			if *payment.PayeeType == "EventRegistration" {
				// Legacy flow: update existing registration
				s.registrationRepo.UpdatePaymentStatus(uint(*payment.PayeeId), "paid", fmt.Sprintf("MOMO_%d", payment.ID))
			} else if *payment.PayeeType == "EventPayment" {
				// New flow: create registration after payment success
				if err := s.createEventRegistrationFromPayment(payment); err != nil {
					fmt.Printf("Error creating event registration for payment %d: %v\n", payment.ID, err)
				}
			} else if *payment.PayeeType == "SchoolBillPayment" {
				// School bill payment: update school bill balance
				if err := s.processSchoolBillPaymentSuccess(payment); err != nil {
					fmt.Printf("Error processing school bill payment %d: %v\n", payment.ID, err)
				}
			}
		}

		// Create finance transaction for successful payment
		if err := s.createFinanceTransaction(payment); err != nil {
			fmt.Printf("Error creating finance transaction for payment %d: %v\n", payment.ID, err)
		}

		// Update school bill (reduce balance, increase amount paid) for event payments
		if err := s.updateSchoolBillAfterPayment(payment); err != nil {
			fmt.Printf("Error updating school bill for payment %d: %v\n", payment.ID, err)
		}

		return &MomoStatusResponse{
			Status:        "successful",
			BankStatus:    "APPROVED",
			TransStatus:   hubtelStatus.TransStatus,
			Message:       "Payment completed successfully",
			TransactionID: hubtelStatus.TransactionID,
		}, nil
	} else if hubtelStatus.Status == "failed" {
		s.paymentRepo.UpdateStatus(payment.ID, "failed", "DECLINED", &hubtelStatus.TransStatus)
		return &MomoStatusResponse{
			Status:      "failed",
			BankStatus:  "DECLINED",
			TransStatus: hubtelStatus.TransStatus,
			Message:     hubtelStatus.Message,
		}, nil
	}

	// Still pending - check if retries exceeded
	if payment.Retries != nil && *payment.Retries > 69 {
		failedStatus := "failed"
		s.paymentRepo.UpdateStatus(payment.ID, failedStatus, "TIMEOUT", nil)
		return &MomoStatusResponse{
			Status:  "failed",
			Message: "Payment verification timeout. Please try again.",
		}, nil
	}

	return &MomoStatusResponse{
		Status:      "pending",
		TransStatus: hubtelStatus.TransStatus,
		Message:     "Payment is being processed. Please approve on your phone.",
	}, nil
}

// HandleHubtelCallback handles the callback from Hubtel
func (s *MomoPaymentService) HandleHubtelCallback(reference string, status string, transactionID string) error {
	payment, err := s.paymentRepo.FindByMomoTransactionID(reference)
	if err != nil {
		return fmt.Errorf("payment not found for reference: %s", reference)
	}

	// Map Hubtel status to our status
	var paymentStatus string
	var bankStatus string
	switch status {
	case "Success", "Successful":
		paymentStatus = "successful"
		bankStatus = "APPROVED"
	case "Failed", "Declined":
		paymentStatus = "failed"
		bankStatus = "DECLINED"
	default:
		paymentStatus = "pending"
		bankStatus = "PENDING"
	}

	// Update payment status
	s.paymentRepo.UpdateStatus(payment.ID, paymentStatus, bankStatus, &status)

	// If successful, handle based on payee type and create finance transaction
	if paymentStatus == "successful" && payment.PayeeType != nil {
		if *payment.PayeeType == "EventRegistration" {
			// Legacy flow: update existing registration
			s.registrationRepo.UpdatePaymentStatus(uint(*payment.PayeeId), "paid", fmt.Sprintf("MOMO_%d", payment.ID))
		} else if *payment.PayeeType == "EventPayment" {
			// New flow: create registration after payment success
			if err := s.createEventRegistrationFromPayment(payment); err != nil {
				fmt.Printf("Error creating event registration for payment %d: %v\n", payment.ID, err)
			}
		} else if *payment.PayeeType == "SchoolBillPayment" {
			// School bill payment: update school bill balance
			if err := s.processSchoolBillPaymentSuccess(payment); err != nil {
				fmt.Printf("Error processing school bill payment %d: %v\n", payment.ID, err)
			}
		}

		// Create finance transaction for successful payment
		if err := s.createFinanceTransaction(payment); err != nil {
			fmt.Printf("Error creating finance transaction for payment %d: %v\n", payment.ID, err)
		}

		// Update school bill (reduce balance, increase amount paid) for event payments
		if err := s.updateSchoolBillAfterPayment(payment); err != nil {
			fmt.Printf("Error updating school bill for payment %d: %v\n", payment.ID, err)
		}
	}

	return nil
}

// getStatusMessage returns a human-readable message for the status
func (s *MomoPaymentService) getStatusMessage(status string) string {
	switch status {
	case "successful":
		return "Payment completed successfully"
	case "failed":
		return "Payment failed. Please try again."
	case "pending":
		return "Payment is being processed. Please approve on your phone."
	default:
		return "Unknown payment status"
	}
}

// createFinanceTransaction creates a finance transaction for a successful payment
// and updates the momo_payment with the finance_transaction_ids
func (s *MomoPaymentService) createFinanceTransaction(payment *models.MomoPayment) error {
	if payment == nil {
		return fmt.Errorf("payment is nil")
	}

	// Check if finance transaction already exists for this payment
	if payment.FinanceTransactionIds != nil {
		var existingIds []uint
		if err := json.Unmarshal(*payment.FinanceTransactionIds, &existingIds); err == nil && len(existingIds) > 0 {
			// Already has finance transactions, skip creation
			return nil
		}
	}

	// Build title and description
	feeName := "Payment"
	if payment.FeeName != nil {
		feeName = *payment.FeeName
	}
	title := fmt.Sprintf("%s Payment", feeName)
	description := fmt.Sprintf("MoMo payment for %s", feeName)

	// Get payment mode info
	paymentMode := "Mobile Money"
	modeInfo := ""
	if payment.MomoNetwork != nil {
		modeInfo = fmt.Sprintf("%s - ", *payment.MomoNetwork)
	}
	if payment.MomoNumber != nil {
		modeInfo += *payment.MomoNumber
	}

	// Get reference number
	referenceNo := ""
	if payment.MomoTransactionId != nil {
		referenceNo = *payment.MomoTransactionId
	}

	// Get finance account for payments (find or use default)
	var financeAccountId *int64
	var financeAccount models.FinanceAccount
	// Try to find an income account
	err := s.db.Where("account_type = ? AND is_income = ? AND (is_deleted = ? OR is_deleted IS NULL)",
		"income", true, false).First(&financeAccount).Error
	if err == nil {
		id := int64(financeAccount.ID)
		financeAccountId = &id
	}

	// Determine finance type - normalize payment types for consistency
	financeType := payment.PayeeType
	if payment.PayeeType != nil {
		switch *payment.PayeeType {
		case "SchoolBillPayment":
			// Use "SchoolBill" to be consistent with cash payments
			schoolBillType := "SchoolBill"
			financeType = &schoolBillType
		case "EventPayment":
			// Use "EventRegistration" for consistency (this should already be updated by createEventRegistrationFromPayment)
			eventRegType := "EventRegistration"
			financeType = &eventRegType
		}
	}

	// Ensure SchoolId is set - try to extract from payment details if not on payment object
	schoolId := payment.SchoolId
	if (schoolId == nil || *schoolId == 0) && payment.PaymentDetails != nil {
		var details map[string]interface{}
		if err := json.Unmarshal(*payment.PaymentDetails, &details); err == nil {
			if sid, ok := details["school_id"].(float64); ok && sid > 0 {
				sidInt := int64(sid)
				schoolId = &sidInt
			}
		}
	}

	// Create the finance transaction
	financeTransaction := &models.FinanceTransaction{
		Title:            &title,
		Description:      &description,
		Amount:           payment.Amount,
		FinanceAccountId: financeAccountId,
		TransactionDate:  time.Now(),
		FinanceId:        payment.PayeeId,
		FinanceType:      financeType,
		SchoolId:         schoolId, // Include school ID for proper tracking
		PaymentMode:      &paymentMode,
		ModeInfo:         &modeInfo,
		ReferenceNo:      &referenceNo,
		UserId:           payment.UserId,
	}

	// Copy payment details if available
	if payment.PaymentDetails != nil {
		financeTransaction.PaymentDetails = payment.PaymentDetails
	}

	// Set owner fields from the finance entity (payee)
	// First try to get from the payment itself (if it has owner fields set)
	if payment.OwnerType != nil && payment.OwnerId != nil && *payment.OwnerId > 0 {
		financeTransaction.SetOwner(*payment.OwnerType, *payment.OwnerId)
	} else if financeType != nil && payment.PayeeId != nil {
		// Look up ownership from the finance entity
		ownerInfo := repositories.GetOwnerFromFinance(s.db, *financeType, *payment.PayeeId)
		if ownerInfo != nil {
			financeTransaction.SetOwner(ownerInfo.OwnerType, ownerInfo.OwnerID)
		}
	}

	// Save the finance transaction
	if err := s.db.Create(financeTransaction).Error; err != nil {
		return fmt.Errorf("failed to create finance transaction: %w", err)
	}

	// Update momo_payment with the finance_transaction_ids as JSON array
	transactionIds := []uint{financeTransaction.ID}
	idsJSON, err := json.Marshal(transactionIds)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction ids: %w", err)
	}
	jsonIds := datatypes.JSON(idsJSON)

	// Update trans_status to Done (like Ruby code)
	transStatus := "Done"
	if err := s.paymentRepo.Update(payment.ID, map[string]interface{}{
		"finance_transaction_ids": jsonIds,
		"trans_status":            transStatus,
	}); err != nil {
		return fmt.Errorf("failed to update payment with transaction ids: %w", err)
	}

	fmt.Printf("Created finance transaction %d for payment %d\n", financeTransaction.ID, payment.ID)
	return nil
}

// createEventRegistrationFromPayment creates an event registration after successful payment
// This is used for the payment-first flow where registration is created only after payment succeeds
func (s *MomoPaymentService) createEventRegistrationFromPayment(payment *models.MomoPayment) error {
	if payment == nil || payment.PaymentDetails == nil {
		return fmt.Errorf("payment or payment details is nil")
	}

	// Parse payment details to get registration data
	var details map[string]interface{}
	if err := json.Unmarshal(*payment.PaymentDetails, &details); err != nil {
		return fmt.Errorf("failed to parse payment details: %w", err)
	}

	// Extract registration data from payment details
	eventCode, ok := details["event_code"].(string)
	if !ok || eventCode == "" {
		return fmt.Errorf("event_code not found in payment details")
	}

	schoolID, ok := details["school_id"].(float64)
	if !ok || schoolID == 0 {
		return fmt.Errorf("school_id not found in payment details")
	}

	numberOfAttendees := 1
	if num, ok := details["number_of_attendees"].(float64); ok && num > 0 {
		numberOfAttendees = int(num)
	}

	paymentPhone := ""
	if phone, ok := details["payment_phone"].(string); ok {
		paymentPhone = phone
	}

	paymentMethod := ""
	if method, ok := details["payment_method"].(string); ok {
		paymentMethod = method
	}

	// Get event by code
	event, err := s.eventRepo.FindByCode(eventCode)
	if err != nil {
		return fmt.Errorf("event not found for code %s: %w", eventCode, err)
	}

	// Create the registration
	paidStatus := "paid"
	paymentRef := fmt.Sprintf("MOMO_%d", payment.ID)
	schoolIDInt := int64(schoolID)
	registration := &models.EventRegistration{
		EventId:           int64(event.ID),
		SchoolId:          schoolIDInt,
		NumberOfAttendees: &numberOfAttendees,
		PaymentStatus:     &paidStatus,
		PaymentReference:  &paymentRef,
		PaymentMethod:     &paymentMethod,
		PaymentPhone:      &paymentPhone,
		RegistrationDate:  time.Now(),
	}

	// Set owner fields from the Event
	if event.OwnerType != nil && event.OwnerId != nil && *event.OwnerId > 0 {
		registration.SetOwner(*event.OwnerType, *event.OwnerId)
	}

	if err := s.registrationRepo.Create(registration); err != nil {
		return fmt.Errorf("failed to create registration: %w", err)
	}

	// Update payment's payee_id to the new registration ID for tracking
	newPayeeId := int64(registration.ID)
	newPayeeType := "EventRegistration"
	if err := s.paymentRepo.Update(payment.ID, map[string]interface{}{
		"payee_id":   newPayeeId,
		"payee_type": newPayeeType,
	}); err != nil {
		fmt.Printf("Warning: failed to update payment payee_id: %v\n", err)
	}

	// Also update in-memory payment object for subsequent operations (like createFinanceTransaction)
	payment.PayeeId = &newPayeeId
	payment.PayeeType = &newPayeeType

	// Ensure SchoolId is set on the payment from payment details if not already set
	if payment.SchoolId == nil || *payment.SchoolId == 0 {
		schoolIDInt := int64(schoolID)
		payment.SchoolId = &schoolIDInt
	}

	fmt.Printf("Created event registration %d for payment %d (event: %s, school: %d)\n",
		registration.ID, payment.ID, eventCode, int(schoolID))
	return nil
}

// updateSchoolBillAfterPayment updates the school bill (AmountPaid and Balance) after a successful event payment
// This should be called whenever an event registration payment succeeds
func (s *MomoPaymentService) updateSchoolBillAfterPayment(payment *models.MomoPayment) error {
	if payment == nil || payment.PaymentDetails == nil {
		return nil // Not an error, just nothing to update
	}

	// Parse payment details to get event code
	var details map[string]interface{}
	if err := json.Unmarshal(*payment.PaymentDetails, &details); err != nil {
		return fmt.Errorf("failed to parse payment details: %w", err)
	}

	// Get event code from payment details
	eventCode, ok := details["event_code"].(string)
	if !ok || eventCode == "" {
		// No event code means this isn't an event payment
		return nil
	}

	// Get school ID from payment details or payment itself
	var schoolID int64
	if sid, ok := details["school_id"].(float64); ok && sid > 0 {
		schoolID = int64(sid)
	} else if payment.SchoolId != nil {
		schoolID = *payment.SchoolId
	}
	if schoolID == 0 {
		return fmt.Errorf("school_id not found for payment %d", payment.ID)
	}

	// Find the event by code
	event, err := s.eventRepo.FindByCode(eventCode)
	if err != nil {
		return fmt.Errorf("event not found for code %s: %w", eventCode, err)
	}

	// Check if event has a linked bill
	if event.BillId == nil {
		fmt.Printf("Event %d (%s) has no linked bill, skipping bill update\n", event.ID, eventCode)
		return nil
	}

	// Find the school bill for this school and bill
	var schoolBill models.SchoolBill
	err = s.db.Where("school_id = ? AND bill_id = ? AND deleted_at IS NULL", schoolID, *event.BillId).First(&schoolBill).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("No school bill found for school %d and bill %d, skipping bill update\n", schoolID, *event.BillId)
			return nil
		}
		return fmt.Errorf("failed to find school bill: %w", err)
	}

	// Get payment amount
	paymentAmount := 0.0
	if payment.Amount != nil {
		paymentAmount = *payment.Amount
	}
	if paymentAmount <= 0 {
		return nil // No amount to apply
	}

	// Calculate new values
	currentAmountPaid := 0.0
	if schoolBill.AmountPaid != nil {
		currentAmountPaid = *schoolBill.AmountPaid
	}
	currentBalance := 0.0
	if schoolBill.Balance != nil {
		currentBalance = *schoolBill.Balance
	}

	newAmountPaid := currentAmountPaid + paymentAmount
	newBalance := currentBalance - paymentAmount
	if newBalance < 0 {
		newBalance = 0 // Don't go negative
	}

	// Update the school bill
	if err := s.db.Model(&schoolBill).Updates(map[string]interface{}{
		"amount_paid": newAmountPaid,
		"balance":     newBalance,
	}).Error; err != nil {
		return fmt.Errorf("failed to update school bill: %w", err)
	}

	fmt.Printf("Updated school bill %d: AmountPaid %.2f -> %.2f, Balance %.2f -> %.2f (payment %d, amount %.2f)\n",
		schoolBill.ID, currentAmountPaid, newAmountPaid, currentBalance, newBalance, payment.ID, paymentAmount)
	return nil
}

// processSchoolBillPaymentSuccess handles successful school bill MoMo payments
// PayeeID contains the school_bill_id, and we update the balance directly
func (s *MomoPaymentService) processSchoolBillPaymentSuccess(payment *models.MomoPayment) error {
	if payment == nil || payment.PayeeId == nil {
		return fmt.Errorf("payment or payee_id is nil")
	}

	schoolBillID := uint(*payment.PayeeId)

	// Get the school bill
	var schoolBill models.SchoolBill
	if err := s.db.First(&schoolBill, schoolBillID).Error; err != nil {
		return fmt.Errorf("school bill not found: %w", err)
	}

	// Get payment amount
	paymentAmount := 0.0
	if payment.Amount != nil {
		paymentAmount = *payment.Amount
	}
	if paymentAmount <= 0 {
		return nil // No amount to apply
	}

	// Calculate new values
	currentAmountPaid := 0.0
	if schoolBill.AmountPaid != nil {
		currentAmountPaid = *schoolBill.AmountPaid
	}
	currentBalance := 0.0
	if schoolBill.Balance != nil {
		currentBalance = *schoolBill.Balance
	}

	newAmountPaid := currentAmountPaid + paymentAmount
	newBalance := currentBalance - paymentAmount
	if newBalance < 0 {
		newBalance = 0 // Don't go negative
	}

	// Determine if fully paid
	isPaid := newBalance <= 0

	// Update the school bill
	if err := s.db.Model(&schoolBill).Updates(map[string]interface{}{
		"amount_paid": newAmountPaid,
		"balance":     newBalance,
		"is_paid":     isPaid,
	}).Error; err != nil {
		return fmt.Errorf("failed to update school bill: %w", err)
	}

	fmt.Printf("School bill payment %d successful: Bill %d - AmountPaid %.2f -> %.2f, Balance %.2f -> %.2f\n",
		payment.ID, schoolBillID, currentAmountPaid, newAmountPaid, currentBalance, newBalance)
	return nil
}

// GetPaymentByID retrieves a payment by ID
func (s *MomoPaymentService) GetPaymentByID(id uint) (*models.MomoPayment, error) {
	return s.paymentRepo.FindByID(id)
}

// GetPaymentByPayee retrieves the latest payment for a payee
func (s *MomoPaymentService) GetPaymentByPayee(payeeID int64, payeeType string) (*models.MomoPayment, error) {
	return s.paymentRepo.FindByPayeeAndType(payeeID, payeeType)
}

// CheckAndUpdatePendingPayments checks all pending payments and updates their status from Hubtel
func (s *MomoPaymentService) CheckAndUpdatePendingPayments() error {
	// Get gateway credentials
	credentials, err := s.GetGatewayCredentials()
	if err != nil {
		// If no credentials, skip the check (development mode)
		fmt.Printf("Skipping payment status check: %v\n", err)
		return nil
	}

	// Get all pending payments
	pendingPayments, err := s.paymentRepo.GetPendingPayments()
	if err != nil {
		return fmt.Errorf("failed to get pending payments: %w", err)
	}

	if len(pendingPayments) == 0 {
		return nil
	}

	fmt.Printf("Found %d pending payments to check\n", len(pendingPayments))

	// Check status for each pending payment
	for _, payment := range pendingPayments {
		if payment.MomoTransactionId == nil {
			continue
		}

		// Check if payment is older than 10 minutes - auto-fail if still pending
		// Use CreatedAt instead of TransactionDate (TransactionDate is DATE type without time component)
		paymentAge := time.Since(payment.CreatedAt)
		if paymentAge > 10*time.Minute {
			fmt.Printf("Payment %d is older than 10 minutes (%.1f mins, created: %s), marking as failed\n", payment.ID, paymentAge.Minutes(), payment.CreatedAt.Format(time.RFC3339))
			s.paymentRepo.UpdateStatus(payment.ID, "failed", "TIMEOUT", nil)
			s.saveGatewayResponse(payment.ID, map[string]interface{}{
				"error":   "Payment timeout - exceeded 10 minutes",
				"age_sec": paymentAge.Seconds(),
			})
			continue
		}

		// Query Hubtel for transaction status
		status, err := s.queryHubtelTransactionStatus(credentials, *payment.MomoTransactionId)

		// Always save raw Hubtel response to gateway_response (even on error)
		if status != nil && status.RawBody != nil {
			s.saveRawGatewayResponse(payment.ID, status.RawBody)
		}

		if err != nil {
			fmt.Printf("Error checking status for payment %d: %v\n", payment.ID, err)
			// If we didn't get a raw body, save the error message
			if status == nil || status.RawBody == nil {
				s.saveGatewayResponse(payment.ID, map[string]interface{}{"error": err.Error()})
			}
			s.paymentRepo.IncrementRetries(payment.ID)
			continue
		}

		// Always increment retries on every check (like Ruby code)
		s.paymentRepo.IncrementRetries(payment.ID)

		// Check if max retries exceeded (like Ruby: retries > 69 -> failed)
		currentRetries := 0
		if payment.Retries != nil {
			currentRetries = *payment.Retries + 1 // +1 because we just incremented
		}
		if currentRetries > 69 && (status == nil || status.Status == "pending") {
			fmt.Printf("Payment %d exceeded max retries (%d), marking as failed\n", payment.ID, currentRetries)
			s.paymentRepo.UpdateStatus(payment.ID, "failed", "TIMEOUT", nil)
			continue
		}

		// Update payment based on Hubtel response
		if status != nil && status.Status != "pending" {
			s.paymentRepo.UpdateStatus(payment.ID, status.Status, status.BankStatus, &status.TransStatus)

			// If successful, handle based on payee type
			if status.Status == "successful" && payment.PayeeType != nil {
				if *payment.PayeeType == "EventRegistration" {
					// Legacy flow: update existing registration
					s.registrationRepo.UpdatePaymentStatus(uint(*payment.PayeeId), "paid", fmt.Sprintf("MOMO_%d", payment.ID))
					fmt.Printf("Payment %d marked as successful, registration updated\n", payment.ID)
				} else if *payment.PayeeType == "EventPayment" {
					// New flow: create registration after payment success
					if err := s.createEventRegistrationFromPayment(&payment); err != nil {
						fmt.Printf("Error creating event registration for payment %d: %v\n", payment.ID, err)
					} else {
						fmt.Printf("Payment %d marked as successful, registration created\n", payment.ID)
					}
				} else if *payment.PayeeType == "SchoolBillPayment" {
					// School bill payment: update school bill balance
					if err := s.processSchoolBillPaymentSuccess(&payment); err != nil {
						fmt.Printf("Error processing school bill payment %d: %v\n", payment.ID, err)
					} else {
						fmt.Printf("Payment %d marked as successful, school bill updated\n", payment.ID)
					}
				}

				// Create finance transaction for successful payment
				if err := s.createFinanceTransaction(&payment); err != nil {
					fmt.Printf("Error creating finance transaction for payment %d: %v\n", payment.ID, err)
				}

				// Update school bill (reduce balance, increase amount paid) for event payments
				if err := s.updateSchoolBillAfterPayment(&payment); err != nil {
					fmt.Printf("Error updating school bill for payment %d: %v\n", payment.ID, err)
				}
			} else if status.Status == "failed" {
				fmt.Printf("Payment %d marked as failed: %s\n", payment.ID, status.Message)
			}
		} else {
			fmt.Printf("Payment %d still pending (retry %d)\n", payment.ID, currentRetries)
		}
	}

	return nil
}

// HubtelStatusResponse represents the response from Hubtel status check API
type HubtelStatusResponse struct {
	ResponseCode string `json:"ResponseCode"`
	Message      string `json:"Message"`
	Data         struct {
		TransactionStatus string  `json:"Status"`
		Amount            float64 `json:"Amount"`
		ClientReference   string  `json:"ClientReference"`
		Description       string  `json:"Description"`
		ExternalReference string  `json:"ExternalTransactionId"`
	} `json:"Data"`
}

// queryHubtelTransactionStatus queries Hubtel API for transaction status
// Ruby uses POST with JSON body: Faraday.post(requestUrl, crd_params.to_json, headers)
func (s *MomoPaymentService) queryHubtelTransactionStatus(credentials *GatewayCredentials, transactionRef string) (*MomoStatusResponse, error) {
	// Build Hubtel status check URL (no query params - Ruby sends clientReference in body)
	statusURL := fmt.Sprintf("https://api-txnstatus.hubtel.com/transactions/%s/status", credentials.AccountID)

	// Build request body like Ruby: {clientReference: momo_transaction_id}
	requestBody := map[string]string{
		"clientReference": transactionRef,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	fmt.Printf("Hubtel status check URL: %s\n", statusURL)
	fmt.Printf("Hubtel status check Body: %s\n", string(jsonBody))

	// Make HTTP POST request (Ruby uses Faraday.post)
	httpReq, err := http.NewRequest("POST", statusURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("No-Cache", "true")
	httpReq.Header.Set("Authorization", "Basic "+credentials.SecretKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("hubtel request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log response for debugging
	fmt.Printf("Hubtel query status response (HTTP %d): %s\n", resp.StatusCode, string(body))

	// Check for non-2xx status codes - return raw body with error for saving
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &MomoStatusResponse{
			Status:  "pending",
			Message: fmt.Sprintf("Hubtel HTTP %d", resp.StatusCode),
			RawBody: body, // Include raw body even on error
		}, fmt.Errorf("hubtel returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var hubtelResp HubtelStatusResponse
	if err := json.Unmarshal(body, &hubtelResp); err != nil {
		return &MomoStatusResponse{
			Status:  "pending",
			Message: "Failed to parse response",
			RawBody: body, // Include raw body even on parse error
		}, fmt.Errorf("failed to parse response (raw: %s): %w", string(body), err)
	}

	// Map Hubtel status to our status
	var paymentStatus, bankStatus, message string
	switch hubtelResp.Data.TransactionStatus {
	case "Paid", "Success", "Successful":
		paymentStatus = "successful"
		bankStatus = "APPROVED"
		message = "Payment completed successfully"
	case "Failed", "Declined", "Cancelled", "Error", "Expired":
		paymentStatus = "failed"
		bankStatus = "DECLINED"
		message = "Payment was declined or cancelled"
	case "Pending", "":
		paymentStatus = "pending"
		bankStatus = "PENDING"
		message = "Payment is still being processed"
	default:
		paymentStatus = "pending"
		bankStatus = hubtelResp.Data.TransactionStatus
		message = hubtelResp.Message
	}

	return &MomoStatusResponse{
		Status:        paymentStatus,
		BankStatus:    bankStatus,
		TransStatus:   hubtelResp.Data.TransactionStatus,
		Message:       message,
		TransactionID: hubtelResp.Data.ExternalReference,
		RawBody:       body, // Include raw response body for saving
	}, nil
}
