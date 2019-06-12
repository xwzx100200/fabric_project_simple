/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"os"
	"testing"
)

// Runner provides common data for running integration tests.
type Runner struct {
	Org1Name           string
	Org2Name           string
	Org1AdminUser      string
	Org2AdminUser      string
	Org1User           string
	Org2User           string
	ChannelID          string
	CCPath             string
	sdk                *fabsdk.FabricSDK
	testSetup          *BaseSetupImpl
	installExampleCC   bool
	exampleChaincodeID string
}

// New constructs a Runner instance using defaults.
func New() *Runner {
	r := Runner{
		Org1Name:      GetSDKOrgs()[0],
		Org2Name:      GetSDKOrgs()[1],
		Org1AdminUser: GetSDKAdmins()[0],
		Org2AdminUser: GetSDKAdmins()[1],
		Org1User:      GetSDKUsers()[0],
		Org2User:      GetSDKUsers()[1],
		ChannelID:     GetSDKChannelID(),
		CCPath:        GetChainCodeName(),
	}

	return &r
}

// NewWithExampleCC constructs a Runner instance using defaults and configures to install example CC.
func NewWithCC() *Runner {
	r := New()
	r.installExampleCC = true

	return r
}

// Run executes the test suite against ExampleCC.
func (r *Runner) Run(m *testing.M) {
	gr := m.Run()
	r.teardown()
	os.Exit(gr)
}

// SDK returns the instantiated SDK instance. Panics if nil.
func (r *Runner) GetSDK() *fabsdk.FabricSDK {
	if r.sdk == nil {
		panic("SDK not instantiated")
	}

	return r.sdk
}

// GetSetup returns the integration test setup.
func (r *Runner) GetSetup() *BaseSetupImpl {
	return r.testSetup
}

// GetChaincodeID returns the generated chaincode ID for example CC.
func (r *Runner) GetChaincodeID() string {
	return r.exampleChaincodeID
}

// Initialize prepares for the test run.
func (r *Runner) Initialize() {
	r.testSetup = &BaseSetupImpl{
		ChannelID:         r.ChannelID,
		OrgID:             r.Org1Name,
		ChannelConfigFile: GetChannelConfigPath(r.ChannelID + ".tx"),
	}

	sdk, err := fabsdk.New(fetchConfigBackend())
	if err != nil {
		panic(fmt.Sprintf("Failed to create new SDK: %s", err))
	}
	r.sdk = sdk

	// Delete all private keys from the crypto suite store
	// and users from the user store
	CleanupUserData(nil, sdk)

	if err := r.testSetup.Initialize(sdk); err != nil {
		panic(err.Error())
	}

	if r.installExampleCC {
		r.exampleChaincodeID = GenerateExampleID(false)
		if err := PrepareExampleCC(sdk, fabsdk.WithUser("Admin"), r.testSetup.OrgID, r.exampleChaincodeID); err != nil {
			panic(fmt.Sprintf("PrepareExampleCC return error: %s", err))
		}
	}
}

func (r *Runner) teardown() {
	CleanupUserData(nil, r.sdk)
	r.sdk.Close()
}
