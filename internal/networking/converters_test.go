/*
 * Copyright © Siemens 2024 - 2025. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package networking

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_SuccessConvertUInt32BIToIP(t *testing.T) {
	ip := uint32(2130706433)
	expectedResult := "127.0.0.1"
	result := IPFromUInt32BI(ip)

	assert.Equal(t, expectedResult, result, "IPFromUInt32BI result should match expected IP")
}

func Test_SuccessConvertUInt32LIToIP(t *testing.T) {
	ip := uint32(16777343)
	expectedResult := "127.0.0.1"
	result := IPFromUInt32LI(ip)

	assert.Equal(t, expectedResult, result, "IPFromUInt32LI result should match expected IP")
}

func Test_SuccessConvertIPToUInt32BI(t *testing.T) {
	ip := "127.0.0.1"
	expectedResult := uint32(2130706433)
	result := IPToUInt32BI(ip)

	assert.Equal(t, expectedResult, result, "IPToUInt32BI result should match expected value")
}

func Test_SuccessConvertIPToUInt32LI(t *testing.T) {
	ip := "127.0.0.1"
	expectedResult := uint32(16777343)
	result := IPToUInt32LI(ip)

	assert.Equal(t, expectedResult, result, "IPToUInt32LI result should match expected value")
}

func Test_SuccessParseNetMaskSize(t *testing.T) {
	ip := "255.255.255.0"
	expectedResult := uint32(24)
	result := ParseNetMaskSize(ip)

	assert.Equal(t, expectedResult, result, "ParseNetMaskSize result should match expected size")
}

func Test_SuccessParseNetMask(t *testing.T) {
	cidrPrefix := uint32(24)
	expectedResult := "255.255.255.0"
	result := ParseNetMask(cidrPrefix)

	assert.Equal(t, expectedResult, result, "ParseNetMask result should match expected mask")
}

func Test_FailureParseNetMask_InvalidCIDRPrefix(t *testing.T) {
	cidrPrefix := uint32(33)
	expectedResult := ""
	result := ParseNetMask(cidrPrefix)

	assert.Equal(t, expectedResult, result, "ParseNetMask result should return empty string when get invalid cidr prefix")
}
