package actions

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewConfigController),
	fx.Provide(NewLedgerController),
	fx.Provide(NewScriptController),
	fx.Provide(NewAccountController),
	fx.Provide(NewTransactionController),
	fx.Provide(NewContractController),
	fx.Provide(NewGuardController),
	fx.Provide(NewAuthController),
	fx.Provide(NewLensController),
	fx.Provide(NewWalletController),
	fx.Provide(NewPaymentController),
)
