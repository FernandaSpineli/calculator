// Package api exposes the calculator as a RESTful HTTP API built on Gin.
package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/FernandaSpineli/calculator/internal/calculator"
	"github.com/FernandaSpineli/calculator/internal/store"
)

const basePath = "/api/v1"

type handler struct {
	store *store.Memory
}

// NewRouter wires every route to its handler.
func NewRouter(s *store.Memory) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	h := &handler{store: s}
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	v1 := r.Group(basePath)
	v1.GET("/operations", h.listOperations)
	v1.GET("/operations/:name", h.getOperation)
	v1.POST("/calculations", h.createCalculation)
	v1.GET("/calculations", h.listCalculations)
	v1.GET("/calculations/:id", h.getCalculation)
	v1.DELETE("/calculations/:id", h.deleteCalculation)

	return r
}

type createCalculationRequest struct {
	Operation string    `json:"operation" binding:"required"`
	Operands  []float64 `json:"operands"`
	AngleUnit string    `json:"angle_unit"`
}

func (h *handler) listOperations(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"operations": calculator.Operations()})
}

func (h *handler) getOperation(c *gin.Context) {
	op, ok := calculator.Lookup(c.Param("name"))
	if !ok {
		abort(c, http.StatusNotFound, "operation_not_found", "operation "+strconv.Quote(c.Param("name"))+" does not exist")
		return
	}
	c.JSON(http.StatusOK, op)
}

func (h *handler) createCalculation(c *gin.Context) {
	var req createCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abort(c, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	unit, err := calculator.ParseAngleUnit(req.AngleUnit)
	if err != nil {
		abort(c, http.StatusUnprocessableEntity, "invalid_angle_unit", err.Error())
		return
	}
	result, err := calculator.Evaluate(req.Operation, req.Operands, unit)
	if err != nil {
		status, code := evaluationError(err)
		abort(c, status, code, err.Error())
		return
	}

	calc := store.Calculation{Operation: req.Operation, Operands: req.Operands, Result: result}
	if op, _ := calculator.Lookup(req.Operation); op.UsesAngleUnit {
		calc.AngleUnit = string(unit)
	}
	calc = h.store.Create(calc)

	c.Header("Location", basePath+"/calculations/"+strconv.FormatInt(calc.ID, 10))
	c.JSON(http.StatusCreated, calc)
}

func (h *handler) listCalculations(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"calculations": h.store.List()})
}

func (h *handler) getCalculation(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	calc, err := h.store.Get(id)
	if err != nil {
		abort(c, http.StatusNotFound, "calculation_not_found", err.Error())
		return
	}
	c.JSON(http.StatusOK, calc)
}

func (h *handler) deleteCalculation(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.store.Delete(id); err != nil {
		abort(c, http.StatusNotFound, "calculation_not_found", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		abort(c, http.StatusBadRequest, "invalid_id", "id must be a positive integer")
		return 0, false
	}
	return id, true
}

var evaluationErrorCodes = []struct {
	err  error
	code string
}{
	{calculator.ErrUnknownOperation, "unknown_operation"},
	{calculator.ErrOperandCount, "wrong_operand_count"},
	{calculator.ErrDivisionByZero, "division_by_zero"},
	{calculator.ErrDomain, "domain_error"},
	{calculator.ErrOverflow, "overflow"},
}

// evaluationError maps a calculator error to an HTTP status and error code.
func evaluationError(err error) (int, string) {
	for _, e := range evaluationErrorCodes {
		if errors.Is(err, e.err) {
			return http.StatusUnprocessableEntity, e.code
		}
	}
	return http.StatusInternalServerError, "internal_error"
}

func abort(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
