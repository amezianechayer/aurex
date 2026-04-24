package compiler

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/amezianechayer/corren-vm/core"
	"github.com/amezianechayer/corren-vm/script/parser"
	"github.com/amezianechayer/corren-vm/vm/program"
	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type parseVisitor struct {
	elistener       *ErrorListener
	instructions    []byte
	resources       []program.Resource
	var_idx         map[string]core.Address
	needed_balances map[core.Address]map[core.Address]struct{}
}

func (p *parseVisitor) findConstant(constant program.Constant) (*core.Address, bool) {
	for i := 0; i < len(p.resources); i++ {
		if c, ok := p.resources[i].(program.Constant); ok {
			if core.ValueEquals(c.Inner, constant.Inner) {
				addr := core.Address(i)
				return &addr, true
			}
		}
	}
	return nil, false
}

func (p *parseVisitor) AllocateResource(res program.Resource) (*core.Address, error) {
	if c, ok := res.(program.Constant); ok {
		idx, ok := p.findConstant(c)
		if ok {
			return idx, nil
		}
	}
	if len(p.resources) >= 65536 {
		return nil, errors.New("number of unique constants exceeded 65536")
	}
	p.resources = append(p.resources, res)
	addr := core.NewAddress(uint16(len(p.resources) - 1))
	return &addr, nil
}

func (p *parseVisitor) PushAddress(addr core.Address) {
	p.instructions = append(p.instructions, program.OP_APUSH)
	bytes := addr.ToBytes()
	p.instructions = append(p.instructions, bytes...)
}

func (p *parseVisitor) PushInteger(val core.Number) {
	p.instructions = append(p.instructions, program.OP_IPUSH)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(val))
	p.instructions = append(p.instructions, bytes...)
}

func (p *parseVisitor) isWorld(addr core.Address) bool {
	idx := int(addr)
	if idx < len(p.resources) {
		if c, ok := p.resources[idx].(program.Constant); ok {
			if acc, ok := c.Inner.(core.Account); ok {
				if string(acc) == "@world" {
					return true
				}
			}
		}
	}
	return false
}

func (p *parseVisitor) VisitVariable(c parser.IExpressionContext, push bool) (core.Type, *core.Address, *CompileError) {
	if ctx, ok := c.(*parser.ExprVariableContext); ok {
		name := ctx.GetVariable().GetText()[1:]
		if idx, ok := p.var_idx[name]; ok {
			res := p.resources[idx]
			if push {
				p.PushAddress(idx)
			}
			return res.GetType(), &idx, nil
		}
		return 0, nil, LogicError(ctx, errors.New("variable not declared"))
	}
	return 0, nil, &CompileError{Msg: "expected variable"}
}

func (p *parseVisitor) VisitScript(c parser.IScriptContext) *CompileError {
	switch c := c.(type) {
	case *parser.ScriptContext:
		for _, v := range c.AllVarDecl() {
			err := p.VisitVarDecl(v)
			if err != nil {
				return err
			}
		}
		for _, stmt := range c.GetStmts() {
			switch c := stmt.(type) {
			case *parser.PrintContext:
				err := p.VisitPrint(c)
				if err != nil {
					return err
				}
			case *parser.FailContext:
				p.instructions = append(p.instructions, program.OP_FAIL)
			case *parser.TransferContext:
				err := p.VisitTransfer(c)
				if err != nil {
					return err
				}
			default:
				return InternalError(c)
			}
		}
	default:
		return InternalError(c)
	}
	return nil
}

func (p *parseVisitor) VisitVarDecl(v parser.IVarDeclContext) *CompileError {
	switch v := v.(type) {
	case *parser.VarTypedContext:
		name := v.GetName().GetText()[1:]
		if _, ok := p.var_idx[name]; ok {
			return LogicError(v, fmt.Errorf("duplicate variable: %v", name))
		}
		var ty core.Type
		switch v.GetTy().GetText() {
		case "account":
			ty = core.TYPE_ACCOUNT
		case "asset":
			ty = core.TYPE_ASSET
		case "number":
			ty = core.TYPE_NUMBER
		case "monetary":
			ty = core.TYPE_MONETARY
		case "portion":
			ty = core.TYPE_PORTION
		default:
			return InternalError(v)
		}
		var addr core.Address
		c_orig := v.GetOrig()
		if c_orig != nil {
			src_ty, src, cerr := p.VisitExpr(c_orig.GetAcc(), false)
			if cerr != nil {
				return cerr
			}
			if src_ty != core.TYPE_ACCOUNT {
				return LogicError(c_orig, errors.New("wrong type: expected account"))
			}
			key := strings.Trim(c_orig.GetKey().GetText(), `"`)
			a, err := p.AllocateResource(program.Metadata{SourceAccount: *src, Key: key, Typ: ty})
			if err != nil {
				return LogicError(c_orig, err)
			}
			addr = *a
		} else {
			a, err := p.AllocateResource(program.Parameter{Typ: ty, Name: name})
			if err != nil {
				return LogicError(v, err)
			}
			addr = *a
		}
		p.var_idx[name] = addr
	default:
		return InternalError(v)
	}
	return nil
}

func (p *parseVisitor) VisitPrint(ctx *parser.PrintContext) *CompileError {
	_, _, err := p.VisitExpr(ctx.GetExpr(), true)
	if err != nil {
		return err
	}
	p.instructions = append(p.instructions, program.OP_PRINT)
	return nil
}

func (p *parseVisitor) VisitExpr(ctx parser.IExpressionContext, push bool) (core.Type, *core.Address, *CompileError) {
	switch ctx := ctx.(type) {
	case *parser.ExprAddSubContext:
		ty, _, err := p.VisitExpr(ctx.GetLhs(), push)
		if err != nil {
			return 0, nil, err
		}
		if ty != core.TYPE_NUMBER {
			return 0, nil, LogicError(ctx, errors.New("tried to do arithmetic with wrong type"))
		}
		ty, _, err = p.VisitExpr(ctx.GetRhs(), push)
		if err != nil {
			return 0, nil, err
		}
		if ty != core.TYPE_NUMBER {
			return 0, nil, LogicError(ctx, errors.New("tried to do arithmetic with wrong type"))
		}
		if push {
			switch ctx.GetOp().GetTokenType() {
			case parser.FaRlLexerOP_ADD:
				p.instructions = append(p.instructions, program.OP_IADD)
			case parser.FaRlLexerOP_SUB:
				p.instructions = append(p.instructions, program.OP_ISUB)
			}
		}
		return core.TYPE_NUMBER, nil, nil
	case *parser.ExprLiteralContext:
		ty, addr, err := p.VisitLit(ctx.GetLit(), push)
		if err != nil {
			return 0, nil, err
		}
		return ty, addr, nil
	case *parser.ExprVariableContext:
		name := ctx.GetVariable().GetText()[1:]
		if idx, ok := p.var_idx[name]; ok {
			res := p.resources[idx]
			if push {
				p.PushAddress(idx)
			}
			return res.GetType(), &idx, nil
		}
		return 0, nil, LogicError(ctx, fmt.Errorf("variable not declared: %v", name))
	default:
		return 0, nil, InternalError(ctx)
	}
}

func (p *parseVisitor) VisitLit(c parser.ILiteralContext, push bool) (core.Type, *core.Address, *CompileError) {
	switch c := c.(type) {
	case *parser.LitAccountContext:
		account := core.Account(c.ACCOUNT().GetText())
		addr, err := p.AllocateResource(program.Constant{Inner: account})
		if err != nil {
			return 0, nil, LogicError(c, err)
		}
		if push {
			p.PushAddress(*addr)
		}
		return core.TYPE_ACCOUNT, addr, nil
	case *parser.LitAssetContext:
		asset := core.Asset(c.GetText())
		addr, err := p.AllocateResource(program.Constant{Inner: asset})
		if err != nil {
			return 0, nil, LogicError(c, err)
		}
		if push {
			p.PushAddress(*addr)
		}
		return core.TYPE_ASSET, addr, nil
	case *parser.LitNumberContext:
		n, err := strconv.ParseUint(c.GetText(), 10, 64)
		if err != nil {
			return 0, nil, LogicError(c, err)
		}
		if push {
			p.PushInteger(core.Number(n))
		}
		return core.TYPE_NUMBER, nil, nil
	case *parser.LitMonetaryContext:
		switch m := c.Monetary().(type) {
		case *parser.MonetaryLitContext:
			asset := m.GetAsset().GetText()
			precision := m.GetPrecision().GetText()
			a, err := strconv.ParseUint(m.GetAmount().GetText(), 10, 64)
			if err != nil {
				return 0, nil, LogicError(c, err)
			}
			monetary := core.Monetary{
				Asset:  core.Asset(asset + "." + precision),
				Amount: a,
			}
			addr, err := p.AllocateResource(program.Constant{Inner: monetary})
			if err != nil {
				return 0, nil, LogicError(c, err)
			}
			if push {
				p.PushAddress(*addr)
			}
			return core.TYPE_MONETARY, addr, nil
		case *parser.MonetaryNoPrecisionContext:
			asset := m.GetAsset().GetText()
			a, err := strconv.ParseUint(m.GetAmount().GetText(), 10, 64)
			if err != nil {
				return 0, nil, LogicError(c, err)
			}
			monetary := core.Monetary{
				Asset:  core.Asset(asset),
				Amount: a,
			}
			addr, err := p.AllocateResource(program.Constant{Inner: monetary})
			if err != nil {
				return 0, nil, LogicError(c, err)
			}
			if push {
				p.PushAddress(*addr)
			}
			return core.TYPE_MONETARY, addr, nil
		default:
			return 0, nil, InternalError(c)
		}
	default:
		return 0, nil, InternalError(c)
	}
}

func (p *parseVisitor) VisitTransfer(c *parser.TransferContext) *CompileError {
	var asset_addr core.Address
	var needed_accounts map[core.Address]struct{}

	if monAll := c.GetMonAll(); monAll != nil {
		// transfer [ASSET *] — send all
		var asset core.Asset
		switch m := monAll.(type) {
		case *parser.MonetaryAllPrecContext:
			asset = core.Asset(m.GetAsset().GetText() + "." + m.GetPrecision().GetText())
		case *parser.MonetaryAllNoPrecContext:
			asset = core.Asset(m.GetAsset().GetText())
		}
		addr, err := p.AllocateResource(program.Constant{Inner: asset})
		if err != nil {
			return LogicError(c, err)
		}
		asset_addr = *addr
		accounts, cerr := p.VisitValueAwareSource(c.GetSrc(), func() {
			p.PushAddress(*addr)
		}, nil)
		if cerr != nil {
			return cerr
		}
		needed_accounts = accounts
	} else if mon := c.GetMon(); mon != nil {
		ty, mon_addr, err := p.VisitExpr(mon, false)
		if err != nil {
			return err
		}
		if ty != core.TYPE_MONETARY {
			return LogicError(c, errors.New("wrong type for monetary value"))
		}
		asset_addr = *mon_addr
		accounts, cerr := p.VisitValueAwareSource(c.GetSrc(), func() {
			p.PushAddress(*mon_addr)
			p.instructions = append(p.instructions, program.OP_ASSET)
		}, mon_addr)
		if cerr != nil {
			return cerr
		}
		needed_accounts = accounts
	}

	for acc := range needed_accounts {
		if b, ok := p.needed_balances[acc]; ok {
			b[asset_addr] = struct{}{}
		} else {
			p.needed_balances[acc] = map[core.Address]struct{}{
				asset_addr: {},
			}
		}
	}

	return p.VisitDestination(c.GetDest())
}

type CompileArtifacts struct {
	Source  string
	Tokens  []antlr.Token
	Errors  []CompileError
	Program *program.Program
}

func CompileFull(input string) CompileArtifacts {
	artifacts := CompileArtifacts{
		Source: input,
	}

	elistener := &ErrorListener{}

	is := antlr.NewInputStream(input)
	lexer := parser.NewFaRlLexer(is)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(elistener)

	stream := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)
	p := parser.NewFaRlParser(stream)
	p.RemoveErrorListeners()
	p.AddErrorListener(elistener)

	p.BuildParseTrees = true

	tree := p.Script()

	artifacts.Tokens = stream.GetAllTokens()
	artifacts.Errors = append(artifacts.Errors, elistener.Errors...)

	if len(elistener.Errors) != 0 {
		return artifacts
	}

	visitor := parseVisitor{
		elistener:       elistener,
		instructions:    make([]byte, 0),
		resources:       make([]program.Resource, 0),
		var_idx:         make(map[string]core.Address),
		needed_balances: make(map[core.Address]map[core.Address]struct{}),
	}

	err := visitor.VisitScript(tree)

	if err != nil {
		artifacts.Errors = append(artifacts.Errors, *err)
		return artifacts
	}

	artifacts.Program = &program.Program{
		Instructions:   visitor.instructions,
		Resources:      visitor.resources,
		NeededBalances: visitor.needed_balances,
	}

	return artifacts
}

func Compile(input string) (*program.Program, error) {
	artifacts := CompileFull(input)
	if len(artifacts.Errors) > 0 {
		err := CompileErrorList{
			Errors: artifacts.Errors,
			Source: artifacts.Source,
		}
		return nil, &err
	}
	return artifacts.Program, nil
}
