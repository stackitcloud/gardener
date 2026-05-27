// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package helper

import (
	"net"
	"slices"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
)

// GetDNSRecordType returns the appropriate DNS record type (A/AAAA or CNAME) for the given address.
func GetDNSRecordType(address string) extensionsv1alpha1.DNSRecordType {
	if ip := net.ParseIP(address); ip != nil {
		if ip.To4() != nil {
			return extensionsv1alpha1.DNSRecordTypeA
		}
		return extensionsv1alpha1.DNSRecordTypeAAAA
	}
	return extensionsv1alpha1.DNSRecordTypeCNAME
}

// GetDNSRecordTTL returns the value of the given ttl, or 120 if nil.
func GetDNSRecordTTL(ttl *int64) int64 {
	if ttl != nil {
		return *ttl
	}
	return 120
}

// TODO: Refactor description and add tests
// IPFamilyToDNSRecordType maps a Gardener IP family to the corresponding DNSRecord type.
func IPFamilyToDNSRecordType(family gardencorev1beta1.IPFamily) extensionsv1alpha1.DNSRecordType {
	if family == gardencorev1beta1.IPFamilyIPv6 {
		return extensionsv1alpha1.DNSRecordTypeAAAA
	}
	return extensionsv1alpha1.DNSRecordTypeA
}

// FilterAddressesByIPFamily returns the subset of addresses matching the first IP family in ipFamilies that has at
// least one match, along with that family's DNS record type.
func FilterAddressesByIPFamily(addresses []string, ipFamilies []gardencorev1beta1.IPFamily) ([]string, extensionsv1alpha1.DNSRecordType, bool) {
	for _, family := range ipFamilies {
		recordType := IPFamilyToDNSRecordType(family)
		filtered := slices.DeleteFunc(slices.Clone(addresses), func(addr string) bool {
			return GetDNSRecordType(addr) != recordType
		})
		if len(filtered) > 0 {
			return filtered, recordType, true
		}
	}
	return nil, "", false
}
