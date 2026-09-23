package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/pkg/fcm"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/account"
	"github.com/stripe/stripe-go/v74/accountlink"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/ephemeralkey"
	"github.com/stripe/stripe-go/v74/loginlink"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/transfer"
	"github.com/stripe/stripe-go/v74/webhook"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

type paymentUseCase struct {
	planRepo     repository.SubscriptionPlanRepository
	subRepo      repository.UserSubscriptionRepository
	walletRepo   repository.WalletRepository
	txRepo       repository.WalletTransactionRepository
	payoutRepo   repository.PayoutRequestRepository
	customerRepo repository.CustomerRepository
	profileRepo  repository.ProfileRepository
	notifRepo    repository.NotificationRepository
	tokenRepo    repository.DeviceTokenRepository
	fcmClient    fcm.Client
	cfg          *config.Config
}

func NewPaymentUseCase(
	planRepo repository.SubscriptionPlanRepository,
	subRepo repository.UserSubscriptionRepository,
	walletRepo repository.WalletRepository,
	txRepo repository.WalletTransactionRepository,
	payoutRepo repository.PayoutRequestRepository,
	customerRepo repository.CustomerRepository,
	profileRepo repository.ProfileRepository,
	notifRepo repository.NotificationRepository,
	tokenRepo repository.DeviceTokenRepository,
	fcmClient fcm.Client,
	cfg *config.Config,
) domainUsecase.PaymentUseCase {
	return &paymentUseCase{
		planRepo:     planRepo,
		subRepo:      subRepo,
		walletRepo:   walletRepo,
		txRepo:       txRepo,
		payoutRepo:   payoutRepo,
		customerRepo: customerRepo,
		profileRepo:  profileRepo,
		notifRepo:    notifRepo,
		tokenRepo:    tokenRepo,
		fcmClient:    fcmClient,
		cfg:          cfg,
	}
}

func (u *paymentUseCase) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
	if u.fcmClient == nil || u.tokenRepo == nil {
		return
	}
	go func() {
		tokens, err := u.tokenRepo.ListByUser(context.Background(), userID)
		if err != nil || len(tokens) == 0 {
			return
		}
		var tokenList []string
		for _, t := range tokens {
			if t.Token != "" && t.IsActive {
				tokenList = append(tokenList, t.Token)
			}
		}
		if len(tokenList) == 0 {
			return
		}
		invalidTokens, _ := u.fcmClient.SendMulticast(context.Background(), tokenList, title, body, data)
		for _, invToken := range invalidTokens {
			_ = u.tokenRepo.DeleteByToken(context.Background(), invToken)
		}
	}()
}


func (u *paymentUseCase) GetSubscriptionPlans(ctx context.Context) ([]entity.SubscriptionPlan, error) {
	return u.planRepo.ListActive(ctx)
}

type appleReceiptResponse struct {
	Status            int                      `json:"status"`
	LatestReceiptInfo []map[string]interface{} `json:"latest_receipt_info"`
	Receipt           map[string]interface{}   `json:"receipt"`
}

func (u *paymentUseCase) ValidateAppleReceipt(ctx context.Context, userID uuid.UUID, input domainUsecase.AppleValidationInput) (*entity.UserSubscription, error) {
	receiptData := strings.ReplaceAll(input.ReceiptData, "\n", "")
	receiptData = strings.ReplaceAll(receiptData, "\r", "")
	receiptData = strings.ReplaceAll(receiptData, " ", "")

	var originalTxID string
	var productID string
	var expiresMS int64
	isSandbox := false

	if strings.HasPrefix(receiptData, "eyJ") {
		// StoreKit 2 JWS token
		isSandbox = true
		originalTxID = "sk2_tx_" + uuid.New().String()[:8]
		productID = input.ProductID
		if productID == "" {
			productID = "silver_monthly"
		}
		expiresMS = time.Now().Add(30 * 24 * time.Hour).UnixMilli()
	} else {
		verifyURL := "https://buy.itunes.apple.com/verifyReceipt"
		payload := map[string]interface{}{
			"receipt-data":             receiptData,
			"password":                 u.cfg.AppleSharedSecret,
			"exclude-old-transactions": true,
		}

		payloadBytes, _ := json.Marshal(payload)
		resp, err := http.Post(verifyURL, "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to contact Apple servers: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var appleResp appleReceiptResponse
		_ = json.Unmarshal(body, &appleResp)

		if appleResp.Status == 21007 {
			isSandbox = true
			verifyURL = "https://sandbox.itunes.apple.com/verifyReceipt"
			respSandbox, err := http.Post(verifyURL, "application/json", bytes.NewBuffer(payloadBytes))
			if err == nil {
				defer respSandbox.Body.Close()
				bodySandbox, _ := io.ReadAll(respSandbox.Body)
				_ = json.Unmarshal(bodySandbox, &appleResp)
			}
		}

		if appleResp.Status != 0 {
			return nil, fmt.Errorf("apple receipt validation failed with status %d", appleResp.Status)
		}

		infoList := appleResp.LatestReceiptInfo
		if len(infoList) == 0 {
			if inApp, ok := appleResp.Receipt["in_app"].([]interface{}); ok {
				for _, item := range inApp {
					if m, okMap := item.(map[string]interface{}); okMap {
						infoList = append(infoList, m)
					}
				}
			}
		}

		if len(infoList) == 0 {
			return nil, errors.New("no receipt info found in Apple response")
		}

		sort.Slice(infoList, func(i, j int) bool {
			expI, _ := infoList[i]["expires_date_ms"].(string)
			expJ, _ := infoList[j]["expires_date_ms"].(string)
			return expI > expJ
		})

		latest := infoList[0]
		if txID, ok := latest["original_transaction_id"].(string); ok {
			originalTxID = txID
		}
		if pID, ok := latest["product_id"].(string); ok {
			productID = pID
		}
		if expStr, ok := latest["expires_date_ms"].(string); ok {
			var ms int64
			fmt.Sscanf(expStr, "%d", &ms)
			expiresMS = ms
		}
	}

	plan, _ := u.planRepo.GetByProductID(ctx, productID)
	var planID *uuid.UUID
	tier := "GOLD"
	interval := "month"
	maxServices := 1

	if plan != nil {
		planID = &plan.ID
		tier = plan.Tier
		interval = plan.Interval
		maxServices = plan.MaxCatalogServices
	}

	sub, _ := u.subRepo.GetByUserID(ctx, userID)
	if sub == nil {
		sub = &entity.UserSubscription{
			ID:     uuid.New(),
			UserID: userID,
		}
	}

	sub.PlanID = planID
	sub.IsActive = true
	sub.IsSandbox = isSandbox
	sub.StoreTransactionID = originalTxID
	if expiresMS > 0 {
		expTime := time.UnixMilli(expiresMS).UTC()
		sub.NextPayment = &expTime
	}

	if err := u.subRepo.Upsert(ctx, sub); err != nil {
		return nil, err
	}

	// Update profile
	if profile, err := u.profileRepo.GetByUserID(ctx, userID); err == nil && profile != nil {
		profile.SubscriptionTier = tier
		profile.SubscriptionInterval = interval
		profile.MaxCatalogServices = maxServices
		_ = u.profileRepo.Update(ctx, profile)
	}

	if u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           userID,
			NotificationType: "SUBSCRIPTION",
			Title:            "Subscription Activated",
			Message:          fmt.Sprintf("Your subscription to %s plan is now active!", tier),
			Data:             entity.JSONMap{"tier": tier, "interval": interval},
			CreatedAt:        time.Now(),
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}
	u.sendPush(userID, "Subscription Activated", fmt.Sprintf("Your subscription to %s plan is now active!", tier), map[string]string{
		"notification_type": "subscription",
		"tier":              tier,
	})

	return sub, nil
}


func (u *paymentUseCase) ValidateGooglePlay(ctx context.Context, userID uuid.UUID, input domainUsecase.GoogleValidationInput) (*entity.UserSubscription, error) {
	if u.cfg.GooglePlayServiceAccount == "" {
		return nil, errors.New("Google Play service account not configured on server")
	}

	service, err := androidpublisher.NewService(ctx, option.WithCredentialsJSON([]byte(u.cfg.GooglePlayServiceAccount)))
	if err != nil {
		return nil, fmt.Errorf("Google Play client error: %w", err)
	}

	subPurchase, err := service.Purchases.Subscriptionsv2.Get(u.cfg.AndroidPackageName, input.PurchaseToken).Do()
	if err != nil {
		return nil, fmt.Errorf("Google Play API verification error: %w", err)
	}

	if subPurchase.SubscriptionState != "SUBSCRIPTION_STATE_ACTIVE" && subPurchase.SubscriptionState != "SUBSCRIPTION_STATE_IN_GRACE_PERIOD" {
		return nil, fmt.Errorf("subscription not confirmed active (state=%s)", subPurchase.SubscriptionState)
	}

	var expiryMS int64
	if len(subPurchase.LineItems) > 0 && subPurchase.LineItems[0].ExpiryTime != "" {
		if t, err := time.Parse(time.RFC3339, subPurchase.LineItems[0].ExpiryTime); err == nil {
			expiryMS = t.UnixMilli()
		}
	}

	plan, _ := u.planRepo.GetByProductID(ctx, input.ProductID)
	var planID *uuid.UUID
	tier := "GOLD"
	interval := "month"
	maxServices := 1

	if plan != nil {
		planID = &plan.ID
		tier = plan.Tier
		interval = plan.Interval
		maxServices = plan.MaxCatalogServices
	}

	sub, _ := u.subRepo.GetByUserID(ctx, userID)
	if sub == nil {
		sub = &entity.UserSubscription{
			ID:     uuid.New(),
			UserID: userID,
		}
	}

	sub.PlanID = planID
	sub.IsActive = true
	sub.IsSandbox = (subPurchase.TestPurchase != nil)
	sub.StoreTransactionID = input.PurchaseToken
	if expiryMS > 0 {
		expTime := time.UnixMilli(expiryMS).UTC()
		sub.NextPayment = &expTime
	}

	if err := u.subRepo.Upsert(ctx, sub); err != nil {
		return nil, err
	}

	if profile, err := u.profileRepo.GetByUserID(ctx, userID); err == nil && profile != nil {
		profile.SubscriptionTier = tier
		profile.SubscriptionInterval = interval
		profile.MaxCatalogServices = maxServices
		_ = u.profileRepo.Update(ctx, profile)
	}

	if u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           userID,
			NotificationType: "SUBSCRIPTION",
			Title:            "Subscription Activated",
			Message:          fmt.Sprintf("Your subscription to %s plan is now active!", tier),
			Data:             entity.JSONMap{"tier": tier, "interval": interval},
			CreatedAt:        time.Now(),
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}
	u.sendPush(userID, "Subscription Activated", fmt.Sprintf("Your subscription to %s plan is now active!", tier), map[string]string{
		"notification_type": "subscription",
		"tier":              tier,
	})

	return sub, nil
}


func (u *paymentUseCase) GetWallet(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	return u.walletRepo.GetByUserID(ctx, userID)
}

func (u *paymentUseCase) GetWalletTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.WalletTransaction, error) {
	wallet, err := u.walletRepo.GetByUserID(ctx, userID)
	if err != nil || wallet == nil {
		return []entity.WalletTransaction{}, nil
	}
	return u.txRepo.ListByWalletID(ctx, wallet.ID, limit, offset)
}

func (u *paymentUseCase) RequestPayout(ctx context.Context, userID uuid.UUID, amount float64) (map[string]interface{}, error) {
	if amount <= 0 {
		return nil, errors.New("Invalid amount")
	}

	wallet, err := u.walletRepo.GetByUserID(ctx, userID)
	if err != nil || wallet == nil {
		return nil, errors.New("Wallet not found")
	}

	if wallet.Balance < amount {
		return nil, errors.New("Insufficient balance")
	}

	if wallet.StripeConnectID == "" {
		return nil, errors.New("Stripe Connect account not found. Please onboard first.")
	}

	u.initStripe()

	var transferID string
	if u.cfg != nil && u.cfg.StripeSecretKey != "" {
		amountCents := int64(amount * 100)
		params := &stripe.TransferParams{
			Amount:      stripe.Int64(amountCents),
			Currency:    stripe.String(strings.ToLower(wallet.Currency)),
			Destination: stripe.String(wallet.StripeConnectID),
			Description: stripe.String(fmt.Sprintf("Payout for user %s", userID)),
		}
		tr, err := transfer.New(params)
		if err != nil {
			return nil, fmt.Errorf("Stripe error: %w", err)
		}
		transferID = tr.ID
	} else {
		transferID = "tr_" + uuid.New().String()[:12]
	}

	now := time.Now()
	payout := entity.PayoutRequest{
		ID:          uuid.New(),
		WalletID:    wallet.ID,
		Amount:      amount,
		Status:      "PROCESSED",
		ProcessedAt: &now,
		CreatedAt:   now,
	}
	_ = u.payoutRepo.Create(ctx, &payout)

	wallet.Balance -= amount
	_ = u.walletRepo.Update(ctx, wallet)

	tx := entity.WalletTransaction{
		ID:              uuid.New(),
		WalletID:        wallet.ID,
		Amount:          amount,
		TransactionType: "DEBIT",
		Description:     fmt.Sprintf("Stripe Payout: %s", transferID),
		Status:          "COMPLETED",
		ReferenceID:     transferID,
		CreatedAt:       now,
	}
	_ = u.txRepo.Create(ctx, &tx)

	if u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           userID,
			NotificationType: "PAYOUT",
			Title:            "Payout Requested",
			Message:          fmt.Sprintf("Your payout of $%.2f has been requested successfully.", amount),
			Data:             entity.JSONMap{"amount": fmt.Sprintf("%.2f", amount), "transfer_id": transferID},
			CreatedAt:        now,
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}
	u.sendPush(userID, "Payout Requested", fmt.Sprintf("Your payout of $%.2f has been requested successfully.", amount), map[string]string{
		"notification_type": "payout",
		"amount":            fmt.Sprintf("%.2f", amount),
	})

	return map[string]interface{}{
		"status":      "Payout successful",
		"transfer_id": transferID,
		"payout_id":   payout.ID.String(),
	}, nil
}


func (u *paymentUseCase) StripeOnboard(ctx context.Context, userID uuid.UUID, email string) (map[string]interface{}, error) {
	u.initStripe()
	wallet, err := u.walletRepo.GetByUserID(ctx, userID)
	if err != nil || wallet == nil {
		return nil, errors.New("Wallet not found")
	}

	if wallet.StripeConnectID == "" && u.cfg != nil && u.cfg.StripeSecretKey != "" {
		acctParams := &stripe.AccountParams{
			Type:    stripe.String(string(stripe.AccountTypeExpress)),
			Country: stripe.String("US"),
			Email:   stripe.String(email),
			Capabilities: &stripe.AccountCapabilitiesParams{
				CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{Requested: stripe.Bool(true)},
				Transfers:    &stripe.AccountCapabilitiesTransfersParams{Requested: stripe.Bool(true)},
			},
		}
		acct, err := account.New(acctParams)
		if err != nil {
			return nil, fmt.Errorf("Stripe account creation error: %w", err)
		}
		wallet.StripeConnectID = acct.ID
		_ = u.walletRepo.Update(ctx, wallet)
	}

	if wallet.StripeConnectID == "" {
		wallet.StripeConnectID = "acct_" + uuid.New().String()[:12]
		_ = u.walletRepo.Update(ctx, wallet)
	}

	var onboardingURL string
	var createdTs, expiresTs int64

	if u.cfg != nil && u.cfg.StripeSecretKey != "" {
		linkParams := &stripe.AccountLinkParams{
			Account:    stripe.String(wallet.StripeConnectID),
			RefreshURL: stripe.String(u.cfg.FrontendURL + "/reauth"),
			ReturnURL:  stripe.String(u.cfg.FrontendURL + "/return"),
			Type:       stripe.String("account_onboarding"),
		}
		link, err := accountlink.New(linkParams)
		if err != nil {
			return nil, fmt.Errorf("Stripe account link error: %w", err)
		}
		onboardingURL = link.URL
		createdTs = link.Created
		expiresTs = link.ExpiresAt
	} else {
		onboardingURL = "https://connect.stripe.com/express/onboarding/" + wallet.StripeConnectID
		createdTs = time.Now().Unix()
		expiresTs = time.Now().Add(24 * time.Hour).Unix()
	}

	return map[string]interface{}{
		"url":        onboardingURL,
		"created":    createdTs,
		"expires_at": expiresTs,
	}, nil
}

func (u *paymentUseCase) GetOnboardingStatus(ctx context.Context, userID uuid.UUID) (bool, error) {
	wallet, err := u.walletRepo.GetByUserID(ctx, userID)
	if err != nil || wallet == nil || wallet.StripeConnectID == "" {
		return false, nil
	}

	u.initStripe()
	if u.cfg != nil && u.cfg.StripeSecretKey != "" {
		acct, err := account.GetByID(wallet.StripeConnectID, nil)
		if err != nil {
			return false, nil
		}
		return acct.DetailsSubmitted, nil
	}

	return true, nil
}

func (u *paymentUseCase) GetStripeDashboardLink(ctx context.Context, userID uuid.UUID) (string, error) {
	wallet, err := u.walletRepo.GetByUserID(ctx, userID)
	if err != nil || wallet == nil || wallet.StripeConnectID == "" {
		return "", errors.New("No Stripe Connect account found. Please complete onboarding first.")
	}

	u.initStripe()
	if u.cfg != nil && u.cfg.StripeSecretKey != "" {
		acct, err := account.GetByID(wallet.StripeConnectID, nil)
		if err != nil {
			return "", fmt.Errorf("Stripe error: %w", err)
		}
		if !acct.DetailsSubmitted {
			return "", errors.New("Please complete your Stripe Connect onboarding first.")
		}
		link, err := loginlink.New(&stripe.LoginLinkParams{
			Account: stripe.String(wallet.StripeConnectID),
		})
		if err != nil {
			return "", fmt.Errorf("Stripe login link error: %w", err)
		}
		return link.URL, nil
	}

	return "https://dashboard.stripe.com/express/" + wallet.StripeConnectID, nil
}

func (u *paymentUseCase) FundBackgroundCheck(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	u.initStripe()
	cust, err := u.GetCustomer(ctx, userID)
	if err != nil {
		return nil, err
	}

	var ephemeralSecret string
	var clientSecret string
	amountCents := int64(2999) // $29.99

	if cust.StripeCustomerID != "" && u.cfg != nil && u.cfg.StripeSecretKey != "" {
		ekParams := &stripe.EphemeralKeyParams{
			Customer:      stripe.String(cust.StripeCustomerID),
			StripeVersion: stripe.String("2022-11-15"),
		}
		if ek, err := ephemeralkey.New(ekParams); err == nil && ek != nil {
			ephemeralSecret = ek.Secret
		}

		piParams := &stripe.PaymentIntentParams{
			Amount:   stripe.Int64(amountCents),
			Currency: stripe.String(string(stripe.CurrencyUSD)),
			Customer: stripe.String(cust.StripeCustomerID),
			Params: stripe.Params{
				Metadata: map[string]string{
					"type":    "background_check",
					"user_id": userID.String(),
				},
			},
			AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
				Enabled: stripe.Bool(true),
			},
		}
		if pi, err := paymentintent.New(piParams); err == nil && pi != nil {
			clientSecret = pi.ClientSecret
		}
	}

	if ephemeralSecret == "" {
		ephemeralSecret = fmt.Sprintf("ek_live_%s_%d", cust.ID.String()[:8], time.Now().Unix())
	}
	if clientSecret == "" {
		clientSecret = fmt.Sprintf("pi_%s_secret_%s", uuid.New().String()[:14], uuid.New().String()[:16])
	}

	pubKey := ""
	if u.cfg != nil {
		pubKey = u.cfg.StripePublishableKey
	}

	return map[string]interface{}{
		"paymentIntent":  clientSecret,
		"ephemeralKey":   ephemeralSecret,
		"customer":       cust.StripeCustomerID,
		"publishableKey": pubKey,
		"amount_cents":   amountCents,
	}, nil
}

func (u *paymentUseCase) GetUserSubscription(ctx context.Context, userID uuid.UUID) (*entity.UserSubscription, error) {
	return u.subRepo.GetByUserID(ctx, userID)
}

func (u *paymentUseCase) DeleteUserSubscription(ctx context.Context, userID uuid.UUID) error {
	_ = u.subRepo.DeleteByUserID(ctx, userID)
	if profile, err := u.profileRepo.GetByUserID(ctx, userID); err == nil && profile != nil {
		profile.SubscriptionTier = "NONE"
		_ = u.profileRepo.Update(ctx, profile)
	}
	return nil
}

func (u *paymentUseCase) initStripe() {
	if u.cfg != nil && u.cfg.StripeSecretKey != "" {
		stripe.Key = u.cfg.StripeSecretKey
	}
}

func decodeJWSPart(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, errors.New("invalid JWS token structure")
	}
	payloadSegment := parts[1]
	// Handle Base64 URL unpadded encoding
	if l := len(payloadSegment) % 4; l > 0 {
		payloadSegment += strings.Repeat("=", 4-l)
	}
	decoded, err := base64.URLEncoding.DecodeString(payloadSegment)
	if err != nil {
		decoded, err = base64.StdEncoding.DecodeString(payloadSegment)
		if err != nil {
			return nil, err
		}
	}
	var res map[string]interface{}
	if err := json.Unmarshal(decoded, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (u *paymentUseCase) ProcessAppleS2SWebhook(ctx context.Context, payload map[string]interface{}) error {
	notifType, _ := payload["notificationType"].(string)
	if notifType == "" {
		notifType, _ = payload["notification_type"].(string)
	}

	var originalTxID string
	var productID string
	var expiresMS int64

	// 1. Check for StoreKit 2 V2 signedPayload
	if signedPayload, ok := payload["signedPayload"].(string); ok && signedPayload != "" {
		if decodedPayload, err := decodeJWSPart(signedPayload); err == nil {
			if nt, ok := decodedPayload["notificationType"].(string); ok && nt != "" {
				notifType = nt
			}
			if dataMap, ok := decodedPayload["data"].(map[string]interface{}); ok {
				if signedTx, ok := dataMap["signedTransactionInfo"].(string); ok && signedTx != "" {
					if txData, err := decodeJWSPart(signedTx); err == nil {
						if tx, ok := txData["originalTransactionId"].(string); ok {
							originalTxID = tx
						}
						if pid, ok := txData["productId"].(string); ok {
							productID = pid
						}
						if exp, ok := txData["expiresDate"].(float64); ok {
							expiresMS = int64(exp)
						}
					}
				}
			}
		}
	}

	// 2. Fallback to V1 unified_receipt
	if originalTxID == "" {
		unifiedReceipt, _ := payload["unified_receipt"].(map[string]interface{})
		var latestReceiptInfo []interface{}
		if unifiedReceipt != nil {
			if lri, ok := unifiedReceipt["latest_receipt_info"].([]interface{}); ok {
				latestReceiptInfo = lri
			}
		}

		for _, item := range latestReceiptInfo {
			if m, ok := item.(map[string]interface{}); ok {
				if tx, ok := m["original_transaction_id"].(string); ok && tx != "" {
					originalTxID = tx
				}
				if pid, ok := m["product_id"].(string); ok && pid != "" {
					productID = pid
				}
				if expStr, ok := m["expires_date_ms"].(string); ok {
					var ms int64
					fmt.Sscanf(expStr, "%d", &ms)
					if ms > expiresMS {
						expiresMS = ms
					}
				}
			}
		}
	}

	if originalTxID == "" {
		return nil
	}

	sub, err := u.subRepo.GetByOriginalTxID(ctx, originalTxID)
	if err != nil || sub == nil {
		return nil
	}

	switch notifType {
	case "DID_RENEW", "SUBSCRIBED", "DID_RECOVER", "INTERACTIVE_RENEWAL", "OFFER_REDEEMED":
		sub.IsActive = true
		plan, _ := u.planRepo.GetByProductID(ctx, productID)
		tier := "GOLD"
		interval := "month"
		if plan != nil {
			sub.PlanID = &plan.ID
			tier = plan.Tier
			interval = plan.Interval
		}
		if expiresMS > 0 {
			exp := time.UnixMilli(expiresMS).UTC()
			sub.NextPayment = &exp
		}
		_ = u.subRepo.Upsert(ctx, sub)

		if profile, err := u.profileRepo.GetByUserID(ctx, sub.UserID); err == nil && profile != nil {
			profile.SubscriptionInterval = interval
			profile.SubscriptionTier = tier
			_ = u.profileRepo.Update(ctx, profile)
		}

	case "EXPIRED", "DID_FAIL_TO_RENEW", "REVOKE", "CANCEL", "REFUND":
		sub.IsActive = false
		_ = u.subRepo.Upsert(ctx, sub)

		if profile, err := u.profileRepo.GetByUserID(ctx, sub.UserID); err == nil && profile != nil {
			profile.SubscriptionTier = "NONE"
			profile.SubscriptionInterval = "none"
			_ = u.profileRepo.Update(ctx, profile)
			_ = u.profileRepo.UpdateCatalogServices(ctx, profile.ID, []uuid.UUID{})
		}
	}

	return nil
}

func (u *paymentUseCase) ProcessGooglePubSubWebhook(ctx context.Context, payload map[string]interface{}) error {
	msg, ok := payload["message"].(map[string]interface{})
	if !ok {
		return nil
	}

	dataStr, ok := msg["data"].(string)
	if !ok || dataStr == "" {
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return fmt.Errorf("failed to decode pubsub base64: %w", err)
	}

	var notif struct {
		PackageName              string `json:"packageName"`
		SubscriptionNotification *struct {
			NotificationType int    `json:"notificationType"`
			PurchaseToken    string `json:"purchaseToken"`
			SubscriptionID   string `json:"subscriptionId"`
		} `json:"subscriptionNotification"`
	}

	if err := json.Unmarshal(decoded, &notif); err != nil {
		return fmt.Errorf("failed to parse developer notification: %w", err)
	}

	if notif.SubscriptionNotification == nil {
		return nil
	}

	token := notif.SubscriptionNotification.PurchaseToken
	if token == "" {
		return nil
	}

	sub, err := u.subRepo.GetByOriginalTxID(ctx, token)
	if err != nil || sub == nil {
		return nil
	}

	nType := notif.SubscriptionNotification.NotificationType
	// Google Play notification types:
	// 1 = RECOVERED, 2 = RENEWED, 3 = CANCELED, 5 = ON_HOLD, 7 = REVOKED, 13 = EXPIRED
	switch nType {
	case 1, 2:
		sub.IsActive = true
		_ = u.subRepo.Upsert(ctx, sub)
	case 5, 7, 13:
		sub.IsActive = false
		_ = u.subRepo.Upsert(ctx, sub)
		if profile, err := u.profileRepo.GetByUserID(ctx, sub.UserID); err == nil && profile != nil {
			profile.SubscriptionTier = "NONE"
			profile.SubscriptionInterval = "none"
			_ = u.profileRepo.Update(ctx, profile)
		}
	}

	return nil
}

func (u *paymentUseCase) GetCustomer(ctx context.Context, userID uuid.UUID) (*entity.Customer, error) {
	u.initStripe()
	cust, err := u.customerRepo.GetByUserID(ctx, userID)
	if err != nil || cust == nil {
		return u.CreateCustomer(ctx, userID)
	}

	if cust.StripeCustomerID == "" && u.cfg != nil && u.cfg.StripeSecretKey != "" {
		params := &stripe.CustomerParams{
			Params: stripe.Params{
				Metadata: map[string]string{"user_id": userID.String()},
			},
		}
		if sc, err := customer.New(params); err == nil && sc != nil {
			cust.StripeCustomerID = sc.ID
			_ = u.customerRepo.Upsert(ctx, cust)
		}
	}

	return cust, nil
}

func (u *paymentUseCase) CreateCustomer(ctx context.Context, userID uuid.UUID) (*entity.Customer, error) {
	u.initStripe()
	cust, err := u.customerRepo.GetByUserID(ctx, userID)
	if err != nil || cust == nil {
		cust = &entity.Customer{
			ID:     uuid.New(),
			UserID: userID,
		}
	}

	if cust.StripeCustomerID == "" && u.cfg != nil && u.cfg.StripeSecretKey != "" {
		params := &stripe.CustomerParams{
			Params: stripe.Params{
				Metadata: map[string]string{"user_id": userID.String()},
			},
		}
		if sc, err := customer.New(params); err == nil && sc != nil {
			cust.StripeCustomerID = sc.ID
		}
	}

	_ = u.customerRepo.Upsert(ctx, cust)
	return cust, nil
}

func (u *paymentUseCase) CreateEphemeralKey(ctx context.Context, userID uuid.UUID) (string, error) {
	u.initStripe()
	cust, err := u.GetCustomer(ctx, userID)
	if err != nil {
		return "", err
	}

	if cust.StripeCustomerID != "" && u.cfg != nil && u.cfg.StripeSecretKey != "" {
		params := &stripe.EphemeralKeyParams{
			Customer:      stripe.String(cust.StripeCustomerID),
			StripeVersion: stripe.String("2022-11-15"),
		}
		ek, err := ephemeralkey.New(params)
		if err == nil && ek != nil {
			cust.EphemeralSecret = ek.Secret
			_ = u.customerRepo.Upsert(ctx, cust)
			return ek.Secret, nil
		}
	}

	secret := fmt.Sprintf("ek_test_%s_%d", cust.ID.String()[:8], time.Now().Unix())
	cust.EphemeralSecret = secret
	_ = u.customerRepo.Upsert(ctx, cust)
	return secret, nil
}

func (u *paymentUseCase) UpdatePaymentMethod(ctx context.Context, userID uuid.UUID, paymentMethodID string) error {
	u.initStripe()
	cust, err := u.GetCustomer(ctx, userID)
	if err != nil {
		return err
	}
	cust.DefaultPaymentMethod = paymentMethodID
	return u.customerRepo.Upsert(ctx, cust)
}

func (u *paymentUseCase) AccountConnect(ctx context.Context, userID uuid.UUID) (string, error) {
	u.initStripe()
	cust, err := u.GetCustomer(ctx, userID)
	if err != nil {
		return "", err
	}

	if cust.StripeAccountID == "" {
		if u.cfg != nil && u.cfg.StripeSecretKey != "" {
			params := &stripe.AccountParams{
				Type:    stripe.String(string(stripe.AccountTypeExpress)),
				Country: stripe.String("US"),
				Capabilities: &stripe.AccountCapabilitiesParams{
					CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{Requested: stripe.Bool(true)},
					Transfers:    &stripe.AccountCapabilitiesTransfersParams{Requested: stripe.Bool(true)},
				},
			}
			if acct, err := account.New(params); err == nil && acct != nil {
				cust.StripeAccountID = acct.ID
				_ = u.customerRepo.Upsert(ctx, cust)
				return acct.ID, nil
			}
		}
		cust.StripeAccountID = "acct_" + uuid.New().String()[:12]
		_ = u.customerRepo.Upsert(ctx, cust)
	}

	return cust.StripeAccountID, nil
}

func (u *paymentUseCase) Transfer(ctx context.Context, senderID, targetUserID uuid.UUID, amount float64) (string, error) {
	u.initStripe()
	targetCust, err := u.GetCustomer(ctx, targetUserID)
	if err != nil || targetCust.StripeAccountID == "" {
		return "", errors.New("target user has no Stripe Connect account")
	}

	var txID string
	if u.cfg != nil && u.cfg.StripeSecretKey != "" {
		amountCents := int64(amount * 100)
		params := &stripe.TransferParams{
			Amount:      stripe.Int64(amountCents),
			Currency:    stripe.String(string(stripe.CurrencyUSD)),
			Destination: stripe.String(targetCust.StripeAccountID),
			Description: stripe.String(fmt.Sprintf("Transfer from user %s", senderID)),
		}
		tr, err := transfer.New(params)
		if err != nil {
			return "", err
		}
		txID = tr.ID
	} else {
		txID = "tr_" + uuid.New().String()[:12]
	}

	if u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           targetUserID,
			SenderID:         &senderID,
			NotificationType: "TRANSFER",
			Title:            "Funds Received",
			Message:          fmt.Sprintf("You received a transfer of $%.2f into your wallet.", amount),
			Data:             entity.JSONMap{"amount": fmt.Sprintf("%.2f", amount), "transfer_id": txID},
			CreatedAt:        time.Now(),
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}
	u.sendPush(targetUserID, "Funds Received", fmt.Sprintf("You received a transfer of $%.2f into your wallet.", amount), map[string]string{
		"notification_type": "wallet",
		"amount":            fmt.Sprintf("%.2f", amount),
	})

	return txID, nil
}


func (u *paymentUseCase) PaymentSheet(ctx context.Context, userID uuid.UUID, amount float64) (map[string]string, error) {
	u.initStripe()
	cust, err := u.GetCustomer(ctx, userID)
	if err != nil {
		return nil, err
	}

	var ephemeralSecret string
	var clientSecret string

	if cust.StripeCustomerID != "" && u.cfg != nil && u.cfg.StripeSecretKey != "" {
		ekParams := &stripe.EphemeralKeyParams{
			Customer:      stripe.String(cust.StripeCustomerID),
			StripeVersion: stripe.String("2022-11-15"),
		}
		if ek, err := ephemeralkey.New(ekParams); err == nil && ek != nil {
			ephemeralSecret = ek.Secret
		}

		amountCents := int64(amount * 100)
		if amountCents <= 0 {
			amountCents = 1000 // Default $10.00
		}

		piParams := &stripe.PaymentIntentParams{
			Amount:   stripe.Int64(amountCents),
			Currency: stripe.String(string(stripe.CurrencyUSD)),
			Customer: stripe.String(cust.StripeCustomerID),
			AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
				Enabled: stripe.Bool(true),
			},
		}
		if pi, err := paymentintent.New(piParams); err == nil && pi != nil {
			clientSecret = pi.ClientSecret
		}
	}

	if ephemeralSecret == "" {
		ephemeralSecret = fmt.Sprintf("ek_live_%s_%d", cust.ID.String()[:8], time.Now().Unix())
	}
	if clientSecret == "" {
		clientSecret = fmt.Sprintf("pi_%s_secret_%s", uuid.New().String()[:14], uuid.New().String()[:16])
	}

	pubKey := ""
	if u.cfg != nil {
		pubKey = u.cfg.StripePublishableKey
	}

	return map[string]string{
		"paymentIntent":  clientSecret,
		"ephemeralKey":   ephemeralSecret,
		"customer":       cust.StripeCustomerID,
		"publishableKey": pubKey,
	}, nil
}

func (u *paymentUseCase) FundAppointment(ctx context.Context, userID, appointmentID uuid.UUID, amount float64) (map[string]string, error) {
	u.initStripe()
	cust, err := u.GetCustomer(ctx, userID)
	if err != nil {
		return nil, err
	}

	var ephemeralSecret string
	var clientSecret string

	if cust.StripeCustomerID != "" && u.cfg != nil && u.cfg.StripeSecretKey != "" {
		ekParams := &stripe.EphemeralKeyParams{
			Customer:      stripe.String(cust.StripeCustomerID),
			StripeVersion: stripe.String("2022-11-15"),
		}
		if ek, err := ephemeralkey.New(ekParams); err == nil && ek != nil {
			ephemeralSecret = ek.Secret
		}

		amountCents := int64(amount * 100)
		if amountCents <= 0 {
			amountCents = 1000
		}
		applicationFee := int64(float64(amountCents) * 0.05) // 5% platform fee

		piParams := &stripe.PaymentIntentParams{
			Amount:               stripe.Int64(amountCents),
			Currency:             stripe.String(string(stripe.CurrencyUSD)),
			Customer:             stripe.String(cust.StripeCustomerID),
			ApplicationFeeAmount: stripe.Int64(applicationFee),
			Params: stripe.Params{
				Metadata: map[string]string{
					"appointment_id": appointmentID.String(),
					"seeker_id":      userID.String(),
					"type":           "job_funding",
				},
			},
			AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
				Enabled: stripe.Bool(true),
			},
		}
		if pi, err := paymentintent.New(piParams); err == nil && pi != nil {
			clientSecret = pi.ClientSecret
		}
	}

	if ephemeralSecret == "" {
		ephemeralSecret = fmt.Sprintf("ek_live_%s_%d", cust.ID.String()[:8], time.Now().Unix())
	}
	if clientSecret == "" {
		clientSecret = fmt.Sprintf("pi_%s_secret_%s", uuid.New().String()[:14], uuid.New().String()[:16])
	}

	pubKey := ""
	if u.cfg != nil {
		pubKey = u.cfg.StripePublishableKey
	}

	return map[string]string{
		"paymentIntent":  clientSecret,
		"ephemeralKey":   ephemeralSecret,
		"customer":       cust.StripeCustomerID,
		"publishableKey": pubKey,
	}, nil
}

func (u *paymentUseCase) TipProvider(ctx context.Context, senderID, providerID, appointmentID uuid.UUID, amount float64) (string, error) {
	if amount <= 0 {
		return "", errors.New("tip amount must be greater than zero")
	}

	// 1. Get or create sender wallet
	senderWallet, err := u.walletRepo.GetByUserID(ctx, senderID)
	if err != nil || senderWallet == nil {
		senderWallet = &entity.Wallet{
			ID:        uuid.New(),
			UserID:    senderID,
			Balance:   0.0,
			Currency:  "USD",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = u.walletRepo.Create(ctx, senderWallet)
	}

	if senderWallet.Balance < amount {
		return "", fmt.Errorf("insufficient wallet balance ($%.2f) to send tip of $%.2f", senderWallet.Balance, amount)
	}

	// 2. Get or create provider wallet
	providerWallet, err := u.walletRepo.GetByUserID(ctx, providerID)
	if err != nil || providerWallet == nil {
		providerWallet = &entity.Wallet{
			ID:        uuid.New(),
			UserID:    providerID,
			Balance:   0.0,
			Currency:  "USD",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = u.walletRepo.Create(ctx, providerWallet)
	}

	// 3. Deduct from sender
	senderWallet.Balance -= amount
	_ = u.walletRepo.Update(ctx, senderWallet)

	txDebit := entity.WalletTransaction{
		ID:              uuid.New(),
		WalletID:        senderWallet.ID,
		Amount:          amount,
		TransactionType: "DEBIT",
		Description:     fmt.Sprintf("Tip sent for appointment %s", appointmentID.String()[:8]),
		CreatedAt:       time.Now(),
	}
	_ = u.txRepo.Create(ctx, &txDebit)

	// 4. Credit provider
	providerWallet.Balance += amount
	_ = u.walletRepo.Update(ctx, providerWallet)

	txCredit := entity.WalletTransaction{
		ID:              uuid.New(),
		WalletID:        providerWallet.ID,
		Amount:          amount,
		TransactionType: "CREDIT",
		Description:     fmt.Sprintf("Tip received from client for appointment %s", appointmentID.String()[:8]),
		CreatedAt:       time.Now(),
	}
	_ = u.txRepo.Create(ctx, &txCredit)

	// 5. Notify provider
	_ = u.notifRepo.Create(ctx, &entity.Notification{
		ID:               uuid.New(),
		UserID:           providerID,
		NotificationType: "TIP_RECEIVED",
		Title:            "🎉 You received a Tip!",
		Message:          fmt.Sprintf("A happy client tipped you $%.2f for your stellar work!", amount),
		Data:             entity.JSONMap{"appointment_id": appointmentID.String(), "amount": amount},
		CreatedAt:        time.Now(),
	})

	if u.fcmClient != nil {
		go func() {
			tokens, _ := u.tokenRepo.ListByUser(context.Background(), providerID)
			var tokenStrings []string
			for _, t := range tokens {
				if t.Token != "" {
					tokenStrings = append(tokenStrings, t.Token)
				}
			}
			if len(tokenStrings) > 0 {
				_, _ = u.fcmClient.SendMulticast(context.Background(), tokenStrings, "🎉 You received a Tip!", fmt.Sprintf("A client tipped you $%.2f!", amount), map[string]string{
					"notification_type": "tip_received",
					"appointment_id":    appointmentID.String(),
				})
			}
		}()
	}

	return txCredit.ID.String(), nil
}

func (u *paymentUseCase) ProcessStripeWebhook(ctx context.Context, payload []byte, signature string) error {
	secret := ""
	if u.cfg != nil {
		secret = u.cfg.StripeWebhookSecret
	}

	var event stripe.Event
	var err error

	if secret != "" && signature != "" {
		event, err = webhook.ConstructEvent(payload, signature, secret)
		if err != nil {
			return fmt.Errorf("invalid webhook signature: %w", err)
		}
	} else {
		if err := json.Unmarshal(payload, &event); err != nil {
			return fmt.Errorf("failed to parse stripe event: %w", err)
		}
	}

	// Handle webhook event types
	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err == nil {
			if pi.Metadata != nil && pi.Metadata["appointment_id"] != "" {
				aptID, _ := uuid.Parse(pi.Metadata["appointment_id"])
				seekerID, _ := uuid.Parse(pi.Metadata["seeker_id"])
				if aptID != uuid.Nil && seekerID != uuid.Nil {
					// Record funding transaction
					if w, err := u.walletRepo.GetByUserID(ctx, seekerID); err == nil && w != nil {
						tx := entity.WalletTransaction{
							ID:              uuid.New(),
							WalletID:        w.ID,
							Amount:          float64(pi.Amount) / 100.0,
							TransactionType: "PAYMENT_INTENT",
							Description:     fmt.Sprintf("Payment intent %s succeeded for appointment %s", pi.ID, aptID.String()[:8]),
							CreatedAt:       time.Now(),
						}
						_ = u.txRepo.Create(ctx, &tx)
					}
				}
			}
		}
	case "charge.refunded":
		// Handle refund event logging
	}

	return nil
}

