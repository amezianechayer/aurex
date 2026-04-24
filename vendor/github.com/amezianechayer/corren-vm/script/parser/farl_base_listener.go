// Code generated from FaRl.g4 by ANTLR 4.9.2. DO NOT EDIT.

package parser // FaRl

import "github.com/antlr/antlr4/runtime/Go/antlr"

// BaseFaRlListener is a complete listener for a parse tree produced by FaRlParser.
type BaseFaRlListener struct{}

var _ FaRlListener = &BaseFaRlListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseFaRlListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseFaRlListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseFaRlListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseFaRlListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterMonetaryLit is called when production MonetaryLit is entered.
func (s *BaseFaRlListener) EnterMonetaryLit(ctx *MonetaryLitContext) {}

// ExitMonetaryLit is called when production MonetaryLit is exited.
func (s *BaseFaRlListener) ExitMonetaryLit(ctx *MonetaryLitContext) {}

// EnterMonetaryNoPrecision is called when production MonetaryNoPrecision is entered.
func (s *BaseFaRlListener) EnterMonetaryNoPrecision(ctx *MonetaryNoPrecisionContext) {}

// ExitMonetaryNoPrecision is called when production MonetaryNoPrecision is exited.
func (s *BaseFaRlListener) ExitMonetaryNoPrecision(ctx *MonetaryNoPrecisionContext) {}

// EnterMonetaryAllPrec is called when production MonetaryAllPrec is entered.
func (s *BaseFaRlListener) EnterMonetaryAllPrec(ctx *MonetaryAllPrecContext) {}

// ExitMonetaryAllPrec is called when production MonetaryAllPrec is exited.
func (s *BaseFaRlListener) ExitMonetaryAllPrec(ctx *MonetaryAllPrecContext) {}

// EnterMonetaryAllNoPrec is called when production MonetaryAllNoPrec is entered.
func (s *BaseFaRlListener) EnterMonetaryAllNoPrec(ctx *MonetaryAllNoPrecContext) {}

// ExitMonetaryAllNoPrec is called when production MonetaryAllNoPrec is exited.
func (s *BaseFaRlListener) ExitMonetaryAllNoPrec(ctx *MonetaryAllNoPrecContext) {}

// EnterLitAccount is called when production LitAccount is entered.
func (s *BaseFaRlListener) EnterLitAccount(ctx *LitAccountContext) {}

// ExitLitAccount is called when production LitAccount is exited.
func (s *BaseFaRlListener) ExitLitAccount(ctx *LitAccountContext) {}

// EnterLitAsset is called when production LitAsset is entered.
func (s *BaseFaRlListener) EnterLitAsset(ctx *LitAssetContext) {}

// ExitLitAsset is called when production LitAsset is exited.
func (s *BaseFaRlListener) ExitLitAsset(ctx *LitAssetContext) {}

// EnterLitNumber is called when production LitNumber is entered.
func (s *BaseFaRlListener) EnterLitNumber(ctx *LitNumberContext) {}

// ExitLitNumber is called when production LitNumber is exited.
func (s *BaseFaRlListener) ExitLitNumber(ctx *LitNumberContext) {}

// EnterLitMonetary is called when production LitMonetary is entered.
func (s *BaseFaRlListener) EnterLitMonetary(ctx *LitMonetaryContext) {}

// ExitLitMonetary is called when production LitMonetary is exited.
func (s *BaseFaRlListener) ExitLitMonetary(ctx *LitMonetaryContext) {}

// EnterExprAddSub is called when production ExprAddSub is entered.
func (s *BaseFaRlListener) EnterExprAddSub(ctx *ExprAddSubContext) {}

// ExitExprAddSub is called when production ExprAddSub is exited.
func (s *BaseFaRlListener) ExitExprAddSub(ctx *ExprAddSubContext) {}

// EnterExprLiteral is called when production ExprLiteral is entered.
func (s *BaseFaRlListener) EnterExprLiteral(ctx *ExprLiteralContext) {}

// ExitExprLiteral is called when production ExprLiteral is exited.
func (s *BaseFaRlListener) ExitExprLiteral(ctx *ExprLiteralContext) {}

// EnterExprVariable is called when production ExprVariable is entered.
func (s *BaseFaRlListener) EnterExprVariable(ctx *ExprVariableContext) {}

// ExitExprVariable is called when production ExprVariable is exited.
func (s *BaseFaRlListener) ExitExprVariable(ctx *ExprVariableContext) {}

// EnterAllotmentPortionConst is called when production allotmentPortionConst is entered.
func (s *BaseFaRlListener) EnterAllotmentPortionConst(ctx *AllotmentPortionConstContext) {}

// ExitAllotmentPortionConst is called when production allotmentPortionConst is exited.
func (s *BaseFaRlListener) ExitAllotmentPortionConst(ctx *AllotmentPortionConstContext) {}

// EnterAllotmentPortionConstPercent is called when production allotmentPortionConstPercent is entered.
func (s *BaseFaRlListener) EnterAllotmentPortionConstPercent(ctx *AllotmentPortionConstPercentContext) {
}

// ExitAllotmentPortionConstPercent is called when production allotmentPortionConstPercent is exited.
func (s *BaseFaRlListener) ExitAllotmentPortionConstPercent(ctx *AllotmentPortionConstPercentContext) {
}

// EnterAllotmentPortionVar is called when production allotmentPortionVar is entered.
func (s *BaseFaRlListener) EnterAllotmentPortionVar(ctx *AllotmentPortionVarContext) {}

// ExitAllotmentPortionVar is called when production allotmentPortionVar is exited.
func (s *BaseFaRlListener) ExitAllotmentPortionVar(ctx *AllotmentPortionVarContext) {}

// EnterAllotmentPortionRemaining is called when production allotmentPortionRemaining is entered.
func (s *BaseFaRlListener) EnterAllotmentPortionRemaining(ctx *AllotmentPortionRemainingContext) {}

// ExitAllotmentPortionRemaining is called when production allotmentPortionRemaining is exited.
func (s *BaseFaRlListener) ExitAllotmentPortionRemaining(ctx *AllotmentPortionRemainingContext) {}

// EnterDestinationInOrder is called when production destinationInOrder is entered.
func (s *BaseFaRlListener) EnterDestinationInOrder(ctx *DestinationInOrderContext) {}

// ExitDestinationInOrder is called when production destinationInOrder is exited.
func (s *BaseFaRlListener) ExitDestinationInOrder(ctx *DestinationInOrderContext) {}

// EnterDestinationMaxed is called when production destinationMaxed is entered.
func (s *BaseFaRlListener) EnterDestinationMaxed(ctx *DestinationMaxedContext) {}

// ExitDestinationMaxed is called when production destinationMaxed is exited.
func (s *BaseFaRlListener) ExitDestinationMaxed(ctx *DestinationMaxedContext) {}

// EnterDestinationAllotment is called when production destinationAllotment is entered.
func (s *BaseFaRlListener) EnterDestinationAllotment(ctx *DestinationAllotmentContext) {}

// ExitDestinationAllotment is called when production destinationAllotment is exited.
func (s *BaseFaRlListener) ExitDestinationAllotment(ctx *DestinationAllotmentContext) {}

// EnterDestAccount is called when production DestAccount is entered.
func (s *BaseFaRlListener) EnterDestAccount(ctx *DestAccountContext) {}

// ExitDestAccount is called when production DestAccount is exited.
func (s *BaseFaRlListener) ExitDestAccount(ctx *DestAccountContext) {}

// EnterDestMaxed is called when production DestMaxed is entered.
func (s *BaseFaRlListener) EnterDestMaxed(ctx *DestMaxedContext) {}

// ExitDestMaxed is called when production DestMaxed is exited.
func (s *BaseFaRlListener) ExitDestMaxed(ctx *DestMaxedContext) {}

// EnterDestInOrder is called when production DestInOrder is entered.
func (s *BaseFaRlListener) EnterDestInOrder(ctx *DestInOrderContext) {}

// ExitDestInOrder is called when production DestInOrder is exited.
func (s *BaseFaRlListener) ExitDestInOrder(ctx *DestInOrderContext) {}

// EnterDestAllotment is called when production DestAllotment is entered.
func (s *BaseFaRlListener) EnterDestAllotment(ctx *DestAllotmentContext) {}

// ExitDestAllotment is called when production DestAllotment is exited.
func (s *BaseFaRlListener) ExitDestAllotment(ctx *DestAllotmentContext) {}

// EnterSourceAllotment is called when production sourceAllotment is entered.
func (s *BaseFaRlListener) EnterSourceAllotment(ctx *SourceAllotmentContext) {}

// ExitSourceAllotment is called when production sourceAllotment is exited.
func (s *BaseFaRlListener) ExitSourceAllotment(ctx *SourceAllotmentContext) {}

// EnterSourceMaxed is called when production sourceMaxed is entered.
func (s *BaseFaRlListener) EnterSourceMaxed(ctx *SourceMaxedContext) {}

// ExitSourceMaxed is called when production sourceMaxed is exited.
func (s *BaseFaRlListener) ExitSourceMaxed(ctx *SourceMaxedContext) {}

// EnterSourceInOrder is called when production sourceInOrder is entered.
func (s *BaseFaRlListener) EnterSourceInOrder(ctx *SourceInOrderContext) {}

// ExitSourceInOrder is called when production sourceInOrder is exited.
func (s *BaseFaRlListener) ExitSourceInOrder(ctx *SourceInOrderContext) {}

// EnterSrcAccount is called when production SrcAccount is entered.
func (s *BaseFaRlListener) EnterSrcAccount(ctx *SrcAccountContext) {}

// ExitSrcAccount is called when production SrcAccount is exited.
func (s *BaseFaRlListener) ExitSrcAccount(ctx *SrcAccountContext) {}

// EnterSrcMaxed is called when production SrcMaxed is entered.
func (s *BaseFaRlListener) EnterSrcMaxed(ctx *SrcMaxedContext) {}

// ExitSrcMaxed is called when production SrcMaxed is exited.
func (s *BaseFaRlListener) ExitSrcMaxed(ctx *SrcMaxedContext) {}

// EnterSrcInOrder is called when production SrcInOrder is entered.
func (s *BaseFaRlListener) EnterSrcInOrder(ctx *SrcInOrderContext) {}

// ExitSrcInOrder is called when production SrcInOrder is exited.
func (s *BaseFaRlListener) ExitSrcInOrder(ctx *SrcInOrderContext) {}

// EnterSrc is called when production Src is entered.
func (s *BaseFaRlListener) EnterSrc(ctx *SrcContext) {}

// ExitSrc is called when production Src is exited.
func (s *BaseFaRlListener) ExitSrc(ctx *SrcContext) {}

// EnterSrcAllotment is called when production SrcAllotment is entered.
func (s *BaseFaRlListener) EnterSrcAllotment(ctx *SrcAllotmentContext) {}

// ExitSrcAllotment is called when production SrcAllotment is exited.
func (s *BaseFaRlListener) ExitSrcAllotment(ctx *SrcAllotmentContext) {}

// EnterType_ is called when production type_ is entered.
func (s *BaseFaRlListener) EnterType_(ctx *Type_Context) {}

// ExitType_ is called when production type_ is exited.
func (s *BaseFaRlListener) ExitType_(ctx *Type_Context) {}

// EnterOrigin is called when production origin is entered.
func (s *BaseFaRlListener) EnterOrigin(ctx *OriginContext) {}

// ExitOrigin is called when production origin is exited.
func (s *BaseFaRlListener) ExitOrigin(ctx *OriginContext) {}

// EnterVarTyped is called when production VarTyped is entered.
func (s *BaseFaRlListener) EnterVarTyped(ctx *VarTypedContext) {}

// ExitVarTyped is called when production VarTyped is exited.
func (s *BaseFaRlListener) ExitVarTyped(ctx *VarTypedContext) {}

// EnterMetaValueExpr is called when production MetaValueExpr is entered.
func (s *BaseFaRlListener) EnterMetaValueExpr(ctx *MetaValueExprContext) {}

// ExitMetaValueExpr is called when production MetaValueExpr is exited.
func (s *BaseFaRlListener) ExitMetaValueExpr(ctx *MetaValueExprContext) {}

// EnterMetaValueRatio is called when production MetaValueRatio is entered.
func (s *BaseFaRlListener) EnterMetaValueRatio(ctx *MetaValueRatioContext) {}

// ExitMetaValueRatio is called when production MetaValueRatio is exited.
func (s *BaseFaRlListener) ExitMetaValueRatio(ctx *MetaValueRatioContext) {}

// EnterMetadataEntry is called when production metadataEntry is entered.
func (s *BaseFaRlListener) EnterMetadataEntry(ctx *MetadataEntryContext) {}

// ExitMetadataEntry is called when production metadataEntry is exited.
func (s *BaseFaRlListener) ExitMetadataEntry(ctx *MetadataEntryContext) {}

// EnterPrint is called when production Print is entered.
func (s *BaseFaRlListener) EnterPrint(ctx *PrintContext) {}

// ExitPrint is called when production Print is exited.
func (s *BaseFaRlListener) ExitPrint(ctx *PrintContext) {}

// EnterFail is called when production Fail is entered.
func (s *BaseFaRlListener) EnterFail(ctx *FailContext) {}

// ExitFail is called when production Fail is exited.
func (s *BaseFaRlListener) ExitFail(ctx *FailContext) {}

// EnterTransfer is called when production Transfer is entered.
func (s *BaseFaRlListener) EnterTransfer(ctx *TransferContext) {}

// ExitTransfer is called when production Transfer is exited.
func (s *BaseFaRlListener) ExitTransfer(ctx *TransferContext) {}

// EnterSetTxMeta is called when production SetTxMeta is entered.
func (s *BaseFaRlListener) EnterSetTxMeta(ctx *SetTxMetaContext) {}

// ExitSetTxMeta is called when production SetTxMeta is exited.
func (s *BaseFaRlListener) ExitSetTxMeta(ctx *SetTxMetaContext) {}

// EnterSetTxMetaBlock is called when production SetTxMetaBlock is entered.
func (s *BaseFaRlListener) EnterSetTxMetaBlock(ctx *SetTxMetaBlockContext) {}

// ExitSetTxMetaBlock is called when production SetTxMetaBlock is exited.
func (s *BaseFaRlListener) ExitSetTxMetaBlock(ctx *SetTxMetaBlockContext) {}

// EnterSetAccountMeta is called when production SetAccountMeta is entered.
func (s *BaseFaRlListener) EnterSetAccountMeta(ctx *SetAccountMetaContext) {}

// ExitSetAccountMeta is called when production SetAccountMeta is exited.
func (s *BaseFaRlListener) ExitSetAccountMeta(ctx *SetAccountMetaContext) {}

// EnterScript is called when production script is entered.
func (s *BaseFaRlListener) EnterScript(ctx *ScriptContext) {}

// ExitScript is called when production script is exited.
func (s *BaseFaRlListener) ExitScript(ctx *ScriptContext) {}
