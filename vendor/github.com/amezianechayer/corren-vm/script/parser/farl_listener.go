// Code generated from FaRl.g4 by ANTLR 4.9.2. DO NOT EDIT.

package parser // FaRl

import "github.com/antlr/antlr4/runtime/Go/antlr"

// FaRlListener is a complete listener for a parse tree produced by FaRlParser.
type FaRlListener interface {
	antlr.ParseTreeListener

	// EnterMonetaryLit is called when entering the MonetaryLit production.
	EnterMonetaryLit(c *MonetaryLitContext)

	// EnterMonetaryNoPrecision is called when entering the MonetaryNoPrecision production.
	EnterMonetaryNoPrecision(c *MonetaryNoPrecisionContext)

	// EnterMonetaryAllPrec is called when entering the MonetaryAllPrec production.
	EnterMonetaryAllPrec(c *MonetaryAllPrecContext)

	// EnterMonetaryAllNoPrec is called when entering the MonetaryAllNoPrec production.
	EnterMonetaryAllNoPrec(c *MonetaryAllNoPrecContext)

	// EnterLitAccount is called when entering the LitAccount production.
	EnterLitAccount(c *LitAccountContext)

	// EnterLitAsset is called when entering the LitAsset production.
	EnterLitAsset(c *LitAssetContext)

	// EnterLitNumber is called when entering the LitNumber production.
	EnterLitNumber(c *LitNumberContext)

	// EnterLitMonetary is called when entering the LitMonetary production.
	EnterLitMonetary(c *LitMonetaryContext)

	// EnterExprAddSub is called when entering the ExprAddSub production.
	EnterExprAddSub(c *ExprAddSubContext)

	// EnterExprLiteral is called when entering the ExprLiteral production.
	EnterExprLiteral(c *ExprLiteralContext)

	// EnterExprVariable is called when entering the ExprVariable production.
	EnterExprVariable(c *ExprVariableContext)

	// EnterAllotmentPortionConst is called when entering the allotmentPortionConst production.
	EnterAllotmentPortionConst(c *AllotmentPortionConstContext)

	// EnterAllotmentPortionConstPercent is called when entering the allotmentPortionConstPercent production.
	EnterAllotmentPortionConstPercent(c *AllotmentPortionConstPercentContext)

	// EnterAllotmentPortionVar is called when entering the allotmentPortionVar production.
	EnterAllotmentPortionVar(c *AllotmentPortionVarContext)

	// EnterAllotmentPortionRemaining is called when entering the allotmentPortionRemaining production.
	EnterAllotmentPortionRemaining(c *AllotmentPortionRemainingContext)

	// EnterDestinationInOrder is called when entering the destinationInOrder production.
	EnterDestinationInOrder(c *DestinationInOrderContext)

	// EnterDestinationMaxed is called when entering the destinationMaxed production.
	EnterDestinationMaxed(c *DestinationMaxedContext)

	// EnterDestinationAllotment is called when entering the destinationAllotment production.
	EnterDestinationAllotment(c *DestinationAllotmentContext)

	// EnterDestAccount is called when entering the DestAccount production.
	EnterDestAccount(c *DestAccountContext)

	// EnterDestMaxed is called when entering the DestMaxed production.
	EnterDestMaxed(c *DestMaxedContext)

	// EnterDestInOrder is called when entering the DestInOrder production.
	EnterDestInOrder(c *DestInOrderContext)

	// EnterDestAllotment is called when entering the DestAllotment production.
	EnterDestAllotment(c *DestAllotmentContext)

	// EnterSourceAllotment is called when entering the sourceAllotment production.
	EnterSourceAllotment(c *SourceAllotmentContext)

	// EnterSourceMaxed is called when entering the sourceMaxed production.
	EnterSourceMaxed(c *SourceMaxedContext)

	// EnterSourceInOrder is called when entering the sourceInOrder production.
	EnterSourceInOrder(c *SourceInOrderContext)

	// EnterSrcAccount is called when entering the SrcAccount production.
	EnterSrcAccount(c *SrcAccountContext)

	// EnterSrcMaxed is called when entering the SrcMaxed production.
	EnterSrcMaxed(c *SrcMaxedContext)

	// EnterSrcInOrder is called when entering the SrcInOrder production.
	EnterSrcInOrder(c *SrcInOrderContext)

	// EnterSrc is called when entering the Src production.
	EnterSrc(c *SrcContext)

	// EnterSrcAllotment is called when entering the SrcAllotment production.
	EnterSrcAllotment(c *SrcAllotmentContext)

	// EnterType_ is called when entering the type_ production.
	EnterType_(c *Type_Context)

	// EnterOrigin is called when entering the origin production.
	EnterOrigin(c *OriginContext)

	// EnterVarTyped is called when entering the VarTyped production.
	EnterVarTyped(c *VarTypedContext)

	// EnterMetaValueExpr is called when entering the MetaValueExpr production.
	EnterMetaValueExpr(c *MetaValueExprContext)

	// EnterMetaValueRatio is called when entering the MetaValueRatio production.
	EnterMetaValueRatio(c *MetaValueRatioContext)

	// EnterMetadataEntry is called when entering the metadataEntry production.
	EnterMetadataEntry(c *MetadataEntryContext)

	// EnterPrint is called when entering the Print production.
	EnterPrint(c *PrintContext)

	// EnterFail is called when entering the Fail production.
	EnterFail(c *FailContext)

	// EnterTransfer is called when entering the Transfer production.
	EnterTransfer(c *TransferContext)

	// EnterSetTxMeta is called when entering the SetTxMeta production.
	EnterSetTxMeta(c *SetTxMetaContext)

	// EnterSetTxMetaBlock is called when entering the SetTxMetaBlock production.
	EnterSetTxMetaBlock(c *SetTxMetaBlockContext)

	// EnterSetAccountMeta is called when entering the SetAccountMeta production.
	EnterSetAccountMeta(c *SetAccountMetaContext)

	// EnterScript is called when entering the script production.
	EnterScript(c *ScriptContext)

	// ExitMonetaryLit is called when exiting the MonetaryLit production.
	ExitMonetaryLit(c *MonetaryLitContext)

	// ExitMonetaryNoPrecision is called when exiting the MonetaryNoPrecision production.
	ExitMonetaryNoPrecision(c *MonetaryNoPrecisionContext)

	// ExitMonetaryAllPrec is called when exiting the MonetaryAllPrec production.
	ExitMonetaryAllPrec(c *MonetaryAllPrecContext)

	// ExitMonetaryAllNoPrec is called when exiting the MonetaryAllNoPrec production.
	ExitMonetaryAllNoPrec(c *MonetaryAllNoPrecContext)

	// ExitLitAccount is called when exiting the LitAccount production.
	ExitLitAccount(c *LitAccountContext)

	// ExitLitAsset is called when exiting the LitAsset production.
	ExitLitAsset(c *LitAssetContext)

	// ExitLitNumber is called when exiting the LitNumber production.
	ExitLitNumber(c *LitNumberContext)

	// ExitLitMonetary is called when exiting the LitMonetary production.
	ExitLitMonetary(c *LitMonetaryContext)

	// ExitExprAddSub is called when exiting the ExprAddSub production.
	ExitExprAddSub(c *ExprAddSubContext)

	// ExitExprLiteral is called when exiting the ExprLiteral production.
	ExitExprLiteral(c *ExprLiteralContext)

	// ExitExprVariable is called when exiting the ExprVariable production.
	ExitExprVariable(c *ExprVariableContext)

	// ExitAllotmentPortionConst is called when exiting the allotmentPortionConst production.
	ExitAllotmentPortionConst(c *AllotmentPortionConstContext)

	// ExitAllotmentPortionConstPercent is called when exiting the allotmentPortionConstPercent production.
	ExitAllotmentPortionConstPercent(c *AllotmentPortionConstPercentContext)

	// ExitAllotmentPortionVar is called when exiting the allotmentPortionVar production.
	ExitAllotmentPortionVar(c *AllotmentPortionVarContext)

	// ExitAllotmentPortionRemaining is called when exiting the allotmentPortionRemaining production.
	ExitAllotmentPortionRemaining(c *AllotmentPortionRemainingContext)

	// ExitDestinationInOrder is called when exiting the destinationInOrder production.
	ExitDestinationInOrder(c *DestinationInOrderContext)

	// ExitDestinationMaxed is called when exiting the destinationMaxed production.
	ExitDestinationMaxed(c *DestinationMaxedContext)

	// ExitDestinationAllotment is called when exiting the destinationAllotment production.
	ExitDestinationAllotment(c *DestinationAllotmentContext)

	// ExitDestAccount is called when exiting the DestAccount production.
	ExitDestAccount(c *DestAccountContext)

	// ExitDestMaxed is called when exiting the DestMaxed production.
	ExitDestMaxed(c *DestMaxedContext)

	// ExitDestInOrder is called when exiting the DestInOrder production.
	ExitDestInOrder(c *DestInOrderContext)

	// ExitDestAllotment is called when exiting the DestAllotment production.
	ExitDestAllotment(c *DestAllotmentContext)

	// ExitSourceAllotment is called when exiting the sourceAllotment production.
	ExitSourceAllotment(c *SourceAllotmentContext)

	// ExitSourceMaxed is called when exiting the sourceMaxed production.
	ExitSourceMaxed(c *SourceMaxedContext)

	// ExitSourceInOrder is called when exiting the sourceInOrder production.
	ExitSourceInOrder(c *SourceInOrderContext)

	// ExitSrcAccount is called when exiting the SrcAccount production.
	ExitSrcAccount(c *SrcAccountContext)

	// ExitSrcMaxed is called when exiting the SrcMaxed production.
	ExitSrcMaxed(c *SrcMaxedContext)

	// ExitSrcInOrder is called when exiting the SrcInOrder production.
	ExitSrcInOrder(c *SrcInOrderContext)

	// ExitSrc is called when exiting the Src production.
	ExitSrc(c *SrcContext)

	// ExitSrcAllotment is called when exiting the SrcAllotment production.
	ExitSrcAllotment(c *SrcAllotmentContext)

	// ExitType_ is called when exiting the type_ production.
	ExitType_(c *Type_Context)

	// ExitOrigin is called when exiting the origin production.
	ExitOrigin(c *OriginContext)

	// ExitVarTyped is called when exiting the VarTyped production.
	ExitVarTyped(c *VarTypedContext)

	// ExitMetaValueExpr is called when exiting the MetaValueExpr production.
	ExitMetaValueExpr(c *MetaValueExprContext)

	// ExitMetaValueRatio is called when exiting the MetaValueRatio production.
	ExitMetaValueRatio(c *MetaValueRatioContext)

	// ExitMetadataEntry is called when exiting the metadataEntry production.
	ExitMetadataEntry(c *MetadataEntryContext)

	// ExitPrint is called when exiting the Print production.
	ExitPrint(c *PrintContext)

	// ExitFail is called when exiting the Fail production.
	ExitFail(c *FailContext)

	// ExitTransfer is called when exiting the Transfer production.
	ExitTransfer(c *TransferContext)

	// ExitSetTxMeta is called when exiting the SetTxMeta production.
	ExitSetTxMeta(c *SetTxMetaContext)

	// ExitSetTxMetaBlock is called when exiting the SetTxMetaBlock production.
	ExitSetTxMetaBlock(c *SetTxMetaBlockContext)

	// ExitSetAccountMeta is called when exiting the SetAccountMeta production.
	ExitSetAccountMeta(c *SetAccountMetaContext)

	// ExitScript is called when exiting the script production.
	ExitScript(c *ScriptContext)
}
