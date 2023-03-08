package client

import (
	"fleetmanager/api/errors"
	mockclient "fleetmanager/mocks/client"
	"github.com/golang/mock/gomock"
	vpcmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
	"testing"
)

type ApiCallReturn struct {
	value interface{}
	err   error
}

func TestCreateVpc(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	errorCall := errors.NewError("error call")

	tests := []struct {
		name                    string
		argCidr                 string
		argName                 string
		argEnterpriseProject    string
		expectedCreateVpcReturn ApiCallReturn
		expectedReturn          ApiCallReturn
	}{
		{
			name:                    "client.CreateVpc return nil error with vpc client no exception",
			argCidr:                 "192.168.0.0/16",
			argName:                 "mockName",
			argEnterpriseProject:    "0",
			expectedCreateVpcReturn: ApiCallReturn{&vpcmodel.CreateVpcResponse{Vpc: &vpcmodel.Vpc{Id: "mock_vpc_id"}}, nil},
			expectedReturn:          ApiCallReturn{"mock_vpc_id", nil},
		},
		{
			name:                    "client.CreateVpc return error with vpc client exception",
			argCidr:                 "192.168.0.0/16",
			argName:                 "mockName",
			argEnterpriseProject:    "0",
			expectedCreateVpcReturn: ApiCallReturn{nil, errorCall},
			expectedReturn:          ApiCallReturn{"", errorCall},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mockclient.NewMockVpcV2(mockCtrl)

			// record
			mockClient.EXPECT().CreateVpc(&vpcmodel.CreateVpcRequest{
				Body: &vpcmodel.CreateVpcRequestBody{
					Vpc: &vpcmodel.CreateVpcOption{
						Cidr:                &tt.argCidr,
						Name:                &tt.argName,
						EnterpriseProjectId: &tt.argEnterpriseProject,
					},
				},
			}).Return(tt.expectedCreateVpcReturn.value, tt.expectedCreateVpcReturn.err)
			output, err := CreateVpc(mockClient, &tt.argCidr, &tt.argName, &tt.argEnterpriseProject)
			if output != tt.expectedReturn.value {
				t.Errorf("expected return nil output on CreateVpc, got not nil")
			}
			if err != tt.expectedReturn.err {
				t.Errorf("expected return nil error on CreateVpc, got error")
			}
		})
	}
}

func TestCreateSecurityGroup(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	errorCall := errors.NewError("error call")

	tests := []struct {
		name                              string
		argName                           string
		argEnterpriseProject              string
		expectedCreateSecurityGroupReturn ApiCallReturn
		expectedReturn                    ApiCallReturn
	}{
		{
			name:                              "client.CreateSecurityGroup return nil error with vpc client no exception",
			argName:                           "mockName",
			argEnterpriseProject:              "0",
			expectedCreateSecurityGroupReturn: ApiCallReturn{&vpcmodel.CreateSecurityGroupResponse{SecurityGroup: &vpcmodel.SecurityGroup{Id: "mock_sg_id"}}, nil},
			expectedReturn:                    ApiCallReturn{"mock_sg_id", nil},
		},
		{
			name:                              "client.CreateSecurityGroup return error with vpc client exception",
			argName:                           "mockName",
			argEnterpriseProject:              "0",
			expectedCreateSecurityGroupReturn: ApiCallReturn{nil, errorCall},
			expectedReturn:                    ApiCallReturn{"", errorCall},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mockclient.NewMockVpcV2(mockCtrl)

			// record
			mockClient.EXPECT().CreateSecurityGroup(&vpcmodel.CreateSecurityGroupRequest{
				Body: &vpcmodel.CreateSecurityGroupRequestBody{
					SecurityGroup: &vpcmodel.CreateSecurityGroupOption{
						Name:                tt.argName,
						EnterpriseProjectId: &tt.argEnterpriseProject,
					},
				},
			}).Return(tt.expectedCreateSecurityGroupReturn.value, tt.expectedCreateSecurityGroupReturn.err)
			output, err := CreateSecurityGroup(mockClient, &tt.argName, &tt.argEnterpriseProject)
			if output != tt.expectedReturn.value {
				t.Errorf("expected return nil output on CreateSecurityGroup, got not nil")
			}
			if err != tt.expectedReturn.err {
				t.Errorf("expected return nil error on CreateSecurityGroup, got error")
			}
		})
	}
}

func TestGetSecurityGroupRules(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockRules = &[]vpcmodel.SecurityGroupRule{{Id: "mock_sg_rule_id"}}
	var errorCall = errors.NewError("error call")

	tests := []struct {
		name                                   string
		argSecurityGroupId                     string
		argVpcClient                           *mockclient.MockVpcV2
		expectedListSecurityGroupRulesResponse *vpcmodel.ListSecurityGroupRulesResponse
		expectedListSecurityGroupRulesError    error
		expectedRsp                            *[]vpcmodel.SecurityGroupRule
		expectedError                          error
	}{
		{
			name:                                   "client.GetSecurityGroupRules return nil error with vpc client no exception",
			argSecurityGroupId:                     "mockSecurityGroupId",
			argVpcClient:                           mockclient.NewMockVpcV2(mockCtrl),
			expectedListSecurityGroupRulesResponse: &vpcmodel.ListSecurityGroupRulesResponse{SecurityGroupRules: mockRules},
			expectedListSecurityGroupRulesError:    nil,
			expectedRsp:                            mockRules,
			expectedError:                          nil,
		},
		{
			name:                                   "client.GetSecurityGroupRules return error with vpc client exception",
			argSecurityGroupId:                     "mockSecurityGroupId",
			argVpcClient:                           mockclient.NewMockVpcV2(mockCtrl),
			expectedListSecurityGroupRulesResponse: nil,
			expectedListSecurityGroupRulesError:    errorCall,
			expectedRsp:                            nil,
			expectedError:                          errorCall,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// record
			tt.argVpcClient.EXPECT().ListSecurityGroupRules(&vpcmodel.ListSecurityGroupRulesRequest{
				SecurityGroupId: &tt.argSecurityGroupId,
			}).Return(tt.expectedListSecurityGroupRulesResponse, tt.expectedListSecurityGroupRulesError)
			output, err := GetSecurityGroupRules(tt.argVpcClient, &tt.argSecurityGroupId)

			if output != tt.expectedRsp {
				t.Errorf("expected return nil output on GetSecurityGroupRules, got not nil")
			}
			if err != tt.expectedError {
				t.Errorf("expected return nil error on GetSecurityGroupRules, got error")
			}
		})
	}
}
