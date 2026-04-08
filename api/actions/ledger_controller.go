package actions

import (
	"github.com/amezianechayer/corren/ledger"
	"github.com/gin-gonic/gin"
)

// LedgerController -
type LedgerController struct {
	BaseController
}

// NewLedgerController -
func NewLedgerController() *LedgerController {
	return &LedgerController{}
}

// CreateLedgerController -
func CreateLedgerController() *LedgerController {
	return NewLedgerController()
}

// GetStats -
func (ctl *LedgerController) GetStats(c *gin.Context) {
	l, _ := c.Get("ledger")

	stats, err := l.(*ledger.Ledger).Stats()

	c.JSON(200, gin.H{
		"ok":    err == nil,
		"stats": stats,
	})
}
