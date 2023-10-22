// PetStore vTODO 204f6b26587305ef3a4c043b8636035ada3889ef
// --
// Code generated by webrpc-gen@v0.14.0-dev with golang generator. DO NOT EDIT.
//
// gospeak .
package proto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// WebRPC description and code-gen version
func WebRPCVersion() string {
	return "v1"
}

// Schema version of your RIDL schema
func WebRPCSchemaVersion() string {
	return "vTODO"
}

// Schema hash generated from your RIDL schema
func WebRPCSchemaHash() string {
	return "204f6b26587305ef3a4c043b8636035ada3889ef"
}

//
// Types
//



const (
	Status_approved Status = 0
	Status_pending Status = 1
	Status_closed Status = 2
	Status_new Status = 3
)

var Status_name = map[int]string{
	0: "approved",
	1: "pending",
	2: "closed",
	3: "new",
}

var Status_value = map[string]int{
	"approved": 0,
	"pending": 1,
	"closed": 2,
	"new": 3,
}

func (x Status) String() string {
	return Status_name[int(x)]
}

func (x Status) MarshalText() ([]byte, error) {
	return []byte(Status_name[int(x)]), nil
}

func (x *Status) UnmarshalText(b []byte) error {
	*x = Status(Status_value[string(b)])
	return nil
}


//
// Server
//

type WebRPCServer interface {
	http.Handler
}

type petStoreServer struct {
	PetStore
	OnError func(r *http.Request, rpcErr *WebRPCError)
}

func NewPetStoreServer(svc PetStore) *petStoreServer {
	return &petStoreServer{
		PetStore: svc,
	}
}

func (s *petStoreServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		// In case of a panic, serve a HTTP 500 error and then panic.
		if rr := recover(); rr != nil {
			s.sendErrorJSON(w, r, ErrWebrpcServerPanic.WithCause(fmt.Errorf("%v", rr)))
			panic(rr)
		}
	}()

	ctx := r.Context()
	ctx = context.WithValue(ctx, HTTPResponseWriterCtxKey, w)
	ctx = context.WithValue(ctx, HTTPRequestCtxKey, r)
	ctx = context.WithValue(ctx, ServiceNameCtxKey, "PetStore")

	var handler func(ctx context.Context, w http.ResponseWriter, r *http.Request)
	switch r.URL.Path {
	case "/rpc/PetStore/CreatePet": handler = s.serveCreatePetJSON
	case "/rpc/PetStore/DeletePet": handler = s.serveDeletePetJSON
	case "/rpc/PetStore/GetPet": handler = s.serveGetPetJSON
	case "/rpc/PetStore/ListPets": handler = s.serveListPetsJSON
	case "/rpc/PetStore/UpdatePet": handler = s.serveUpdatePetJSON
	default:
		err := ErrWebrpcBadRoute.WithCause(fmt.Errorf("no handler for path %q", r.URL.Path))
		s.sendErrorJSON(w, r, err)
		return
	}

	if r.Method != "POST" {
		w.Header().Add("Allow", "POST") // RFC 9110.
		err := ErrWebrpcBadMethod.WithCause(fmt.Errorf("unsupported method %q (only POST is allowed)", r.Method))
		s.sendErrorJSON(w, r, err)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if i := strings.Index(contentType, ";"); i >= 0 {
		contentType = contentType[:i]
	}
	contentType = strings.TrimSpace(strings.ToLower(contentType))

	switch contentType  {
	case "application/json":
		handler(ctx, w, r)
	default:
		err := ErrWebrpcBadRequest.WithCause(fmt.Errorf("unexpected Content-Type: %q", r.Header.Get("Content-Type")))
		s.sendErrorJSON(w, r, err)
	}
}

func (s *petStoreServer) serveCreatePetJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx = context.WithValue(ctx, MethodNameCtxKey, "CreatePet")

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		s.sendErrorJSON(w, r, ErrWebrpcBadRequest.WithCause(fmt.Errorf("failed to read request data: %w", err)))
		return
	}
	defer r.Body.Close()

	reqPayload := struct {
		Arg0 *Pet `json:"new"`
	}{}
	if err := json.Unmarshal(reqBody, &reqPayload); err != nil {
		s.sendErrorJSON(w, r, ErrWebrpcBadRequest.WithCause(fmt.Errorf("failed to unmarshal request data: %w", err)))
		return
	}

	// Call service method implementation.
	ret0, err := s.PetStore.CreatePet(ctx, reqPayload.Arg0)
	if err != nil {
		rpcErr, ok := err.(WebRPCError)
		if !ok {
			rpcErr = ErrWebrpcEndpoint.WithCause(err)
		}
		s.sendErrorJSON(w, r, rpcErr)
		return
	}

	respPayload := struct {
		Ret0 *Pet `json:"pet"`
	}{ret0}
	respBody, err := json.Marshal(respPayload)
	if err != nil {
		s.sendErrorJSON(w, r, ErrWebrpcBadResponse.WithCause(fmt.Errorf("failed to marshal json response: %w", err)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func (s *petStoreServer) serveDeletePetJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx = context.WithValue(ctx, MethodNameCtxKey, "DeletePet")

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		s.sendErrorJSON(w, r, ErrWebrpcBadRequest.WithCause(fmt.Errorf("failed to read request data: %w", err)))
		return
	}
	defer r.Body.Close()

	reqPayload := struct {
		Arg0 int64 `json:"ID"`
	}{}
	if err := json.Unmarshal(reqBody, &reqPayload); err != nil {
		s.sendErrorJSON(w, r, ErrWebrpcBadRequest.WithCause(fmt.Errorf("failed to unmarshal request data: %w", err)))
		return
	}

	// Call service method implementation.
	err = s.PetStore.DeletePet(ctx, reqPayload.Arg0)
	if err != nil {
		rpcErr, ok := err.(WebRPCError)
		if !ok {
			rpcErr = ErrWebrpcEndpoint.WithCause(err)
		}
		s.sendErrorJSON(w, r, rpcErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (s *petStoreServer) serveGetPetJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx = context.WithValue(ctx, MethodNameCtxKey, "GetPet")

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		s.sendErrorJSON(w, r, ErrWebrpcBadRequest.WithCause(fmt.Errorf("failed to read request data: %w", err)))
		return
	}
	defer r.Body.Close()

	reqPayload := struct {
		Arg0 int64 `json:"ID"`
	}{}
	if err := json.Unmarshal(reqBody, &reqPayload); err != nil {
		s.sendErrorJSON(w, r, ErrWebrpcBadRequest.WithCause(fmt.Errorf("failed to unmarshal request data: %w", err)))
		return
	}

	// Call service method implementation.
	ret0, err := s.PetStore.GetPet(ctx, reqPayload.Arg0)
	if err != nil {
		rpcErr, ok := err.(WebRPCError)
		if !ok {
			rpcErr = ErrWebrpcEndpoint.WithCause(err)
		}
		s.sendErrorJSON(w, r, rpcErr)
		return
	}

	respPayload := struct {
		Ret0 *Pet `json:"pet"`
	}{ret0}
	respBody, err := json.Marshal(respPayload)
	if err != nil {
		s.sendErrorJSON(w, r, ErrWebrpcBadResponse.WithCause(fmt.Errorf("failed to marshal json response: %w", err)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func (s *petStoreServer) serveListPetsJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx = context.WithValue(ctx, MethodNameCtxKey, "ListPets")

	// Call service method implementation.
	ret0, err := s.PetStore.ListPets(ctx)
	if err != nil {
		rpcErr, ok := err.(WebRPCError)
		if !ok {
			rpcErr = ErrWebrpcEndpoint.WithCause(err)
		}
		s.sendErrorJSON(w, r, rpcErr)
		return
	}

	respPayload := struct {
		Ret0 []*Pet `json:"pets"`
	}{ret0}
	respBody, err := json.Marshal(respPayload)
	if err != nil {
		s.sendErrorJSON(w, r, ErrWebrpcBadResponse.WithCause(fmt.Errorf("failed to marshal json response: %w", err)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func (s *petStoreServer) serveUpdatePetJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx = context.WithValue(ctx, MethodNameCtxKey, "UpdatePet")

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		s.sendErrorJSON(w, r, ErrWebrpcBadRequest.WithCause(fmt.Errorf("failed to read request data: %w", err)))
		return
	}
	defer r.Body.Close()

	reqPayload := struct {
		Arg0 int64 `json:"ID"`
		Arg1 *Pet `json:"update"`
	}{}
	if err := json.Unmarshal(reqBody, &reqPayload); err != nil {
		s.sendErrorJSON(w, r, ErrWebrpcBadRequest.WithCause(fmt.Errorf("failed to unmarshal request data: %w", err)))
		return
	}

	// Call service method implementation.
	ret0, err := s.PetStore.UpdatePet(ctx, reqPayload.Arg0, reqPayload.Arg1)
	if err != nil {
		rpcErr, ok := err.(WebRPCError)
		if !ok {
			rpcErr = ErrWebrpcEndpoint.WithCause(err)
		}
		s.sendErrorJSON(w, r, rpcErr)
		return
	}

	respPayload := struct {
		Ret0 *Pet `json:"pet"`
	}{ret0}
	respBody, err := json.Marshal(respPayload)
	if err != nil {
		s.sendErrorJSON(w, r, ErrWebrpcBadResponse.WithCause(fmt.Errorf("failed to marshal json response: %w", err)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}


func (s *petStoreServer) sendErrorJSON(w http.ResponseWriter, r *http.Request, rpcErr WebRPCError) {
	if s.OnError != nil {
		 s.OnError(r, &rpcErr)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(rpcErr.HTTPStatus)

	respBody, _ := json.Marshal(rpcErr)
	w.Write(respBody)
}
func RespondWithError(w http.ResponseWriter, err error) {
	rpcErr, ok := err.(WebRPCError)
	if !ok {
		rpcErr = ErrWebrpcEndpoint.WithCause(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(rpcErr.HTTPStatus)

	respBody, _ := json.Marshal(rpcErr)
	w.Write(respBody)
}

//
// Helpers
//

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "webrpc context value " + k.name
}

var (
	HTTPResponseWriterCtxKey = &contextKey{"HTTPResponseWriter"}

	HTTPRequestCtxKey = &contextKey{"HTTPRequest"}

	ServiceNameCtxKey = &contextKey{"ServiceName"}

	MethodNameCtxKey = &contextKey{"MethodName"}
)

func ServiceNameFromContext(ctx context.Context) string {
	service, _ := ctx.Value(ServiceNameCtxKey).(string)
	return service
}

func MethodNameFromContext(ctx context.Context) string {
	method, _ := ctx.Value(MethodNameCtxKey).(string)
	return method
}

func RequestFromContext(ctx context.Context) *http.Request {
	r, _ := ctx.Value(HTTPRequestCtxKey).(*http.Request)
	return r
}
func ResponseWriterFromContext(ctx context.Context) http.ResponseWriter {
	w, _ := ctx.Value(HTTPResponseWriterCtxKey).(http.ResponseWriter)
	return w
}


//
// Errors
//

type WebRPCError struct {
	Name       string `json:"error"`
	Code       int    `json:"code"`
	Message    string `json:"msg"`
	Cause      string `json:"cause,omitempty"`
	HTTPStatus int    `json:"status"`
	cause      error
}

var _ error = WebRPCError{}

func (e WebRPCError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s %d: %s: %v", e.Name, e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s %d: %s", e.Name, e.Code, e.Message)
}

func (e WebRPCError) Is(target error) bool {
	if rpcErr, ok := target.(WebRPCError); ok {
		return rpcErr.Code == e.Code
	}
	if legacyErr, ok := target.(legacyError); ok {
		return legacyErr.Code == e.Code
	}
	return errors.Is(e.cause, target)
}

func (e WebRPCError) Unwrap() error {
	return e.cause
}

func (e WebRPCError) WithCause(cause error) WebRPCError {
	err := e
	err.cause = cause
	err.Cause = cause.Error()
	return err
}

// Deprecated: Use .WithCause() method on WebRPCError.
func ErrorWithCause(rpcErr WebRPCError, cause error) WebRPCError {
	return rpcErr.WithCause(cause)
}

// Webrpc errors
var (
	ErrWebrpcEndpoint = WebRPCError{Code: 0, Name: "WebrpcEndpoint", Message: "endpoint error", HTTPStatus: 400}
	ErrWebrpcRequestFailed = WebRPCError{Code: -1, Name: "WebrpcRequestFailed", Message: "request failed", HTTPStatus: 400}
	ErrWebrpcBadRoute = WebRPCError{Code: -2, Name: "WebrpcBadRoute", Message: "bad route", HTTPStatus: 404}
	ErrWebrpcBadMethod = WebRPCError{Code: -3, Name: "WebrpcBadMethod", Message: "bad method", HTTPStatus: 405}
	ErrWebrpcBadRequest = WebRPCError{Code: -4, Name: "WebrpcBadRequest", Message: "bad request", HTTPStatus: 400}
	ErrWebrpcBadResponse = WebRPCError{Code: -5, Name: "WebrpcBadResponse", Message: "bad response", HTTPStatus: 500}
	ErrWebrpcServerPanic = WebRPCError{Code: -6, Name: "WebrpcServerPanic", Message: "server panic", HTTPStatus: 500}
	ErrWebrpcInternalError = WebRPCError{Code: -7, Name: "WebrpcInternalError", Message: "internal error", HTTPStatus: 500}
)

//
// Legacy errors
//

// Deprecated: Use fmt.Errorf() or WebRPCError.
func Errorf(err legacyError, format string, args ...interface{}) WebRPCError {
	return err.WebRPCError.WithCause(fmt.Errorf(format, args...))
}

// Deprecated: Use .WithCause() method on WebRPCError.
func WrapError(err legacyError, cause error, format string, args ...interface{}) WebRPCError {
	return err.WebRPCError.WithCause(fmt.Errorf("%v: %w", fmt.Errorf(format, args...), cause))
}

// Deprecated: Use fmt.Errorf() or WebRPCError.
func Failf(format string, args ...interface{}) WebRPCError {
	return Errorf(ErrFail, format, args...)
}

// Deprecated: Use .WithCause() method on WebRPCError.
func WrapFailf(cause error, format string, args ...interface{}) WebRPCError {
	return WrapError(ErrFail, cause, format, args...)
}

// Deprecated: Use fmt.Errorf() or WebRPCError.
func ErrorNotFound(format string, args ...interface{}) WebRPCError {
	return Errorf(ErrNotFound, format, args...)
}

// Deprecated: Use fmt.Errorf() or WebRPCError.
func ErrorInvalidArgument(argument string, validationMsg string) WebRPCError {
	return Errorf(ErrInvalidArgument, argument+" "+validationMsg)
}

// Deprecated: Use fmt.Errorf() or WebRPCError.
func ErrorRequiredArgument(argument string) WebRPCError {
	return ErrorInvalidArgument(argument, "is required")
}

// Deprecated: Use fmt.Errorf() or WebRPCError.
func ErrorInternal(format string, args ...interface{}) WebRPCError {
	return Errorf(ErrInternal, format, args...)
}

type legacyError struct { WebRPCError }

// Legacy errors (webrpc v0.10.0 and earlier). Will be removed.
var (
	// Deprecated. Define errors in RIDL schema.
	ErrCanceled = legacyError{WebRPCError{Code: -10000, Name: "ErrCanceled", Message: "canceled", HTTPStatus: 408 /* RequestTimeout */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrUnknown = legacyError{WebRPCError{Code: -10001, Name: "ErrUnknown", Message: "unknown", HTTPStatus: 400 /* Bad Request */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrFail = legacyError{WebRPCError{Code: -10002, Name: "ErrFail", Message: "fail", HTTPStatus: 422 /* Unprocessable Entity */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrInvalidArgument = legacyError{WebRPCError{Code: -10003, Name: "ErrInvalidArgument", Message: "invalid argument", HTTPStatus: 400 /* BadRequest */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrDeadlineExceeded = legacyError{WebRPCError{Code: -10004, Name: "ErrDeadlineExceeded", Message: "deadline exceeded", HTTPStatus: 408 /* RequestTimeout */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrNotFound = legacyError{WebRPCError{Code: -10005, Name: "ErrNotFound", Message: "not found", HTTPStatus: 404 /* Not Found */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrBadRoute = legacyError{WebRPCError{Code: -10006, Name: "ErrBadRoute", Message: "bad route", HTTPStatus: 404 /* Not Found */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrAlreadyExists = legacyError{WebRPCError{Code: -10007, Name: "ErrAlreadyExists", Message: "already exists", HTTPStatus: 409 /* Conflict */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrPermissionDenied = legacyError{WebRPCError{Code: -10008, Name: "ErrPermissionDenied", Message: "permission denied", HTTPStatus: 403 /* Forbidden */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrUnauthenticated = legacyError{WebRPCError{Code: -10009, Name: "ErrUnauthenticated", Message: "unauthenticated", HTTPStatus: 401 /* Unauthorized */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrResourceExhausted = legacyError{WebRPCError{Code: -10010, Name: "ErrResourceExhausted", Message: "resource exhausted", HTTPStatus: 403 /* Forbidden */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrFailedPrecondition = legacyError{WebRPCError{Code: -10011, Name: "ErrFailedPrecondition", Message: "failed precondition", HTTPStatus: 412 /* Precondition Failed */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrAborted = legacyError{WebRPCError{Code: -10012, Name: "ErrAborted", Message: "aborted", HTTPStatus: 409 /* Conflict */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrOutOfRange = legacyError{WebRPCError{Code: -10013, Name: "ErrOutOfRange", Message: "out of range", HTTPStatus: 400 /* Bad Request */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrUnimplemented = legacyError{WebRPCError{Code: -10014, Name: "ErrUnimplemented", Message: "unimplemented", HTTPStatus: 501 /* Not Implemented */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrInternal = legacyError{WebRPCError{Code: -10015, Name: "ErrInternal", Message: "internal", HTTPStatus: 500 /* Internal Server Error */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrUnavailable = legacyError{WebRPCError{Code: -10016, Name: "ErrUnavailable", Message: "unavailable", HTTPStatus: 503 /* Service Unavailable */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrDataLoss = legacyError{WebRPCError{Code: -10017, Name: "ErrDataLoss", Message: "data loss", HTTPStatus: 500 /* Internal Server Error */ }}
	// Deprecated. Define errors in RIDL schema.
	ErrNone = legacyError{WebRPCError{Code: -10018, Name: "ErrNone", Message: "", HTTPStatus: 200 /* OK */ }}
)

