package handlers

import "github.com/gin-gonic/gin"

type GuestBookHandler struct {
}

func NewGuestBookHandler() *GuestBookHandler {
	return &GuestBookHandler{}
}

func (h *GuestBookHandler) CreateEntry(c *gin.Context) {

}

func (h *GuestBookHandler) GetEntries(c *gin.Context) {

}
func (h *GuestBookHandler) GetEntryByID(c *gin.Context) {

}
func (h *GuestBookHandler) UpdateEntry(c *gin.Context) {

}
func (h *GuestBookHandler) DeleteEntry(c *gin.Context) {

}
