// Code generated by "stringer -type=Capability"; DO NOT EDIT.

package providers

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[CanAutoDNSSEC-0]
	_ = x[CanConcur-1]
	_ = x[CanGetZones-2]
	_ = x[CanUseAKAMAICDN-3]
	_ = x[CanUseAlias-4]
	_ = x[CanUseAzureAlias-5]
	_ = x[CanUseCAA-6]
	_ = x[CanUseDHCID-7]
	_ = x[CanUseDS-8]
	_ = x[CanUseDSForChildren-9]
	_ = x[CanUseLOC-10]
	_ = x[CanUseNAPTR-11]
	_ = x[CanUsePTR-12]
	_ = x[CanUseRoute53Alias-13]
	_ = x[CanUseSOA-14]
	_ = x[CanUseSRV-15]
	_ = x[CanUseSSHFP-16]
	_ = x[CanUseTLSA-17]
	_ = x[DocCreateDomains-18]
	_ = x[DocDualHost-19]
	_ = x[DocOfficiallySupported-20]
}

const _Capability_name = "CanAutoDNSSECCanConcurCanGetZonesCanUseAKAMAICDNCanUseAliasCanUseAzureAliasCanUseCAACanUseDHCIDCanUseDSCanUseDSForChildrenCanUseLOCCanUseNAPTRCanUsePTRCanUseRoute53AliasCanUseSOACanUseSRVCanUseSSHFPCanUseTLSADocCreateDomainsDocDualHostDocOfficiallySupported"

var _Capability_index = [...]uint16{0, 13, 22, 33, 48, 59, 75, 84, 95, 103, 122, 131, 142, 151, 169, 178, 187, 198, 208, 224, 235, 257}

func (i Capability) String() string {
	if i >= Capability(len(_Capability_index)-1) {
		return "Capability(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Capability_name[_Capability_index[i]:_Capability_index[i+1]]
}
