// Code generated by "stringer -type=PlaidLinkStatus -output=plaid_link.strings.go"; DO NOT EDIT.

package models

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[PlaidLinkStatusUnknown-0]
	_ = x[PlaidLinkStatusPending-1]
	_ = x[PlaidLinkStatusSetup-2]
	_ = x[PlaidLinkStatusError-3]
	_ = x[PlaidLinkStatusPendingExpiration-4]
	_ = x[PlaidLinkStatusRevoked-5]
}

const _PlaidLinkStatus_name = "PlaidLinkStatusUnknownPlaidLinkStatusPendingPlaidLinkStatusSetupPlaidLinkStatusErrorPlaidLinkStatusPendingExpirationPlaidLinkStatusRevoked"

var _PlaidLinkStatus_index = [...]uint8{0, 22, 44, 64, 84, 116, 138}

func (i PlaidLinkStatus) String() string {
	if i >= PlaidLinkStatus(len(_PlaidLinkStatus_index)-1) {
		return "PlaidLinkStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _PlaidLinkStatus_name[_PlaidLinkStatus_index[i]:_PlaidLinkStatus_index[i+1]]
}