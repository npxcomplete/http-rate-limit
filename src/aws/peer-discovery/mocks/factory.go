package peer_discovery

import (
	"reflect"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
)

// Mock implementations generated manually to avoid using mockgen in this environment.

type MockMetadataAPI struct {
	ctrl     *gomock.Controller
	recorder *MockMetadataAPIMockRecorder
}

type MockMetadataAPIMockRecorder struct {
	mock *MockMetadataAPI
}

func NewMockMetadataAPI(ctrl *gomock.Controller) *MockMetadataAPI {
	mock := &MockMetadataAPI{ctrl: ctrl}
	mock.recorder = &MockMetadataAPIMockRecorder{mock}
	return mock
}

func (m *MockMetadataAPI) EXPECT() *MockMetadataAPIMockRecorder {
	return m.recorder
}

func (m *MockMetadataAPI) GetMetadata(path string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetadata", path)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockMetadataAPIMockRecorder) GetMetadata(path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetadata", reflect.TypeOf((*MockMetadataAPI)(nil).GetMetadata), path)
}

// AutoScaling

type MockAutoScalingAPI struct {
	ctrl     *gomock.Controller
	recorder *MockAutoScalingAPIMockRecorder
}

type MockAutoScalingAPIMockRecorder struct {
	mock *MockAutoScalingAPI
}

func NewMockAutoScalingAPI(ctrl *gomock.Controller) *MockAutoScalingAPI {
	mock := &MockAutoScalingAPI{ctrl: ctrl}
	mock.recorder = &MockAutoScalingAPIMockRecorder{mock}
	return mock
}

func (m *MockAutoScalingAPI) EXPECT() *MockAutoScalingAPIMockRecorder {
	return m.recorder
}

func (m *MockAutoScalingAPI) DescribeAutoScalingInstances(in *autoscaling.DescribeAutoScalingInstancesInput) (*autoscaling.DescribeAutoScalingInstancesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeAutoScalingInstances", in)
	ret0, _ := ret[0].(*autoscaling.DescribeAutoScalingInstancesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockAutoScalingAPIMockRecorder) DescribeAutoScalingInstances(in interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeAutoScalingInstances", reflect.TypeOf((*MockAutoScalingAPI)(nil).DescribeAutoScalingInstances), in)
}

func (m *MockAutoScalingAPI) DescribeAutoScalingGroups(in *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeAutoScalingGroups", in)
	ret0, _ := ret[0].(*autoscaling.DescribeAutoScalingGroupsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockAutoScalingAPIMockRecorder) DescribeAutoScalingGroups(in interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeAutoScalingGroups", reflect.TypeOf((*MockAutoScalingAPI)(nil).DescribeAutoScalingGroups), in)
}

// EC2

type MockEC2API struct {
	ctrl     *gomock.Controller
	recorder *MockEC2APIMockRecorder
}

type MockEC2APIMockRecorder struct {
	mock *MockEC2API
}

func NewMockEC2API(ctrl *gomock.Controller) *MockEC2API {
	mock := &MockEC2API{ctrl: ctrl}
	mock.recorder = &MockEC2APIMockRecorder{mock}
	return mock
}

func (m *MockEC2API) EXPECT() *MockEC2APIMockRecorder {
	return m.recorder
}

func (m *MockEC2API) DescribeInstances(in *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeInstances", in)
	ret0, _ := ret[0].(*ec2.DescribeInstancesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockEC2APIMockRecorder) DescribeInstances(in interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeInstances", reflect.TypeOf((*MockEC2API)(nil).DescribeInstances), in)
}

// Factory

type MockClientFactory struct {
	ctrl     *gomock.Controller
	recorder *MockClientFactoryMockRecorder
}

type MockClientFactoryMockRecorder struct {
	mock *MockClientFactory
}

func NewMockClientFactory(ctrl *gomock.Controller) *MockClientFactory {
	mock := &MockClientFactory{ctrl: ctrl}
	mock.recorder = &MockClientFactoryMockRecorder{mock}
	return mock
}

func (m *MockClientFactory) EXPECT() *MockClientFactoryMockRecorder {
	return m.recorder
}

func (m *MockClientFactory) Metadata() metadataAPI {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Metadata")
	ret0, _ := ret[0].(metadataAPI)
	return ret0
}

func (mr *MockClientFactoryMockRecorder) Metadata() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Metadata", reflect.TypeOf((*MockClientFactory)(nil).Metadata))
}

func (m *MockClientFactory) AutoScaling() autoscalingAPI {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AutoScaling")
	ret0, _ := ret[0].(autoscalingAPI)
	return ret0
}

func (mr *MockClientFactoryMockRecorder) AutoScaling() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AutoScaling", reflect.TypeOf((*MockClientFactory)(nil).AutoScaling))
}

func (m *MockClientFactory) EC2() ec2API {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EC2")
	ret0, _ := ret[0].(ec2API)
	return ret0
}

func (mr *MockClientFactoryMockRecorder) EC2() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EC2", reflect.TypeOf((*MockClientFactory)(nil).EC2))
}
