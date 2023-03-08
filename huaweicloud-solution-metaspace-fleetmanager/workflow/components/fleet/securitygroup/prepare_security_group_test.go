package securitygroup

import (
	"fleetmanager/api/errors"
	"fleetmanager/client"
	"fleetmanager/config"
	"fleetmanager/logger"
	mockconfig "fleetmanager/mocks/config"
	mockdirecter "fleetmanager/mocks/workflow/directer"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"github.com/golang/mock/gomock"
	vpc "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2"
	vpcmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
	"testing"
)

type ApiCallReturn struct {
	value interface{}
	err   error
}

func TestExecute(t *testing.T) {
	originGetSecurityGroup := getSecurityGroup
	originCreateSecurityGroup := createSecurityGroup
	originFindSecurityGroupRules := findSecurityGroupRules
	originDeleteRules := deleteRules
	defer func() {
		getSecurityGroup = originGetSecurityGroup
		createSecurityGroup = originCreateSecurityGroup
		findSecurityGroupRules = originFindSecurityGroupRules
		deleteRules = originDeleteRules
	}()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	errorCall := errors.NewError("error call")

	tests := []struct {
		name                                 string
		preparedSecurityGroupId              string
		expectedGetSecurityGroupReturn       ApiCallReturn
		expectedCreateSecurityGroupReturn    ApiCallReturn
		expectedFindSecurityGroupRulesReturn ApiCallReturn
		expectedDeleteRules                  ApiCallReturn
		expectedReturn                       ApiCallReturn
	}{
		{
			name:                                 "prepare_security_group execute success: no prepared security_group",
			preparedSecurityGroupId:              "",
			expectedGetSecurityGroupReturn:       ApiCallReturn{"", nil},
			expectedCreateSecurityGroupReturn:    ApiCallReturn{"mock_id", nil},
			expectedFindSecurityGroupRulesReturn: ApiCallReturn{[]string{}, nil},
			expectedDeleteRules:                  ApiCallReturn{nil, nil},
			expectedReturn:                       ApiCallReturn{nil, nil},
		},
		{
			name:                                 "prepare_security_group execute success: prepared security_group",
			preparedSecurityGroupId:              "prepared_security_group",
			expectedGetSecurityGroupReturn:       ApiCallReturn{"", nil},
			expectedCreateSecurityGroupReturn:    ApiCallReturn{"mock_id", nil},
			expectedFindSecurityGroupRulesReturn: ApiCallReturn{[]string{}, nil},
			expectedDeleteRules:                  ApiCallReturn{nil, nil},
			expectedReturn:                       ApiCallReturn{nil, nil},
		},
		{
			name:                                 "prepare_security_group execute error: get security_group error",
			preparedSecurityGroupId:              "",
			expectedGetSecurityGroupReturn:       ApiCallReturn{"", errorCall},
			expectedCreateSecurityGroupReturn:    ApiCallReturn{"mock_id", nil},
			expectedFindSecurityGroupRulesReturn: ApiCallReturn{[]string{}, nil},
			expectedDeleteRules:                  ApiCallReturn{nil, nil},
			expectedReturn:                       ApiCallReturn{nil, errorCall},
		},
		{
			name:                                 "prepare_security_group execute error: findSecurityGroupRules error",
			preparedSecurityGroupId:              "",
			expectedGetSecurityGroupReturn:       ApiCallReturn{"", nil},
			expectedCreateSecurityGroupReturn:    ApiCallReturn{"mock_id", nil},
			expectedFindSecurityGroupRulesReturn: ApiCallReturn{[]string{}, errorCall},
			expectedDeleteRules:                  ApiCallReturn{nil, nil},
			expectedReturn:                       ApiCallReturn{nil, errorCall},
		},
		{
			name:                                 "prepare_security_group execute error: deleteRules error",
			preparedSecurityGroupId:              "",
			expectedGetSecurityGroupReturn:       ApiCallReturn{"", nil},
			expectedCreateSecurityGroupReturn:    ApiCallReturn{"mock_id", nil},
			expectedFindSecurityGroupRulesReturn: ApiCallReturn{[]string{}, nil},
			expectedDeleteRules:                  ApiCallReturn{nil, errorCall},
			expectedReturn:                       ApiCallReturn{nil, errorCall},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDirecter := mockdirecter.NewMockDirecter(mockCtrl)
			task := PrepareSecurityGroupTask{
				components.BaseTask{
					Logger:   logger.NewDebugLogger(),
					Directer: mockDirecter,
				},
			}

			mockConfig := mockconfig.NewMockConfig(mockCtrl)
			context := &directer.WorkflowContext{Config: mockConfig}

			mockDirecter.EXPECT().GetContext().Return(context).MaxTimes(8)
			mockDirecter.EXPECT().Process(gomock.Any())

			// record
			mockConfig.EXPECT().Get(directer.WfKeyResourceProjectId).Return(&config.Entry{Val: "mock_res_pid"})
			mockConfig.EXPECT().Get(directer.WfKeyRegion).Return(&config.Entry{Val: "mock_region"})
			mockConfig.EXPECT().Get(directer.WfKeySecurityGroupName).Return(&config.Entry{Val: "mock_vpc_name"})
			mockConfig.EXPECT().Get(directer.WfKeyResourceAgencyName).Return(&config.Entry{Val: "mock_res_agency_name"})
			mockConfig.EXPECT().Get(directer.WfKeyResourceDomainId).Return(&config.Entry{Val: "mock_res_did"})
			mockConfig.EXPECT().Get(directer.WfKeySecurityGroupId).Return(&config.Entry{Val: tt.preparedSecurityGroupId})
			if tt.expectedGetSecurityGroupReturn.err == nil && tt.expectedFindSecurityGroupRulesReturn.err == nil &&
				tt.expectedDeleteRules.err == nil {
				if tt.preparedSecurityGroupId == "" {
					mockConfig.EXPECT().Set(directer.WfKeySecurityGroupId, tt.expectedCreateSecurityGroupReturn.value.(string)).Return(nil).Times(1)
				} else {
					mockConfig.EXPECT().Set(directer.WfKeySecurityGroupId, tt.preparedSecurityGroupId).Return(nil).Times(1)
				}
			}

			getSecurityGroup = func(regionId string, projectId string, agencyName string, resDomainId string,
				vpcName string) (result string, err error) {
				result = tt.expectedGetSecurityGroupReturn.value.(string)
				err = tt.expectedGetSecurityGroupReturn.err
				return
			}

			createSecurityGroup = func(regionId string, projectId string, agencyName string, resDomainId string,
				securityGroupName string, enterpriseProject string) (result string, err error) {
				result = tt.expectedCreateSecurityGroupReturn.value.(string)
				err = tt.expectedCreateSecurityGroupReturn.err
				return
			}

			findSecurityGroupRules = func(regionId string, projectId string, agencyName string, resDomainId string,
				securityGroupId string, filtedGroupRule string) (result *[]string, err error) {
				tmp := tt.expectedFindSecurityGroupRulesReturn.value.([]string)
				result = &tmp
				err = tt.expectedFindSecurityGroupRulesReturn.err
				return
			}

			deleteRules = func(regionId string, projectId string, agencyName string, resDomainId string,
				ruleIds []string) (err error) {
				err = tt.expectedDeleteRules.err
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

func TestCreateSecurityGroup(t *testing.T) {
	originClientGetAgencyVpcClient := clientGetAgencyVpcClient
	originCreateSecurityGroup := createSecurityGroup
	originClientCreateSecurityGroup := clientCreateSecurityGroup
	defer func() {
		clientGetAgencyVpcClient = originClientGetAgencyVpcClient
		createSecurityGroup = originCreateSecurityGroup
		clientCreateSecurityGroup = originClientCreateSecurityGroup
	}()

	errorCall := errors.NewError("error call")
	tests := []struct {
		name                                   string
		regionId                               string
		projectId                              string
		agencyName                             string
		resDomainId                            string
		securityGroupName                      string
		enterpriseProject                      string
		expectedGetAgencyVpcClientReturnClient *vpc.VpcClient
		expectedGetAgencyVpcClientReturnError  error
		expectedCreateSecurityGroupReturn      ApiCallReturn
	}{
		{
			name:                                   "create_security_group execute success: with no error calls",
			expectedGetAgencyVpcClientReturnClient: nil,
			expectedGetAgencyVpcClientReturnError:  nil,
			expectedCreateSecurityGroupReturn:      ApiCallReturn{"", nil},
		},
		{
			name:                                   "create_security_group execute success: with client error calls",
			expectedGetAgencyVpcClientReturnClient: nil,
			expectedGetAgencyVpcClientReturnError:  errorCall,
			expectedCreateSecurityGroupReturn:      ApiCallReturn{"", errorCall},
		},
		{
			name:                                   "create_security_group execute success: with create security group error calls",
			expectedGetAgencyVpcClientReturnClient: nil,
			expectedGetAgencyVpcClientReturnError:  nil,
			expectedCreateSecurityGroupReturn:      ApiCallReturn{"", errorCall},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientGetAgencyVpcClient = func(regionId string, projectId string, agencyName string,
				resDomainId string) (*vpc.VpcClient, error) {
				return tt.expectedGetAgencyVpcClientReturnClient, tt.expectedGetAgencyVpcClientReturnError
			}

			clientCreateSecurityGroup = func(vpcClient client.VpcV2, name *string, enterpriseProject *string) (string, error) {
				return tt.expectedCreateSecurityGroupReturn.value.(string), tt.expectedCreateSecurityGroupReturn.err
			}

			s1, e1 := createSecurityGroup(tt.regionId, tt.projectId, tt.agencyName, tt.resDomainId, tt.securityGroupName, tt.enterpriseProject)
			if tt.expectedCreateSecurityGroupReturn.value != s1 {
				t.Errorf("expected return %v, got %v", tt.expectedCreateSecurityGroupReturn.value, s1)
			}
			if tt.expectedCreateSecurityGroupReturn.err != e1 {
				t.Errorf("expected return %v, got %v", tt.expectedCreateSecurityGroupReturn.err, e1)
			}
		})
	}
}

func TestDeleteRules(t *testing.T) {
	originClientGetAgencyVpcClient := clientGetAgencyVpcClient
	originDeleteRules := deleteRules
	originClientDeleteSecurityGroupRule := clientDeleteSecurityGroupRule
	defer func() {
		clientGetAgencyVpcClient = originClientGetAgencyVpcClient
		deleteRules = originDeleteRules
		clientDeleteSecurityGroupRule = originClientDeleteSecurityGroupRule
	}()

	errorCall := errors.NewError("error call")
	tests := []struct {
		name                                   string
		regionId                               string
		projectId                              string
		agencyName                             string
		resDomainId                            string
		ruleIds                                []string
		expectedGetAgencyVpcClientReturnClient *vpc.VpcClient
		expectedGetAgencyVpcClientReturnError  error
		expectedDeleteSecurityGroupRuleError   error
		expectedDeleteRulesReturn              ApiCallReturn
	}{
		{
			name:                                   "delete_security_group_rules execute success: with no error calls",
			expectedGetAgencyVpcClientReturnClient: nil,
			expectedGetAgencyVpcClientReturnError:  nil,
			expectedDeleteSecurityGroupRuleError:   nil,
			expectedDeleteRulesReturn:              ApiCallReturn{"", nil},
		},
		{
			name:                                   "delete_security_group_rules execute success: with client error calls",
			expectedGetAgencyVpcClientReturnClient: nil,
			expectedGetAgencyVpcClientReturnError:  errorCall,
			expectedDeleteSecurityGroupRuleError:   nil,
			expectedDeleteRulesReturn:              ApiCallReturn{"", errorCall},
		},
		{
			name:                                   "delete_security_group_rules execute success: with delete security group rules error calls",
			ruleIds:                                []string{"ruleId"},
			expectedGetAgencyVpcClientReturnClient: nil,
			expectedGetAgencyVpcClientReturnError:  nil,
			expectedDeleteSecurityGroupRuleError:   errorCall,
			expectedDeleteRulesReturn:              ApiCallReturn{"", errorCall},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientGetAgencyVpcClient = func(regionId string, projectId string, agencyName string,
				resDomainId string) (*vpc.VpcClient, error) {
				return tt.expectedGetAgencyVpcClientReturnClient, tt.expectedGetAgencyVpcClientReturnError
			}

			clientDeleteSecurityGroupRule = func(vpcClient *vpc.VpcClient, id string) error {
				return tt.expectedDeleteSecurityGroupRuleError
			}

			e1 := deleteRules(tt.regionId, tt.projectId, tt.agencyName, tt.resDomainId, tt.ruleIds)
			if tt.expectedDeleteRulesReturn.err != e1 {
				t.Errorf("expected return %v, got %v", tt.expectedDeleteRulesReturn.err, e1)
			}
		})
	}
}

func TestFindSecurityGroupRules(t *testing.T) {
	originClientGetAgencyVpcClient := clientGetAgencyVpcClient
	originFindSecurityGroupRules := findSecurityGroupRules
	originClientGetSecurityGroupRules := clientGetSecurityGroupRules
	defer func() {
		clientGetAgencyVpcClient = originClientGetAgencyVpcClient
		findSecurityGroupRules = originFindSecurityGroupRules
		clientGetSecurityGroupRules = originClientGetSecurityGroupRules
	}()

	errorCall := errors.NewError("error call")
	tests := []struct {
		name                                   string
		regionId                               string
		projectId                              string
		agencyName                             string
		resDomainId                            string
		securityGroupId                        string
		filtedGroupRule                        string
		expectedGetAgencyVpcClientReturnClient *vpc.VpcClient
		expectedGetAgencyVpcClientReturnError  error
		expectedGetSecurityGroupRules          ApiCallReturn
		expectedFindSecurityGroupRulesValue    []string
		expectedFindSecurityGroupRulesError    error
	}{
		{
			name:                                   "findSecurityGroupRules execute success: with no error calls",
			expectedGetAgencyVpcClientReturnClient: nil,
			expectedGetAgencyVpcClientReturnError:  nil,
			expectedGetSecurityGroupRules: ApiCallReturn{[]vpcmodel.SecurityGroupRule{
				{Id: "ruleID"},
			}, nil},
			expectedFindSecurityGroupRulesValue: []string{"ruleID"},
			expectedFindSecurityGroupRulesError: nil,
		},
		{
			name:                                   "findSecurityGroupRules execute success: with client error calls",
			expectedGetAgencyVpcClientReturnClient: nil,
			expectedGetAgencyVpcClientReturnError:  errorCall,
			expectedGetSecurityGroupRules:          ApiCallReturn{"", nil},
			expectedFindSecurityGroupRulesValue:    []string{},
			expectedFindSecurityGroupRulesError:    errorCall,
		},
		{
			name:                                   "findSecurityGroupRules execute success: with get security group rules error calls",
			expectedGetAgencyVpcClientReturnClient: nil,
			expectedGetAgencyVpcClientReturnError:  nil,
			expectedGetSecurityGroupRules: ApiCallReturn{[]vpcmodel.SecurityGroupRule{
				{Id: "ruleID"},
			}, errorCall},
			expectedFindSecurityGroupRulesValue: []string{},
			expectedFindSecurityGroupRulesError: errorCall,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientGetAgencyVpcClient = func(regionId string, projectId string, agencyName string,
				resDomainId string) (*vpc.VpcClient, error) {
				return tt.expectedGetAgencyVpcClientReturnClient, tt.expectedGetAgencyVpcClientReturnError
			}

			clientGetSecurityGroupRules = func(vpcClient client.VpcV2, securityGroupId *string) (*[]vpcmodel.SecurityGroupRule, error) {
				ruleRsp := tt.expectedGetSecurityGroupRules.value.([]vpcmodel.SecurityGroupRule)
				return &ruleRsp, tt.expectedGetSecurityGroupRules.err
			}

			s1, e1 := findSecurityGroupRules(tt.regionId, tt.projectId, tt.agencyName, tt.resDomainId, tt.securityGroupId, tt.filtedGroupRule)
			for key, value := range tt.expectedFindSecurityGroupRulesValue {
				if (*s1)[key] != value {
					t.Errorf("expected return %v, got %v", &tt.expectedFindSecurityGroupRulesValue, s1)
				}
			}
			if tt.expectedFindSecurityGroupRulesError != e1 {
				t.Errorf("expected return %v, got %v", tt.expectedFindSecurityGroupRulesError, e1)
			}
		})
	}
}
