package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var (
	ErrUnknown           = errors.New("An unknown error occurred")
	ErrUnauthorized      = errors.New("Unauthorized")
	ErrRequiresMasterKey = errors.New("Operation requires Master key")

	ErrAccountAlreadyLinked              = 208
	ErrCacheMiss                         = 120
	ErrCommandUnavailable                = 108
	ErrConnectionFailed                  = 100
	ErrDuplicateValue                    = 137
	ErrExceededQuota                     = 140
	ErrFacebookAccountAlreadyLinked      = 208
	ErrFacebookIDMissing                 = 250
	ErrFacebookInvalidSession            = 251
	ErrFileDeleteFailure                 = 153
	ErrIncorrectType                     = 111
	ErrInternalServer                    = 1
	ErrInvalidACL                        = 123
	ErrInvalidChannelName                = 112
	ErrInvalidClassName                  = 103
	ErrInvalidDeviceToken                = 114
	ErrInvalidEmailAddress               = 125
	ErrInvalidFileName                   = 122
	ErrInvalidImageData                  = 150
	ErrInvalidJSON                       = 107
	ErrInvalidKeyName                    = 105
	ErrInvalidLinkedSession              = 251
	ErrInvalidNestedKey                  = 121
	ErrInvalidPointer                    = 106
	ErrInvalidProductIDentifier          = 146
	ErrInvalidPurchaseReceipt            = 144
	ErrInvalidQuery                      = 102
	ErrInvalidRoleName                   = 139
	ErrInvalidServerResponse             = 148
	ErrLinkedIDMissing                   = 250
	ErrMissingObjectID                   = 104
	ErrObjectNotFound                    = 101
	ErrObjectTooLarge                    = 116
	ErrOperationForbidden                = 119
	ErrPaymentDisabled                   = 145
	ErrProductDownloadFileSystemFailure  = 149
	ErrProductNotFoundInAppStore         = 147
	ErrPushMisconfigured                 = 115
	ErrReceiptMissing                    = 143
	ErrTimeout                           = 124
	ErrUnsavedFile                       = 151
	ErrUserCannotBeAlteredWithoutSession = 206
	ErrUserCanOnlyBeCreatedThroughSignUp = 207
	ErrUserEmailMissing                  = 204
	ErrUserEmailTaken                    = 203
	ErrUserIDMismatch                    = 209
	ErrUsernameMissing                   = 200
	ErrUsernameTaken                     = 202
	ErrUserPasswordMissing               = 201
	ErrUserWithEmailNotFound             = 205
	ErrScriptError                       = 141
	ErrValidationError                   = 142
)

// Error represents a Parse API error.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"error"`
}

func (e Error) Error() string {
	// TODO(tmc): improve formatting
	return fmt.Sprintf("parse.com error %v: %s", e.Code, e.Message)
}

func unmarshalError(r io.Reader) (error, bool) {
	err := &Error{}
	if r == nil {
		return ErrUnauthorized, false
	}
	if marshalErr := json.NewDecoder(r).Decode(&err); marshalErr != nil {
		return ErrUnknown, false
	}
	return err, false
}
