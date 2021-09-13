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

package database

import (
	"fmt"

	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	"github.com/apiclarity/apiclarity/backend/pkg/utils"
	speculatorspec "github.com/apiclarity/speculator/pkg/spec"
)

const (
	apiEventTableName = "api_events"

	// NOTE: when changing one of the column names change also the gorm label in APIEvent
	timeColumnName                     = "time"
	methodColumnName                   = "method"
	pathColumnName                     = "path"
	providedPathIdColumnName           = "providedPathID"
	reconstructedPathIdColumnName      = "reconstructedPathId"
	queryColumnName                    = "query"
	statusCodeColumnName               = "statusCode"
	sourceIPColumnName                 = "sourceIP"
	destinationIPColumnName            = "destinationIP"
	destinationPortColumnName          = "destinationPort"
	hasReconstructedSpecDiffColumnName = "hasReconstructedSpecDiff"
	hasProvidedSpecDiffColumnName      = "hasProvidedSpecDiff"
	hasSpecDiffColumnName              = "hasSpecDiff" // hasProvidedSpecDiff || hasReconstructedSpecDiff
	hostSpecNameColumnName             = "hostSpecName"
	newReconstructedSpecColumnName     = "newReconstructedSpec"
	oldReconstructedSpecColumnName     = "oldReconstructedSpec"
	newProvidedSpecColumnName          = "newProvidedSpec"
	oldProvidedSpecColumnName          = "oldProvidedSpec"
	apiInfoIdColumnName                = "apiInfoId"
	isNonApiColumnName                 = "isNonApi"
	eventTypeColumnName                = "eventType"
)

var specDiffColumns = []string{newReconstructedSpecColumnName, oldReconstructedSpecColumnName, newProvidedSpecColumnName, oldProvidedSpecColumnName}

type APIEvent struct {
	// will be populated after inserting to DB
	ID uint `gorm:"primarykey" faker:"-"`
	//CreatedAt time.Time
	//UpdatedAt time.Time

	Time                     strfmt.DateTime   `json:"time" gorm:"column:time" faker:"-"`
	Method                   models.HTTPMethod `json:"method,omitempty" gorm:"column:method" faker:"oneof: GET, PUT, POST, DELETE"`
	Path                     string            `json:"path,omitempty" gorm:"column:path" faker:"oneof: /news, /customers, /jokes"`
	ProvidedPathID           string            `json:"providedPathId,omitempty" gorm:"column:providedPathId" faker:"-"`
	ReconstructedPathID      string            `json:"reconstructedPathId,omitempty" gorm:"column:reconstructedPathId" faker:"-"`
	Query                    string            `json:"query,omitempty" gorm:"column:query" faker:"oneof: name=ferret&color=purple, foo=bar, -"`
	StatusCode               int64             `json:"statusCode,omitempty" gorm:"column:statusCode" faker:"oneof: 200, 401, 404, 500"`
	SourceIP                 string            `json:"sourceIP,omitempty" gorm:"column:sourceIP" faker:"sourceIP"`
	DestinationIP            string            `json:"destinationIP,omitempty" gorm:"column:destinationIP" faker:"destinationIP"`
	DestinationPort          int64             `json:"destinationPort,omitempty" gorm:"column:destinationPort" faker:"oneof: 80, 443"`
	HasReconstructedSpecDiff bool              `json:"hasReconstructedSpecDiff,omitempty" gorm:"column:hasReconstructedSpecDiff"`
	HasProvidedSpecDiff      bool              `json:"hasProvidedSpecDiff,omitempty" gorm:"column:hasProvidedSpecDiff"`
	HasSpecDiff              bool              `json:"hasSpecDiff,omitempty" gorm:"column:hasSpecDiff"`
	HostSpecName             string            `json:"hostSpecName,omitempty" gorm:"column:hostSpecName" faker:"oneof: test.com, example.com, kaki.org"`
	IsNonApi                 bool              `json:"isNonApi,omitempty" gorm:"column:isNonApi" faker:"-"`

	// Spec diff info
	// New reconstructed spec json string
	NewReconstructedSpec string `json:"newReconstructedSpec,omitempty" gorm:"column:newReconstructedSpec" faker:"-"`
	// Old reconstructed spec json string
	OldReconstructedSpec string `json:"oldReconstructedSpec,omitempty" gorm:"column:oldReconstructedSpec" faker:"-"`
	// New provided spec json string
	NewProvidedSpec string `json:"newProvidedSpec,omitempty" gorm:"column:newProvidedSpec" faker:"-"`
	// Old provided spec json string
	OldProvidedSpec string `json:"oldProvidedSpec,omitempty" gorm:"column:oldProvidedSpec" faker:"-"`

	// ID for the relevant APIInfo
	APIInfoID uint `json:"apiInfoId,omitempty" gorm:"column:apiInfoId" faker:"-"`
	// We'll not always have a corresponding API info (e.g. non-API resources) so the type is needed also for the event
	EventType models.APIType `json:"eventType,omitempty" gorm:"column:eventType"faker:"oneof: INTERNAL, EXTERNAL"`
}

type hostGroup struct {
	HostSpecName string
	Port         int64
	ApiType      string
	ApiInfoId    uint32
	Count        int
}

const dashboardTopAPIsNum = 5

func GroupByAPIInfo(db *gorm.DB) ([]hostGroup, error) {
	var results []hostGroup

	rows, err := db.
		// filters out non APIs
		Not(isNonApiColumnName+" = ?", true).
		Select(
			FieldInTable(apiEventTableName, hostSpecNameColumnName) +
				", " + FieldInTable(apiEventTableName, destinationPortColumnName) +
				", " + FieldInTable(apiEventTableName, apiInfoIdColumnName) +
				", " + FieldInTable(apiEventTableName, eventTypeColumnName) +
				", COUNT(*) AS count").
		Group(FieldInTable(apiEventTableName, apiInfoIdColumnName)).
		Order("count desc").
		Limit(dashboardTopAPIsNum).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get top API event counts: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Warnf("Failed to close rows: %v", err)
		}
	}()

	for rows.Next() {
		group := hostGroup{}
		if err := rows.Scan(&group.HostSpecName, &group.Port, &group.ApiInfoId, &group.ApiType, &group.Count); err != nil {
			return nil, fmt.Errorf("failed to get fields: %v", err)
		}
		log.Debugf("Fetched fields: %+v", group)
		results = append(results, group)
	}

	return results, nil
}

func (APIEvent) TableName() string {
	return apiEventTableName
}

func APIEventFromDB(event *APIEvent) *models.APIEvent {
	return &models.APIEvent{
		APIInfoID:                uint32(event.APIInfoID),
		APIType:                  event.EventType,
		DestinationIP:            event.DestinationIP,
		DestinationPort:          event.DestinationPort,
		HasProvidedSpecDiff:      &event.HasProvidedSpecDiff,
		HasReconstructedSpecDiff: &event.HasReconstructedSpecDiff,
		HostSpecName:             event.HostSpecName,
		ID:                       uint32(event.ID),
		Method:                   event.Method,
		Path:                     event.Path,
		Query:                    event.Query,
		SourceIP:                 event.SourceIP,
		StatusCode:               event.StatusCode,
		Time:                     event.Time,
	}
}

func GetAPIEventsTable() *gorm.DB {
	return DB.Table(apiEventTableName)
}

func CreateAPIEvent(event *APIEvent) {
	if result := GetAPIEventsTable().Create(event); result.Error != nil {
		log.Errorf("Failed to create event: %v", result.Error)
	} else {
		log.Infof("Event created %+v", event)
	}
}

func GetAPIEventsAndTotal(params operations.GetAPIEventsParams) ([]APIEvent, int64, error) {
	var apiEvents []APIEvent
	var count int64

	tx := SetAPIEventsFilters(GetAPIEventsTable(), getAPIEventsParamsToFilters(params), true)
	// get total count item with the current filters
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// get specific page ordered items with the current filters
	if err := tx.Scopes(Paginate(params.Page, params.PageSize)).
		Order(CreateSortOrder(params.SortKey, params.SortDir)).
		Omit(specDiffColumns...).
		Find(&apiEvents).Error; err != nil {
		return nil, 0, err
	}

	return apiEvents, count, nil
}

func getAPIEventsParamsToFilters(params operations.GetAPIEventsParams) *APIEventsFilters {
	return &APIEventsFilters{
		DestinationIPIsNot:   params.DestinationIPIsNot,
		DestinationIPIs:      params.DestinationIPIs,
		DestinationPortIsNot: params.DestinationPortIsNot,
		DestinationPortIs:    params.DestinationPortIs,
		EndTime:              params.EndTime,
		ShowNonAPI:           params.ShowNonAPI,
		HasSpecDiffIs:        params.HasSpecDiffIs,
		MethodIs:             params.MethodIs,
		PathContains:         params.PathContains,
		PathEnd:              params.PathEnd,
		PathIsNot:            params.PathIsNot,
		PathIs:               params.PathIs,
		PathStart:            params.PathStart,
		SourceIPIsNot:        params.SourceIPIsNot,
		SourceIPIs:           params.SourceIPIs,
		SpecContains:         params.SpecContains,
		SpecEnd:              params.SpecEnd,
		SpecIsNot:            params.SpecIsNot,
		SpecIs:               params.SpecIs,
		SpecStart:            params.SpecStart,
		StartTime:            params.StartTime,
		StatusCodeGte:        params.StatusCodeGte,
		StatusCodeIsNot:      params.StatusCodeIsNot,
		StatusCodeIs:         params.StatusCodeIs,
		StatusCodeLte:        params.StatusCodeLte,
	}
}

func GetAPIEvent(eventID uint32) (*APIEvent, error) {
	var apiEvent APIEvent

	if err := GetAPIEventsTable().Omit(specDiffColumns...).First(&apiEvent, eventID).Error; err != nil {
		return nil, err
	}

	return &apiEvent, nil
}

func GetAPIEventReconstructedSpecDiff(eventID uint32) (*APIEvent, error) {
	var apiEvent APIEvent

	if err := GetAPIEventsTable().Select(newReconstructedSpecColumnName, oldReconstructedSpecColumnName).First(&apiEvent, eventID).Error; err != nil {
		return nil, err
	}

	return &apiEvent, nil
}

func GetAPIEventProvidedSpecDiff(eventID uint32) (*APIEvent, error) {
	var apiEvent APIEvent

	if err := GetAPIEventsTable().Select(newProvidedSpecColumnName, oldProvidedSpecColumnName).First(&apiEvent, eventID).Error; err != nil {
		return nil, err
	}

	return &apiEvent, nil
}

func GetAPIEventsLatestDiffs(latestDiffsNum int) ([]APIEvent, error) {
	var latestDiffs []APIEvent
	if err := GetAPIEventsTable().Where(hasSpecDiffColumnName + " = true").
		Order("time desc").Limit(latestDiffsNum).Scan(&latestDiffs).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest diffs from events table. %v", err)
	}

	return latestDiffs, nil
}

type APIEventsFilters struct {
	DestinationIPIsNot    []string
	DestinationIPIs       []string
	DestinationPortIsNot  []string
	DestinationPortIs     []string
	EndTime               strfmt.DateTime
	ShowNonAPI            bool
	HasSpecDiffIs         *bool
	MethodIs              []string
	ReconstructedPathIDIs []string
	ProvidedPathIDIs      []string
	PathContains          []string
	PathEnd               *string
	PathIsNot             []string
	PathIs                []string
	PathStart             *string
	SourceIPIsNot         []string
	SourceIPIs            []string
	SpecContains          []string
	SpecEnd               *string
	SpecIsNot             []string
	SpecIs                []string
	SpecStart             *string
	StartTime             strfmt.DateTime
	StatusCodeGte         *string
	StatusCodeIsNot       []string
	StatusCodeIs          []string
	StatusCodeLte         *string
}

func SetAPIEventsFilters(tx *gorm.DB, filters *APIEventsFilters, shouldSetTimeFilters bool) *gorm.DB {
	if shouldSetTimeFilters {
		// time filter
		tx = tx.Where(CreateTimeFilter(filters.StartTime, filters.EndTime))
	}

	// methods filter
	tx = FilterIs(tx, methodColumnName, filters.MethodIs)

	// path ID filters
	tx = FilterIs(tx, providedPathIdColumnName, filters.ProvidedPathIDIs)
	tx = FilterIs(tx, reconstructedPathIdColumnName, filters.ReconstructedPathIDIs)

	// path filters
	tx = FilterIs(tx, pathColumnName, filters.PathIs)
	tx = FilterIsNot(tx, pathColumnName, filters.PathIsNot)
	tx = FilterContains(tx, pathColumnName, filters.PathContains)
	tx = FilterStartsWith(tx, pathColumnName, filters.PathStart)
	tx = FilterEndsWith(tx, pathColumnName, filters.PathEnd)

	// status codes filters
	tx = FilterIs(tx, statusCodeColumnName, filters.StatusCodeIs)
	tx = FilterIsNot(tx, statusCodeColumnName, filters.StatusCodeIsNot)
	tx = FilterGte(tx, statusCodeColumnName, filters.StatusCodeGte)
	tx = FilterLte(tx, statusCodeColumnName, filters.StatusCodeLte)

	// source IPs filters
	tx = FilterIs(tx, sourceIPColumnName, filters.SourceIPIs)
	tx = FilterIsNot(tx, sourceIPColumnName, filters.SourceIPIsNot)
	// destination IPs filters
	tx = FilterIs(tx, destinationIPColumnName, filters.DestinationIPIs)
	tx = FilterIsNot(tx, destinationIPColumnName, filters.DestinationIPIsNot)
	// destination ports filters
	tx = FilterIs(tx, destinationPortColumnName, filters.DestinationPortIs)
	tx = FilterIsNot(tx, destinationPortColumnName, filters.DestinationPortIsNot)

	// has spec diff filter
	tx = FilterIsBool(tx, hasSpecDiffColumnName, filters.HasSpecDiffIs)

	// host spec name filters
	tx = FilterIs(tx, hostSpecNameColumnName, filters.SpecIs)
	tx = FilterIsNot(tx, hostSpecNameColumnName, filters.SpecIsNot)
	tx = FilterContains(tx, hostSpecNameColumnName, filters.SpecContains)
	tx = FilterStartsWith(tx, hostSpecNameColumnName, filters.SpecStart)
	tx = FilterEndsWith(tx, hostSpecNameColumnName, filters.SpecEnd)

	// ignore non APIs
	if !filters.ShowNonAPI {
		tx.Where(fmt.Sprintf("%s = ?", isNonApiColumnName), false)
	}

	return tx
}

// SetAPIEventsReconstructedPathId will set reconstructed path ID for all events with the provided paths, host and port
func SetAPIEventsReconstructedPathId(approvedReview []*speculatorspec.ApprovedSpecReviewPathItem, host string, port string) error {
	err := GetAPIEventsTable().Transaction(func(tx *gorm.DB) error {
		for _, item := range approvedReview {
			tx := FilterIs(tx, pathColumnName, utils.MapToSlice(item.Paths))
			tx = FilterIs(tx, hostSpecNameColumnName, []string{host})
			tx = FilterIs(tx, destinationPortColumnName, []string{port})

			if err := tx.Model(&APIEvent{}).Updates(map[string]interface{}{reconstructedPathIdColumnName: item.PathUUID}).Error; err != nil {
				// return any error will rollback
				return err
			}
		}

		// return nil will commit the whole transaction
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
