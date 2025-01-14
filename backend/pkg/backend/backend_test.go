// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backend

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/golang/mock/gomock"
	"gotest.tools/assert"

	"github.com/openclarity/apiclarity/api/server/models"
	_database "github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/k8smonitor"
	"github.com/openclarity/apiclarity/backend/pkg/modules"
	pluginsmodels "github.com/openclarity/apiclarity/plugins/api/server/models"
	_spec "github.com/openclarity/speculator/pkg/spec"
	_speculator "github.com/openclarity/speculator/pkg/speculator"
)

func Test_isNonAPI(t *testing.T) {
	type args struct {
		trace *_spec.Telemetry
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "content type is not application/json expected to classify as non API",
			args: args{
				trace: &_spec.Telemetry{
					Response: &_spec.Response{
						Common: &_spec.Common{
							Headers: []*_spec.Header{
								{
									Key:   contentTypeHeaderName,
									Value: "non-api",
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "REST API",
			args: args{
				trace: &_spec.Telemetry{
					Response: &_spec.Response{
						Common: &_spec.Common{
							Headers: []*_spec.Header{
								{
									Key:   contentTypeHeaderName,
									Value: contentTypeApplicationJSON,
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "no headers expected to classify as API",
			args: args{
				trace: &_spec.Telemetry{
					Response: &_spec.Response{
						Common: &_spec.Common{},
					},
				},
			},
			want: false,
		},
		{
			name: "content type is application/hal+json - classify as API",
			args: args{
				trace: &_spec.Telemetry{
					Response: &_spec.Response{
						Common: &_spec.Common{
							Headers: []*_spec.Header{
								{
									Key:   contentTypeHeaderName,
									Value: "application/hal+json",
								},
							},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNonAPI(tt.args.trace); got != tt.want {
				t.Errorf("isNonAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getHostname(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "no scheme",
			args: args{
				host: "example.com:8080",
			},
			want: "example.com",
		},
		{
			name: "with scheme",
			args: args{
				host: "acap://example.com:8080",
			},
			want: "example.com",
		},
		{
			name: "only host",
			args: args{
				host: "example.com",
			},
			want: "example.com",
		},
		{
			name: "hostname is empty",
			args: args{
				host: "https://",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "failed to parse host",
			args: args{
				host: "1 2 3",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getHostname(tt.args.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHostname() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getHostname() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertAPIDiffType(t *testing.T) {
	type args struct {
		diffType _spec.DiffType
	}
	tests := []struct {
		name string
		args args
		want models.DiffType
	}{
		{
			name: "unknown type - default DiffTypeNODIFF",
			args: args{
				diffType: "unknown type",
			},
			want: models.DiffTypeNODIFF,
		},
		{
			name: "DiffTypeNoDiff",
			args: args{
				diffType: _spec.DiffTypeNoDiff,
			},
			want: models.DiffTypeNODIFF,
		},
		{
			name: "DiffTypeZombieDiff",
			args: args{
				diffType: _spec.DiffTypeZombieDiff,
			},
			want: models.DiffTypeZOMBIEDIFF,
		},
		{
			name: "DiffTypeShadowDiff",
			args: args{
				diffType: _spec.DiffTypeShadowDiff,
			},
			want: models.DiffTypeSHADOWDIFF,
		},
		{
			name: "DiffTypeGeneralDiff",
			args: args{
				diffType: _spec.DiffTypeGeneralDiff,
			},
			want: models.DiffTypeGENERALDIFF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertAPIDiffType(tt.args.diffType); got != tt.want {
				t.Errorf("convertAPIDiffType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getHighestPrioritySpecDiffType(t *testing.T) {
	type args struct {
		providedDiff      models.DiffType
		reconstructedDiff models.DiffType
	}
	tests := []struct {
		name string
		args args
		want models.DiffType
	}{
		{
			name: "Zombie over Shadow",
			args: args{
				providedDiff:      models.DiffTypeZOMBIEDIFF,
				reconstructedDiff: models.DiffTypeSHADOWDIFF,
			},
			want: models.DiffTypeZOMBIEDIFF,
		},
		{
			name: "Same type",
			args: args{
				providedDiff:      models.DiffTypeGENERALDIFF,
				reconstructedDiff: models.DiffTypeGENERALDIFF,
			},
			want: models.DiffTypeGENERALDIFF,
		},
		{
			name: "reconstructed unknown type",
			args: args{
				providedDiff:      models.DiffTypeNODIFF,
				reconstructedDiff: "unknown type",
			},
			want: models.DiffTypeNODIFF,
		},
		{
			name: "provided unknown type",
			args: args{
				providedDiff:      "unknown type",
				reconstructedDiff: models.DiffTypeNODIFF,
			},
			want: models.DiffTypeNODIFF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHighestPrioritySpecDiffType(tt.args.providedDiff, tt.args.reconstructedDiff); got != tt.want {
				t.Errorf("getHighestPrioritySpecDiffType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackend_handleHTTPTrace(t *testing.T) {
	mockCtrlDatabase := gomock.NewController(t)
	defer mockCtrlDatabase.Finish()
	mockDatabase := _database.NewMockDatabase(mockCtrlDatabase)

	mockCtrlAPIEventTable := gomock.NewController(t)
	defer mockCtrlAPIEventTable.Finish()
	mockAPIEventTable := _database.NewMockAPIEventsTable(mockCtrlAPIEventTable)

	mockCtrlAPIInventoryTable := gomock.NewController(t)
	defer mockCtrlAPIInventoryTable.Finish()
	mockAPIInventoryTable := _database.NewMockAPIInventoryTable(mockCtrlAPIInventoryTable)

	mockCtrlModules := gomock.NewController(t)
	defer mockCtrlModules.Finish()
	mockModules := modules.NewMockModule(mockCtrlModules)

	speculatorWithProvidedSpec := _speculator.CreateSpeculator(_speculator.Config{})
	speculatorWithProvidedSpec.Specs[specKey] = _spec.CreateDefaultSpec(host, port, _spec.OperationGeneratorConfig{})
	err := speculatorWithProvidedSpec.LoadProvidedSpec(specKey, []byte(providedSpec), map[string]string{})
	assert.NilError(t, err)

	speculatorWithApprovedSpec := _speculator.CreateSpeculator(_speculator.Config{})
	speculatorWithApprovedSpec.Specs[specKey] = _spec.CreateDefaultSpec(host, port, _spec.OperationGeneratorConfig{})
	ApprovedSpecReview := &_spec.ApprovedSpecReview{
		PathToPathItem: map[string]*spec.PathItem{
			"/api/1/foo": &_spec.NewTestPathItem().WithOperation(http.MethodPost, nil).PathItem,
			"/api/2/foo": &_spec.NewTestPathItem().WithOperation(http.MethodGet, nil).PathItem,
		},
		PathItemsReview: []*_spec.ApprovedSpecReviewPathItem{
			{
				ReviewPathItem: _spec.ReviewPathItem{
					ParameterizedPath: "/api/{param1}/foo",
					Paths:             map[string]bool{"/api/1/foo": true, "/api/2/foo": true},
				},
			},
		},
	}
	err = speculatorWithApprovedSpec.ApplyApprovedReview(specKey, ApprovedSpecReview)
	assert.NilError(t, err)

	type fields struct {
		speculator              *_speculator.Speculator
		monitor                 *k8smonitor.Monitor
		dbHandler               _database.Database
		expectDatabase          func(database *_database.MockDatabase)
		expectAPIEventTable     func(apiEventTable *_database.MockAPIEventsTable)
		expectAPIInventoryTable func(apiInventoryTable *_database.MockAPIInventoryTable)
	}
	type args struct {
		trace *pluginsmodels.Telemetry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "good run",
			fields: fields{
				speculator: _speculator.CreateSpeculator(_speculator.Config{}),
				monitor:    nil, // TODO turn monitor into interface so we can use it in tests. for now we assume to run locally (no monitor)
				dbHandler:  mockDatabase,
				expectDatabase: func(database *_database.MockDatabase) {
					database.EXPECT().APIInventoryTable().Return(mockAPIInventoryTable)
					database.EXPECT().APIEventsTable().Return(mockAPIEventTable)
				},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {
					apiInventoryTable.EXPECT().FirstOrCreate(gomock.Any())
				},
				expectAPIEventTable: func(apiEventTable *_database.MockAPIEventsTable) {
					apiEventTable.EXPECT().CreateAPIEvent(NewEventMatcher(createDefaultTestEvent().event))
				},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   host,
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: false,
		},
		{
			name: "Host field is empty, get host from headers",
			fields: fields{
				speculator: _speculator.CreateSpeculator(_speculator.Config{}),
				monitor:    nil,
				dbHandler:  mockDatabase,
				expectDatabase: func(database *_database.MockDatabase) {
					database.EXPECT().APIInventoryTable().Return(mockAPIInventoryTable)
					database.EXPECT().APIEventsTable().Return(mockAPIEventTable)
				},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {
					apiInventoryTable.EXPECT().FirstOrCreate(gomock.Any())
				},
				expectAPIEventTable: func(apiEventTable *_database.MockAPIEventsTable) {
					apiEventTable.EXPECT().CreateAPIEvent(NewEventMatcher(createDefaultTestEvent().event))
				},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers: []*pluginsmodels.Header{
								{
									Key:   "host",
									Value: host,
								},
							},
							Time:    0,
							Version: "1.1",
						},
						Host:   "",
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: false,
		},
		{
			name: "no host name found",
			fields: fields{
				speculator:              _speculator.CreateSpeculator(_speculator.Config{}),
				monitor:                 nil,
				dbHandler:               mockDatabase,
				expectDatabase:          func(database *_database.MockDatabase) {},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {},
				expectAPIEventTable:     func(apiEventTable *_database.MockAPIEventsTable) {},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   "",
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid destination address",
			fields: fields{
				speculator:              _speculator.CreateSpeculator(_speculator.Config{}),
				monitor:                 nil,
				dbHandler:               mockDatabase,
				expectDatabase:          func(database *_database.MockDatabase) {},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {},
				expectAPIEventTable:     func(apiEventTable *_database.MockAPIEventsTable) {},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   "1.1.1.1",
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   host,
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid source address",
			fields: fields{
				speculator:              _speculator.CreateSpeculator(_speculator.Config{}),
				monitor:                 nil,
				dbHandler:               mockDatabase,
				expectDatabase:          func(database *_database.MockDatabase) {},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {},
				expectAPIEventTable:     func(apiEventTable *_database.MockAPIEventsTable) {},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   host,
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2",
				},
			},
			wantErr: true,
		},
		{
			name: "non api",
			fields: fields{
				speculator: _speculator.CreateSpeculator(_speculator.Config{}),
				monitor:    nil,
				dbHandler:  mockDatabase,
				expectDatabase: func(database *_database.MockDatabase) {
					database.EXPECT().APIEventsTable().Return(mockAPIEventTable)
				},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {},
				expectAPIEventTable: func(apiEventTable *_database.MockAPIEventsTable) {
					apiEventTable.EXPECT().CreateAPIEvent(NewEventMatcher(createDefaultTestEvent().WithIsNonAPI(true).event))
				},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   host,
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers: []*pluginsmodels.Header{
								{
									Key:   contentTypeHeaderName,
									Value: "xml",
								},
							},
							Time:    0,
							Version: "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: false,
		},
		{
			name: "has provided spec diff",
			fields: fields{
				speculator: speculatorWithProvidedSpec,
				monitor:    nil,
				dbHandler:  mockDatabase,
				expectDatabase: func(database *_database.MockDatabase) {
					database.EXPECT().APIInventoryTable().Return(mockAPIInventoryTable)
					database.EXPECT().APIEventsTable().Return(mockAPIEventTable)
				},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {
					apiInventoryTable.EXPECT().FirstOrCreate(gomock.Any())
				},
				expectAPIEventTable: func(apiEventTable *_database.MockAPIEventsTable) {
					apiEventTable.EXPECT().CreateAPIEvent(NewEventMatcher(createDefaultTestEvent().WithHasProvidedSpecDiff(true).WithSpecDiffType(models.DiffTypeSHADOWDIFF).event))
				},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   host,
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: false,
		},
		{
			name: "has reconstructed spec diff",
			fields: fields{
				speculator: speculatorWithApprovedSpec,
				monitor:    nil,
				dbHandler:  mockDatabase,
				expectDatabase: func(database *_database.MockDatabase) {
					database.EXPECT().APIInventoryTable().Return(mockAPIInventoryTable)
					database.EXPECT().APIEventsTable().Return(mockAPIEventTable)
				},
				expectAPIInventoryTable: func(apiInventoryTable *_database.MockAPIInventoryTable) {
					apiInventoryTable.EXPECT().FirstOrCreate(gomock.Any())
				},
				expectAPIEventTable: func(apiEventTable *_database.MockAPIEventsTable) {
					apiEventTable.EXPECT().CreateAPIEvent(NewEventMatcher(createDefaultTestEvent().WithHasReconstructedSpecDiff(true).WithSpecDiffType(models.DiffTypeSHADOWDIFF).event))
				},
			},
			args: args{
				trace: &pluginsmodels.Telemetry{
					DestinationAddress:   destinationAddress,
					DestinationNamespace: "foo",
					Request: &pluginsmodels.Request{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						Host:   host,
						Method: "GET",
						Path:   "/test?foo=bar",
					},
					RequestID: "1",
					Response: &pluginsmodels.Response{
						Common: &pluginsmodels.Common{
							TruncatedBody: false,
							Body:          []byte{},
							Headers:       []*pluginsmodels.Header{},
							Time:          0,
							Version:       "1.1",
						},
						StatusCode: "200",
					},
					Scheme:        "http",
					SourceAddress: "2.2.2.2:80",
				},
			},
			wantErr: false,
		},
	}
	ctx := context.Background()

	for _, tt := range tests {
		tt.fields.expectDatabase(mockDatabase)
		tt.fields.expectAPIInventoryTable(mockAPIInventoryTable)
		tt.fields.expectAPIEventTable(mockAPIEventTable)
		mockModules.EXPECT().EventNotify(ctx, gomock.Any()).AnyTimes()
		t.Run(tt.name, func(t *testing.T) {
			b := &Backend{
				speculator: tt.fields.speculator,
				monitor:    tt.fields.monitor,
				dbHandler:  tt.fields.dbHandler,
				modules:    mockModules,
			}
			if err := b.handleHTTPTrace(ctx, tt.args.trace); (err != nil) != tt.wantErr {
				t.Errorf("handleHTTPTrace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
