/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package gsm

import (
	v1 "networkservice/api/siemens_iedge_dmapi_v1"
	"networkservice/internal/networking/common"

	"testing"

	nm "github.com/Wifx/gonetworkmanager/v2"
	mm "github.com/maltegrosse/go-modemmanager"
	"github.com/stretchr/testify/assert"
)

func Test_validateAPN(t *testing.T) {
	tests := []struct {
		name     string
		apn      string
		expected bool
	}{
		{
			name:     "Valid simple APN",
			apn:      "internet",
			expected: true,
		},
		{
			name:     "Valid APN with domain",
			apn:      "myAPN.provider.com",
			expected: true,
		},
		{
			name:     "Valid APN with subdomain",
			apn:      "mobile.internet.provider.com",
			expected: true,
		},
		{
			name:     "Valid APN with numbers",
			apn:      "internet123.provider456.com",
			expected: true,
		},
		{
			name:     "Valid APN with hyphens",
			apn:      "my-APN.provider-network.com",
			expected: true,
		},
		{
			name:     "Valid single character domain",
			apn:      "a.b.c",
			expected: true,
		},
		{
			name:     "Empty APN",
			apn:      "",
			expected: false,
		},
		{
			name:     "APN starting with dot",
			apn:      ".internet.com",
			expected: false,
		},
		{
			name:     "APN ending with dot",
			apn:      "internet.com.",
			expected: false,
		},
		{
			name:     "APN with consecutive dots",
			apn:      "internet..com",
			expected: false,
		},
		{
			name:     "APN starting with hyphen",
			apn:      "-internet.com",
			expected: false,
		},
		{
			name:     "APN ending with hyphen",
			apn:      "internet.com-",
			expected: false,
		},
		{
			name:     "APN with consecutive hyphens",
			apn:      "internet--mobile.com",
			expected: false,
		},
		{
			name:     "APN with invalid characters",
			apn:      "internet@provider.com",
			expected: false,
		},
		{
			name:     "APN with spaces",
			apn:      "internet provider.com",
			expected: false,
		},
		{
			name:     "APN with underscore",
			apn:      "internet_mobile.com",
			expected: false,
		},
		{
			name:     "APN too long (over 100 characters)",
			apn:      "wverylongapnnamethatexceedsthelimitofhundredcharactersandshouldfailvalidationbecauseitistooloyytrytryt",
			expected: false,
		},
		{
			name:     "Valid APN exactly 100 characters",
			apn:      "verylongapnnamethatisexactlyhundredcharactersandshouldjustpassvalidationbecauseitisnottoolo",
			expected: true,
		},
		{
			name:     "Valid APN with maximum label length (63 chars)",
			apn:      "a1234567890123456789012345678901234567890123456789012345678901.com",
			expected: true,
		},
		{
			name:     "Invalid APN with label too long (over 63 chars)",
			apn:      "a1234567890123456789012345678901234567890123456789012345678901234.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateAPN(tt.apn)
			assert.Equal(t, tt.expected, result, "validateAPN(%s) = %v, expected %v", tt.apn, result, tt.expected)
		})
	}
}

func Test_validatePIN(t *testing.T) {
	tests := []struct {
		name     string
		pin      string
		expected bool
	}{
		{
			name:     "Valid 4-digit PIN",
			pin:      "1234",
			expected: true,
		},
		{
			name:     "Valid 6-digit PIN",
			pin:      "123456",
			expected: true,
		},
		{
			name:     "Valid 8-digit PIN",
			pin:      "12345678",
			expected: true,
		},
		{
			name:     "Valid PIN with leading zeros",
			pin:      "0123",
			expected: true,
		},
		{
			name:     "Valid PIN all zeros",
			pin:      "0000",
			expected: true,
		},
		{
			name:     "Valid PIN all nines",
			pin:      "9999",
			expected: true,
		},
		{
			name:     "PIN too short (3 digits)",
			pin:      "123",
			expected: false,
		},
		{
			name:     "PIN too long (9 digits)",
			pin:      "123456789",
			expected: false,
		},
		{
			name:     "PIN with letters",
			pin:      "12a4",
			expected: false,
		},
		{
			name:     "PIN with special characters",
			pin:      "12#4",
			expected: false,
		},
		{
			name:     "PIN with spaces",
			pin:      "12 34",
			expected: false,
		},
		{
			name:     "PIN with leading space",
			pin:      " 1234",
			expected: false,
		},
		{
			name:     "PIN with trailing space",
			pin:      "1234 ",
			expected: false,
		},
		{
			name:     "PIN with leading and trailing spaces",
			pin:      " 1234 ",
			expected: false,
		},
		{
			name:     "PIN with hyphen",
			pin:      "12-34",
			expected: false,
		},
		{
			name:     "PIN with dot",
			pin:      "12.34",
			expected: false,
		},
		{
			name:     "PIN with plus sign",
			pin:      "+1234",
			expected: false,
		},
		{
			name:     "Valid 5-digit PIN",
			pin:      "12345",
			expected: true,
		},
		{
			name:     "Valid 7-digit PIN",
			pin:      "1234567",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validatePIN(tt.pin)
			assert.Equal(t, tt.expected, result, "validatePIN(%s) = %v, expected %v", tt.pin, result, tt.expected)
		})
	}
}

func Test_validateUserName(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{
			name:     "Empty username",
			username: "",
			expected: true, // Optional field, empty is valid
		},
		{
			name:     "Valid simple username",
			username: "user123",
			expected: true,
		},
		{
			name:     "Valid username with letters and numbers",
			username: "testUser123",
			expected: true,
		},
		{
			name:     "Valid username with dots",
			username: "user.name",
			expected: true,
		},
		{
			name:     "Valid username with underscores",
			username: "user_name",
			expected: true,
		},
		{
			name:     "Valid username with hyphens",
			username: "user-name",
			expected: true,
		},
		{
			name:     "Valid username with mixed allowed characters",
			username: "test.user_123-name",
			expected: true,
		},
		{
			name:     "Valid single character username",
			username: "a",
			expected: true,
		},
		{
			name:     "Valid username exactly 64 characters",
			username: "a123456789012345678901234567890123456789012345678901234567890123",
			expected: true,
		},
		{
			name:     "Valid username with leading space",
			username: " user123",
			expected: false,
		},
		{
			name:     "Valid username with trailing space",
			username: "user123 ",
			expected: false,
		},
		{
			name:     "Valid username with leading and trailing spaces",
			username: " user123 ",
			expected: false,
		},
		{
			name:     "Username with only spaces",
			username: "    ",
			expected: false,
		},
		{
			name:     "Invalid username too long (over 64 chars)",
			username: "a1234567890123456789012345678901234567890123456789012345678901234",
			expected: false,
		},
		{
			name:     "Invalid username with spaces in middle",
			username: "user name",
			expected: false,
		},
		{
			name:     "Invalid username with special characters",
			username: "user@domain.com",
			expected: false,
		},
		{
			name:     "Invalid username with hash",
			username: "user#123",
			expected: false,
		},
		{
			name:     "Invalid username with percent",
			username: "user%123",
			expected: false,
		},
		{
			name:     "Invalid username with ampersand",
			username: "user&name",
			expected: false,
		},
		{
			name:     "Invalid username with asterisk",
			username: "user*123",
			expected: false,
		},
		{
			name:     "Invalid username with plus",
			username: "user+name",
			expected: false,
		},
		{
			name:     "Invalid username with equals",
			username: "user=name",
			expected: false,
		},
		{
			name:     "Invalid username with brackets",
			username: "user[123]",
			expected: false,
		},
		{
			name:     "Invalid username with braces",
			username: "user{123}",
			expected: false,
		},
		{
			name:     "Invalid username with parentheses",
			username: "user(123)",
			expected: false,
		},
		{
			name:     "Invalid username with quotes",
			username: "user\"name",
			expected: false,
		},
		{
			name:     "Invalid username with single quotes",
			username: "user'name",
			expected: false,
		},
		{
			name:     "Invalid username with semicolon",
			username: "user;name",
			expected: false,
		},
		{
			name:     "Invalid username with colon",
			username: "user:name",
			expected: false,
		},
		{
			name:     "Invalid username with comma",
			username: "user,name",
			expected: false,
		},
		{
			name:     "Invalid username with less than",
			username: "user<name",
			expected: false,
		},
		{
			name:     "Invalid username with greater than",
			username: "user>name",
			expected: false,
		},
		{
			name:     "Invalid username with question mark",
			username: "user?name",
			expected: false,
		},
		{
			name:     "Invalid username with forward slash",
			username: "user/name",
			expected: false,
		},
		{
			name:     "Invalid username with backslash",
			username: "user\\name",
			expected: false,
		},
		{
			name:     "Invalid username with pipe",
			username: "user|name",
			expected: false,
		},
		{
			name:     "Valid username starting with number",
			username: "123user",
			expected: true,
		},
		{
			name:     "Valid username starting with dot",
			username: ".username",
			expected: true,
		},
		{
			name:     "Valid username starting with underscore",
			username: "_username",
			expected: true,
		},
		{
			name:     "Valid username starting with hyphen",
			username: "-username",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateUserName(tt.username)
			assert.Equal(t, tt.expected, result, "validateUserName(%s) = %v, expected %v", tt.username, result, tt.expected)
		})
	}
}

func Test_validatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "Valid simple password",
			password: "password123",
			expected: true,
		},
		{
			name:     "Valid password with uppercase and lowercase",
			password: "Password123",
			expected: true,
		},
		{
			name:     "Valid password with special characters",
			password: "Pass@123!",
			expected: true,
		},
		{
			name:     "Valid password with all allowed special characters",
			password: "P@ss#$%^&*()_+-=[]{}|;':\",./<>?",
			expected: true,
		},
		{
			name:     "Valid password with numbers only",
			password: "123456789",
			expected: true,
		},
		{
			name:     "Valid password with letters only",
			password: "abcdefghijklmnopqrstuvwxyz",
			expected: true,
		},
		{
			name:     "Valid single character password",
			password: "a",
			expected: true,
		},
		{
			name:     "Valid password exactly 64 characters",
			password: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ab",
			expected: true,
		},
		{
			name:     "Valid password with leading space",
			password: " Password123",
			expected: false,
		},
		{
			name:     "Valid password with trailing space",
			password: "Password123 ",
			expected: false,
		},
		{
			name:     "Valid password with leading and trailing spaces",
			password: " Password123 ",
			expected: false,
		},
		{
			name:     "Valid password with exclamation mark",
			password: "Password!",
			expected: true,
		},
		{
			name:     "Valid password with at symbol",
			password: "user@domain",
			expected: true,
		},
		{
			name:     "Valid password with hash",
			password: "pass#word",
			expected: true,
		},
		{
			name:     "Valid password with dollar sign",
			password: "money$123",
			expected: true,
		},
		{
			name:     "Valid password with percent",
			password: "100%sure",
			expected: true,
		},
		{
			name:     "Valid password with caret",
			password: "power^2",
			expected: true,
		},
		{
			name:     "Valid password with ampersand",
			password: "me&you",
			expected: true,
		},
		{
			name:     "Valid password with asterisk",
			password: "star*light",
			expected: true,
		},
		{
			name:     "Valid password with parentheses",
			password: "test(123)",
			expected: true,
		},
		{
			name:     "Valid password with underscore",
			password: "test_password",
			expected: true,
		},
		{
			name:     "Valid password with plus",
			password: "one+two",
			expected: true,
		},
		{
			name:     "Valid password with minus/hyphen",
			password: "test-password",
			expected: true,
		},
		{
			name:     "Valid password with equals",
			password: "a=b",
			expected: true,
		},
		{
			name:     "Valid password with brackets",
			password: "array[0]",
			expected: true,
		},
		{
			name:     "Valid password with braces",
			password: "object{key}",
			expected: true,
		},
		{
			name:     "Valid password with semicolon",
			password: "line;end",
			expected: true,
		},
		{
			name:     "Valid password with single quote",
			password: "it's",
			expected: true,
		},
		{
			name:     "Valid password with double quote",
			password: "say\"hello\"",
			expected: true,
		},
		{
			name:     "Valid password with backslash",
			password: "path\\file",
			expected: true,
		},
		{
			name:     "Valid password with pipe",
			password: "cmd|grep",
			expected: true,
		},
		{
			name:     "Valid password with comma",
			password: "a,b,c",
			expected: true,
		},
		{
			name:     "Valid password with period",
			password: "file.txt",
			expected: true,
		},
		{
			name:     "Valid password with forward slash",
			password: "path/file",
			expected: true,
		},
		{
			name:     "Valid password with less than",
			password: "a<b",
			expected: true,
		},
		{
			name:     "Valid password with greater than",
			password: "a>b",
			expected: true,
		},
		{
			name:     "Valid password with question mark",
			password: "why?",
			expected: true,
		},
		{
			name:     "Empty password",
			password: "",
			expected: false,
		},
		{
			name:     "Password with only spaces",
			password: "    ",
			expected: false, // Becomes empty after trimming
		},
		{
			name:     "Password too long (over 64 characters)",
			password: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abc",
			expected: false,
		},
		{
			name:     "Password with space in middle",
			password: "pass word",
			expected: false,
		},
		{
			name:     "Password with tab character",
			password: "pass\tword",
			expected: false,
		},
		{
			name:     "Password with newline character",
			password: "pass\nword",
			expected: false,
		},
		{
			name:     "Password with carriage return",
			password: "pass\rword",
			expected: false,
		},
		{
			name:     "Password with tilde",
			password: "pass~word",
			expected: false,
		},
		{
			name:     "Password with backtick",
			password: "pass`word",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validatePassword(tt.password)
			assert.Equal(t, tt.expected, result, "validatePassword(%s) = %v, expected %v", tt.password, result, tt.expected)
		})
	}
}

func Test_validateGSMConfig(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		gsmConfig *v1.Interface_GsmConf
		wantErr   bool
	}{
		{
			name: "Valid GSM config",
			gsmConfig: &v1.Interface_GsmConf{
				Apn:      "internet.provider.com",
				Pin:      "1234",
				Username: "user_name",
				Password: "Password123!",
			},
			wantErr: false,
		},
		{
			name: "Empty Pin in GSM config",
			gsmConfig: &v1.Interface_GsmConf{
				Apn: "valid.apn",
				Pin: "",
			},
			wantErr: false,
		},
		{
			name: "Valid APN and invalid Pin in GSM config",
			gsmConfig: &v1.Interface_GsmConf{
				Apn: "valid.apn",
				Pin: "12rfg4",
			},
			wantErr: true,
		},
		{
			name: "Invalid APN and Empty PIN in GSM config",
			gsmConfig: &v1.Interface_GsmConf{
				Apn: "invalid..apn",
				Pin: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid APN and PIN in GSM config",
			gsmConfig: &v1.Interface_GsmConf{
				Apn: "invalid..apn",
				Pin: "1234",
			},
			wantErr: true,
		},
		{
			name: "Invalid PIN in GSM config",
			gsmConfig: &v1.Interface_GsmConf{
				Apn: "internet.provider.com",
				Pin: "12a4",
			},
			wantErr: true,
		},
		{
			name: "Empty Password in GSM config",
			gsmConfig: &v1.Interface_GsmConf{
				Apn:      "internet.provider.com",
				Pin:      "1234",
				Username: "username",
				Password: "",
			},
			wantErr: true,
		},
		{
			name: "Empty Username in GSM config",
			gsmConfig: &v1.Interface_GsmConf{
				Apn:      "internet.provider.com",
				Pin:      "1234",
				Username: "",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "Invalid Username in GSM config",
			gsmConfig: &v1.Interface_GsmConf{
				Apn:      "internet.provider.com",
				Pin:      "1234",
				Username: "user name",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "Invalid Password in GSM config",
			gsmConfig: &v1.Interface_GsmConf{
				Apn:      "internet.provider.com",
				Pin:      "1234",
				Username: "username",
				Password: "pass word",
			},
			wantErr: true,
		},
		{
			name: "PinWithLeadingAndTrailingSpaces",
			gsmConfig: &v1.Interface_GsmConf{
				Apn: "internet.provider.com",
				Pin: " 1234 ",
			},
			wantErr: false,
		},
		{
			name: "PinWithLeadingSpace",
			gsmConfig: &v1.Interface_GsmConf{
				Apn: "internet.provider.com",
				Pin: " 1234",
			},
			wantErr: false,
		},
		{
			name: "PinWithTrailingSpace",
			gsmConfig: &v1.Interface_GsmConf{
				Apn: "internet.provider.com",
				Pin: "1234 ",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connecionSettings := &v1.ConnectionSettings_Gsm{
				Gsm: tt.gsmConfig,
			}
			gotErr := HasValidGSMConfig(connecionSettings)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("validateGSMConfig() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("validateGSMConfig() succeeded unexpectedly")
			}
		})
	}
}

func Test_isValidGsmConfigUserNamePassword(t *testing.T) {
	tests := []struct {
		name      string
		gsmConfig *v1.Interface_GsmConf
		wantErr   bool
	}{
		{
			name: "Valid username and password",
			gsmConfig: &v1.Interface_GsmConf{
				Username: "valid_user",
				Password: "ValidPass123!",
			},
			wantErr: false,
		},
		{
			name: "Username provided but password empty",
			gsmConfig: &v1.Interface_GsmConf{
				Username: "user_without_password",
				Password: "",
			},
			wantErr: true,
		},
		{
			name: "Password provided but username empty",
			gsmConfig: &v1.Interface_GsmConf{
				Username: "",
				Password: "password_without_user",
			},
			wantErr: true,
		},
		{
			name: "Both username and	 password empty",
			gsmConfig: &v1.Interface_GsmConf{
				Username: "",
				Password: "",
			},
			wantErr: false,
		},
		{
			name: "Invalid username format",
			gsmConfig: &v1.Interface_GsmConf{
				Username: "invalid user",
				Password: "ValidPass123!",
			},
			wantErr: true,
		},
		{
			name: "Invalid password format",
			gsmConfig: &v1.Interface_GsmConf{
				Username: "valid_user",
				Password: "invalid pass",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connectionSettings := &v1.ConnectionSettings_Gsm{
				Gsm: tt.gsmConfig,
			}
			err := isValidGsmConfigUserNamePassword(connectionSettings)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("isValidGsmConfigUserNamePassword() failed: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("isValidGsmConfigUserNamePassword() succeeded unexpectedly")
			}
		})
	}
}

// extractStringFromSettings test cases

func Test_extractStringFromSettings_ReturnsValueWhenPresent(t *testing.T) {
	settings := nm.ConnectionSettings{
		"gsm": map[string]interface{}{
			"apn": "my-internet",
		},
	}

	result := extractStringFromSettings(settings, "gsm", "apn")
	assert.Equal(t, "my-internet", result)
}

func Test_extractStringFromSettings_ReturnsEmptyWhenSectionMissing(t *testing.T) {
	settings := nm.ConnectionSettings{}

	result := extractStringFromSettings(settings, "gsm", "apn")
	assert.Equal(t, "", result)
}

func Test_extractStringFromSettings_ReturnsEmptyWhenKeyMissing(t *testing.T) {
	settings := nm.ConnectionSettings{
		"gsm": map[string]interface{}{},
	}

	result := extractStringFromSettings(settings, "gsm", "apn")
	assert.Equal(t, "", result)
}

func Test_extractStringFromSettings_ReturnsEmptyWhenValueNotString(t *testing.T) {
	settings := nm.ConnectionSettings{
		"gsm": map[string]interface{}{
			"apn": 12345,
		},
	}

	result := extractStringFromSettings(settings, "gsm", "apn")
	assert.Equal(t, "", result)
}

func Test_ValidateGSMConfigFileContent(t *testing.T) {
	tests := []struct {
		name    string
		parsed  map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name:   "Valid config with APN only",
			parsed: map[string]string{"Apn": "internet"},
		},
		{
			name:   "Valid config with all fields",
			parsed: map[string]string{"Apn": "internet", "Pin": "1234", "Username": "user1", "Password": "pass1"},
		},
		{
			name:   "Valid config with APN and PIN",
			parsed: map[string]string{"Apn": "internet", "Pin": "5678"},
		},
		{
			name:    "Missing APN",
			parsed:  map[string]string{"Pin": "1234"},
			wantErr: true,
			errMsg:  "missing required APN",
		},
		{
			name:    "Empty APN",
			parsed:  map[string]string{"Apn": ""},
			wantErr: true,
			errMsg:  "missing required APN",
		},
		{
			name:    "Invalid APN format",
			parsed:  map[string]string{"Apn": "invalid apn!"},
			wantErr: true,
			errMsg:  "invalid APN format",
		},
		{
			name:    "Invalid PIN format",
			parsed:  map[string]string{"Apn": "internet", "Pin": "ab"},
			wantErr: true,
			errMsg:  "invalid PIN format",
		},
		{
			name:    "Username without Password",
			parsed:  map[string]string{"Apn": "internet", "Username": "user1"},
			wantErr: true,
			errMsg:  "Username but missing Password",
		},
		{
			name:    "Password without Username",
			parsed:  map[string]string{"Apn": "internet", "Password": "pass1"},
			wantErr: true,
			errMsg:  "Password but missing Username",
		},
		{
			name:    "Invalid Username format",
			parsed:  map[string]string{"Apn": "internet", "Username": "user@!!", "Password": "pass1"},
			wantErr: true,
			errMsg:  "invalid Username format",
		},
		{
			name:    "Username with empty Password",
			parsed:  map[string]string{"Apn": "internet", "Username": "user1", "Password": ""},
			wantErr: true,
			errMsg:  "Username but missing Password",
		},
		{
			name:    "Invalid Password format",
			parsed:  map[string]string{"Apn": "internet", "Username": "user1", "Password": "pass word"},
			wantErr: true,
			errMsg:  "invalid Password format",
		},
		{
			name:   "Empty optional fields are OK",
			parsed: map[string]string{"Apn": "internet", "Pin": "", "Username": "", "Password": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGSMConfigFileContent(tt.parsed)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_mapStateFailedReasonToString(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		reason mm.MMModemStateFailedReason
		want   string
	}{
		{
			name:   "No Error",
			reason: mm.MmModemStateFailedReasonNone,
			want:   "NO_ERROR",
		},
		{
			name:   "Unknown Error",
			reason: mm.MmModemStateFailedReasonUnknown,
			want:   "UNKNOWN_ERROR",
		},
		{
			name:   "SIM Missing",
			reason: mm.MmModemStateFailedReasonSimMissing,
			want:   "SIM_MISSING",
		},
		{
			name:   "SIM Error",
			reason: mm.MmModemStateFailedReasonSimError,
			want:   "SIM_ERROR",
		},
		{
			name:   "Unknown Capabilities",
			reason: mm.MmModemStateFailedReasonUnknownCapabilities,
			want:   "UNKNOWN_CAPABILITIES",
		},
		{
			name:   "eSIM Without Profiles",
			reason: mm.MmModemStateFailedReasonEsimWithoutProfiles,
			want:   "ESIM_NO_PROFILES",
		},
		{
			name:   "Fallback For Unknown Value",
			reason: 111,
			want:   "UNKNOWN_ERROR",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapStateFailedReasonToString(tt.reason)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_mapProtoConnectionState(t *testing.T) {
	testCases := []struct {
		name               string
		connectionState    mm.MMModemState
		expectedProtoState string
		bearerConnected    bool
	}{
		{
			name:               "Registered State",
			connectionState:    mm.MmModemStateRegistered,
			expectedProtoState: "DISCONNECTED",
		},
		{
			name:               "Searching State",
			connectionState:    mm.MmModemStateSearching,
			expectedProtoState: "DISCONNECTED",
		},
		{
			name:               "Disabled State",
			connectionState:    mm.MmModemStateDisabled,
			expectedProtoState: "DISCONNECTED",
		},
		{
			name:               "Locked State",
			connectionState:    mm.MmModemStateLocked,
			expectedProtoState: "SIM_LOCKED",
		},
		{
			name:               "Connected State",
			connectionState:    mm.MmModemStateConnected,
			expectedProtoState: "CONNECTED",
			bearerConnected:    true,
		},
		{
			name:               "Connected State Without Bearer",
			connectionState:    mm.MmModemStateConnected,
			expectedProtoState: "CONNECTING",
			bearerConnected:    false,
		},
		{
			name:               "Failed State",
			connectionState:    mm.MmModemStateFailed,
			expectedProtoState: "DISCONNECTED",
		},
		{
			name:               "Unknown State",
			connectionState:    mm.MmModemStateUnknown,
			expectedProtoState: "UNSPECIFIED",
		},
		{
			name:               "Initializing State",
			connectionState:    mm.MmModemStateInitializing,
			expectedProtoState: "CONNECTING",
		},
		{
			name:               "Enabling State",
			connectionState:    mm.MmModemStateEnabling,
			expectedProtoState: "CONNECTING",
		},
		{
			name:               "Enabled State",
			connectionState:    mm.MmModemStateEnabled,
			expectedProtoState: "DISCONNECTED",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapProtoConnectionState(tc.connectionState, tc.bearerConnected)
			assert.Equal(t, tc.expectedProtoState, result)
		})
	}
}

func Test_mapProtoAccessTechnology(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		techs []mm.MMModemAccessTechnology
		want  string
	}{
		{
			name:  "4G Technology",
			techs: []mm.MMModemAccessTechnology{mm.MmModemAccessTechnologyLte},
			want:  common.AccessTech4G,
		},
		{
			name:  "3G Technology",
			techs: []mm.MMModemAccessTechnology{mm.MmModemAccessTechnologyHspa},
			want:  common.AccessTech3G,
		},
		{
			name: "2G Technology",
			techs: []mm.MMModemAccessTechnology{mm.MmModemAccessTechnologyGsmCompact,
				mm.MmModemAccessTechnologyGprs, mm.MmModemAccessTechnologyEdge},
			want: common.AccessTech2G,
		},
		{
			name:  "Empty Technologies",
			techs: []mm.MMModemAccessTechnology{},
			want:  common.AccessTechUnspecified,
		},
		{
			name:  "Unknown Technology",
			techs: []mm.MMModemAccessTechnology{mm.MmModemAccessTechnologyAny},
			want:  common.AccessTechUnspecified,
		},
		{
			name:  "Multiple Technologies Selects Highest",
			techs: []mm.MMModemAccessTechnology{mm.MmModemAccessTechnologyGsm, mm.MmModemAccessTechnologyLte},
			want:  common.AccessTech4G,
		},
		{
			name:  "Multiple Technologies Selects Highest Regardless Of Order",
			techs: []mm.MMModemAccessTechnology{mm.MmModemAccessTechnologyLte, mm.MmModemAccessTechnologyGsm},
			want:  common.AccessTech4G,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapProtoAccessTechnology(tt.techs)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_mapProtoRegistration(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		reg   mm.MMModem3gppRegistrationState
		want  string
		want2 string
	}{
		{
			name:  "Home Registration",
			reg:   mm.MmModem3gppRegistrationStateHomeSmsOnly,
			want:  common.RegStateRegistered,
			want2: common.RoamingStateHome,
		},
		{
			name:  "Roaming Registration",
			reg:   mm.MmModem3gppRegistrationStateRoamingSmsOnly,
			want:  common.RegStateRegistered,
			want2: common.RoamingStateRoaming,
		},
		{
			name:  "Searching Registration",
			reg:   mm.MmModem3gppRegistrationStateSearching,
			want:  common.RegStateSearching,
			want2: common.RoamingStateUnspecified,
		},
		{
			name:  "Denied Registration",
			reg:   mm.MmModem3gppRegistrationStateDenied,
			want:  common.RegStateDenied,
			want2: common.RoamingStateUnspecified,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2 := mapProtoRegistration(tt.reg)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want2, got2)
		})
	}
}
