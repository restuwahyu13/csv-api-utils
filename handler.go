package main

import (
	"fmt"
	"net/http"
	"path"

	"github.com/goccy/go-json"
)

type InterfaceHandler interface {
	Ping(rw http.ResponseWriter, r *http.Request)
	Merge(rw http.ResponseWriter, r *http.Request)
	Split(rw http.ResponseWriter, r *http.Request)
}

type StructHandler struct {
	service InterfaceService
}

func NewHandler(service InterfaceService) InterfaceHandler {
	return &StructHandler{service: service}
}

/*
* ================================
* PING HANDLER
* ================================
 */

func (h *StructHandler) Ping(rw http.ResponseWriter, r *http.Request) {
	APIResponse(rw, &ApiResponse{StatCode: http.StatusOK, StatMessage: "Pong"})
}

/*
* ================================
* MERGE HANDLER
* ================================
 */

func (h *StructHandler) Merge(rw http.ResponseWriter, r *http.Request) {
	var (
		req CSVMergePayload = CSVMergePayload{}
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		APIResponse(rw, &ApiResponse{StatCode: http.StatusUnprocessableEntity, ErrMessage: err.Error()})
		return
	}

	if !path.IsAbs(req.InputDir) || !path.IsAbs(req.OutputDir) {
		APIResponse(rw, &ApiResponse{StatCode: http.StatusUnprocessableEntity, ErrMessage: "InputDir or OutputDir must be absolute path"})
		return
	}

	_, err := h.service.Merge(&req)
	if err != nil {
		APIResponse(rw, &ApiResponse{StatCode: http.StatusUnprocessableEntity, ErrMessage: err.Error()})
		return
	}

	APIResponse(rw, &ApiResponse{StatCode: http.StatusOK, StatMessage: fmt.Sprintf("CSV file output location: %s", req.OutputDir)})
}

/*
* ================================
* SPLIT HANDLER
* ================================
 */

func (h *StructHandler) Split(rw http.ResponseWriter, r *http.Request) {
	var (
		req CSVSplitPayload = CSVSplitPayload{}
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		APIResponse(rw, &ApiResponse{StatCode: http.StatusUnprocessableEntity, ErrMessage: err.Error()})
		return
	}

	if !path.IsAbs(req.InputFile) || !path.IsAbs(req.InputFile) {
		APIResponse(rw, &ApiResponse{StatCode: http.StatusUnprocessableEntity, ErrMessage: "InputFile or OutputFile must be absolute path"})
		return
	} else if path.Ext(req.InputFile) != ".csv" {
		APIResponse(rw, &ApiResponse{StatCode: http.StatusUnprocessableEntity, ErrMessage: "InputFile not csv"})
		return
	}

	_, err := h.service.Split(&req)
	if err != nil {
		APIResponse(rw, &ApiResponse{StatCode: http.StatusUnprocessableEntity, ErrMessage: err.Error()})
		return
	}

	APIResponse(rw, &ApiResponse{StatCode: http.StatusOK, StatMessage: fmt.Sprintf("CSV file output location: %s", req.OutputFile)})
}
