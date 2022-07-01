// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package fuzzer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	oapicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/api3/global"
	"github.com/openclarity/apiclarity/api3/notifications"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/clients"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/config"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/logging"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/model"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/restapi"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/tools"
)

const (
	ModuleDescription = "Runs a set of tests against API endpoints to discover insecure implementations"
	ModuleVersion     = "0.0.0"
	EmptyJSON         = "{}"
	NbMaxServicePart  = 2
)

type pluginFuzzer struct {
	runningState restapi.TestingModuleState
	httpHandler  http.Handler

	config       *config.Config
	model        *model.Model
	fuzzerClient clients.Client

	accessor core.BackendAccessor
	info     *core.ModuleInfo
}

//nolint:gochecknoinits // was needed for the module implementation of ApiClarity
func init() {
	core.RegisterModule(newFuzzer)
}

//nolint:ireturn,nolintlint // was needed for the module implementation of ApiClarity
func newFuzzer(ctx context.Context, moduleName string, accessor core.BackendAccessor) (core.Module, error) {
	logging.InitLogger()
	logging.Logf("[Fuzzer] Start():: -->")

	// Use default values
	plugin := pluginFuzzer{
		httpHandler:  nil,
		runningState: restapi.TestingModuleState{APIsInCache: 0, Version: ModuleVersion},
		config:       config.GetConfig(),
		fuzzerClient: nil,
		model:        nil,
		accessor:     accessor,
		info: &core.ModuleInfo{
			Name:        moduleName,
			Description: ModuleDescription,
		},
	}

	plugin.config.Dump()

	plugin.httpHandler = restapi.HandlerWithOptions(&pluginFuzzerHTTPHandler{fuzzer: &plugin}, restapi.ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + moduleName})

	// Initialize the model
	plugin.model = model.NewModel(accessor)
	err := plugin.model.Init(ctx)
	if err != nil {
		logging.Errorf("[Fuzzer] Error, failed to init model.")
	}
	logging.Logf("[Fuzzer] Model creation ok")

	// Create the client according to the configuration
	plugin.fuzzerClient, err = clients.NewClient(plugin.config, accessor)
	if err != nil {
		logging.Errorf("[testing] Error, failed to create a client")
		return nil, fmt.Errorf("ignoring fuzzer module due to missing fuzzer client")
	}
	logging.Logf("[testing] Client (%v) creation, ok", plugin.config.GetDeploymentType())

	logging.Logf("[Fuzzer] Start():: <--")

	return &plugin, nil
}

func (p *pluginFuzzer) Info() core.ModuleInfo {
	return *p.info
}

func (p *pluginFuzzer) EventNotify(ctx context.Context, event *core.Event) {
	// Fuzzer doesn't use this
	// Logf("[Fuzzer] received a new trace for API(%s) EventID(%v)", event.APIInfoID, event.ID)
}

/*
*
*  Manage notifications
*
 */

func (p *pluginFuzzer) sendAPIFindingsNotification(ctx context.Context, apiID uint, findings []oapicommon.APIFinding) error {
	apiFindingsNotification := notifications.ApiFindingsNotification{
		NotificationType: "ApiFindingsNotification",
		Items:            &findings,
	}
	notification := notifications.APIClarityNotification{}
	err := notification.FromApiFindingsNotification(apiFindingsNotification)
	if err != nil {
		return fmt.Errorf("failed to create 'APIFindings' notification, err=(%v)", err)
	}

	err = p.accessor.Notify(ctx, p.info.Name, apiID, notification)

	return err //nolint:wrapcheck // really want to return the result of the notify
}

func (p *pluginFuzzer) sendTestReportNotification(ctx context.Context, apiID uint, report restapi.ShortTestReport) error {
	globalReportTags := tools.ConvertLocalToGlobalReportTag(report.Tags)
	testReportNotification := notifications.TestReportNotification{
		ApiID:            report.ApiID,
		HighestSeverity:  report.HighestSeverity,
		NotificationType: "TestReportNotification",
		Starttime:        report.Starttime,
		Status:           global.FuzzingStatusEnum(report.Status),
		StatusMessage:    report.StatusMessage,
		Tags:             globalReportTags,
	}
	notification := notifications.APIClarityNotification{}
	err := notification.FromTestReportNotification(testReportNotification)
	if err != nil {
		return fmt.Errorf("failed to create 'TestReport' notification, err=(%v)", err)
	}

	err = p.accessor.Notify(ctx, p.info.Name, apiID, notification)

	return err //nolint:wrapcheck // really want to return the result of the notify
}

func (p *pluginFuzzer) sendTestProgressNotification(ctx context.Context, apiID uint, report restapi.ShortTestProgress) error {
	testProgressNotification := notifications.TestProgressNotification{
		ApiID:            report.ApiID,
		NotificationType: "TestProgressNotification",
		Progress:         report.Progress,
		Starttime:        report.Starttime,
	}
	notification := notifications.APIClarityNotification{}
	err := notification.FromTestProgressNotification(testProgressNotification)
	if err != nil {
		return fmt.Errorf("failed to create 'TestProgress' notification, err=(%v)", err)
	}

	err = p.accessor.Notify(ctx, p.info.Name, apiID, notification)

	return err //nolint:wrapcheck // really want to return the result of the notify
}

/*
*
*  Implement Fuzzer module stuff
*
 */

func (p *pluginFuzzer) FuzzTarget(ctx context.Context, apiID oapicommon.ApiID, params *model.FuzzingInput) (model.FuzzingTimestamp, error) {
	// Checks
	if p.fuzzerClient == nil {
		return model.ZeroTime, &PluginError{"No deployment client running"}
	}
	if params == nil {
		return model.ZeroTime, &InvalidParameterError{"No input parameter"}
	}

	// Retrieve the API (it will give the endpoint and the port)
	api, err := p.model.GetAPI(ctx, uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] FuzzTarget(): can't retrieve API (%v)", apiID)
		return model.ZeroTime, &NotFoundError{msg: ""}
	}

	logging.Logf("[Fuzzer] FuzzTarget(): API_id (%v) => API (%v)", apiID, api)

	// Construct the URI of the enpoint to fuzz
	serviceToTest := api.Name
	if len(api.Namespace) > 0 {
		serviceToTest = fmt.Sprintf("%s.%s", serviceToTest, api.Namespace)
	}
	fullServiceURI := fmt.Sprintf("http://%s:%v", serviceToTest, api.Port)

	// Get auth material, if provided
	securityParam, err := tools.GetAuthStringFromParam(params.Auth)
	if err != nil {
		logging.Errorf("[Fuzzer] FuzzTarget(): can't get auth material for (%v)", apiID)
		return model.ZeroTime, &InvalidParameterError{msg: err.Error()}
	}

	// Get time budget
	timeBudget, err := tools.GetTimeBudgetFromParam(params.Depth)
	if err != nil {
		logging.Errorf("[Fuzzer] FuzzTarget(): can't get depth param (%v)", apiID)
		return model.ZeroTime, &InvalidParameterError{msg: err.Error()}
	}

	// Fuzz it!

	timestamp, err := p.model.StartAPIFuzzing(ctx, uint(apiID), params)
	if err != nil {
		logging.Errorf("[Fuzzer] FuzzTarget(): can't start fuzzing for API (%v)", apiID)
		return model.ZeroTime, &PluginError{msg: err.Error()}
	}

	err = p.fuzzerClient.TriggerFuzzingJob(apiID, fullServiceURI, securityParam, timeBudget)
	if err != nil {
		logging.Errorf("[Fuzzer] FuzzTarget(): can't trigger fuzzing job for API (%v), err=(%v)", apiID, err)
		fuzzerError := fmt.Errorf("can't start fuzzing job for API (%v), err=(%v)", apiID, err)
		_ = p.model.StopAPIFuzzing(ctx, uint(apiID), fuzzerError)
		return model.ZeroTime, &PluginError{msg: err.Error()}
	}

	// Success
	return timestamp, nil
}

func (p *pluginFuzzer) StopFuzzing(ctx context.Context, apiID oapicommon.ApiID) error {
	// Checks
	if p.fuzzerClient == nil {
		// Must not happen, as we had be able to start a Fuzzing...
		return &PluginError{"No deployment client running"}
	}

	// Retrieve the API
	api, err := p.model.GetAPI(ctx, uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] StopFuzzing(): can't retrieve API (%v)", apiID)
		return &NotFoundError{msg: ""}
	}

	logging.Logf("[Fuzzer] StopFuzzing(): API_id (%v) => API (%v)", apiID, api)

	// Check we are yet in fuzzing
	if !api.InFuzzing {
		logging.Errorf("[Fuzzer] StopFuzzing(%v): API (%v) not in Fuzzing", apiID, apiID)
		return &InvalidParameterError{msg: err.Error()}
	}

	err = p.model.StopAPIFuzzing(ctx, uint(apiID), nil)
	if err != nil {
		logging.Errorf("[Fuzzer] StopFuzzing(): failed to stop Fuzzing for API (%v), error=%v", apiID, err)
		return &PluginError{msg: err.Error()}
	}

	err = p.fuzzerClient.StopFuzzingJob(apiID)
	if err != nil {
		logging.Errorf("[Fuzzer] StopFuzzing(): can't trigger fuzzing job for API (%v), err=(%v)", apiID, err)
		return &PluginError{msg: err.Error()}
	}

	// Success
	return nil
}

type pluginFuzzerHTTPHandler struct {
	fuzzer *pluginFuzzer
}

func httpError(writer http.ResponseWriter, err error) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusBadRequest)
	if err2 := json.NewEncoder(writer).Encode(map[string]interface{}{"error": err.Error()}); err2 != nil {
		// we can't send the error... we can't fo anything else, here, except logging the error
		logging.Errorf("[Fuzzer] Can't encode the error (%v)", err2)
		logging.Errorf("[Fuzzer] The original error is (%v)", err)
	}
}

func httpResponse(writer http.ResponseWriter, statusCode int, data interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	if err := json.NewEncoder(writer).Encode(data); err != nil {
		httpError(writer, err)
	}
}

//
// Return the version for the fuzzer module.
//
func (*pluginFuzzerHTTPHandler) GetVersion(writer http.ResponseWriter, req *http.Request) {
	logging.Debugf("[Fuzzer] GetVersion(): -->")
	if err := json.NewEncoder(writer).Encode(restapi.Version{Version: ModuleVersion}); err != nil {
		httpError(writer, err)
	}
}

func (p *pluginFuzzerHTTPHandler) GetState(writer http.ResponseWriter, req *http.Request) {
	state := p.fuzzer.runningState
	httpResponse(writer, http.StatusOK, state)
}

//
// Retrieve the last update status for the API.
//
func (p *pluginFuzzerHTTPHandler) GetUpdateStatus(writer http.ResponseWriter, req *http.Request, apiID int64) {
	logging.Debugf("[Fuzzer] GetUpdateStatus(%v): -->", apiID)

	api, err := p.fuzzer.model.GetAPI(req.Context(), uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] GetUpdateStatus(%v): Can't retrieve api_id=(%v), error=(%v)", apiID, apiID, err)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}

	result := api.GetLastStatus()
	if result == nil {
		logging.Errorf("[Fuzzer] GetUpdateStatus(%v): No test available for this API", apiID)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(result)
	if err != nil {
		logging.Errorf("[Fuzzer] GetUpdateStatus(%v): Failed to encode response, error=(%v)", apiID, err)
	}
}

//
// Receive last status update.
//
func (p *pluginFuzzerHTTPHandler) PostUpdateStatus(writer http.ResponseWriter, req *http.Request, apiID int64) {
	logging.Debugf("[Fuzzer] PostUpdateStatus(%v): -->", apiID)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logging.Errorf("[Fuzzer] PostUpdateStatus(%v): failed to read the request body, error=%v", apiID, err)
		httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
		return
	}
	logging.Logf("[Fuzzer] PostUpdateStatus(%v): Received a request of size=(%v)", apiID, len(body))

	/*
	* Decode the result
	 */
	var data restapi.FuzzingStatusAndReport
	err = json.Unmarshal(body, &data)
	if err != nil {
		logging.Errorf("[Fuzzer] PostUpdateStatus(%v): failed to decode the request body, error=%v", apiID, err)
		httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
		return
	}

	/*
	* Add the new status to the last Test
	 */
	api, err := p.fuzzer.model.GetAPI(req.Context(), uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] PostUpdateStatus(%v): Can't retrieve api_id=(%v), error=(%v)", apiID, apiID, err)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}
	api.AddNewStatusReport(data)
	err = api.StoreReportData(req.Context(), p.fuzzer.accessor, p.fuzzer.info.Name, data)
	if err != nil {
		logging.Errorf("[Fuzzer] PostUpdateStatus(%v): Can't store report data, error=(%v)", apiID, err)
		// Not fatal, we can continue
	}

	/*
	* If the status indicate a completion, close the job
	 */
	if data.Progress == 100 && data.Status != "IN_PROGRESS" {
		err = p.fuzzer.model.StopAPIFuzzing(req.Context(), uint(apiID), nil)
		if err != nil {
			logging.Errorf("[Fuzzer] PostUpdateStatus(%v): failed to stop fuzzing status, error=%v", apiID, err)
		}
	}

	/*
	* ... Then check for notifications to send
	 */
	shortReport, err := api.GetLastShortStatus()
	if err != nil {
		logging.Errorf("[Fuzzer] PostUpdateStatus(%v): Can't get short report for this API", apiID)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}
	if shortReport.Status == restapi.INPROGRESS {
		err = p.fuzzer.sendTestProgressNotification(
			req.Context(),
			uint(apiID),
			restapi.ShortTestProgress{
				ApiID:     &apiID,
				Progress:  data.Progress,
				Starttime: shortReport.Starttime,
			},
		)
		if err != nil {
			// Log the error, but do not block the process as the error seems external to the Fuzzer modue
			logging.Errorf("[Fuzzer] PostUpdateStatus(%v): Failed to send 'TestProgress' notification, err=(%v)", apiID, err)
		}
	} else if shortReport.Status == restapi.DONE || shortReport.Status == restapi.ERROR {
		// The
		err = p.fuzzer.sendTestReportNotification(
			req.Context(),
			uint(apiID),
			*shortReport,
		)
		if err != nil {
			// Log the error, but do not block the process as the error seems external to the Fuzzer modue
			logging.Errorf("[Fuzzer] PostUpdateStatus(%v): Failed to send 'TestReport' notification, err=(%v)", apiID, err)
		}
		lastFindings := api.GetLastAPIFindings()
		err = p.fuzzer.sendAPIFindingsNotification(
			req.Context(),
			uint(apiID),
			*lastFindings,
		)
		if err != nil {
			// Log the error, but do not block the process as the error seems external to the Fuzzer modue
			logging.Errorf("[Fuzzer] PostUpdateStatus(%v): Failed to send 'APIFindings' notification, err=(%v)", apiID, err)
		}
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusNoContent)
}

//
// Return the findings list for the lastest Test.
//
func (p *pluginFuzzerHTTPHandler) GetAPIFindings(writer http.ResponseWriter, req *http.Request, apiID int64, params restapi.GetAPIFindingsParams) {
	logging.Debugf("[Fuzzer] GetFindings(%v): -->", apiID)
	api, err := p.fuzzer.model.GetAPI(req.Context(), uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] GetFindings(%v): Can't retrieve api_id=(%v), error=(%v)", apiID, apiID, err)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}
	lastFindings := api.GetLastAPIFindings()
	result := oapicommon.APIFindings{
		Items: lastFindings,
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(result)
	if err != nil {
		logging.Errorf("[Fuzzer] GetFindings(%v): Failed to encode response, error=(%v)", apiID, err)
		httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
	}
}

//
// Receive findings for last Test.
//
func (p *pluginFuzzerHTTPHandler) PostRawfindings(writer http.ResponseWriter, req *http.Request, apiID int64) {
	logging.Debugf("[Fuzzer] PostRawfindings(%v): -->", apiID)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logging.Errorf("[Fuzzer] PostRawfindings(%v): can't read body content, error=(%v)", apiID, err)
		httpResponse(writer, http.StatusBadRequest, EmptyJSON)
		return
	}
	logging.Debugf(string(body))
	// Only store it, but do nothing with it (the real list of findings will be extracted from report)
	api, err := p.fuzzer.model.GetAPI(req.Context(), uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] PostRawfindings(%v): Can't retrieve api_id=(%v), error=(%v)", apiID, apiID, err)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}
	err = api.StoreLastFindingsData(req.Context(), p.fuzzer.accessor, p.fuzzer.info.Name, body)
	if err != nil {
		logging.Errorf("[Fuzzer] PostRawfindings(%v): Can't store findings data, error=(%v)", apiID, err)
		// Not fatal, we can continue
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusNoContent)
}

//
// Send the list of Tests for the API.
//
func (p *pluginFuzzerHTTPHandler) GetTests(writer http.ResponseWriter, req *http.Request, apiID int64) {
	logging.Debugf("[Fuzzer] GetTests(%v): -->", apiID)

	api, err := p.fuzzer.model.GetAPI(req.Context(), uint(apiID))
	if err != nil {
		fmt.Printf("[Fuzzer] GetTests(%v):: can't retrieve API (%v)", apiID, apiID)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}

	// Logf("[Fuzzer] GetTests():: API_id (%v) => API (%v)", apiID, api)

	count := 0
	// tests := api.tests
	result := restapi.Tests{
		Items: api.GetTestsList(),
		Total: &count,
	}
	count = len(*(result.Items))

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(result)
	if err != nil {
		logging.Errorf("[Fuzzer] GetTests(%v): Failed to encode response, error=(%v)", apiID, err)
	}
}

func (p *pluginFuzzerHTTPHandler) GetReport(writer http.ResponseWriter, req *http.Request, apiID int64, timestamp int64) {
	logging.Debugf("[Fuzzer] GetReport(%v): -->", apiID)

	api, err := p.fuzzer.model.GetAPI(req.Context(), uint(apiID))
	if err != nil {
		fmt.Printf("[Fuzzer] GetTests(%v): can't retrieve API (%v)", apiID, apiID)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}

	result := api.GetTestByTimestamp(timestamp)
	if result == nil {
		fmt.Printf("[Fuzzer] GetReport(%v): can't retrieve Report with timestamp (%v)", apiID, timestamp)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(result)
	if err != nil {
		logging.Errorf("[Fuzzer] GetReport(%v): Failed to encode response, error=(%v)", apiID, err)
	}
}

func (p *pluginFuzzerHTTPHandler) GetAnnotatedSpec(writer http.ResponseWriter, req *http.Request, apiID int64) {
	logging.Logf("[Fuzzer] GetAnnotatedSpec(%v): --> <--", apiID)
	httpResponse(writer, http.StatusNotImplemented, EmptyJSON)
}

//
// Return the progress status of the on going test.
//
func (p *pluginFuzzerHTTPHandler) GetTestProgress(writer http.ResponseWriter, req *http.Request, apiID int64) {
	logging.Debugf("[Fuzzer] GetTestProgress(%v): -->  <--", apiID)

	// Retrieve the API
	api, err := p.fuzzer.model.GetAPI(req.Context(), uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] GetTestProgress(%v):: can't retrieve API (%v)", apiID, apiID)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}

	test := api.GetLastTest()
	if test == nil {
		logging.Errorf("[Fuzzer] GetTestProgress(%v): Can't get last test for API (%v)", apiID, apiID)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}
	report := test.Report

	if report.Status == restapi.INPROGRESS {
		testProgress := restapi.ShortTestProgress{
			ApiID:     &apiID,
			Progress:  report.Progress,
			Starttime: *test.Starttime,
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		err = json.NewEncoder(writer).Encode(testProgress)
		if err != nil {
			logging.Errorf("[Fuzzer] GetTestProgress(%v): Failed to encode response, error=(%v)", apiID, err)
		}
	} else {
		logging.Errorf("[Fuzzer] GetTestProgress(%v): API (%v) is not in Fuzzing", apiID, apiID)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
	}
}

//
// Start a test.
//
func (p *pluginFuzzerHTTPHandler) StartTest(writer http.ResponseWriter, req *http.Request, apiID int64) {
	logging.Debugf("[Fuzzer] StartTest(%v): -->  <--", apiID)

	// Decode the restapi.TestInput requesBody
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logging.Errorf("[Fuzzer] StartTest(%v): can't read body content, error=(%v)", apiID, err)
		httpResponse(writer, http.StatusBadRequest, EmptyJSON)
		return
	}
	logging.Debugf(string(body))
	var testInput restapi.TestInput
	err = json.Unmarshal(body, &testInput)
	if err != nil {
		logging.Errorf("[Fuzzer] StartTest(%v): failed to decode the request body, error=%v", apiID, err)
		httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
		return
	}

	// Get the specs here as it need ctx and accessor
	specsInfo, err := tools.GetAPISpecsInfo(req.Context(), p.fuzzer.accessor, uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] StartTest(%v): can't retrieve specs error=(%v)", apiID, err)
		httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
		return
	}

	// Store everything we need on a FuzzingInput struct
	fuzzingInput := model.FuzzingInput{
		Depth:     testInput.Depth,
		Auth:      testInput.Auth,
		SpecsInfo: specsInfo,
	}

	timestamp, err := p.fuzzer.FuzzTarget(req.Context(), apiID, &fuzzingInput)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		//nolint: errorlint // no wrapped error here
		switch err2 := err.(type) {
		case *NotFoundError:
			httpResponse(writer, http.StatusNotFound, EmptyJSON)
		case *InvalidParameterError:
			httpResponse(writer, http.StatusBadRequest, EmptyJSON)
		case *PluginError:
			httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
		case *NotSupportedError:
			httpResponse(writer, http.StatusBadRequest, EmptyJSON)
		default:
			logging.Errorf("[Fuzzer] StartTest(%v): unexpected error=(%v)", apiID, err2)
			httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
		}
		return
	}

	// Success
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	result := restapi.TestHandle{
		ApiID:     &apiID,
		Timestamp: &timestamp,
	}
	err = json.NewEncoder(writer).Encode(&result)
	if err != nil {
		logging.Errorf("[Fuzzer] StartTest(%v): Failed to encode response, error=(%v)", apiID, err)
	}
}

//
// Stop an ongoing test.
//
func (p *pluginFuzzerHTTPHandler) StopTest(writer http.ResponseWriter, req *http.Request, apiID int64) {
	logging.Debugf("[Fuzzer] StopTest(%v): -->  <--", apiID)

	err := p.fuzzer.StopFuzzing(req.Context(), apiID)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		//nolint: errorlint // no wrapped error here
		switch err2 := err.(type) {
		case *NotFoundError:
			httpResponse(writer, http.StatusNotFound, EmptyJSON)
		case *InvalidParameterError:
			httpResponse(writer, http.StatusBadRequest, EmptyJSON)
		case *PluginError:
			httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
		case *NotSupportedError:
			httpResponse(writer, http.StatusBadRequest, EmptyJSON)
		default:
			logging.Errorf("[Fuzzer] StopTest(%v): unexpected error=(%v)", apiID, err2)
			httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
		}
		return
	}

	// Success
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusNoContent)
}

//
// Return the report of the last test.
//
func (p *pluginFuzzerHTTPHandler) GetTestReport(writer http.ResponseWriter, req *http.Request, apiID int64) {
	logging.Debugf("[Fuzzer] GetTestReport(%v): -->", apiID)

	// Retrieve the API (it will give the endpoint and the port)
	api, err := p.fuzzer.model.GetAPI(req.Context(), uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] GetTestReport(%v): can't retrieve API (%v)", apiID, apiID)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}

	// Retrieve last status
	shortReport, err := api.GetLastShortStatus()
	if err != nil {
		logging.Errorf("[Fuzzer] GetTestReport(%v): Can't get short report for API (%v)", apiID, apiID)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}

	// Retrieve last status
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(shortReport)
	if err != nil {
		logging.Errorf("[Fuzzer] GetUpdateStatus(%v): Failed to encode response, error=(%v)", apiID, err)
	}
}

func (p *pluginFuzzer) HTTPHandler() http.Handler {
	return p.httpHandler
}
