package subnet

import (
	"fleetmanager/api/errors"
	"fleetmanager/config"
	"fleetmanager/logger"
	mockconfig "fleetmanager/mocks/config"
	mockdirecter "fleetmanager/mocks/workflow/directer"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"github.com/golang/mock/gomock"
	vpc "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2"
	"reflect"
	"testing"
)

type ApiCallReturn struct {
	value interface{}
	err   error
}

func TestExecute(t *testing.T) {
	originGetSubnet := getSubnet
	originCreateSubnet := createSubnet
	originWaitSubnetReady := waitSubnetReady
	defer func() {
		getSubnet = originGetSubnet
		createSubnet = originCreateSubnet
		waitSubnetReady = originWaitSubnetReady
	}()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDirecter := mockdirecter.NewMockDirecter(mockCtrl)

	task := PrepareSubnetTask{
		components.BaseTask{
			Logger:   logger.NewDebugLogger(),
			Directer: mockDirecter,
		},
	}

	tests := []struct {
		name                          string
		preparedSubnetId              string
		expectedGetSubnetReturn       ApiCallReturn
		expectedCreateSubnetReturn    ApiCallReturn
		expectedWaitSubnetReadyReturn ApiCallReturn
		expectedReturn                ApiCallReturn
	}{
		{
			name:                          "prepare_vpc execute success: no prepared subnet",
			preparedSubnetId:              "",
			expectedGetSubnetReturn:       ApiCallReturn{"", nil},
			expectedCreateSubnetReturn:    ApiCallReturn{"", nil},
			expectedWaitSubnetReadyReturn: ApiCallReturn{nil, nil},
			expectedReturn:                ApiCallReturn{nil, nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := mockconfig.NewMockConfig(mockCtrl)
			context := &directer.WorkflowContext{Config: mockConfig}
			mockDirecter.EXPECT().GetContext().Return(context).MaxTimes(11)
			mockDirecter.EXPECT().Process(gomock.Any())

			// record
			mockConfig.EXPECT().Get(directer.WfKeyResourceProjectId).Return(&config.Entry{Val: "mock_res_pid"})
			mockConfig.EXPECT().Get(directer.WfKeyRegion).Return(&config.Entry{Val: "mock_region"})
			mockConfig.EXPECT().Get(directer.WfKeySubnetName).Return(&config.Entry{Val: "mock_subnet_name"})
			mockConfig.EXPECT().Get(directer.WfKeySubnetCidr).Return(&config.Entry{Val: "mock_subnet_cidr"})
			mockConfig.EXPECT().Get(directer.WfKeySubnetGatewayIp).Return(&config.Entry{Val: "mock_gateway_ip"})
			mockConfig.EXPECT().Get(directer.WfKeyVpcId).Return(&config.Entry{Val: "mock_vpc_id"})
			mockConfig.EXPECT().Get(directer.WfKeySubnetId).Return(&config.Entry{Val: tt.preparedSubnetId})
			mockConfig.EXPECT().Get(directer.WfKeyResourceAgencyName).Return(&config.Entry{Val: "mock_res_agency_name"})
			mockConfig.EXPECT().Get(directer.WfKeyResourceDomainId).Return(&config.Entry{Val: "mock_res_did"})

			if tt.expectedGetSubnetReturn.err == nil {
				if tt.preparedSubnetId == "" {
					mockConfig.EXPECT().Get(directer.WfDnsConfig).Return(&config.Entry{Val: "mock_config"})
					mockConfig.EXPECT().Set(directer.WfKeySubnetId, tt.expectedCreateSubnetReturn.value.(string)).Return(nil).Times(1)
				} else {
					mockConfig.EXPECT().Set(directer.WfKeySubnetId, tt.preparedSubnetId).Return(nil).Times(1)
				}
			}

			getSubnet = func(regionId string, projectId string, agencyName string, resDomainId string,
				subnetName string, vpc_id string) (subnetId string, vpcId string, err error) {
				subnetId = tt.expectedGetSubnetReturn.value.(string)
				err = tt.expectedGetSubnetReturn.err
				vpcId = ""
				return
			}

			createSubnet = func(regionId string, projectId string, agencyName string, resDomainId string,
				subnetName string, subnetCidr string, gatewayIp string, vpcId string,
				dnsConfig string) (result string, err error) {
				result = tt.expectedCreateSubnetReturn.value.(string)
				err = tt.expectedCreateSubnetReturn.err
				return
			}

			waitSubnetReady = func(regionId string, projectId string, agencyName string, resDomainId string,
				subnetId string) (err error) {
				err = tt.expectedWaitSubnetReadyReturn.err
				return
			}

			output, err := task.Execute(nil)
			if output != tt.expectedReturn.value {
				t.Errorf("expected return nil output on Execute, got not nil")
			}
			if err != tt.expectedReturn.err {
				t.Errorf("expected return nil error on Execute, got error")
			}
		})
	}
}

func TestGetSubnetDnsConfig(t *testing.T) {
	tests := []struct {
		name                 string
		dnsConfig            string
		expectedPrimaryDns   string
		expectedSecondaryDns string
		expectedDnsList      []string
	}{
		{
			name:                 "get_subnet_dns_config execute success: with empty dns config",
			dnsConfig:            "",
			expectedPrimaryDns:   "",
			expectedSecondaryDns: "",
			expectedDnsList:      nil,
		},
		{
			name:                 "get_subnet_dns_config execute success: with one dns config",
			dnsConfig:            "127.0.0.1",
			expectedPrimaryDns:   "127.0.0.1",
			expectedSecondaryDns: "",
			expectedDnsList:      []string{"127.0.0.1"},
		},
		{
			name:                 "get_subnet_dns_config execute success: with two dns config",
			dnsConfig:            "127.0.0.1,127.0.0.2",
			expectedPrimaryDns:   "127.0.0.1",
			expectedSecondaryDns: "127.0.0.2",
			expectedDnsList:      []string{"127.0.0.1", "127.0.0.2"},
		},
		{
			name:                 "get_subnet_dns_config execute success: with three dns config",
			dnsConfig:            "127.0.0.1,127.0.0.2,127.0.0.3",
			expectedPrimaryDns:   "127.0.0.1",
			expectedSecondaryDns: "127.0.0.2",
			expectedDnsList:      []string{"127.0.0.1", "127.0.0.2"},
		},
		{
			name:                 "get_subnet_dns_config execute success: with invalid config: 127.0.0.1, ",
			dnsConfig:            "127.0.0.1,",
			expectedPrimaryDns:   "127.0.0.1",
			expectedSecondaryDns: "",
			expectedDnsList:      []string{"127.0.0.1", ""},
		},
		{
			name:                 "get_subnet_dns_config execute success: with invalid config: ,127.0.0.1 ",
			dnsConfig:            ",127.0.0.1",
			expectedPrimaryDns:   "",
			expectedSecondaryDns: "127.0.0.1",
			expectedDnsList:      []string{"", "127.0.0.1"},
		},
	}

	var stringEqual = func(r1 *string, r2 string) bool {
		if r1 == nil {
			return r2 == ""
		}
		return *r1 == r2
	}

	var stringSliceEqual = func(r1 *[]string, r2 []string) bool {
		if r1 == nil {
			return r2 == nil
		}
		if r2 == nil {
			return r1 == nil
		}

		return reflect.DeepEqual(*r1, r2)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r1, r2, r3 := getSubnetDnsConfig(tt.dnsConfig)
			if !stringEqual(r1, tt.expectedPrimaryDns) {
				t.Errorf("expected return %v, got %v", tt.expectedPrimaryDns, r1)
			}
			if !stringEqual(r2, tt.expectedSecondaryDns) {
				t.Errorf("expected return %v, got %v", tt.expectedSecondaryDns, r2)
			}
			if !stringSliceEqual(r3, tt.expectedDnsList) {
				t.Errorf("expected return %v, got %v", tt.expectedDnsList, r3)
			}
		})
	}
}

func TestCreateSubnet(t *testing.T) {
	originClientGetAgencyVpcClient := clientGetAgencyVpcClient
	originCreateSubnet := createSubnet
	originClientCreateSubnet := clientCreateSubnet
	defer func() {
		clientGetAgencyVpcClient = originClientGetAgencyVpcClient
		createSubnet = originCreateSubnet
		clientCreateSubnet = originClientCreateSubnet
	}()

	errorCall := errors.NewError("error call")
	tests := []struct {
		name                                 string
		regionId                             string
		projectId                            string
		agencyName                           string
		resDomainId                          string
		subnetName                           string
		subnetCidr                           string
		gatewayIp                            string
		vpcId                                string
		dnsConfig                            string
		clientGetAgencyVpcClientReturnClient *vpc.VpcClient
		clientGetAgencyVpcClientReturnError  error
		clientCreateSubnetReturnString       string
		clientCreateSubnetReturnError        error
	}{
		{
			name:                                 "create_subnet execute success: with no error calls",
			clientGetAgencyVpcClientReturnClient: nil,
			clientGetAgencyVpcClientReturnError:  nil,
			clientCreateSubnetReturnString:       "",
			clientCreateSubnetReturnError:        nil,
		},
		{
			name:                                 "create_subnet execute success: with client error calls",
			clientGetAgencyVpcClientReturnClient: nil,
			clientGetAgencyVpcClientReturnError:  errorCall,
			clientCreateSubnetReturnString:       "",
			clientCreateSubnetReturnError:        errorCall,
		},
		{
			name:                                 "create_subnet execute success: with create subnet error calls",
			clientGetAgencyVpcClientReturnClient: nil,
			clientGetAgencyVpcClientReturnError:  nil,
			clientCreateSubnetReturnString:       "",
			clientCreateSubnetReturnError:        errorCall,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientGetAgencyVpcClient = func(regionId string, projectId string, agencyName string,
				resDomainId string) (*vpc.VpcClient, error) {
				return tt.clientGetAgencyVpcClientReturnClient, tt.clientGetAgencyVpcClientReturnError
			}

			clientCreateSubnet = func(vpcClient *vpc.VpcClient, vpcId *string, gatewayIp *string, cidr *string, name *string,
				primaryDns *string, secondaryDns *string, dnsList *[]string) (string, error) {
				return tt.clientCreateSubnetReturnString, tt.clientCreateSubnetReturnError
			}

			s1, e1 := createSubnet(tt.regionId, tt.projectId, tt.agencyName, tt.resDomainId, tt.subnetName,
				tt.subnetCidr, tt.gatewayIp, tt.vpcId, tt.dnsConfig)
			if tt.clientCreateSubnetReturnString != s1 {
				t.Errorf("expected return %v, got %v", tt.clientCreateSubnetReturnString, s1)
			}
			if tt.clientCreateSubnetReturnError != e1 {
				t.Errorf("expected return %v, got %v", tt.clientCreateSubnetReturnError, e1)
			}
		})
	}
}
