package middlewares

import (
	"github.com/amezianechayer/corren/ledger"
	"github.com/gin-gonic/gin"
)

func LedgerMiddleware(resolver *ledger.Resolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("ledger")

		if name == "" {
			return
		}

		l, err := resolver.GetLedger(name)

		if err != nil {
			c.JSON(400, gin.H{
				"ok":  false,
				"err": err.Error(),
			})
		}

		c.Set("ledger", l)
	}
}
