package integration

import (
	"crypto/sha256"
	M "factors/model"
	U "factors/util"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type LineItem struct {
	SKU string `json:"string"`
}

type CheckoutObject struct {
	ID                  float64        `json:"id"`
	Token               string         `json:"token"`
	CartToken           string         `json:"cart_token"`
	Email               string         `json:"email"`
	UserID              string         `json:"user_id"`
	Gateway             string         `json:"gateway"`
	CreatedAt           string         `json:"created_at"`
	UpdatedAt           string         `json:"updated_at"`
	Currency            string         `json:"currency"`
	PresentmentCurrency string         `json:"presentment_currency"`
	TotalDiscounts      string         `json:"total_discounts"`
	TotalLineItemsPrice string         `json:"total_line_items_price"`
	TotalPrice          string         `json:"total_price"`
	SubtotalPrice       string         `json:"subtotal_price"`
	LineItems           []LineItem     `json:"line_items"`
	Customer            CustomerObject `json:"customer"`
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

// Returns eventName, customerEventId, userId, isNewUser, eventProperties, userProperties, timestamp, err
func GetTrackDetailsFromCheckoutObject(
	projectId uint64, actionType int64, shouldHashEmail bool, checkoutObject *CheckoutObject) (
	string, string, bool, U.PropertiesMap, U.PropertiesMap, int64, error) {
	custUserId := ""
	if checkoutObject.Email != "" {
		custUserId = checkoutObject.Email
	} else if checkoutObject.Customer.Email != "" {
		custUserId = checkoutObject.Customer.Email
	} else if checkoutObject.UserID != "" {
		custUserId = checkoutObject.UserID
	} else if checkoutObject.ID > 0 {
		custUserId = fmt.Sprintf("%f", checkoutObject.Customer.ID)
	}

	if custUserId == "" {
		return "", "", false, nil, nil, 0, fmt.Errorf("Missing email in CheckoutObject")
	}
	if shouldHashEmail {
		h := sha256.New()
		h.Write([]byte(custUserId))
		custUserId = fmt.Sprintf("%x", h.Sum(nil))
	}

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

	isNewUser := false
	userId := ""
	user, errCode := M.GetUserLatestByCustomerUserId(projectId, custUserId)
	switch errCode {
	case http.StatusInternalServerError:
		return "", "", false, nil, nil, 0, fmt.Errorf(
			"Getting user by email failed.")

	case http.StatusNotFound:
		user = &M.User{ProjectId: projectId,
			CustomerUserId: custUserId,
			JoinTimestamp:  eventTimestamp,
		}
		_, errCode := M.CreateUser(user)
		if errCode != http.StatusCreated {
			return "", "", false, nil, nil, 0, fmt.Errorf("Creating user by email failed.")
		}
		userId = user.ID
		isNewUser = true

	case http.StatusFound:
		userId = user.ID
	}

	userProperties := U.PropertiesMap{}
	if shouldHashEmail {
		userProperties["emailHash"] = custUserId
	} else {
		userProperties[U.UP_EMAIL] = custUserId
	}
	eventProperties := U.PropertiesMap{
		"gateway":  checkoutObject.Gateway,
		"currency": checkoutObject.Currency,
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
	ID                  float64        `json:"id"`
	Email               string         `json:"email"`
	ClosedAt            string         `json:"closed_at"`
	CreatedAt           string         `json:"created_at"`
	UpdatedAt           string         `json:"updated_at"`
	Number              float64        `json:"number"`
	Token               string         `json:"token"`
	Gateway             string         `json:"gateway"`
	TotalPrice          string         `json:"total_price"`
	SubtotalPrice       string         `json:"subtotal_price"`
	TotalDiscounts      string         `json:"total_discounts"`
	TotalLineItemsPrice string         `json:"total_line_items_price"`
	Currency            string         `json:"currency"`
	Confirmed           bool           `json:"confirmed"`
	CartToken           string         `json:"cart_token"`
	Name                string         `json:"name"`
	CancelledAt         string         `json:"cancelled_at"`
	CancelReason        string         `json:"cancel_reason"`
	UserID              string         `json:"user_id"`
	OrderNumber         float64        `json:"order_number"`
	ProcessingMethod    string         `json:"processing_method"`
	CheckoutId          float64        `json:"checkout_id"`
	SourceName          string         `json:"source_name"`
	Customer            CustomerObject `json:"customer"`
}

// Returns eventName, userId, isNewUser, eventProperties, userProperties, timestamp, err
func GetTrackDetailsFromOrderObject(
	projectId uint64, actionType int64, shouldHashEmail bool, orderObject *OrderObject) (
	string, string, bool, U.PropertiesMap, U.PropertiesMap, int64, error) {
	custUserId := ""

	if orderObject.Email != "" {
		custUserId = orderObject.Email
	} else if orderObject.Customer.Email != "" {
		custUserId = orderObject.Customer.Email
	} else if orderObject.UserID != "" {
		custUserId = orderObject.UserID
	} else if orderObject.ID > 0 {
		custUserId = fmt.Sprintf("%f", orderObject.Customer.ID)
	}

	if custUserId == "" {
		return "", "", false, nil, nil, 0, fmt.Errorf("Missing email in OrderObject")
	}

	if shouldHashEmail {
		h := sha256.New()
		h.Write([]byte(custUserId))
		custUserId = fmt.Sprintf("%x", h.Sum(nil))
	}

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

	isNewUser := false
	userId := ""
	user, errCode := M.GetUserLatestByCustomerUserId(projectId, custUserId)
	switch errCode {
	case http.StatusInternalServerError:
		return "", "", false, nil, nil, 0, fmt.Errorf(
			"Getting user by email failed.")

	case http.StatusNotFound:
		user = &M.User{ProjectId: projectId,
			CustomerUserId: custUserId,
			JoinTimestamp:  eventTimestamp,
		}
		_, errCode := M.CreateUser(user)
		if errCode != http.StatusCreated {
			return "", "", false, nil, nil, 0, fmt.Errorf("Creating user by email failed.")
		}
		userId = user.ID
		isNewUser = true

	case http.StatusFound:
		userId = user.ID
	}

	userProperties := U.PropertiesMap{}
	if shouldHashEmail {
		userProperties["emailHash"] = custUserId
	} else {
		userProperties[U.UP_EMAIL] = custUserId
	}
	eventProperties := U.PropertiesMap{
		"gateway":      orderObject.Gateway,
		"currency":     orderObject.Currency,
		"number":       orderObject.Number,
		"order_number": orderObject.OrderNumber,
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
