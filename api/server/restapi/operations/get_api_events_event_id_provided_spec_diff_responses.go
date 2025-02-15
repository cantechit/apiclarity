// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/apiclarity/apiclarity/api/server/models"
)

// GetAPIEventsEventIDProvidedSpecDiffOKCode is the HTTP code returned for type GetAPIEventsEventIDProvidedSpecDiffOK
const GetAPIEventsEventIDProvidedSpecDiffOKCode int = 200

/*GetAPIEventsEventIDProvidedSpecDiffOK Success

swagger:response getApiEventsEventIdProvidedSpecDiffOK
*/
type GetAPIEventsEventIDProvidedSpecDiffOK struct {

	/*
	  In: Body
	*/
	Payload *models.APIEventSpecDiff `json:"body,omitempty"`
}

// NewGetAPIEventsEventIDProvidedSpecDiffOK creates GetAPIEventsEventIDProvidedSpecDiffOK with default headers values
func NewGetAPIEventsEventIDProvidedSpecDiffOK() *GetAPIEventsEventIDProvidedSpecDiffOK {

	return &GetAPIEventsEventIDProvidedSpecDiffOK{}
}

// WithPayload adds the payload to the get Api events event Id provided spec diff o k response
func (o *GetAPIEventsEventIDProvidedSpecDiffOK) WithPayload(payload *models.APIEventSpecDiff) *GetAPIEventsEventIDProvidedSpecDiffOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get Api events event Id provided spec diff o k response
func (o *GetAPIEventsEventIDProvidedSpecDiffOK) SetPayload(payload *models.APIEventSpecDiff) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAPIEventsEventIDProvidedSpecDiffOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetAPIEventsEventIDProvidedSpecDiffDefault unknown error

swagger:response getApiEventsEventIdProvidedSpecDiffDefault
*/
type GetAPIEventsEventIDProvidedSpecDiffDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.APIResponse `json:"body,omitempty"`
}

// NewGetAPIEventsEventIDProvidedSpecDiffDefault creates GetAPIEventsEventIDProvidedSpecDiffDefault with default headers values
func NewGetAPIEventsEventIDProvidedSpecDiffDefault(code int) *GetAPIEventsEventIDProvidedSpecDiffDefault {
	if code <= 0 {
		code = 500
	}

	return &GetAPIEventsEventIDProvidedSpecDiffDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get API events event ID provided spec diff default response
func (o *GetAPIEventsEventIDProvidedSpecDiffDefault) WithStatusCode(code int) *GetAPIEventsEventIDProvidedSpecDiffDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get API events event ID provided spec diff default response
func (o *GetAPIEventsEventIDProvidedSpecDiffDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get API events event ID provided spec diff default response
func (o *GetAPIEventsEventIDProvidedSpecDiffDefault) WithPayload(payload *models.APIResponse) *GetAPIEventsEventIDProvidedSpecDiffDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get API events event ID provided spec diff default response
func (o *GetAPIEventsEventIDProvidedSpecDiffDefault) SetPayload(payload *models.APIResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAPIEventsEventIDProvidedSpecDiffDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
