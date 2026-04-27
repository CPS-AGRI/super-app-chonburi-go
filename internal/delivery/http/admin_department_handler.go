package http

import (
	"strconv"
	"super-app-chonburi-go/internal/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type AdminDepartmentHandler struct {
	useCase domain.AdminDepartmentUseCase
}

func NewAdminDepartmentHandler(useCase domain.AdminDepartmentUseCase) *AdminDepartmentHandler {
	return &AdminDepartmentHandler{useCase: useCase}
}

func (h *AdminDepartmentHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/departments")
	group.Get("/", h.GetDepartments)
	group.Get("/:id", h.GetDepartmentByID)
	group.Post("/", h.CreateDepartment)
	group.Put("/:id", h.UpdateDepartment)
	group.Delete("/:id", h.DeleteDepartment)
}

func (h *AdminDepartmentHandler) GetDepartments(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("pageNumber", "1"))
	size, _ := strconv.Atoi(c.Query("pageSize", "20"))
	name := c.Query("name", "")

	query := domain.AdminDepartmentQuery{
		PageNumber: page,
		PageSize:   size,
		Name:       name,
	}

	response, err := h.useCase.GetDepartments(query)
	if err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	return SuccessResponse(c, response)
}

func (h *AdminDepartmentHandler) GetDepartmentByID(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return ErrorResponse(c, "invalid department ID", fiber.StatusBadRequest)
	}

	department, err := h.useCase.GetDepartmentByID(id)
	if err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusNotFound)
	}

	return SuccessResponse(c, department)
}

func (h *AdminDepartmentHandler) CreateDepartment(c fiber.Ctx) error {
	var department domain.AdminDepartment
	if err := c.Bind().JSON(&department); err != nil {
		return ErrorResponse(c, "invalid request body", fiber.StatusBadRequest)
	}

	if err := h.useCase.CreateDepartment(&department); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	return SuccessResponse(c, department, fiber.StatusCreated)
}

func (h *AdminDepartmentHandler) UpdateDepartment(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return ErrorResponse(c, "invalid department ID", fiber.StatusBadRequest)
	}

	var department domain.AdminDepartment
	if err := c.Bind().JSON(&department); err != nil {
		return ErrorResponse(c, "invalid request body", fiber.StatusBadRequest)
	}

	department.ID = id
	if err := h.useCase.UpdateDepartment(&department); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	return SuccessResponse(c, department)
}

func (h *AdminDepartmentHandler) DeleteDepartment(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return ErrorResponse(c, "invalid department ID", fiber.StatusBadRequest)
	}

	if err := h.useCase.DeleteDepartment(id); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	return SuccessResponse(c, nil)
}
