package integration

import (
	"crypto/sha256"
	"factors/model/model"
	"factors/model/store"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type LineItem struct {
	SKU string `json:"sku"`
}

type NoteAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CheckoutObject struct {
	ID                  float64         `json:"id"`
	Token               string          `json:"token"`
	CartToken           string          `json:"cart_token"`
	Email               string          `json:"email"`
	UserID              float64         `json:"user_id"`
	Gateway             string          `json:"gateway"`
	CreatedAt           string          `json:"created_at"`
	UpdatedAt           string          `json:"updated_at"`
	Currency            string          `json:"currency"`
	PresentmentCurrency string          `json:"presentment_currency"`
	TotalDiscounts      string          `json:"total_discounts"`
	TotalLineItemsPrice string          `json:"total_line_items_price"`
	TotalPrice          string          `json:"total_price"`
	SubtotalPrice       string          `json:"subtotal_price"`
	LineItems           []LineItem      `json:"line_items"`
	NoteAttributes      []NoteAttribute `json:"note_attributes"`
	Customer            CustomerObject  `json:"customer"`
}

type CustomerObject struct {
	ID    float64 `json:"id"`
	Email string  `json:"email"`
}

const ACTION_SHOPIFY_CHECKOUT_CREATED = 1
const ACTION_SHOPIFY_CHECKOUT_UPDATED = 2
const ACTION_SHOPIFY_ORDER_CREATED = 3
const ACTION_SHOPIFY_ORDER_UPDATED = 4
const ACTION_SHOPIFY_ORDER_PAID = 5
const ACTION_SHOPIFY_ORDER_CANCELLED = 6
const ACTION_SHOPIFY_CART_UPDATED = 7

const FACTORS_CUSTOMER_USER_ID_TYPE = "factorsUserId"

func factorsUserIdFromAttributes(attributes []NoteAttribute) string {
	for _, attr := range attributes {
		if attr.Name == "factors-user-id" {
			return attr.Value
		}
	}
	return ""
}

// Returns eventName, customerEventId, userId, isNewUser, eventProperties, userProperties, timestamp, err
func GetTrackDetailsFromCheckoutObject(
	projectId uint64, actionType int64, shouldHashEmail bool, checkoutObject *CheckoutObject) (
	string, string, bool, U.PropertiesMap, U.PropertiesMap, int64, error) {
	var eventTime time.Time
	var eventName string
	var err error
	if actionType == ACTION_SHOPIFY_CHECKOUT_CREATED {
		eventName = U.EVENT_NAME_SHOPIFY_CHECKOUT_CREATED
		eventTime, err = time.Parse(time.RFC3339, checkoutObject.CreatedAt)
		if err != nil {
			return "", "", false, nil, nil, 0, fmt.Errorf(
				fmt.Sprintf("Failed to parse time %s", checkoutObject.CreatedAt))
		}
	} else if actionType == ACTION_SHOPIFY_CHECKOUT_UPDATED {
		eventName = U.EVENT_NAME_SHOPIFY_CHECKOUT_UPDATED
		eventTime, err = time.Parse(time.RFC3339, checkoutObject.UpdatedAt)
		if err != nil {
			return "", "", false, nil, nil, 0, fmt.Errorf(
				fmt.Sprintf("Failed to parse time %s", checkoutObject.UpdatedAt))
		}
	}
	eventTimestamp := eventTime.Unix()

	userId := factorsUserIdFromAttributes(checkoutObject.NoteAttributes)
	custUserIdType := ""
	// In case of identified users.
	custUserId := ""
	isNewUser := false

	if userId == "" {
		if checkoutObject.Email != "" {
			custUserId = checkoutObject.Email
			custUserIdType = "checkoutEmail"
		} else if checkoutObject.Customer.Email != "" {
			custUserId = checkoutObject.Customer.Email
			custUserIdType = "checkoutCustomerEmail"
		} else if checkoutObject.UserID > 0 {
			custUserId = fmt.Sprintf("%f", checkoutObject.UserID)
			custUserIdType = "checkoutUserId"
		} else if checkoutObject.ID > 0 {
			custUserId = fmt.Sprintf("%f", checkoutObject.Customer.ID)
			custUserIdType = "checkoutCustomerId"
		}

		if custUserId == "" {
			return "", "", false, nil, nil, 0, fmt.Errorf("Missing email in CheckoutObject")
		}
		if shouldHashEmail {
			h := sha256.New()
			h.Write([]byte(custUserId))
			custUserId = fmt.Sprintf("%x", h.Sum(nil))
		}
		user, errCode := store.GetStore().GetUserLatestByCustomerUserId(projectId, custUserId)
		switch errCode {
		case http.StatusInternalServerError:
			return "", "", false, nil, nil, 0, fmt.Errorf(
				"Getting user by email failed.")

		case http.StatusNotFound:
			user = &model.User{ProjectId: projectId,
				CustomerUserId: custUserId,
				JoinTimestamp:  eventTimestamp,
				Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
			}
			_, errCode := store.GetStore().CreateUser(user)
			if errCode != http.StatusCreated {
				return "", "", false, nil, nil, 0, fmt.Errorf("Creating user by email failed.")
			}
			userId = user.ID
			isNewUser = true

		case http.StatusFound:
			userId = user.ID
		}
	} else {
		custUserIdType = FACTORS_CUSTOMER_USER_ID_TYPE
	}

	userProperties := U.PropertiesMap{
		"userIdType": custUserIdType,
	}
	eventProperties := U.PropertiesMap{
		"gateway":    checkoutObject.Gateway,
		"currency":   checkoutObject.Currency,
		"userIdType": custUserIdType,
	}
	if custUserId != "" {
		if shouldHashEmail {
			userProperties[fmt.Sprintf("%s%s", custUserIdType, "Hash")] = custUserId
			eventProperties[fmt.Sprintf("%s%s", custUserIdType, "Hash")] = custUserId
		} else {
			userProperties[custUserIdType] = custUserId
			eventProperties[custUserIdType] = custUserId
		}
	}
	if f, err := strconv.ParseFloat(checkoutObject.TotalPrice, 64); err == nil {
		eventProperties["total_price"] = f
	}
	if f, err := strconv.ParseFloat(checkoutObject.SubtotalPrice, 64); err == nil {
		eventProperties["subtotal_price"] = f
	}
	if f, err := strconv.ParseFloat(checkoutObject.TotalLineItemsPrice, 64); err == nil {
		eventProperties["total_line_items_price"] = f
	}
	if f, err := strconv.ParseFloat(checkoutObject.TotalDiscounts, 64); err == nil {
		eventProperties["total_discounts"] = f
	}

	return eventName, userId, isNewUser, eventProperties, userProperties, eventTimestamp, nil
}

type OrderObject struct {
	ID                  float64         `json:"id"`
	Email               string          `json:"email"`
	ClosedAt            string          `json:"closed_at"`
	CreatedAt           string          `json:"created_at"`
	UpdatedAt           string          `json:"updated_at"`
	Number              float64         `json:"number"`
	Token               string          `json:"token"`
	Gateway             string          `json:"gateway"`
	TotalPrice          string          `json:"total_price"`
	SubtotalPrice       string          `json:"subtotal_price"`
	TotalDiscounts      string          `json:"total_discounts"`
	TotalLineItemsPrice string          `json:"total_line_items_price"`
	Currency            string          `json:"currency"`
	Confirmed           bool            `json:"confirmed"`
	CartToken           string          `json:"cart_token"`
	Name                string          `json:"name"`
	CancelledAt         string          `json:"cancelled_at"`
	CancelReason        string          `json:"cancel_reason"`
	UserID              float64         `json:"user_id"`
	OrderNumber         float64         `json:"order_number"`
	ProcessingMethod    string          `json:"processing_method"`
	CheckoutId          float64         `json:"checkout_id"`
	SourceName          string          `json:"source_name"`
	Customer            CustomerObject  `json:"customer"`
	LineItems           []LineItem      `json:"line_items"`
	NoteAttributes      []NoteAttribute `json:"note_attributes"`
}

// Returns eventName, userId, isNewUser, eventProperties, userProperties, timestamp, err
func GetTrackDetailsFromOrderObject(
	projectId uint64, actionType int64, shouldHashEmail bool, orderObject *OrderObject) (
	string, string, bool, U.PropertiesMap, U.PropertiesMap, int64, error) {

	var eventTime time.Time
	var eventName string
	var err error
	if actionType == ACTION_SHOPIFY_ORDER_CREATED {
		eventName = U.EVENT_NAME_SHOPIFY_ORDER_CREATED
		eventTime, err = time.Parse(time.RFC3339, orderObject.CreatedAt)
		if err != nil {
			return "", "", false, nil, nil, 0, fmt.Errorf(
				fmt.Sprintf("Failed to parse time %s", orderObject.CreatedAt))
		}
	} else if actionType == ACTION_SHOPIFY_ORDER_UPDATED {
		eventName = U.EVENT_NAME_SHOPIFY_ORDER_UPDATED
		eventTime, err = time.Parse(time.RFC3339, orderObject.UpdatedAt)
		if err != nil {
			return "", "", false, nil, nil, 0, fmt.Errorf(
				fmt.Sprintf("Failed to parse time %s", orderObject.UpdatedAt))
		}
	} else if actionType == ACTION_SHOPIFY_ORDER_CANCELLED {
		eventName = U.EVENT_NAME_SHOPIFY_ORDER_CANCELLED
		eventTime, err = time.Parse(time.RFC3339, orderObject.CancelledAt)
		if err != nil {
			return "", "", false, nil, nil, 0, fmt.Errorf(
				fmt.Sprintf("Failed to parse time %s", orderObject.CancelledAt))
		}
	} else if actionType == ACTION_SHOPIFY_ORDER_PAID {
		eventName = U.EVENT_NAME_SHOPIFY_ORDER_PAID
		eventTime, err = time.Parse(time.RFC3339, orderObject.UpdatedAt)
		if err != nil {
			return "", "", false, nil, nil, 0, fmt.Errorf(
				fmt.Sprintf("Failed to parse time %s", orderObject.UpdatedAt))
		}
	}
	eventTimestamp := eventTime.Unix()

	userId := factorsUserIdFromAttributes(orderObject.NoteAttributes)
	custUserIdType := ""
	// In case of identified users.
	custUserId := ""
	isNewUser := false

	if userId == "" {
		if orderObject.Email != "" {
			custUserId = orderObject.Email
			custUserIdType = "orderEmail"
		} else if orderObject.Customer.Email != "" {
			custUserId = orderObject.Customer.Email
			custUserIdType = "orderCustomerEmail"
		} else if orderObject.UserID > 0 {
			custUserId = fmt.Sprintf("%f", orderObject.UserID)
			custUserIdType = "orderUserId"
		} else if orderObject.ID > 0 {
			custUserId = fmt.Sprintf("%f", orderObject.Customer.ID)
			custUserIdType = "orderCustomerId"
		}

		if custUserId == "" {
			return "", "", false, nil, nil, 0, fmt.Errorf("Missing email in OrderObject")
		}

		if shouldHashEmail {
			h := sha256.New()
			h.Write([]byte(custUserId))
			custUserId = fmt.Sprintf("%x", h.Sum(nil))
		}
		user, errCode := store.GetStore().GetUserLatestByCustomerUserId(projectId, custUserId)
		switch errCode {
		case http.StatusInternalServerError:
			return "", "", false, nil, nil, 0, fmt.Errorf(
				"Getting user by email failed.")

		case http.StatusNotFound:
			user = &model.User{ProjectId: projectId,
				CustomerUserId: custUserId,
				JoinTimestamp:  eventTimestamp,
				Source:         model.GetRequestSourcePointer(model.UserSourceWeb),
			}
			_, errCode := store.GetStore().CreateUser(user)
			if errCode != http.StatusCreated {
				return "", "", false, nil, nil, 0, fmt.Errorf("Creating user by email failed.")
			}
			userId = user.ID
			isNewUser = true

		case http.StatusFound:
			userId = user.ID
		}
	} else {
		custUserIdType = FACTORS_CUSTOMER_USER_ID_TYPE
	}

	userProperties := U.PropertiesMap{
		"userIdType": custUserIdType,
	}
	eventProperties := U.PropertiesMap{
		"gateway":      orderObject.Gateway,
		"currency":     orderObject.Currency,
		"number":       orderObject.Number,
		"order_number": orderObject.OrderNumber,
		"userIdType":   custUserIdType,
	}
	if custUserId != "" {
		if shouldHashEmail {
			userProperties[fmt.Sprintf("%s%s", custUserIdType, "Hash")] = custUserId
			eventProperties[fmt.Sprintf("%s%s", custUserIdType, "Hash")] = custUserId
		} else {
			userProperties[custUserIdType] = custUserId
			eventProperties[custUserIdType] = custUserId
		}
	}
	if f, err := strconv.ParseFloat(orderObject.TotalPrice, 64); err == nil {
		eventProperties["total_price"] = f
	}
	if f, err := strconv.ParseFloat(orderObject.SubtotalPrice, 64); err == nil {
		eventProperties["subtotal_price"] = f
	}
	if f, err := strconv.ParseFloat(orderObject.TotalLineItemsPrice, 64); err == nil {
		eventProperties["total_line_items_price"] = f
	}
	if f, err := strconv.ParseFloat(orderObject.TotalDiscounts, 64); err == nil {
		eventProperties["total_discounts"] = f
	}

	return eventName, userId, isNewUser, eventProperties, userProperties, eventTimestamp, nil
}

type CartTokenPayload struct {
	Timestamp float64 `json:"timestamp"`
	UserId    string  `json:"user_id"`
	CartToken string  `json:"cart_token"`
}

type CartObject struct {
	ID        string     `json:"id"`
	Token     string     `json:"token"`
	LineItems []LineItem `json:"line_items"`
	Note      string     `json:"note"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

// Returns eventName, customerEventId, userId, isNewUser, eventProperties, userProperties, timestamp, err
func GetTrackDetailsFromCartObject(
	projectId uint64, actionType int64, cartObject *CartObject) (
	string, string, bool, U.PropertiesMap, U.PropertiesMap, int64, error) {
	cartToken := cartObject.Token
	if cartToken == "" {
		return "", "", false, nil, nil, 0, fmt.Errorf("Missing cart token in CartObject")
	}

	userId, errCode := model.GetCacheUserIdForShopifyCartToken(projectId, cartToken)
	if errCode != http.StatusOK {
		return "", "", false, nil, nil, 0, fmt.Errorf(fmt.Sprintf(
			"Missing userId for project_id:%d, cart_token:%s", projectId, cartToken))
	}

	if actionType != ACTION_SHOPIFY_CART_UPDATED {
		return "", "", false, nil, nil, 0, fmt.Errorf("Unknown action type for Cart Object")
	}
	var eventTime time.Time
	var eventName string
	var err error

	eventName = U.EVENT_NAME_SHOPIFY_CART_UPDATED
	eventTime, err = time.Parse(time.RFC3339, cartObject.UpdatedAt)
	if err != nil {
		return "", "", false, nil, nil, 0, fmt.Errorf(
			fmt.Sprintf("Failed to parse time %s", cartObject.UpdatedAt))
	}
	eventTimestamp := eventTime.Unix()

	_, errCode = store.GetStore().GetUser(projectId, userId)
	if errCode != http.StatusFound {
		return "", "", false, nil, nil, 0, fmt.Errorf(
			fmt.Sprintf("Shopify User not found projectId:%d userId:%s for cart_token:%s",
				projectId, userId, cartToken))
	}

	userProperties := U.PropertiesMap{
		"userIdType": FACTORS_CUSTOMER_USER_ID_TYPE,
	}
	eventProperties := U.PropertiesMap{
		"userIdType": FACTORS_CUSTOMER_USER_ID_TYPE,
	}

	return eventName, userId, false, eventProperties, userProperties, eventTimestamp, nil
}
