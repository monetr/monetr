// Code generated by "stringer -type=TellerLinkStatus -output=teller_link.strings.go"; DO NOT EDIT.

package models

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TellerLinkStatusUnknown-0]
	_ = x[TellerLinkStatusPending-1]
	_ = x[TellerLinkStatusSetup-2]
	_ = x[TellerLinkStatusDisconnected-3]
}

const _TellerLinkStatus_name = "TellerLinkStatusUnknownTellerLinkStatusPendingTellerLinkStatusSetupTellerLinkStatusDisconnected"

var _TellerLinkStatus_index = [...]uint8{0, 23, 46, 67, 95}

func (i TellerLinkStatus) String() string {
	if i >= TellerLinkStatus(len(_TellerLinkStatus_index)-1) {
		return "TellerLinkStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TellerLinkStatus_name[_TellerLinkStatus_index[i]:_TellerLinkStatus_index[i+1]]
}