// Code generated from FaRl.g4 by ANTLR 4.9.2. DO NOT EDIT.

package parser // FaRl

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = reflect.Copy
var _ = strconv.Itoa

var parserATN = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 48, 290,
	4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7, 9, 7,
	4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12, 4, 13,
	9, 13, 4, 14, 9, 14, 4, 15, 9, 15, 4, 16, 9, 16, 4, 17, 9, 17, 4, 18, 9,
	18, 4, 19, 9, 19, 4, 20, 9, 20, 4, 21, 9, 21, 4, 22, 9, 22, 3, 2, 3, 2,
	3, 2, 3, 2, 3, 2, 3, 2, 3, 2, 3, 2, 3, 2, 3, 2, 5, 2, 55, 10, 2, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 5, 3, 67, 10, 3,
	3, 4, 3, 4, 3, 4, 3, 4, 5, 4, 73, 10, 4, 3, 5, 3, 5, 3, 5, 5, 5, 78, 10,
	5, 3, 5, 3, 5, 3, 5, 7, 5, 83, 10, 5, 12, 5, 14, 5, 86, 11, 5, 3, 6, 3,
	6, 3, 6, 3, 6, 5, 6, 92, 10, 6, 3, 6, 3, 6, 3, 6, 5, 6, 97, 10, 6, 3, 7,
	3, 7, 3, 7, 3, 7, 3, 7, 6, 7, 104, 10, 7, 13, 7, 14, 7, 105, 3, 7, 3, 7,
	3, 8, 3, 8, 3, 8, 3, 8, 3, 8, 3, 9, 3, 9, 3, 9, 3, 9, 3, 9, 3, 9, 3, 9,
	6, 9, 122, 10, 9, 13, 9, 14, 9, 123, 3, 9, 3, 9, 3, 10, 3, 10, 3, 10, 3,
	10, 5, 10, 132, 10, 10, 3, 11, 3, 11, 3, 11, 3, 11, 3, 11, 3, 11, 3, 11,
	6, 11, 141, 10, 11, 13, 11, 14, 11, 142, 3, 11, 3, 11, 3, 12, 3, 12, 3,
	12, 3, 12, 3, 12, 3, 13, 3, 13, 3, 13, 3, 13, 3, 13, 6, 13, 157, 10, 13,
	13, 13, 14, 13, 158, 3, 13, 3, 13, 3, 14, 3, 14, 3, 14, 5, 14, 166, 10,
	14, 3, 15, 3, 15, 5, 15, 170, 10, 15, 3, 16, 3, 16, 3, 17, 3, 17, 3, 17,
	3, 17, 3, 17, 3, 17, 3, 17, 3, 18, 3, 18, 3, 18, 3, 18, 3, 18, 3, 18, 5,
	18, 187, 10, 18, 3, 19, 3, 19, 5, 19, 191, 10, 19, 3, 20, 3, 20, 3, 20,
	3, 20, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 5, 21, 203, 10, 21, 3,
	21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21,
	3, 21, 3, 21, 3, 21, 5, 21, 219, 10, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3,
	21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21,
	3, 21, 3, 21, 6, 21, 238, 10, 21, 13, 21, 14, 21, 239, 3, 21, 3, 21, 3,
	21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 3, 21, 5, 21,
	254, 10, 21, 3, 22, 7, 22, 257, 10, 22, 12, 22, 14, 22, 260, 11, 22, 3,
	22, 3, 22, 6, 22, 264, 10, 22, 13, 22, 14, 22, 265, 7, 22, 268, 10, 22,
	12, 22, 14, 22, 271, 11, 22, 3, 22, 3, 22, 6, 22, 275, 10, 22, 13, 22,
	14, 22, 276, 3, 22, 7, 22, 280, 10, 22, 12, 22, 14, 22, 283, 11, 22, 3,
	22, 5, 22, 286, 10, 22, 3, 22, 3, 22, 3, 22, 2, 3, 8, 23, 2, 4, 6, 8, 10,
	12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 2, 4, 3,
	2, 29, 30, 3, 2, 23, 28, 2, 305, 2, 54, 3, 2, 2, 2, 4, 66, 3, 2, 2, 2,
	6, 72, 3, 2, 2, 2, 8, 77, 3, 2, 2, 2, 10, 96, 3, 2, 2, 2, 12, 98, 3, 2,
	2, 2, 14, 109, 3, 2, 2, 2, 16, 114, 3, 2, 2, 2, 18, 131, 3, 2, 2, 2, 20,
	133, 3, 2, 2, 2, 22, 146, 3, 2, 2, 2, 24, 151, 3, 2, 2, 2, 26, 165, 3,
	2, 2, 2, 28, 169, 3, 2, 2, 2, 30, 171, 3, 2, 2, 2, 32, 173, 3, 2, 2, 2,
	34, 180, 3, 2, 2, 2, 36, 190, 3, 2, 2, 2, 38, 192, 3, 2, 2, 2, 40, 253,
	3, 2, 2, 2, 42, 258, 3, 2, 2, 2, 44, 45, 7, 33, 2, 2, 45, 46, 7, 47, 2,
	2, 46, 47, 7, 39, 2, 2, 47, 48, 7, 48, 2, 2, 48, 49, 7, 48, 2, 2, 49, 55,
	7, 34, 2, 2, 50, 51, 7, 33, 2, 2, 51, 52, 7, 47, 2, 2, 52, 53, 7, 48, 2,
	2, 53, 55, 7, 34, 2, 2, 54, 44, 3, 2, 2, 2, 54, 50, 3, 2, 2, 2, 55, 3,
	3, 2, 2, 2, 56, 57, 7, 33, 2, 2, 57, 58, 7, 47, 2, 2, 58, 59, 7, 39, 2,
	2, 59, 60, 7, 48, 2, 2, 60, 61, 7, 41, 2, 2, 61, 67, 7, 34, 2, 2, 62, 63,
	7, 33, 2, 2, 63, 64, 7, 47, 2, 2, 64, 65, 7, 41, 2, 2, 65, 67, 7, 34, 2,
	2, 66, 56, 3, 2, 2, 2, 66, 62, 3, 2, 2, 2, 67, 5, 3, 2, 2, 2, 68, 73, 7,
	46, 2, 2, 69, 73, 7, 47, 2, 2, 70, 73, 7, 48, 2, 2, 71, 73, 5, 2, 2, 2,
	72, 68, 3, 2, 2, 2, 72, 69, 3, 2, 2, 2, 72, 70, 3, 2, 2, 2, 72, 71, 3,
	2, 2, 2, 73, 7, 3, 2, 2, 2, 74, 75, 8, 5, 1, 2, 75, 78, 5, 6, 4, 2, 76,
	78, 7, 45, 2, 2, 77, 74, 3, 2, 2, 2, 77, 76, 3, 2, 2, 2, 78, 84, 3, 2,
	2, 2, 79, 80, 12, 5, 2, 2, 80, 81, 9, 2, 2, 2, 81, 83, 5, 8, 5, 6, 82,
	79, 3, 2, 2, 2, 83, 86, 3, 2, 2, 2, 84, 82, 3, 2, 2, 2, 84, 85, 3, 2, 2,
	2, 85, 9, 3, 2, 2, 2, 86, 84, 3, 2, 2, 2, 87, 97, 7, 44, 2, 2, 88, 91,
	7, 48, 2, 2, 89, 90, 7, 39, 2, 2, 90, 92, 7, 48, 2, 2, 91, 89, 3, 2, 2,
	2, 91, 92, 3, 2, 2, 2, 92, 93, 3, 2, 2, 2, 93, 97, 7, 31, 2, 2, 94, 97,
	7, 45, 2, 2, 95, 97, 7, 32, 2, 2, 96, 87, 3, 2, 2, 2, 96, 88, 3, 2, 2,
	2, 96, 94, 3, 2, 2, 2, 96, 95, 3, 2, 2, 2, 97, 11, 3, 2, 2, 2, 98, 99,
	7, 35, 2, 2, 99, 103, 7, 4, 2, 2, 100, 101, 5, 18, 10, 2, 101, 102, 7,
	4, 2, 2, 102, 104, 3, 2, 2, 2, 103, 100, 3, 2, 2, 2, 104, 105, 3, 2, 2,
	2, 105, 103, 3, 2, 2, 2, 105, 106, 3, 2, 2, 2, 106, 107, 3, 2, 2, 2, 107,
	108, 7, 36, 2, 2, 108, 13, 3, 2, 2, 2, 109, 110, 7, 13, 2, 2, 110, 111,
	5, 8, 5, 2, 111, 112, 7, 12, 2, 2, 112, 113, 5, 18, 10, 2, 113, 15, 3,
	2, 2, 2, 114, 115, 7, 35, 2, 2, 115, 121, 7, 4, 2, 2, 116, 117, 5, 10,
	6, 2, 117, 118, 7, 12, 2, 2, 118, 119, 5, 18, 10, 2, 119, 120, 7, 4, 2,
	2, 120, 122, 3, 2, 2, 2, 121, 116, 3, 2, 2, 2, 122, 123, 3, 2, 2, 2, 123,
	121, 3, 2, 2, 2, 123, 124, 3, 2, 2, 2, 124, 125, 3, 2, 2, 2, 125, 126,
	7, 36, 2, 2, 126, 17, 3, 2, 2, 2, 127, 132, 5, 8, 5, 2, 128, 132, 5, 14,
	8, 2, 129, 132, 5, 12, 7, 2, 130, 132, 5, 16, 9, 2, 131, 127, 3, 2, 2,
	2, 131, 128, 3, 2, 2, 2, 131, 129, 3, 2, 2, 2, 131, 130, 3, 2, 2, 2, 132,
	19, 3, 2, 2, 2, 133, 134, 7, 35, 2, 2, 134, 140, 7, 4, 2, 2, 135, 136,
	5, 10, 6, 2, 136, 137, 7, 11, 2, 2, 137, 138, 5, 26, 14, 2, 138, 139, 7,
	4, 2, 2, 139, 141, 3, 2, 2, 2, 140, 135, 3, 2, 2, 2, 141, 142, 3, 2, 2,
	2, 142, 140, 3, 2, 2, 2, 142, 143, 3, 2, 2, 2, 143, 144, 3, 2, 2, 2, 144,
	145, 7, 36, 2, 2, 145, 21, 3, 2, 2, 2, 146, 147, 7, 13, 2, 2, 147, 148,
	5, 8, 5, 2, 148, 149, 7, 11, 2, 2, 149, 150, 5, 26, 14, 2, 150, 23, 3,
	2, 2, 2, 151, 152, 7, 35, 2, 2, 152, 156, 7, 4, 2, 2, 153, 154, 5, 26,
	14, 2, 154, 155, 7, 4, 2, 2, 155, 157, 3, 2, 2, 2, 156, 153, 3, 2, 2, 2,
	157, 158, 3, 2, 2, 2, 158, 156, 3, 2, 2, 2, 158, 159, 3, 2, 2, 2, 159,
	160, 3, 2, 2, 2, 160, 161, 7, 36, 2, 2, 161, 25, 3, 2, 2, 2, 162, 166,
	5, 8, 5, 2, 163, 166, 5, 22, 12, 2, 164, 166, 5, 24, 13, 2, 165, 162, 3,
	2, 2, 2, 165, 163, 3, 2, 2, 2, 165, 164, 3, 2, 2, 2, 166, 27, 3, 2, 2,
	2, 167, 170, 5, 26, 14, 2, 168, 170, 5, 20, 11, 2, 169, 167, 3, 2, 2, 2,
	169, 168, 3, 2, 2, 2, 170, 29, 3, 2, 2, 2, 171, 172, 9, 3, 2, 2, 172, 31,
	3, 2, 2, 2, 173, 174, 7, 22, 2, 2, 174, 175, 7, 37, 2, 2, 175, 176, 5,
	8, 5, 2, 176, 177, 7, 3, 2, 2, 177, 178, 7, 43, 2, 2, 178, 179, 7, 38,
	2, 2, 179, 33, 3, 2, 2, 2, 180, 181, 7, 16, 2, 2, 181, 182, 7, 45, 2, 2,
	182, 183, 7, 40, 2, 2, 183, 186, 5, 30, 16, 2, 184, 185, 7, 42, 2, 2, 185,
	187, 5, 32, 17, 2, 186, 184, 3, 2, 2, 2, 186, 187, 3, 2, 2, 2, 187, 35,
	3, 2, 2, 2, 188, 191, 5, 8, 5, 2, 189, 191, 7, 44, 2, 2, 190, 188, 3, 2,
	2, 2, 190, 189, 3, 2, 2, 2, 191, 37, 3, 2, 2, 2, 192, 193, 7, 43, 2, 2,
	193, 194, 7, 42, 2, 2, 194, 195, 5, 36, 19, 2, 195, 39, 3, 2, 2, 2, 196,
	197, 7, 8, 2, 2, 197, 254, 5, 8, 5, 2, 198, 254, 7, 9, 2, 2, 199, 202,
	7, 10, 2, 2, 200, 203, 5, 8, 5, 2, 201, 203, 5, 4, 3, 2, 202, 200, 3, 2,
	2, 2, 202, 201, 3, 2, 2, 2, 203, 204, 3, 2, 2, 2, 204, 205, 7, 37, 2, 2,
	205, 218, 7, 4, 2, 2, 206, 207, 7, 11, 2, 2, 207, 208, 5, 28, 15, 2, 208,
	209, 7, 4, 2, 2, 209, 210, 7, 12, 2, 2, 210, 211, 5, 18, 10, 2, 211, 219,
	3, 2, 2, 2, 212, 213, 7, 12, 2, 2, 213, 214, 5, 18, 10, 2, 214, 215, 7,
	4, 2, 2, 215, 216, 7, 11, 2, 2, 216, 217, 5, 28, 15, 2, 217, 219, 3, 2,
	2, 2, 218, 206, 3, 2, 2, 2, 218, 212, 3, 2, 2, 2, 219, 220, 3, 2, 2, 2,
	220, 221, 7, 4, 2, 2, 221, 222, 7, 38, 2, 2, 222, 254, 3, 2, 2, 2, 223,
	224, 7, 17, 2, 2, 224, 225, 7, 18, 2, 2, 225, 226, 7, 19, 2, 2, 226, 227,
	7, 43, 2, 2, 227, 228, 7, 42, 2, 2, 228, 254, 5, 36, 19, 2, 229, 230, 7,
	17, 2, 2, 230, 231, 7, 18, 2, 2, 231, 232, 7, 19, 2, 2, 232, 233, 7, 35,
	2, 2, 233, 237, 7, 4, 2, 2, 234, 235, 5, 38, 20, 2, 235, 236, 7, 4, 2,
	2, 236, 238, 3, 2, 2, 2, 237, 234, 3, 2, 2, 2, 238, 239, 3, 2, 2, 2, 239,
	237, 3, 2, 2, 2, 239, 240, 3, 2, 2, 2, 240, 241, 3, 2, 2, 2, 241, 242,
	7, 36, 2, 2, 242, 254, 3, 2, 2, 2, 243, 244, 7, 17, 2, 2, 244, 245, 7,
	23, 2, 2, 245, 246, 7, 19, 2, 2, 246, 247, 7, 20, 2, 2, 247, 248, 5, 8,
	5, 2, 248, 249, 7, 21, 2, 2, 249, 250, 7, 43, 2, 2, 250, 251, 7, 42, 2,
	2, 251, 252, 5, 36, 19, 2, 252, 254, 3, 2, 2, 2, 253, 196, 3, 2, 2, 2,
	253, 198, 3, 2, 2, 2, 253, 199, 3, 2, 2, 2, 253, 223, 3, 2, 2, 2, 253,
	229, 3, 2, 2, 2, 253, 243, 3, 2, 2, 2, 254, 41, 3, 2, 2, 2, 255, 257, 7,
	4, 2, 2, 256, 255, 3, 2, 2, 2, 257, 260, 3, 2, 2, 2, 258, 256, 3, 2, 2,
	2, 258, 259, 3, 2, 2, 2, 259, 269, 3, 2, 2, 2, 260, 258, 3, 2, 2, 2, 261,
	263, 5, 34, 18, 2, 262, 264, 7, 4, 2, 2, 263, 262, 3, 2, 2, 2, 264, 265,
	3, 2, 2, 2, 265, 263, 3, 2, 2, 2, 265, 266, 3, 2, 2, 2, 266, 268, 3, 2,
	2, 2, 267, 261, 3, 2, 2, 2, 268, 271, 3, 2, 2, 2, 269, 267, 3, 2, 2, 2,
	269, 270, 3, 2, 2, 2, 270, 272, 3, 2, 2, 2, 271, 269, 3, 2, 2, 2, 272,
	281, 5, 40, 21, 2, 273, 275, 7, 4, 2, 2, 274, 273, 3, 2, 2, 2, 275, 276,
	3, 2, 2, 2, 276, 274, 3, 2, 2, 2, 276, 277, 3, 2, 2, 2, 277, 278, 3, 2,
	2, 2, 278, 280, 5, 40, 21, 2, 279, 274, 3, 2, 2, 2, 280, 283, 3, 2, 2,
	2, 281, 279, 3, 2, 2, 2, 281, 282, 3, 2, 2, 2, 282, 285, 3, 2, 2, 2, 283,
	281, 3, 2, 2, 2, 284, 286, 7, 4, 2, 2, 285, 284, 3, 2, 2, 2, 285, 286,
	3, 2, 2, 2, 286, 287, 3, 2, 2, 2, 287, 288, 7, 2, 2, 3, 288, 43, 3, 2,
	2, 2, 28, 54, 66, 72, 77, 84, 91, 96, 105, 123, 131, 142, 158, 165, 169,
	186, 190, 202, 218, 239, 253, 258, 265, 269, 276, 281, 285,
}
var literalNames = []string{
	"", "','", "", "", "", "", "'print'", "'fail'", "'transfer'", "'from'",
	"'to'", "'max'", "'keep'", "'all'", "'var'", "'set'", "'transaction'",
	"'metadata'", "'of'", "'key'", "'meta'", "'account'", "'asset'", "'number'",
	"'monetary'", "'portion'", "'string'", "'+'", "'-'", "'%'", "'remaining'",
	"'['", "']'", "'{'", "'}'", "'('", "')'", "'.'", "':'", "'*'", "'='",
}
var symbolicNames = []string{
	"", "", "NEWLINE", "WHITESPACE", "MULTILINE_COMMENT", "LINE_COMMENT", "PRINT",
	"FAIL", "TRANSFER", "FROM", "TO", "MAX", "KEEP", "ALL", "VAR", "SET", "TRANSACTION",
	"METADATA", "OF", "KEY", "META", "TY_ACCOUNT", "TY_ASSET", "TY_NUMBER",
	"TY_MONETARY", "TY_PORTION", "TY_STRING", "OP_ADD", "OP_SUB", "PERCENT",
	"PORTION_REMAINING", "LBRACK", "RBRACK", "LBRACE", "RBRACE", "LPAREN",
	"RPAREN", "DOT", "COLON", "STAR", "EQ", "STRING", "RATIO", "VARIABLE_NAME",
	"ACCOUNT", "ASSET", "NUMBER",
}

var ruleNames = []string{
	"monetary", "monetaryAll", "literal", "expression", "allotmentPortion",
	"destinationInOrder", "destinationMaxed", "destinationAllotment", "destination",
	"sourceAllotment", "sourceMaxed", "sourceInOrder", "source", "valueAwareSource",
	"type_", "origin", "varDecl", "metadataValue", "metadataEntry", "statement",
	"script",
}

type FaRlParser struct {
	*antlr.BaseParser
}

// NewFaRlParser produces a new parser instance for the optional input antlr.TokenStream.
//
// The *FaRlParser instance produced may be reused by calling the SetInputStream method.
// The initial parser configuration is expensive to construct, and the object is not thread-safe;
// however, if used within a Golang sync.Pool, the construction cost amortizes well and the
// objects can be used in a thread-safe manner.
func NewFaRlParser(input antlr.TokenStream) *FaRlParser {
	this := new(FaRlParser)
	deserializer := antlr.NewATNDeserializer(nil)
	deserializedATN := deserializer.DeserializeFromUInt16(parserATN)
	decisionToDFA := make([]*antlr.DFA, len(deserializedATN.DecisionToState))
	for index, ds := range deserializedATN.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(ds, index)
	}
	this.BaseParser = antlr.NewBaseParser(input)

	this.Interpreter = antlr.NewParserATNSimulator(this, deserializedATN, decisionToDFA, antlr.NewPredictionContextCache())
	this.RuleNames = ruleNames
	this.LiteralNames = literalNames
	this.SymbolicNames = symbolicNames
	this.GrammarFileName = "FaRl.g4"

	return this
}

// FaRlParser tokens.
const (
	FaRlParserEOF               = antlr.TokenEOF
	FaRlParserT__0              = 1
	FaRlParserNEWLINE           = 2
	FaRlParserWHITESPACE        = 3
	FaRlParserMULTILINE_COMMENT = 4
	FaRlParserLINE_COMMENT      = 5
	FaRlParserPRINT             = 6
	FaRlParserFAIL              = 7
	FaRlParserTRANSFER          = 8
	FaRlParserFROM              = 9
	FaRlParserTO                = 10
	FaRlParserMAX               = 11
	FaRlParserKEEP              = 12
	FaRlParserALL               = 13
	FaRlParserVAR               = 14
	FaRlParserSET               = 15
	FaRlParserTRANSACTION       = 16
	FaRlParserMETADATA          = 17
	FaRlParserOF                = 18
	FaRlParserKEY               = 19
	FaRlParserMETA              = 20
	FaRlParserTY_ACCOUNT        = 21
	FaRlParserTY_ASSET          = 22
	FaRlParserTY_NUMBER         = 23
	FaRlParserTY_MONETARY       = 24
	FaRlParserTY_PORTION        = 25
	FaRlParserTY_STRING         = 26
	FaRlParserOP_ADD            = 27
	FaRlParserOP_SUB            = 28
	FaRlParserPERCENT           = 29
	FaRlParserPORTION_REMAINING = 30
	FaRlParserLBRACK            = 31
	FaRlParserRBRACK            = 32
	FaRlParserLBRACE            = 33
	FaRlParserRBRACE            = 34
	FaRlParserLPAREN            = 35
	FaRlParserRPAREN            = 36
	FaRlParserDOT               = 37
	FaRlParserCOLON             = 38
	FaRlParserSTAR              = 39
	FaRlParserEQ                = 40
	FaRlParserSTRING            = 41
	FaRlParserRATIO             = 42
	FaRlParserVARIABLE_NAME     = 43
	FaRlParserACCOUNT           = 44
	FaRlParserASSET             = 45
	FaRlParserNUMBER            = 46
)

// FaRlParser rules.
const (
	FaRlParserRULE_monetary             = 0
	FaRlParserRULE_monetaryAll          = 1
	FaRlParserRULE_literal              = 2
	FaRlParserRULE_expression           = 3
	FaRlParserRULE_allotmentPortion     = 4
	FaRlParserRULE_destinationInOrder   = 5
	FaRlParserRULE_destinationMaxed     = 6
	FaRlParserRULE_destinationAllotment = 7
	FaRlParserRULE_destination          = 8
	FaRlParserRULE_sourceAllotment      = 9
	FaRlParserRULE_sourceMaxed          = 10
	FaRlParserRULE_sourceInOrder        = 11
	FaRlParserRULE_source               = 12
	FaRlParserRULE_valueAwareSource     = 13
	FaRlParserRULE_type_                = 14
	FaRlParserRULE_origin               = 15
	FaRlParserRULE_varDecl              = 16
	FaRlParserRULE_metadataValue        = 17
	FaRlParserRULE_metadataEntry        = 18
	FaRlParserRULE_statement            = 19
	FaRlParserRULE_script               = 20
)

// IMonetaryContext is an interface to support dynamic dispatch.
type IMonetaryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsMonetaryContext differentiates from other interfaces.
	IsMonetaryContext()
}

type MonetaryContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMonetaryContext() *MonetaryContext {
	var p = new(MonetaryContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_monetary
	return p
}

func (*MonetaryContext) IsMonetaryContext() {}

func NewMonetaryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MonetaryContext {
	var p = new(MonetaryContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_monetary

	return p
}

func (s *MonetaryContext) GetParser() antlr.Parser { return s.parser }

func (s *MonetaryContext) CopyFrom(ctx *MonetaryContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *MonetaryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MonetaryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type MonetaryNoPrecisionContext struct {
	*MonetaryContext
	asset  antlr.Token
	amount antlr.Token
}

func NewMonetaryNoPrecisionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MonetaryNoPrecisionContext {
	var p = new(MonetaryNoPrecisionContext)

	p.MonetaryContext = NewEmptyMonetaryContext()
	p.parser = parser
	p.CopyFrom(ctx.(*MonetaryContext))

	return p
}

func (s *MonetaryNoPrecisionContext) GetAsset() antlr.Token { return s.asset }

func (s *MonetaryNoPrecisionContext) GetAmount() antlr.Token { return s.amount }

func (s *MonetaryNoPrecisionContext) SetAsset(v antlr.Token) { s.asset = v }

func (s *MonetaryNoPrecisionContext) SetAmount(v antlr.Token) { s.amount = v }

func (s *MonetaryNoPrecisionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MonetaryNoPrecisionContext) LBRACK() antlr.TerminalNode {
	return s.GetToken(FaRlParserLBRACK, 0)
}

func (s *MonetaryNoPrecisionContext) RBRACK() antlr.TerminalNode {
	return s.GetToken(FaRlParserRBRACK, 0)
}

func (s *MonetaryNoPrecisionContext) ASSET() antlr.TerminalNode {
	return s.GetToken(FaRlParserASSET, 0)
}

func (s *MonetaryNoPrecisionContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(FaRlParserNUMBER, 0)
}

func (s *MonetaryNoPrecisionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterMonetaryNoPrecision(s)
	}
}

func (s *MonetaryNoPrecisionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitMonetaryNoPrecision(s)
	}
}

type MonetaryLitContext struct {
	*MonetaryContext
	asset     antlr.Token
	precision antlr.Token
	amount    antlr.Token
}

func NewMonetaryLitContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MonetaryLitContext {
	var p = new(MonetaryLitContext)

	p.MonetaryContext = NewEmptyMonetaryContext()
	p.parser = parser
	p.CopyFrom(ctx.(*MonetaryContext))

	return p
}

func (s *MonetaryLitContext) GetAsset() antlr.Token { return s.asset }

func (s *MonetaryLitContext) GetPrecision() antlr.Token { return s.precision }

func (s *MonetaryLitContext) GetAmount() antlr.Token { return s.amount }

func (s *MonetaryLitContext) SetAsset(v antlr.Token) { s.asset = v }

func (s *MonetaryLitContext) SetPrecision(v antlr.Token) { s.precision = v }

func (s *MonetaryLitContext) SetAmount(v antlr.Token) { s.amount = v }

func (s *MonetaryLitContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MonetaryLitContext) LBRACK() antlr.TerminalNode {
	return s.GetToken(FaRlParserLBRACK, 0)
}

func (s *MonetaryLitContext) DOT() antlr.TerminalNode {
	return s.GetToken(FaRlParserDOT, 0)
}

func (s *MonetaryLitContext) RBRACK() antlr.TerminalNode {
	return s.GetToken(FaRlParserRBRACK, 0)
}

func (s *MonetaryLitContext) ASSET() antlr.TerminalNode {
	return s.GetToken(FaRlParserASSET, 0)
}

func (s *MonetaryLitContext) AllNUMBER() []antlr.TerminalNode {
	return s.GetTokens(FaRlParserNUMBER)
}

func (s *MonetaryLitContext) NUMBER(i int) antlr.TerminalNode {
	return s.GetToken(FaRlParserNUMBER, i)
}

func (s *MonetaryLitContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterMonetaryLit(s)
	}
}

func (s *MonetaryLitContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitMonetaryLit(s)
	}
}

func (p *FaRlParser) Monetary() (localctx IMonetaryContext) {
	localctx = NewMonetaryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, FaRlParserRULE_monetary)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(52)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 0, p.GetParserRuleContext()) {
	case 1:
		localctx = NewMonetaryLitContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(42)
			p.Match(FaRlParserLBRACK)
		}
		{
			p.SetState(43)

			var _m = p.Match(FaRlParserASSET)

			localctx.(*MonetaryLitContext).asset = _m
		}
		{
			p.SetState(44)
			p.Match(FaRlParserDOT)
		}
		{
			p.SetState(45)

			var _m = p.Match(FaRlParserNUMBER)

			localctx.(*MonetaryLitContext).precision = _m
		}
		{
			p.SetState(46)

			var _m = p.Match(FaRlParserNUMBER)

			localctx.(*MonetaryLitContext).amount = _m
		}
		{
			p.SetState(47)
			p.Match(FaRlParserRBRACK)
		}

	case 2:
		localctx = NewMonetaryNoPrecisionContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(48)
			p.Match(FaRlParserLBRACK)
		}
		{
			p.SetState(49)

			var _m = p.Match(FaRlParserASSET)

			localctx.(*MonetaryNoPrecisionContext).asset = _m
		}
		{
			p.SetState(50)

			var _m = p.Match(FaRlParserNUMBER)

			localctx.(*MonetaryNoPrecisionContext).amount = _m
		}
		{
			p.SetState(51)
			p.Match(FaRlParserRBRACK)
		}

	}

	return localctx
}

// IMonetaryAllContext is an interface to support dynamic dispatch.
type IMonetaryAllContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsMonetaryAllContext differentiates from other interfaces.
	IsMonetaryAllContext()
}

type MonetaryAllContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMonetaryAllContext() *MonetaryAllContext {
	var p = new(MonetaryAllContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_monetaryAll
	return p
}

func (*MonetaryAllContext) IsMonetaryAllContext() {}

func NewMonetaryAllContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MonetaryAllContext {
	var p = new(MonetaryAllContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_monetaryAll

	return p
}

func (s *MonetaryAllContext) GetParser() antlr.Parser { return s.parser }

func (s *MonetaryAllContext) CopyFrom(ctx *MonetaryAllContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *MonetaryAllContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MonetaryAllContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type MonetaryAllPrecContext struct {
	*MonetaryAllContext
	asset     antlr.Token
	precision antlr.Token
}

func NewMonetaryAllPrecContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MonetaryAllPrecContext {
	var p = new(MonetaryAllPrecContext)

	p.MonetaryAllContext = NewEmptyMonetaryAllContext()
	p.parser = parser
	p.CopyFrom(ctx.(*MonetaryAllContext))

	return p
}

func (s *MonetaryAllPrecContext) GetAsset() antlr.Token { return s.asset }

func (s *MonetaryAllPrecContext) GetPrecision() antlr.Token { return s.precision }

func (s *MonetaryAllPrecContext) SetAsset(v antlr.Token) { s.asset = v }

func (s *MonetaryAllPrecContext) SetPrecision(v antlr.Token) { s.precision = v }

func (s *MonetaryAllPrecContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MonetaryAllPrecContext) LBRACK() antlr.TerminalNode {
	return s.GetToken(FaRlParserLBRACK, 0)
}

func (s *MonetaryAllPrecContext) DOT() antlr.TerminalNode {
	return s.GetToken(FaRlParserDOT, 0)
}

func (s *MonetaryAllPrecContext) STAR() antlr.TerminalNode {
	return s.GetToken(FaRlParserSTAR, 0)
}

func (s *MonetaryAllPrecContext) RBRACK() antlr.TerminalNode {
	return s.GetToken(FaRlParserRBRACK, 0)
}

func (s *MonetaryAllPrecContext) ASSET() antlr.TerminalNode {
	return s.GetToken(FaRlParserASSET, 0)
}

func (s *MonetaryAllPrecContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(FaRlParserNUMBER, 0)
}

func (s *MonetaryAllPrecContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterMonetaryAllPrec(s)
	}
}

func (s *MonetaryAllPrecContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitMonetaryAllPrec(s)
	}
}

type MonetaryAllNoPrecContext struct {
	*MonetaryAllContext
	asset antlr.Token
}

func NewMonetaryAllNoPrecContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MonetaryAllNoPrecContext {
	var p = new(MonetaryAllNoPrecContext)

	p.MonetaryAllContext = NewEmptyMonetaryAllContext()
	p.parser = parser
	p.CopyFrom(ctx.(*MonetaryAllContext))

	return p
}

func (s *MonetaryAllNoPrecContext) GetAsset() antlr.Token { return s.asset }

func (s *MonetaryAllNoPrecContext) SetAsset(v antlr.Token) { s.asset = v }

func (s *MonetaryAllNoPrecContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MonetaryAllNoPrecContext) LBRACK() antlr.TerminalNode {
	return s.GetToken(FaRlParserLBRACK, 0)
}

func (s *MonetaryAllNoPrecContext) STAR() antlr.TerminalNode {
	return s.GetToken(FaRlParserSTAR, 0)
}

func (s *MonetaryAllNoPrecContext) RBRACK() antlr.TerminalNode {
	return s.GetToken(FaRlParserRBRACK, 0)
}

func (s *MonetaryAllNoPrecContext) ASSET() antlr.TerminalNode {
	return s.GetToken(FaRlParserASSET, 0)
}

func (s *MonetaryAllNoPrecContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterMonetaryAllNoPrec(s)
	}
}

func (s *MonetaryAllNoPrecContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitMonetaryAllNoPrec(s)
	}
}

func (p *FaRlParser) MonetaryAll() (localctx IMonetaryAllContext) {
	localctx = NewMonetaryAllContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, FaRlParserRULE_monetaryAll)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(64)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 1, p.GetParserRuleContext()) {
	case 1:
		localctx = NewMonetaryAllPrecContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(54)
			p.Match(FaRlParserLBRACK)
		}
		{
			p.SetState(55)

			var _m = p.Match(FaRlParserASSET)

			localctx.(*MonetaryAllPrecContext).asset = _m
		}
		{
			p.SetState(56)
			p.Match(FaRlParserDOT)
		}
		{
			p.SetState(57)

			var _m = p.Match(FaRlParserNUMBER)

			localctx.(*MonetaryAllPrecContext).precision = _m
		}
		{
			p.SetState(58)
			p.Match(FaRlParserSTAR)
		}
		{
			p.SetState(59)
			p.Match(FaRlParserRBRACK)
		}

	case 2:
		localctx = NewMonetaryAllNoPrecContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(60)
			p.Match(FaRlParserLBRACK)
		}
		{
			p.SetState(61)

			var _m = p.Match(FaRlParserASSET)

			localctx.(*MonetaryAllNoPrecContext).asset = _m
		}
		{
			p.SetState(62)
			p.Match(FaRlParserSTAR)
		}
		{
			p.SetState(63)
			p.Match(FaRlParserRBRACK)
		}

	}

	return localctx
}

// ILiteralContext is an interface to support dynamic dispatch.
type ILiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsLiteralContext differentiates from other interfaces.
	IsLiteralContext()
}

type LiteralContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLiteralContext() *LiteralContext {
	var p = new(LiteralContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_literal
	return p
}

func (*LiteralContext) IsLiteralContext() {}

func NewLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LiteralContext {
	var p = new(LiteralContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_literal

	return p
}

func (s *LiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *LiteralContext) CopyFrom(ctx *LiteralContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *LiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type LitAccountContext struct {
	*LiteralContext
}

func NewLitAccountContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LitAccountContext {
	var p = new(LitAccountContext)

	p.LiteralContext = NewEmptyLiteralContext()
	p.parser = parser
	p.CopyFrom(ctx.(*LiteralContext))

	return p
}

func (s *LitAccountContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LitAccountContext) ACCOUNT() antlr.TerminalNode {
	return s.GetToken(FaRlParserACCOUNT, 0)
}

func (s *LitAccountContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterLitAccount(s)
	}
}

func (s *LitAccountContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitLitAccount(s)
	}
}

type LitAssetContext struct {
	*LiteralContext
}

func NewLitAssetContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LitAssetContext {
	var p = new(LitAssetContext)

	p.LiteralContext = NewEmptyLiteralContext()
	p.parser = parser
	p.CopyFrom(ctx.(*LiteralContext))

	return p
}

func (s *LitAssetContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LitAssetContext) ASSET() antlr.TerminalNode {
	return s.GetToken(FaRlParserASSET, 0)
}

func (s *LitAssetContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterLitAsset(s)
	}
}

func (s *LitAssetContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitLitAsset(s)
	}
}

type LitMonetaryContext struct {
	*LiteralContext
}

func NewLitMonetaryContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LitMonetaryContext {
	var p = new(LitMonetaryContext)

	p.LiteralContext = NewEmptyLiteralContext()
	p.parser = parser
	p.CopyFrom(ctx.(*LiteralContext))

	return p
}

func (s *LitMonetaryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LitMonetaryContext) Monetary() IMonetaryContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IMonetaryContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IMonetaryContext)
}

func (s *LitMonetaryContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterLitMonetary(s)
	}
}

func (s *LitMonetaryContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitLitMonetary(s)
	}
}

type LitNumberContext struct {
	*LiteralContext
}

func NewLitNumberContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LitNumberContext {
	var p = new(LitNumberContext)

	p.LiteralContext = NewEmptyLiteralContext()
	p.parser = parser
	p.CopyFrom(ctx.(*LiteralContext))

	return p
}

func (s *LitNumberContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LitNumberContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(FaRlParserNUMBER, 0)
}

func (s *LitNumberContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterLitNumber(s)
	}
}

func (s *LitNumberContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitLitNumber(s)
	}
}

func (p *FaRlParser) Literal() (localctx ILiteralContext) {
	localctx = NewLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, FaRlParserRULE_literal)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(70)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case FaRlParserACCOUNT:
		localctx = NewLitAccountContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(66)
			p.Match(FaRlParserACCOUNT)
		}

	case FaRlParserASSET:
		localctx = NewLitAssetContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(67)
			p.Match(FaRlParserASSET)
		}

	case FaRlParserNUMBER:
		localctx = NewLitNumberContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(68)
			p.Match(FaRlParserNUMBER)
		}

	case FaRlParserLBRACK:
		localctx = NewLitMonetaryContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(69)
			p.Monetary()
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsExpressionContext differentiates from other interfaces.
	IsExpressionContext()
}

type ExpressionContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_expression
	return p
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) CopyFrom(ctx *ExpressionContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type ExprAddSubContext struct {
	*ExpressionContext
	lhs IExpressionContext
	op  antlr.Token
	rhs IExpressionContext
}

func NewExprAddSubContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ExprAddSubContext {
	var p = new(ExprAddSubContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *ExprAddSubContext) GetOp() antlr.Token { return s.op }

func (s *ExprAddSubContext) SetOp(v antlr.Token) { s.op = v }

func (s *ExprAddSubContext) GetLhs() IExpressionContext { return s.lhs }

func (s *ExprAddSubContext) GetRhs() IExpressionContext { return s.rhs }

func (s *ExprAddSubContext) SetLhs(v IExpressionContext) { s.lhs = v }

func (s *ExprAddSubContext) SetRhs(v IExpressionContext) { s.rhs = v }

func (s *ExprAddSubContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExprAddSubContext) AllExpression() []IExpressionContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IExpressionContext)(nil)).Elem())
	var tst = make([]IExpressionContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IExpressionContext)
		}
	}

	return tst
}

func (s *ExprAddSubContext) Expression(i int) IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ExprAddSubContext) OP_ADD() antlr.TerminalNode {
	return s.GetToken(FaRlParserOP_ADD, 0)
}

func (s *ExprAddSubContext) OP_SUB() antlr.TerminalNode {
	return s.GetToken(FaRlParserOP_SUB, 0)
}

func (s *ExprAddSubContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterExprAddSub(s)
	}
}

func (s *ExprAddSubContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitExprAddSub(s)
	}
}

type ExprLiteralContext struct {
	*ExpressionContext
	lit ILiteralContext
}

func NewExprLiteralContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ExprLiteralContext {
	var p = new(ExprLiteralContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *ExprLiteralContext) GetLit() ILiteralContext { return s.lit }

func (s *ExprLiteralContext) SetLit(v ILiteralContext) { s.lit = v }

func (s *ExprLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExprLiteralContext) Literal() ILiteralContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ILiteralContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ILiteralContext)
}

func (s *ExprLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterExprLiteral(s)
	}
}

func (s *ExprLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitExprLiteral(s)
	}
}

type ExprVariableContext struct {
	*ExpressionContext
	variable antlr.Token
}

func NewExprVariableContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ExprVariableContext {
	var p = new(ExprVariableContext)

	p.ExpressionContext = NewEmptyExpressionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ExpressionContext))

	return p
}

func (s *ExprVariableContext) GetVariable() antlr.Token { return s.variable }

func (s *ExprVariableContext) SetVariable(v antlr.Token) { s.variable = v }

func (s *ExprVariableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExprVariableContext) VARIABLE_NAME() antlr.TerminalNode {
	return s.GetToken(FaRlParserVARIABLE_NAME, 0)
}

func (s *ExprVariableContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterExprVariable(s)
	}
}

func (s *ExprVariableContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitExprVariable(s)
	}
}

func (p *FaRlParser) Expression() (localctx IExpressionContext) {
	return p.expression(0)
}

func (p *FaRlParser) expression(_p int) (localctx IExpressionContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()
	_parentState := p.GetState()
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExpressionContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 6
	p.EnterRecursionRule(localctx, 6, FaRlParserRULE_expression, _p)
	var _la int

	defer func() {
		p.UnrollRecursionContexts(_parentctx)
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(75)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case FaRlParserLBRACK, FaRlParserACCOUNT, FaRlParserASSET, FaRlParserNUMBER:
		localctx = NewExprLiteralContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(73)

			var _x = p.Literal()

			localctx.(*ExprLiteralContext).lit = _x
		}

	case FaRlParserVARIABLE_NAME:
		localctx = NewExprVariableContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(74)

			var _m = p.Match(FaRlParserVARIABLE_NAME)

			localctx.(*ExprVariableContext).variable = _m
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(82)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 4, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			localctx = NewExprAddSubContext(p, NewExpressionContext(p, _parentctx, _parentState))
			localctx.(*ExprAddSubContext).lhs = _prevctx

			p.PushNewRecursionContext(localctx, _startState, FaRlParserRULE_expression)
			p.SetState(77)

			if !(p.Precpred(p.GetParserRuleContext(), 3)) {
				panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
			}
			{
				p.SetState(78)

				var _lt = p.GetTokenStream().LT(1)

				localctx.(*ExprAddSubContext).op = _lt

				_la = p.GetTokenStream().LA(1)

				if !(_la == FaRlParserOP_ADD || _la == FaRlParserOP_SUB) {
					var _ri = p.GetErrorHandler().RecoverInline(p)

					localctx.(*ExprAddSubContext).op = _ri
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
			}
			{
				p.SetState(79)

				var _x = p.expression(4)

				localctx.(*ExprAddSubContext).rhs = _x
			}

		}
		p.SetState(84)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 4, p.GetParserRuleContext())
	}

	return localctx
}

// IAllotmentPortionContext is an interface to support dynamic dispatch.
type IAllotmentPortionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsAllotmentPortionContext differentiates from other interfaces.
	IsAllotmentPortionContext()
}

type AllotmentPortionContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAllotmentPortionContext() *AllotmentPortionContext {
	var p = new(AllotmentPortionContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_allotmentPortion
	return p
}

func (*AllotmentPortionContext) IsAllotmentPortionContext() {}

func NewAllotmentPortionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AllotmentPortionContext {
	var p = new(AllotmentPortionContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_allotmentPortion

	return p
}

func (s *AllotmentPortionContext) GetParser() antlr.Parser { return s.parser }

func (s *AllotmentPortionContext) CopyFrom(ctx *AllotmentPortionContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *AllotmentPortionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AllotmentPortionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type AllotmentPortionRemainingContext struct {
	*AllotmentPortionContext
}

func NewAllotmentPortionRemainingContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AllotmentPortionRemainingContext {
	var p = new(AllotmentPortionRemainingContext)

	p.AllotmentPortionContext = NewEmptyAllotmentPortionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*AllotmentPortionContext))

	return p
}

func (s *AllotmentPortionRemainingContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AllotmentPortionRemainingContext) PORTION_REMAINING() antlr.TerminalNode {
	return s.GetToken(FaRlParserPORTION_REMAINING, 0)
}

func (s *AllotmentPortionRemainingContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterAllotmentPortionRemaining(s)
	}
}

func (s *AllotmentPortionRemainingContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitAllotmentPortionRemaining(s)
	}
}

type AllotmentPortionConstContext struct {
	*AllotmentPortionContext
}

func NewAllotmentPortionConstContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AllotmentPortionConstContext {
	var p = new(AllotmentPortionConstContext)

	p.AllotmentPortionContext = NewEmptyAllotmentPortionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*AllotmentPortionContext))

	return p
}

func (s *AllotmentPortionConstContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AllotmentPortionConstContext) RATIO() antlr.TerminalNode {
	return s.GetToken(FaRlParserRATIO, 0)
}

func (s *AllotmentPortionConstContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterAllotmentPortionConst(s)
	}
}

func (s *AllotmentPortionConstContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitAllotmentPortionConst(s)
	}
}

type AllotmentPortionConstPercentContext struct {
	*AllotmentPortionContext
}

func NewAllotmentPortionConstPercentContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AllotmentPortionConstPercentContext {
	var p = new(AllotmentPortionConstPercentContext)

	p.AllotmentPortionContext = NewEmptyAllotmentPortionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*AllotmentPortionContext))

	return p
}

func (s *AllotmentPortionConstPercentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AllotmentPortionConstPercentContext) AllNUMBER() []antlr.TerminalNode {
	return s.GetTokens(FaRlParserNUMBER)
}

func (s *AllotmentPortionConstPercentContext) NUMBER(i int) antlr.TerminalNode {
	return s.GetToken(FaRlParserNUMBER, i)
}

func (s *AllotmentPortionConstPercentContext) PERCENT() antlr.TerminalNode {
	return s.GetToken(FaRlParserPERCENT, 0)
}

func (s *AllotmentPortionConstPercentContext) DOT() antlr.TerminalNode {
	return s.GetToken(FaRlParserDOT, 0)
}

func (s *AllotmentPortionConstPercentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterAllotmentPortionConstPercent(s)
	}
}

func (s *AllotmentPortionConstPercentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitAllotmentPortionConstPercent(s)
	}
}

type AllotmentPortionVarContext struct {
	*AllotmentPortionContext
	por antlr.Token
}

func NewAllotmentPortionVarContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AllotmentPortionVarContext {
	var p = new(AllotmentPortionVarContext)

	p.AllotmentPortionContext = NewEmptyAllotmentPortionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*AllotmentPortionContext))

	return p
}

func (s *AllotmentPortionVarContext) GetPor() antlr.Token { return s.por }

func (s *AllotmentPortionVarContext) SetPor(v antlr.Token) { s.por = v }

func (s *AllotmentPortionVarContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AllotmentPortionVarContext) VARIABLE_NAME() antlr.TerminalNode {
	return s.GetToken(FaRlParserVARIABLE_NAME, 0)
}

func (s *AllotmentPortionVarContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterAllotmentPortionVar(s)
	}
}

func (s *AllotmentPortionVarContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitAllotmentPortionVar(s)
	}
}

func (p *FaRlParser) AllotmentPortion() (localctx IAllotmentPortionContext) {
	localctx = NewAllotmentPortionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, FaRlParserRULE_allotmentPortion)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(94)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case FaRlParserRATIO:
		localctx = NewAllotmentPortionConstContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(85)
			p.Match(FaRlParserRATIO)
		}

	case FaRlParserNUMBER:
		localctx = NewAllotmentPortionConstPercentContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(86)
			p.Match(FaRlParserNUMBER)
		}
		p.SetState(89)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)

		if _la == FaRlParserDOT {
			{
				p.SetState(87)
				p.Match(FaRlParserDOT)
			}
			{
				p.SetState(88)
				p.Match(FaRlParserNUMBER)
			}

		}
		{
			p.SetState(91)
			p.Match(FaRlParserPERCENT)
		}

	case FaRlParserVARIABLE_NAME:
		localctx = NewAllotmentPortionVarContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(92)

			var _m = p.Match(FaRlParserVARIABLE_NAME)

			localctx.(*AllotmentPortionVarContext).por = _m
		}

	case FaRlParserPORTION_REMAINING:
		localctx = NewAllotmentPortionRemainingContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(93)
			p.Match(FaRlParserPORTION_REMAINING)
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// IDestinationInOrderContext is an interface to support dynamic dispatch.
type IDestinationInOrderContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Get_destination returns the _destination rule contexts.
	Get_destination() IDestinationContext

	// Set_destination sets the _destination rule contexts.
	Set_destination(IDestinationContext)

	// GetDests returns the dests rule context list.
	GetDests() []IDestinationContext

	// SetDests sets the dests rule context list.
	SetDests([]IDestinationContext)

	// IsDestinationInOrderContext differentiates from other interfaces.
	IsDestinationInOrderContext()
}

type DestinationInOrderContext struct {
	*antlr.BaseParserRuleContext
	parser       antlr.Parser
	_destination IDestinationContext
	dests        []IDestinationContext
}

func NewEmptyDestinationInOrderContext() *DestinationInOrderContext {
	var p = new(DestinationInOrderContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_destinationInOrder
	return p
}

func (*DestinationInOrderContext) IsDestinationInOrderContext() {}

func NewDestinationInOrderContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DestinationInOrderContext {
	var p = new(DestinationInOrderContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_destinationInOrder

	return p
}

func (s *DestinationInOrderContext) GetParser() antlr.Parser { return s.parser }

func (s *DestinationInOrderContext) Get_destination() IDestinationContext { return s._destination }

func (s *DestinationInOrderContext) Set_destination(v IDestinationContext) { s._destination = v }

func (s *DestinationInOrderContext) GetDests() []IDestinationContext { return s.dests }

func (s *DestinationInOrderContext) SetDests(v []IDestinationContext) { s.dests = v }

func (s *DestinationInOrderContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(FaRlParserLBRACE, 0)
}

func (s *DestinationInOrderContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(FaRlParserNEWLINE)
}

func (s *DestinationInOrderContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(FaRlParserNEWLINE, i)
}

func (s *DestinationInOrderContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(FaRlParserRBRACE, 0)
}

func (s *DestinationInOrderContext) AllDestination() []IDestinationContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IDestinationContext)(nil)).Elem())
	var tst = make([]IDestinationContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IDestinationContext)
		}
	}

	return tst
}

func (s *DestinationInOrderContext) Destination(i int) IDestinationContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IDestinationContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IDestinationContext)
}

func (s *DestinationInOrderContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestinationInOrderContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DestinationInOrderContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterDestinationInOrder(s)
	}
}

func (s *DestinationInOrderContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitDestinationInOrder(s)
	}
}

func (p *FaRlParser) DestinationInOrder() (localctx IDestinationInOrderContext) {
	localctx = NewDestinationInOrderContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, FaRlParserRULE_destinationInOrder)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(96)
		p.Match(FaRlParserLBRACE)
	}
	{
		p.SetState(97)
		p.Match(FaRlParserNEWLINE)
	}
	p.SetState(101)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == FaRlParserMAX || _la == FaRlParserLBRACK || (((_la-33)&-(0x1f+1)) == 0 && ((1<<uint((_la-33)))&((1<<(FaRlParserLBRACE-33))|(1<<(FaRlParserVARIABLE_NAME-33))|(1<<(FaRlParserACCOUNT-33))|(1<<(FaRlParserASSET-33))|(1<<(FaRlParserNUMBER-33)))) != 0) {
		{
			p.SetState(98)

			var _x = p.Destination()

			localctx.(*DestinationInOrderContext)._destination = _x
		}
		localctx.(*DestinationInOrderContext).dests = append(localctx.(*DestinationInOrderContext).dests, localctx.(*DestinationInOrderContext)._destination)
		{
			p.SetState(99)
			p.Match(FaRlParserNEWLINE)
		}

		p.SetState(103)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(105)
		p.Match(FaRlParserRBRACE)
	}

	return localctx
}

// IDestinationMaxedContext is an interface to support dynamic dispatch.
type IDestinationMaxedContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetMax returns the max rule contexts.
	GetMax() IExpressionContext

	// GetDest returns the dest rule contexts.
	GetDest() IDestinationContext

	// SetMax sets the max rule contexts.
	SetMax(IExpressionContext)

	// SetDest sets the dest rule contexts.
	SetDest(IDestinationContext)

	// IsDestinationMaxedContext differentiates from other interfaces.
	IsDestinationMaxedContext()
}

type DestinationMaxedContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
	max    IExpressionContext
	dest   IDestinationContext
}

func NewEmptyDestinationMaxedContext() *DestinationMaxedContext {
	var p = new(DestinationMaxedContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_destinationMaxed
	return p
}

func (*DestinationMaxedContext) IsDestinationMaxedContext() {}

func NewDestinationMaxedContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DestinationMaxedContext {
	var p = new(DestinationMaxedContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_destinationMaxed

	return p
}

func (s *DestinationMaxedContext) GetParser() antlr.Parser { return s.parser }

func (s *DestinationMaxedContext) GetMax() IExpressionContext { return s.max }

func (s *DestinationMaxedContext) GetDest() IDestinationContext { return s.dest }

func (s *DestinationMaxedContext) SetMax(v IExpressionContext) { s.max = v }

func (s *DestinationMaxedContext) SetDest(v IDestinationContext) { s.dest = v }

func (s *DestinationMaxedContext) MAX() antlr.TerminalNode {
	return s.GetToken(FaRlParserMAX, 0)
}

func (s *DestinationMaxedContext) TO() antlr.TerminalNode {
	return s.GetToken(FaRlParserTO, 0)
}

func (s *DestinationMaxedContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *DestinationMaxedContext) Destination() IDestinationContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IDestinationContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IDestinationContext)
}

func (s *DestinationMaxedContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestinationMaxedContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DestinationMaxedContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterDestinationMaxed(s)
	}
}

func (s *DestinationMaxedContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitDestinationMaxed(s)
	}
}

func (p *FaRlParser) DestinationMaxed() (localctx IDestinationMaxedContext) {
	localctx = NewDestinationMaxedContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, FaRlParserRULE_destinationMaxed)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(107)
		p.Match(FaRlParserMAX)
	}
	{
		p.SetState(108)

		var _x = p.expression(0)

		localctx.(*DestinationMaxedContext).max = _x
	}
	{
		p.SetState(109)
		p.Match(FaRlParserTO)
	}
	{
		p.SetState(110)

		var _x = p.Destination()

		localctx.(*DestinationMaxedContext).dest = _x
	}

	return localctx
}

// IDestinationAllotmentContext is an interface to support dynamic dispatch.
type IDestinationAllotmentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Get_allotmentPortion returns the _allotmentPortion rule contexts.
	Get_allotmentPortion() IAllotmentPortionContext

	// Get_destination returns the _destination rule contexts.
	Get_destination() IDestinationContext

	// Set_allotmentPortion sets the _allotmentPortion rule contexts.
	Set_allotmentPortion(IAllotmentPortionContext)

	// Set_destination sets the _destination rule contexts.
	Set_destination(IDestinationContext)

	// GetPortions returns the portions rule context list.
	GetPortions() []IAllotmentPortionContext

	// GetDests returns the dests rule context list.
	GetDests() []IDestinationContext

	// SetPortions sets the portions rule context list.
	SetPortions([]IAllotmentPortionContext)

	// SetDests sets the dests rule context list.
	SetDests([]IDestinationContext)

	// IsDestinationAllotmentContext differentiates from other interfaces.
	IsDestinationAllotmentContext()
}

type DestinationAllotmentContext struct {
	*antlr.BaseParserRuleContext
	parser            antlr.Parser
	_allotmentPortion IAllotmentPortionContext
	portions          []IAllotmentPortionContext
	_destination      IDestinationContext
	dests             []IDestinationContext
}

func NewEmptyDestinationAllotmentContext() *DestinationAllotmentContext {
	var p = new(DestinationAllotmentContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_destinationAllotment
	return p
}

func (*DestinationAllotmentContext) IsDestinationAllotmentContext() {}

func NewDestinationAllotmentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DestinationAllotmentContext {
	var p = new(DestinationAllotmentContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_destinationAllotment

	return p
}

func (s *DestinationAllotmentContext) GetParser() antlr.Parser { return s.parser }

func (s *DestinationAllotmentContext) Get_allotmentPortion() IAllotmentPortionContext {
	return s._allotmentPortion
}

func (s *DestinationAllotmentContext) Get_destination() IDestinationContext { return s._destination }

func (s *DestinationAllotmentContext) Set_allotmentPortion(v IAllotmentPortionContext) {
	s._allotmentPortion = v
}

func (s *DestinationAllotmentContext) Set_destination(v IDestinationContext) { s._destination = v }

func (s *DestinationAllotmentContext) GetPortions() []IAllotmentPortionContext { return s.portions }

func (s *DestinationAllotmentContext) GetDests() []IDestinationContext { return s.dests }

func (s *DestinationAllotmentContext) SetPortions(v []IAllotmentPortionContext) { s.portions = v }

func (s *DestinationAllotmentContext) SetDests(v []IDestinationContext) { s.dests = v }

func (s *DestinationAllotmentContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(FaRlParserLBRACE, 0)
}

func (s *DestinationAllotmentContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(FaRlParserNEWLINE)
}

func (s *DestinationAllotmentContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(FaRlParserNEWLINE, i)
}

func (s *DestinationAllotmentContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(FaRlParserRBRACE, 0)
}

func (s *DestinationAllotmentContext) AllTO() []antlr.TerminalNode {
	return s.GetTokens(FaRlParserTO)
}

func (s *DestinationAllotmentContext) TO(i int) antlr.TerminalNode {
	return s.GetToken(FaRlParserTO, i)
}

func (s *DestinationAllotmentContext) AllAllotmentPortion() []IAllotmentPortionContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IAllotmentPortionContext)(nil)).Elem())
	var tst = make([]IAllotmentPortionContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IAllotmentPortionContext)
		}
	}

	return tst
}

func (s *DestinationAllotmentContext) AllotmentPortion(i int) IAllotmentPortionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IAllotmentPortionContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IAllotmentPortionContext)
}

func (s *DestinationAllotmentContext) AllDestination() []IDestinationContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IDestinationContext)(nil)).Elem())
	var tst = make([]IDestinationContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IDestinationContext)
		}
	}

	return tst
}

func (s *DestinationAllotmentContext) Destination(i int) IDestinationContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IDestinationContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IDestinationContext)
}

func (s *DestinationAllotmentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestinationAllotmentContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DestinationAllotmentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterDestinationAllotment(s)
	}
}

func (s *DestinationAllotmentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitDestinationAllotment(s)
	}
}

func (p *FaRlParser) DestinationAllotment() (localctx IDestinationAllotmentContext) {
	localctx = NewDestinationAllotmentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, FaRlParserRULE_destinationAllotment)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(112)
		p.Match(FaRlParserLBRACE)
	}
	{
		p.SetState(113)
		p.Match(FaRlParserNEWLINE)
	}
	p.SetState(119)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = (((_la-30)&-(0x1f+1)) == 0 && ((1<<uint((_la-30)))&((1<<(FaRlParserPORTION_REMAINING-30))|(1<<(FaRlParserRATIO-30))|(1<<(FaRlParserVARIABLE_NAME-30))|(1<<(FaRlParserNUMBER-30)))) != 0) {
		{
			p.SetState(114)

			var _x = p.AllotmentPortion()

			localctx.(*DestinationAllotmentContext)._allotmentPortion = _x
		}
		localctx.(*DestinationAllotmentContext).portions = append(localctx.(*DestinationAllotmentContext).portions, localctx.(*DestinationAllotmentContext)._allotmentPortion)
		{
			p.SetState(115)
			p.Match(FaRlParserTO)
		}
		{
			p.SetState(116)

			var _x = p.Destination()

			localctx.(*DestinationAllotmentContext)._destination = _x
		}
		localctx.(*DestinationAllotmentContext).dests = append(localctx.(*DestinationAllotmentContext).dests, localctx.(*DestinationAllotmentContext)._destination)
		{
			p.SetState(117)
			p.Match(FaRlParserNEWLINE)
		}

		p.SetState(121)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(123)
		p.Match(FaRlParserRBRACE)
	}

	return localctx
}

// IDestinationContext is an interface to support dynamic dispatch.
type IDestinationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsDestinationContext differentiates from other interfaces.
	IsDestinationContext()
}

type DestinationContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDestinationContext() *DestinationContext {
	var p = new(DestinationContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_destination
	return p
}

func (*DestinationContext) IsDestinationContext() {}

func NewDestinationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DestinationContext {
	var p = new(DestinationContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_destination

	return p
}

func (s *DestinationContext) GetParser() antlr.Parser { return s.parser }

func (s *DestinationContext) CopyFrom(ctx *DestinationContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *DestinationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestinationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type DestAccountContext struct {
	*DestinationContext
}

func NewDestAccountContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestAccountContext {
	var p = new(DestAccountContext)

	p.DestinationContext = NewEmptyDestinationContext()
	p.parser = parser
	p.CopyFrom(ctx.(*DestinationContext))

	return p
}

func (s *DestAccountContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestAccountContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *DestAccountContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterDestAccount(s)
	}
}

func (s *DestAccountContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitDestAccount(s)
	}
}

type DestAllotmentContext struct {
	*DestinationContext
}

func NewDestAllotmentContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestAllotmentContext {
	var p = new(DestAllotmentContext)

	p.DestinationContext = NewEmptyDestinationContext()
	p.parser = parser
	p.CopyFrom(ctx.(*DestinationContext))

	return p
}

func (s *DestAllotmentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestAllotmentContext) DestinationAllotment() IDestinationAllotmentContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IDestinationAllotmentContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IDestinationAllotmentContext)
}

func (s *DestAllotmentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterDestAllotment(s)
	}
}

func (s *DestAllotmentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitDestAllotment(s)
	}
}

type DestMaxedContext struct {
	*DestinationContext
}

func NewDestMaxedContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestMaxedContext {
	var p = new(DestMaxedContext)

	p.DestinationContext = NewEmptyDestinationContext()
	p.parser = parser
	p.CopyFrom(ctx.(*DestinationContext))

	return p
}

func (s *DestMaxedContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestMaxedContext) DestinationMaxed() IDestinationMaxedContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IDestinationMaxedContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IDestinationMaxedContext)
}

func (s *DestMaxedContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterDestMaxed(s)
	}
}

func (s *DestMaxedContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitDestMaxed(s)
	}
}

type DestInOrderContext struct {
	*DestinationContext
}

func NewDestInOrderContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DestInOrderContext {
	var p = new(DestInOrderContext)

	p.DestinationContext = NewEmptyDestinationContext()
	p.parser = parser
	p.CopyFrom(ctx.(*DestinationContext))

	return p
}

func (s *DestInOrderContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DestInOrderContext) DestinationInOrder() IDestinationInOrderContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IDestinationInOrderContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IDestinationInOrderContext)
}

func (s *DestInOrderContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterDestInOrder(s)
	}
}

func (s *DestInOrderContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitDestInOrder(s)
	}
}

func (p *FaRlParser) Destination() (localctx IDestinationContext) {
	localctx = NewDestinationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, FaRlParserRULE_destination)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(129)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 9, p.GetParserRuleContext()) {
	case 1:
		localctx = NewDestAccountContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(125)
			p.expression(0)
		}

	case 2:
		localctx = NewDestMaxedContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(126)
			p.DestinationMaxed()
		}

	case 3:
		localctx = NewDestInOrderContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(127)
			p.DestinationInOrder()
		}

	case 4:
		localctx = NewDestAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(128)
			p.DestinationAllotment()
		}

	}

	return localctx
}

// ISourceAllotmentContext is an interface to support dynamic dispatch.
type ISourceAllotmentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Get_allotmentPortion returns the _allotmentPortion rule contexts.
	Get_allotmentPortion() IAllotmentPortionContext

	// Get_source returns the _source rule contexts.
	Get_source() ISourceContext

	// Set_allotmentPortion sets the _allotmentPortion rule contexts.
	Set_allotmentPortion(IAllotmentPortionContext)

	// Set_source sets the _source rule contexts.
	Set_source(ISourceContext)

	// GetPortions returns the portions rule context list.
	GetPortions() []IAllotmentPortionContext

	// GetSources returns the sources rule context list.
	GetSources() []ISourceContext

	// SetPortions sets the portions rule context list.
	SetPortions([]IAllotmentPortionContext)

	// SetSources sets the sources rule context list.
	SetSources([]ISourceContext)

	// IsSourceAllotmentContext differentiates from other interfaces.
	IsSourceAllotmentContext()
}

type SourceAllotmentContext struct {
	*antlr.BaseParserRuleContext
	parser            antlr.Parser
	_allotmentPortion IAllotmentPortionContext
	portions          []IAllotmentPortionContext
	_source           ISourceContext
	sources           []ISourceContext
}

func NewEmptySourceAllotmentContext() *SourceAllotmentContext {
	var p = new(SourceAllotmentContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_sourceAllotment
	return p
}

func (*SourceAllotmentContext) IsSourceAllotmentContext() {}

func NewSourceAllotmentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SourceAllotmentContext {
	var p = new(SourceAllotmentContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_sourceAllotment

	return p
}

func (s *SourceAllotmentContext) GetParser() antlr.Parser { return s.parser }

func (s *SourceAllotmentContext) Get_allotmentPortion() IAllotmentPortionContext {
	return s._allotmentPortion
}

func (s *SourceAllotmentContext) Get_source() ISourceContext { return s._source }

func (s *SourceAllotmentContext) Set_allotmentPortion(v IAllotmentPortionContext) {
	s._allotmentPortion = v
}

func (s *SourceAllotmentContext) Set_source(v ISourceContext) { s._source = v }

func (s *SourceAllotmentContext) GetPortions() []IAllotmentPortionContext { return s.portions }

func (s *SourceAllotmentContext) GetSources() []ISourceContext { return s.sources }

func (s *SourceAllotmentContext) SetPortions(v []IAllotmentPortionContext) { s.portions = v }

func (s *SourceAllotmentContext) SetSources(v []ISourceContext) { s.sources = v }

func (s *SourceAllotmentContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(FaRlParserLBRACE, 0)
}

func (s *SourceAllotmentContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(FaRlParserNEWLINE)
}

func (s *SourceAllotmentContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(FaRlParserNEWLINE, i)
}

func (s *SourceAllotmentContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(FaRlParserRBRACE, 0)
}

func (s *SourceAllotmentContext) AllFROM() []antlr.TerminalNode {
	return s.GetTokens(FaRlParserFROM)
}

func (s *SourceAllotmentContext) FROM(i int) antlr.TerminalNode {
	return s.GetToken(FaRlParserFROM, i)
}

func (s *SourceAllotmentContext) AllAllotmentPortion() []IAllotmentPortionContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IAllotmentPortionContext)(nil)).Elem())
	var tst = make([]IAllotmentPortionContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IAllotmentPortionContext)
		}
	}

	return tst
}

func (s *SourceAllotmentContext) AllotmentPortion(i int) IAllotmentPortionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IAllotmentPortionContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IAllotmentPortionContext)
}

func (s *SourceAllotmentContext) AllSource() []ISourceContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ISourceContext)(nil)).Elem())
	var tst = make([]ISourceContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ISourceContext)
		}
	}

	return tst
}

func (s *SourceAllotmentContext) Source(i int) ISourceContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISourceContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ISourceContext)
}

func (s *SourceAllotmentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SourceAllotmentContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SourceAllotmentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterSourceAllotment(s)
	}
}

func (s *SourceAllotmentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitSourceAllotment(s)
	}
}

func (p *FaRlParser) SourceAllotment() (localctx ISourceAllotmentContext) {
	localctx = NewSourceAllotmentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, FaRlParserRULE_sourceAllotment)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(131)
		p.Match(FaRlParserLBRACE)
	}
	{
		p.SetState(132)
		p.Match(FaRlParserNEWLINE)
	}
	p.SetState(138)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = (((_la-30)&-(0x1f+1)) == 0 && ((1<<uint((_la-30)))&((1<<(FaRlParserPORTION_REMAINING-30))|(1<<(FaRlParserRATIO-30))|(1<<(FaRlParserVARIABLE_NAME-30))|(1<<(FaRlParserNUMBER-30)))) != 0) {
		{
			p.SetState(133)

			var _x = p.AllotmentPortion()

			localctx.(*SourceAllotmentContext)._allotmentPortion = _x
		}
		localctx.(*SourceAllotmentContext).portions = append(localctx.(*SourceAllotmentContext).portions, localctx.(*SourceAllotmentContext)._allotmentPortion)
		{
			p.SetState(134)
			p.Match(FaRlParserFROM)
		}
		{
			p.SetState(135)

			var _x = p.Source()

			localctx.(*SourceAllotmentContext)._source = _x
		}
		localctx.(*SourceAllotmentContext).sources = append(localctx.(*SourceAllotmentContext).sources, localctx.(*SourceAllotmentContext)._source)
		{
			p.SetState(136)
			p.Match(FaRlParserNEWLINE)
		}

		p.SetState(140)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(142)
		p.Match(FaRlParserRBRACE)
	}

	return localctx
}

// ISourceMaxedContext is an interface to support dynamic dispatch.
type ISourceMaxedContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetMax returns the max rule contexts.
	GetMax() IExpressionContext

	// GetSrc returns the src rule contexts.
	GetSrc() ISourceContext

	// SetMax sets the max rule contexts.
	SetMax(IExpressionContext)

	// SetSrc sets the src rule contexts.
	SetSrc(ISourceContext)

	// IsSourceMaxedContext differentiates from other interfaces.
	IsSourceMaxedContext()
}

type SourceMaxedContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
	max    IExpressionContext
	src    ISourceContext
}

func NewEmptySourceMaxedContext() *SourceMaxedContext {
	var p = new(SourceMaxedContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_sourceMaxed
	return p
}

func (*SourceMaxedContext) IsSourceMaxedContext() {}

func NewSourceMaxedContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SourceMaxedContext {
	var p = new(SourceMaxedContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_sourceMaxed

	return p
}

func (s *SourceMaxedContext) GetParser() antlr.Parser { return s.parser }

func (s *SourceMaxedContext) GetMax() IExpressionContext { return s.max }

func (s *SourceMaxedContext) GetSrc() ISourceContext { return s.src }

func (s *SourceMaxedContext) SetMax(v IExpressionContext) { s.max = v }

func (s *SourceMaxedContext) SetSrc(v ISourceContext) { s.src = v }

func (s *SourceMaxedContext) MAX() antlr.TerminalNode {
	return s.GetToken(FaRlParserMAX, 0)
}

func (s *SourceMaxedContext) FROM() antlr.TerminalNode {
	return s.GetToken(FaRlParserFROM, 0)
}

func (s *SourceMaxedContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *SourceMaxedContext) Source() ISourceContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISourceContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISourceContext)
}

func (s *SourceMaxedContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SourceMaxedContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SourceMaxedContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterSourceMaxed(s)
	}
}

func (s *SourceMaxedContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitSourceMaxed(s)
	}
}

func (p *FaRlParser) SourceMaxed() (localctx ISourceMaxedContext) {
	localctx = NewSourceMaxedContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, FaRlParserRULE_sourceMaxed)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(144)
		p.Match(FaRlParserMAX)
	}
	{
		p.SetState(145)

		var _x = p.expression(0)

		localctx.(*SourceMaxedContext).max = _x
	}
	{
		p.SetState(146)
		p.Match(FaRlParserFROM)
	}
	{
		p.SetState(147)

		var _x = p.Source()

		localctx.(*SourceMaxedContext).src = _x
	}

	return localctx
}

// ISourceInOrderContext is an interface to support dynamic dispatch.
type ISourceInOrderContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Get_source returns the _source rule contexts.
	Get_source() ISourceContext

	// Set_source sets the _source rule contexts.
	Set_source(ISourceContext)

	// GetSources returns the sources rule context list.
	GetSources() []ISourceContext

	// SetSources sets the sources rule context list.
	SetSources([]ISourceContext)

	// IsSourceInOrderContext differentiates from other interfaces.
	IsSourceInOrderContext()
}

type SourceInOrderContext struct {
	*antlr.BaseParserRuleContext
	parser  antlr.Parser
	_source ISourceContext
	sources []ISourceContext
}

func NewEmptySourceInOrderContext() *SourceInOrderContext {
	var p = new(SourceInOrderContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_sourceInOrder
	return p
}

func (*SourceInOrderContext) IsSourceInOrderContext() {}

func NewSourceInOrderContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SourceInOrderContext {
	var p = new(SourceInOrderContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_sourceInOrder

	return p
}

func (s *SourceInOrderContext) GetParser() antlr.Parser { return s.parser }

func (s *SourceInOrderContext) Get_source() ISourceContext { return s._source }

func (s *SourceInOrderContext) Set_source(v ISourceContext) { s._source = v }

func (s *SourceInOrderContext) GetSources() []ISourceContext { return s.sources }

func (s *SourceInOrderContext) SetSources(v []ISourceContext) { s.sources = v }

func (s *SourceInOrderContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(FaRlParserLBRACE, 0)
}

func (s *SourceInOrderContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(FaRlParserNEWLINE)
}

func (s *SourceInOrderContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(FaRlParserNEWLINE, i)
}

func (s *SourceInOrderContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(FaRlParserRBRACE, 0)
}

func (s *SourceInOrderContext) AllSource() []ISourceContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ISourceContext)(nil)).Elem())
	var tst = make([]ISourceContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ISourceContext)
		}
	}

	return tst
}

func (s *SourceInOrderContext) Source(i int) ISourceContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISourceContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ISourceContext)
}

func (s *SourceInOrderContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SourceInOrderContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SourceInOrderContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterSourceInOrder(s)
	}
}

func (s *SourceInOrderContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitSourceInOrder(s)
	}
}

func (p *FaRlParser) SourceInOrder() (localctx ISourceInOrderContext) {
	localctx = NewSourceInOrderContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, FaRlParserRULE_sourceInOrder)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(149)
		p.Match(FaRlParserLBRACE)
	}
	{
		p.SetState(150)
		p.Match(FaRlParserNEWLINE)
	}
	p.SetState(154)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == FaRlParserMAX || _la == FaRlParserLBRACK || (((_la-33)&-(0x1f+1)) == 0 && ((1<<uint((_la-33)))&((1<<(FaRlParserLBRACE-33))|(1<<(FaRlParserVARIABLE_NAME-33))|(1<<(FaRlParserACCOUNT-33))|(1<<(FaRlParserASSET-33))|(1<<(FaRlParserNUMBER-33)))) != 0) {
		{
			p.SetState(151)

			var _x = p.Source()

			localctx.(*SourceInOrderContext)._source = _x
		}
		localctx.(*SourceInOrderContext).sources = append(localctx.(*SourceInOrderContext).sources, localctx.(*SourceInOrderContext)._source)
		{
			p.SetState(152)
			p.Match(FaRlParserNEWLINE)
		}

		p.SetState(156)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(158)
		p.Match(FaRlParserRBRACE)
	}

	return localctx
}

// ISourceContext is an interface to support dynamic dispatch.
type ISourceContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsSourceContext differentiates from other interfaces.
	IsSourceContext()
}

type SourceContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySourceContext() *SourceContext {
	var p = new(SourceContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_source
	return p
}

func (*SourceContext) IsSourceContext() {}

func NewSourceContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SourceContext {
	var p = new(SourceContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_source

	return p
}

func (s *SourceContext) GetParser() antlr.Parser { return s.parser }

func (s *SourceContext) CopyFrom(ctx *SourceContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *SourceContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SourceContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type SrcAccountContext struct {
	*SourceContext
}

func NewSrcAccountContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcAccountContext {
	var p = new(SrcAccountContext)

	p.SourceContext = NewEmptySourceContext()
	p.parser = parser
	p.CopyFrom(ctx.(*SourceContext))

	return p
}

func (s *SrcAccountContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcAccountContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *SrcAccountContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterSrcAccount(s)
	}
}

func (s *SrcAccountContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitSrcAccount(s)
	}
}

type SrcMaxedContext struct {
	*SourceContext
}

func NewSrcMaxedContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcMaxedContext {
	var p = new(SrcMaxedContext)

	p.SourceContext = NewEmptySourceContext()
	p.parser = parser
	p.CopyFrom(ctx.(*SourceContext))

	return p
}

func (s *SrcMaxedContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcMaxedContext) SourceMaxed() ISourceMaxedContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISourceMaxedContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISourceMaxedContext)
}

func (s *SrcMaxedContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterSrcMaxed(s)
	}
}

func (s *SrcMaxedContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitSrcMaxed(s)
	}
}

type SrcInOrderContext struct {
	*SourceContext
}

func NewSrcInOrderContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcInOrderContext {
	var p = new(SrcInOrderContext)

	p.SourceContext = NewEmptySourceContext()
	p.parser = parser
	p.CopyFrom(ctx.(*SourceContext))

	return p
}

func (s *SrcInOrderContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcInOrderContext) SourceInOrder() ISourceInOrderContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISourceInOrderContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISourceInOrderContext)
}

func (s *SrcInOrderContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterSrcInOrder(s)
	}
}

func (s *SrcInOrderContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitSrcInOrder(s)
	}
}

func (p *FaRlParser) Source() (localctx ISourceContext) {
	localctx = NewSourceContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, FaRlParserRULE_source)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(163)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case FaRlParserLBRACK, FaRlParserVARIABLE_NAME, FaRlParserACCOUNT, FaRlParserASSET, FaRlParserNUMBER:
		localctx = NewSrcAccountContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(160)
			p.expression(0)
		}

	case FaRlParserMAX:
		localctx = NewSrcMaxedContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(161)
			p.SourceMaxed()
		}

	case FaRlParserLBRACE:
		localctx = NewSrcInOrderContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(162)
			p.SourceInOrder()
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// IValueAwareSourceContext is an interface to support dynamic dispatch.
type IValueAwareSourceContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsValueAwareSourceContext differentiates from other interfaces.
	IsValueAwareSourceContext()
}

type ValueAwareSourceContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValueAwareSourceContext() *ValueAwareSourceContext {
	var p = new(ValueAwareSourceContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_valueAwareSource
	return p
}

func (*ValueAwareSourceContext) IsValueAwareSourceContext() {}

func NewValueAwareSourceContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ValueAwareSourceContext {
	var p = new(ValueAwareSourceContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_valueAwareSource

	return p
}

func (s *ValueAwareSourceContext) GetParser() antlr.Parser { return s.parser }

func (s *ValueAwareSourceContext) CopyFrom(ctx *ValueAwareSourceContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *ValueAwareSourceContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValueAwareSourceContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type SrcContext struct {
	*ValueAwareSourceContext
}

func NewSrcContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcContext {
	var p = new(SrcContext)

	p.ValueAwareSourceContext = NewEmptyValueAwareSourceContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ValueAwareSourceContext))

	return p
}

func (s *SrcContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcContext) Source() ISourceContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISourceContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISourceContext)
}

func (s *SrcContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterSrc(s)
	}
}

func (s *SrcContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitSrc(s)
	}
}

type SrcAllotmentContext struct {
	*ValueAwareSourceContext
}

func NewSrcAllotmentContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SrcAllotmentContext {
	var p = new(SrcAllotmentContext)

	p.ValueAwareSourceContext = NewEmptyValueAwareSourceContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ValueAwareSourceContext))

	return p
}

func (s *SrcAllotmentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SrcAllotmentContext) SourceAllotment() ISourceAllotmentContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISourceAllotmentContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISourceAllotmentContext)
}

func (s *SrcAllotmentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterSrcAllotment(s)
	}
}

func (s *SrcAllotmentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitSrcAllotment(s)
	}
}

func (p *FaRlParser) ValueAwareSource() (localctx IValueAwareSourceContext) {
	localctx = NewValueAwareSourceContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, FaRlParserRULE_valueAwareSource)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(167)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 13, p.GetParserRuleContext()) {
	case 1:
		localctx = NewSrcContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(165)
			p.Source()
		}

	case 2:
		localctx = NewSrcAllotmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(166)
			p.SourceAllotment()
		}

	}

	return localctx
}

// IType_Context is an interface to support dynamic dispatch.
type IType_Context interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsType_Context differentiates from other interfaces.
	IsType_Context()
}

type Type_Context struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyType_Context() *Type_Context {
	var p = new(Type_Context)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_type_
	return p
}

func (*Type_Context) IsType_Context() {}

func NewType_Context(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Type_Context {
	var p = new(Type_Context)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_type_

	return p
}

func (s *Type_Context) GetParser() antlr.Parser { return s.parser }

func (s *Type_Context) TY_ACCOUNT() antlr.TerminalNode {
	return s.GetToken(FaRlParserTY_ACCOUNT, 0)
}

func (s *Type_Context) TY_ASSET() antlr.TerminalNode {
	return s.GetToken(FaRlParserTY_ASSET, 0)
}

func (s *Type_Context) TY_NUMBER() antlr.TerminalNode {
	return s.GetToken(FaRlParserTY_NUMBER, 0)
}

func (s *Type_Context) TY_MONETARY() antlr.TerminalNode {
	return s.GetToken(FaRlParserTY_MONETARY, 0)
}

func (s *Type_Context) TY_PORTION() antlr.TerminalNode {
	return s.GetToken(FaRlParserTY_PORTION, 0)
}

func (s *Type_Context) TY_STRING() antlr.TerminalNode {
	return s.GetToken(FaRlParserTY_STRING, 0)
}

func (s *Type_Context) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Type_Context) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Type_Context) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterType_(s)
	}
}

func (s *Type_Context) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitType_(s)
	}
}

func (p *FaRlParser) Type_() (localctx IType_Context) {
	localctx = NewType_Context(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, FaRlParserRULE_type_)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(169)
		_la = p.GetTokenStream().LA(1)

		if !(((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<FaRlParserTY_ACCOUNT)|(1<<FaRlParserTY_ASSET)|(1<<FaRlParserTY_NUMBER)|(1<<FaRlParserTY_MONETARY)|(1<<FaRlParserTY_PORTION)|(1<<FaRlParserTY_STRING))) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

	return localctx
}

// IOriginContext is an interface to support dynamic dispatch.
type IOriginContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetKey returns the key token.
	GetKey() antlr.Token

	// SetKey sets the key token.
	SetKey(antlr.Token)

	// GetAcc returns the acc rule contexts.
	GetAcc() IExpressionContext

	// SetAcc sets the acc rule contexts.
	SetAcc(IExpressionContext)

	// IsOriginContext differentiates from other interfaces.
	IsOriginContext()
}

type OriginContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
	acc    IExpressionContext
	key    antlr.Token
}

func NewEmptyOriginContext() *OriginContext {
	var p = new(OriginContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_origin
	return p
}

func (*OriginContext) IsOriginContext() {}

func NewOriginContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OriginContext {
	var p = new(OriginContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_origin

	return p
}

func (s *OriginContext) GetParser() antlr.Parser { return s.parser }

func (s *OriginContext) GetKey() antlr.Token { return s.key }

func (s *OriginContext) SetKey(v antlr.Token) { s.key = v }

func (s *OriginContext) GetAcc() IExpressionContext { return s.acc }

func (s *OriginContext) SetAcc(v IExpressionContext) { s.acc = v }

func (s *OriginContext) META() antlr.TerminalNode {
	return s.GetToken(FaRlParserMETA, 0)
}

func (s *OriginContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(FaRlParserLPAREN, 0)
}

func (s *OriginContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(FaRlParserRPAREN, 0)
}

func (s *OriginContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *OriginContext) STRING() antlr.TerminalNode {
	return s.GetToken(FaRlParserSTRING, 0)
}

func (s *OriginContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OriginContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OriginContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterOrigin(s)
	}
}

func (s *OriginContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitOrigin(s)
	}
}

func (p *FaRlParser) Origin() (localctx IOriginContext) {
	localctx = NewOriginContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, FaRlParserRULE_origin)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(171)
		p.Match(FaRlParserMETA)
	}
	{
		p.SetState(172)
		p.Match(FaRlParserLPAREN)
	}
	{
		p.SetState(173)

		var _x = p.expression(0)

		localctx.(*OriginContext).acc = _x
	}
	{
		p.SetState(174)
		p.Match(FaRlParserT__0)
	}
	{
		p.SetState(175)

		var _m = p.Match(FaRlParserSTRING)

		localctx.(*OriginContext).key = _m
	}
	{
		p.SetState(176)
		p.Match(FaRlParserRPAREN)
	}

	return localctx
}

// IVarDeclContext is an interface to support dynamic dispatch.
type IVarDeclContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsVarDeclContext differentiates from other interfaces.
	IsVarDeclContext()
}

type VarDeclContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyVarDeclContext() *VarDeclContext {
	var p = new(VarDeclContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_varDecl
	return p
}

func (*VarDeclContext) IsVarDeclContext() {}

func NewVarDeclContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *VarDeclContext {
	var p = new(VarDeclContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_varDecl

	return p
}

func (s *VarDeclContext) GetParser() antlr.Parser { return s.parser }

func (s *VarDeclContext) CopyFrom(ctx *VarDeclContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *VarDeclContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VarDeclContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type VarTypedContext struct {
	*VarDeclContext
	name antlr.Token
	ty   IType_Context
	orig IOriginContext
}

func NewVarTypedContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VarTypedContext {
	var p = new(VarTypedContext)

	p.VarDeclContext = NewEmptyVarDeclContext()
	p.parser = parser
	p.CopyFrom(ctx.(*VarDeclContext))

	return p
}

func (s *VarTypedContext) GetName() antlr.Token { return s.name }

func (s *VarTypedContext) SetName(v antlr.Token) { s.name = v }

func (s *VarTypedContext) GetTy() IType_Context { return s.ty }

func (s *VarTypedContext) GetOrig() IOriginContext { return s.orig }

func (s *VarTypedContext) SetTy(v IType_Context) { s.ty = v }

func (s *VarTypedContext) SetOrig(v IOriginContext) { s.orig = v }

func (s *VarTypedContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VarTypedContext) VAR() antlr.TerminalNode {
	return s.GetToken(FaRlParserVAR, 0)
}

func (s *VarTypedContext) COLON() antlr.TerminalNode {
	return s.GetToken(FaRlParserCOLON, 0)
}

func (s *VarTypedContext) VARIABLE_NAME() antlr.TerminalNode {
	return s.GetToken(FaRlParserVARIABLE_NAME, 0)
}

func (s *VarTypedContext) Type_() IType_Context {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IType_Context)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IType_Context)
}

func (s *VarTypedContext) EQ() antlr.TerminalNode {
	return s.GetToken(FaRlParserEQ, 0)
}

func (s *VarTypedContext) Origin() IOriginContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IOriginContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IOriginContext)
}

func (s *VarTypedContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterVarTyped(s)
	}
}

func (s *VarTypedContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitVarTyped(s)
	}
}

func (p *FaRlParser) VarDecl() (localctx IVarDeclContext) {
	localctx = NewVarDeclContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 32, FaRlParserRULE_varDecl)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	localctx = NewVarTypedContext(p, localctx)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(178)
		p.Match(FaRlParserVAR)
	}
	{
		p.SetState(179)

		var _m = p.Match(FaRlParserVARIABLE_NAME)

		localctx.(*VarTypedContext).name = _m
	}
	{
		p.SetState(180)
		p.Match(FaRlParserCOLON)
	}
	{
		p.SetState(181)

		var _x = p.Type_()

		localctx.(*VarTypedContext).ty = _x
	}
	p.SetState(184)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == FaRlParserEQ {
		{
			p.SetState(182)
			p.Match(FaRlParserEQ)
		}
		{
			p.SetState(183)

			var _x = p.Origin()

			localctx.(*VarTypedContext).orig = _x
		}

	}

	return localctx
}

// IMetadataValueContext is an interface to support dynamic dispatch.
type IMetadataValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsMetadataValueContext differentiates from other interfaces.
	IsMetadataValueContext()
}

type MetadataValueContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMetadataValueContext() *MetadataValueContext {
	var p = new(MetadataValueContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_metadataValue
	return p
}

func (*MetadataValueContext) IsMetadataValueContext() {}

func NewMetadataValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MetadataValueContext {
	var p = new(MetadataValueContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_metadataValue

	return p
}

func (s *MetadataValueContext) GetParser() antlr.Parser { return s.parser }

func (s *MetadataValueContext) CopyFrom(ctx *MetadataValueContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *MetadataValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MetadataValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type MetaValueRatioContext struct {
	*MetadataValueContext
}

func NewMetaValueRatioContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MetaValueRatioContext {
	var p = new(MetaValueRatioContext)

	p.MetadataValueContext = NewEmptyMetadataValueContext()
	p.parser = parser
	p.CopyFrom(ctx.(*MetadataValueContext))

	return p
}

func (s *MetaValueRatioContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MetaValueRatioContext) RATIO() antlr.TerminalNode {
	return s.GetToken(FaRlParserRATIO, 0)
}

func (s *MetaValueRatioContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterMetaValueRatio(s)
	}
}

func (s *MetaValueRatioContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitMetaValueRatio(s)
	}
}

type MetaValueExprContext struct {
	*MetadataValueContext
}

func NewMetaValueExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MetaValueExprContext {
	var p = new(MetaValueExprContext)

	p.MetadataValueContext = NewEmptyMetadataValueContext()
	p.parser = parser
	p.CopyFrom(ctx.(*MetadataValueContext))

	return p
}

func (s *MetaValueExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MetaValueExprContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *MetaValueExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterMetaValueExpr(s)
	}
}

func (s *MetaValueExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitMetaValueExpr(s)
	}
}

func (p *FaRlParser) MetadataValue() (localctx IMetadataValueContext) {
	localctx = NewMetadataValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, FaRlParserRULE_metadataValue)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(188)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case FaRlParserLBRACK, FaRlParserVARIABLE_NAME, FaRlParserACCOUNT, FaRlParserASSET, FaRlParserNUMBER:
		localctx = NewMetaValueExprContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(186)
			p.expression(0)
		}

	case FaRlParserRATIO:
		localctx = NewMetaValueRatioContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(187)
			p.Match(FaRlParserRATIO)
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// IMetadataEntryContext is an interface to support dynamic dispatch.
type IMetadataEntryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsMetadataEntryContext differentiates from other interfaces.
	IsMetadataEntryContext()
}

type MetadataEntryContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMetadataEntryContext() *MetadataEntryContext {
	var p = new(MetadataEntryContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_metadataEntry
	return p
}

func (*MetadataEntryContext) IsMetadataEntryContext() {}

func NewMetadataEntryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MetadataEntryContext {
	var p = new(MetadataEntryContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_metadataEntry

	return p
}

func (s *MetadataEntryContext) GetParser() antlr.Parser { return s.parser }

func (s *MetadataEntryContext) STRING() antlr.TerminalNode {
	return s.GetToken(FaRlParserSTRING, 0)
}

func (s *MetadataEntryContext) EQ() antlr.TerminalNode {
	return s.GetToken(FaRlParserEQ, 0)
}

func (s *MetadataEntryContext) MetadataValue() IMetadataValueContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IMetadataValueContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IMetadataValueContext)
}

func (s *MetadataEntryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MetadataEntryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MetadataEntryContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterMetadataEntry(s)
	}
}

func (s *MetadataEntryContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitMetadataEntry(s)
	}
}

func (p *FaRlParser) MetadataEntry() (localctx IMetadataEntryContext) {
	localctx = NewMetadataEntryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 36, FaRlParserRULE_metadataEntry)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(190)
		p.Match(FaRlParserSTRING)
	}
	{
		p.SetState(191)
		p.Match(FaRlParserEQ)
	}
	{
		p.SetState(192)
		p.MetadataValue()
	}

	return localctx
}

// IStatementContext is an interface to support dynamic dispatch.
type IStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsStatementContext differentiates from other interfaces.
	IsStatementContext()
}

type StatementContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStatementContext() *StatementContext {
	var p = new(StatementContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_statement
	return p
}

func (*StatementContext) IsStatementContext() {}

func NewStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StatementContext {
	var p = new(StatementContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_statement

	return p
}

func (s *StatementContext) GetParser() antlr.Parser { return s.parser }

func (s *StatementContext) CopyFrom(ctx *StatementContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *StatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type PrintContext struct {
	*StatementContext
	expr IExpressionContext
}

func NewPrintContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PrintContext {
	var p = new(PrintContext)

	p.StatementContext = NewEmptyStatementContext()
	p.parser = parser
	p.CopyFrom(ctx.(*StatementContext))

	return p
}

func (s *PrintContext) GetExpr() IExpressionContext { return s.expr }

func (s *PrintContext) SetExpr(v IExpressionContext) { s.expr = v }

func (s *PrintContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PrintContext) PRINT() antlr.TerminalNode {
	return s.GetToken(FaRlParserPRINT, 0)
}

func (s *PrintContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *PrintContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterPrint(s)
	}
}

func (s *PrintContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitPrint(s)
	}
}

type TransferContext struct {
	*StatementContext
	mon    IExpressionContext
	monAll IMonetaryAllContext
	src    IValueAwareSourceContext
	dest   IDestinationContext
}

func NewTransferContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *TransferContext {
	var p = new(TransferContext)

	p.StatementContext = NewEmptyStatementContext()
	p.parser = parser
	p.CopyFrom(ctx.(*StatementContext))

	return p
}

func (s *TransferContext) GetMon() IExpressionContext { return s.mon }

func (s *TransferContext) GetMonAll() IMonetaryAllContext { return s.monAll }

func (s *TransferContext) GetSrc() IValueAwareSourceContext { return s.src }

func (s *TransferContext) GetDest() IDestinationContext { return s.dest }

func (s *TransferContext) SetMon(v IExpressionContext) { s.mon = v }

func (s *TransferContext) SetMonAll(v IMonetaryAllContext) { s.monAll = v }

func (s *TransferContext) SetSrc(v IValueAwareSourceContext) { s.src = v }

func (s *TransferContext) SetDest(v IDestinationContext) { s.dest = v }

func (s *TransferContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TransferContext) TRANSFER() antlr.TerminalNode {
	return s.GetToken(FaRlParserTRANSFER, 0)
}

func (s *TransferContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(FaRlParserLPAREN, 0)
}

func (s *TransferContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(FaRlParserNEWLINE)
}

func (s *TransferContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(FaRlParserNEWLINE, i)
}

func (s *TransferContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(FaRlParserRPAREN, 0)
}

func (s *TransferContext) FROM() antlr.TerminalNode {
	return s.GetToken(FaRlParserFROM, 0)
}

func (s *TransferContext) TO() antlr.TerminalNode {
	return s.GetToken(FaRlParserTO, 0)
}

func (s *TransferContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *TransferContext) MonetaryAll() IMonetaryAllContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IMonetaryAllContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IMonetaryAllContext)
}

func (s *TransferContext) ValueAwareSource() IValueAwareSourceContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IValueAwareSourceContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IValueAwareSourceContext)
}

func (s *TransferContext) Destination() IDestinationContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IDestinationContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IDestinationContext)
}

func (s *TransferContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterTransfer(s)
	}
}

func (s *TransferContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitTransfer(s)
	}
}

type SetTxMetaBlockContext struct {
	*StatementContext
}

func NewSetTxMetaBlockContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SetTxMetaBlockContext {
	var p = new(SetTxMetaBlockContext)

	p.StatementContext = NewEmptyStatementContext()
	p.parser = parser
	p.CopyFrom(ctx.(*StatementContext))

	return p
}

func (s *SetTxMetaBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SetTxMetaBlockContext) SET() antlr.TerminalNode {
	return s.GetToken(FaRlParserSET, 0)
}

func (s *SetTxMetaBlockContext) TRANSACTION() antlr.TerminalNode {
	return s.GetToken(FaRlParserTRANSACTION, 0)
}

func (s *SetTxMetaBlockContext) METADATA() antlr.TerminalNode {
	return s.GetToken(FaRlParserMETADATA, 0)
}

func (s *SetTxMetaBlockContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(FaRlParserLBRACE, 0)
}

func (s *SetTxMetaBlockContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(FaRlParserNEWLINE)
}

func (s *SetTxMetaBlockContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(FaRlParserNEWLINE, i)
}

func (s *SetTxMetaBlockContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(FaRlParserRBRACE, 0)
}

func (s *SetTxMetaBlockContext) AllMetadataEntry() []IMetadataEntryContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IMetadataEntryContext)(nil)).Elem())
	var tst = make([]IMetadataEntryContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IMetadataEntryContext)
		}
	}

	return tst
}

func (s *SetTxMetaBlockContext) MetadataEntry(i int) IMetadataEntryContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IMetadataEntryContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IMetadataEntryContext)
}

func (s *SetTxMetaBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterSetTxMetaBlock(s)
	}
}

func (s *SetTxMetaBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitSetTxMetaBlock(s)
	}
}

type SetTxMetaContext struct {
	*StatementContext
}

func NewSetTxMetaContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SetTxMetaContext {
	var p = new(SetTxMetaContext)

	p.StatementContext = NewEmptyStatementContext()
	p.parser = parser
	p.CopyFrom(ctx.(*StatementContext))

	return p
}

func (s *SetTxMetaContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SetTxMetaContext) SET() antlr.TerminalNode {
	return s.GetToken(FaRlParserSET, 0)
}

func (s *SetTxMetaContext) TRANSACTION() antlr.TerminalNode {
	return s.GetToken(FaRlParserTRANSACTION, 0)
}

func (s *SetTxMetaContext) METADATA() antlr.TerminalNode {
	return s.GetToken(FaRlParserMETADATA, 0)
}

func (s *SetTxMetaContext) STRING() antlr.TerminalNode {
	return s.GetToken(FaRlParserSTRING, 0)
}

func (s *SetTxMetaContext) EQ() antlr.TerminalNode {
	return s.GetToken(FaRlParserEQ, 0)
}

func (s *SetTxMetaContext) MetadataValue() IMetadataValueContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IMetadataValueContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IMetadataValueContext)
}

func (s *SetTxMetaContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterSetTxMeta(s)
	}
}

func (s *SetTxMetaContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitSetTxMeta(s)
	}
}

type FailContext struct {
	*StatementContext
}

func NewFailContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *FailContext {
	var p = new(FailContext)

	p.StatementContext = NewEmptyStatementContext()
	p.parser = parser
	p.CopyFrom(ctx.(*StatementContext))

	return p
}

func (s *FailContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FailContext) FAIL() antlr.TerminalNode {
	return s.GetToken(FaRlParserFAIL, 0)
}

func (s *FailContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterFail(s)
	}
}

func (s *FailContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitFail(s)
	}
}

type SetAccountMetaContext struct {
	*StatementContext
}

func NewSetAccountMetaContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SetAccountMetaContext {
	var p = new(SetAccountMetaContext)

	p.StatementContext = NewEmptyStatementContext()
	p.parser = parser
	p.CopyFrom(ctx.(*StatementContext))

	return p
}

func (s *SetAccountMetaContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SetAccountMetaContext) SET() antlr.TerminalNode {
	return s.GetToken(FaRlParserSET, 0)
}

func (s *SetAccountMetaContext) TY_ACCOUNT() antlr.TerminalNode {
	return s.GetToken(FaRlParserTY_ACCOUNT, 0)
}

func (s *SetAccountMetaContext) METADATA() antlr.TerminalNode {
	return s.GetToken(FaRlParserMETADATA, 0)
}

func (s *SetAccountMetaContext) OF() antlr.TerminalNode {
	return s.GetToken(FaRlParserOF, 0)
}

func (s *SetAccountMetaContext) Expression() IExpressionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExpressionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *SetAccountMetaContext) KEY() antlr.TerminalNode {
	return s.GetToken(FaRlParserKEY, 0)
}

func (s *SetAccountMetaContext) STRING() antlr.TerminalNode {
	return s.GetToken(FaRlParserSTRING, 0)
}

func (s *SetAccountMetaContext) EQ() antlr.TerminalNode {
	return s.GetToken(FaRlParserEQ, 0)
}

func (s *SetAccountMetaContext) MetadataValue() IMetadataValueContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IMetadataValueContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IMetadataValueContext)
}

func (s *SetAccountMetaContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterSetAccountMeta(s)
	}
}

func (s *SetAccountMetaContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitSetAccountMeta(s)
	}
}

func (p *FaRlParser) Statement() (localctx IStatementContext) {
	localctx = NewStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, FaRlParserRULE_statement)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(251)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 19, p.GetParserRuleContext()) {
	case 1:
		localctx = NewPrintContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(194)
			p.Match(FaRlParserPRINT)
		}
		{
			p.SetState(195)

			var _x = p.expression(0)

			localctx.(*PrintContext).expr = _x
		}

	case 2:
		localctx = NewFailContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(196)
			p.Match(FaRlParserFAIL)
		}

	case 3:
		localctx = NewTransferContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(197)
			p.Match(FaRlParserTRANSFER)
		}
		p.SetState(200)
		p.GetErrorHandler().Sync(p)
		switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 16, p.GetParserRuleContext()) {
		case 1:
			{
				p.SetState(198)

				var _x = p.expression(0)

				localctx.(*TransferContext).mon = _x
			}

		case 2:
			{
				p.SetState(199)

				var _x = p.MonetaryAll()

				localctx.(*TransferContext).monAll = _x
			}

		}
		{
			p.SetState(202)
			p.Match(FaRlParserLPAREN)
		}
		{
			p.SetState(203)
			p.Match(FaRlParserNEWLINE)
		}
		p.SetState(216)
		p.GetErrorHandler().Sync(p)

		switch p.GetTokenStream().LA(1) {
		case FaRlParserFROM:
			{
				p.SetState(204)
				p.Match(FaRlParserFROM)
			}
			{
				p.SetState(205)

				var _x = p.ValueAwareSource()

				localctx.(*TransferContext).src = _x
			}
			{
				p.SetState(206)
				p.Match(FaRlParserNEWLINE)
			}
			{
				p.SetState(207)
				p.Match(FaRlParserTO)
			}
			{
				p.SetState(208)

				var _x = p.Destination()

				localctx.(*TransferContext).dest = _x
			}

		case FaRlParserTO:
			{
				p.SetState(210)
				p.Match(FaRlParserTO)
			}
			{
				p.SetState(211)

				var _x = p.Destination()

				localctx.(*TransferContext).dest = _x
			}
			{
				p.SetState(212)
				p.Match(FaRlParserNEWLINE)
			}
			{
				p.SetState(213)
				p.Match(FaRlParserFROM)
			}
			{
				p.SetState(214)

				var _x = p.ValueAwareSource()

				localctx.(*TransferContext).src = _x
			}

		default:
			panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		}
		{
			p.SetState(218)
			p.Match(FaRlParserNEWLINE)
		}
		{
			p.SetState(219)
			p.Match(FaRlParserRPAREN)
		}

	case 4:
		localctx = NewSetTxMetaContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(221)
			p.Match(FaRlParserSET)
		}
		{
			p.SetState(222)
			p.Match(FaRlParserTRANSACTION)
		}
		{
			p.SetState(223)
			p.Match(FaRlParserMETADATA)
		}
		{
			p.SetState(224)
			p.Match(FaRlParserSTRING)
		}
		{
			p.SetState(225)
			p.Match(FaRlParserEQ)
		}
		{
			p.SetState(226)
			p.MetadataValue()
		}

	case 5:
		localctx = NewSetTxMetaBlockContext(p, localctx)
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(227)
			p.Match(FaRlParserSET)
		}
		{
			p.SetState(228)
			p.Match(FaRlParserTRANSACTION)
		}
		{
			p.SetState(229)
			p.Match(FaRlParserMETADATA)
		}
		{
			p.SetState(230)
			p.Match(FaRlParserLBRACE)
		}
		{
			p.SetState(231)
			p.Match(FaRlParserNEWLINE)
		}
		p.SetState(235)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = _la == FaRlParserSTRING {
			{
				p.SetState(232)
				p.MetadataEntry()
			}
			{
				p.SetState(233)
				p.Match(FaRlParserNEWLINE)
			}

			p.SetState(237)
			p.GetErrorHandler().Sync(p)
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(239)
			p.Match(FaRlParserRBRACE)
		}

	case 6:
		localctx = NewSetAccountMetaContext(p, localctx)
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(241)
			p.Match(FaRlParserSET)
		}
		{
			p.SetState(242)
			p.Match(FaRlParserTY_ACCOUNT)
		}
		{
			p.SetState(243)
			p.Match(FaRlParserMETADATA)
		}
		{
			p.SetState(244)
			p.Match(FaRlParserOF)
		}
		{
			p.SetState(245)
			p.expression(0)
		}
		{
			p.SetState(246)
			p.Match(FaRlParserKEY)
		}
		{
			p.SetState(247)
			p.Match(FaRlParserSTRING)
		}
		{
			p.SetState(248)
			p.Match(FaRlParserEQ)
		}
		{
			p.SetState(249)
			p.MetadataValue()
		}

	}

	return localctx
}

// IScriptContext is an interface to support dynamic dispatch.
type IScriptContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Get_statement returns the _statement rule contexts.
	Get_statement() IStatementContext

	// Set_statement sets the _statement rule contexts.
	Set_statement(IStatementContext)

	// GetStmts returns the stmts rule context list.
	GetStmts() []IStatementContext

	// SetStmts sets the stmts rule context list.
	SetStmts([]IStatementContext)

	// IsScriptContext differentiates from other interfaces.
	IsScriptContext()
}

type ScriptContext struct {
	*antlr.BaseParserRuleContext
	parser     antlr.Parser
	_statement IStatementContext
	stmts      []IStatementContext
}

func NewEmptyScriptContext() *ScriptContext {
	var p = new(ScriptContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = FaRlParserRULE_script
	return p
}

func (*ScriptContext) IsScriptContext() {}

func NewScriptContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ScriptContext {
	var p = new(ScriptContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = FaRlParserRULE_script

	return p
}

func (s *ScriptContext) GetParser() antlr.Parser { return s.parser }

func (s *ScriptContext) Get_statement() IStatementContext { return s._statement }

func (s *ScriptContext) Set_statement(v IStatementContext) { s._statement = v }

func (s *ScriptContext) GetStmts() []IStatementContext { return s.stmts }

func (s *ScriptContext) SetStmts(v []IStatementContext) { s.stmts = v }

func (s *ScriptContext) EOF() antlr.TerminalNode {
	return s.GetToken(FaRlParserEOF, 0)
}

func (s *ScriptContext) AllStatement() []IStatementContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IStatementContext)(nil)).Elem())
	var tst = make([]IStatementContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IStatementContext)
		}
	}

	return tst
}

func (s *ScriptContext) Statement(i int) IStatementContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IStatementContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IStatementContext)
}

func (s *ScriptContext) AllNEWLINE() []antlr.TerminalNode {
	return s.GetTokens(FaRlParserNEWLINE)
}

func (s *ScriptContext) NEWLINE(i int) antlr.TerminalNode {
	return s.GetToken(FaRlParserNEWLINE, i)
}

func (s *ScriptContext) AllVarDecl() []IVarDeclContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IVarDeclContext)(nil)).Elem())
	var tst = make([]IVarDeclContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IVarDeclContext)
		}
	}

	return tst
}

func (s *ScriptContext) VarDecl(i int) IVarDeclContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IVarDeclContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IVarDeclContext)
}

func (s *ScriptContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ScriptContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ScriptContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.EnterScript(s)
	}
}

func (s *ScriptContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(FaRlListener); ok {
		listenerT.ExitScript(s)
	}
}

func (p *FaRlParser) Script() (localctx IScriptContext) {
	localctx = NewScriptContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 40, FaRlParserRULE_script)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(256)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == FaRlParserNEWLINE {
		{
			p.SetState(253)
			p.Match(FaRlParserNEWLINE)
		}

		p.SetState(258)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(267)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == FaRlParserVAR {
		{
			p.SetState(259)
			p.VarDecl()
		}
		p.SetState(261)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = _la == FaRlParserNEWLINE {
			{
				p.SetState(260)
				p.Match(FaRlParserNEWLINE)
			}

			p.SetState(263)
			p.GetErrorHandler().Sync(p)
			_la = p.GetTokenStream().LA(1)
		}

		p.SetState(269)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(270)

		var _x = p.Statement()

		localctx.(*ScriptContext)._statement = _x
	}
	localctx.(*ScriptContext).stmts = append(localctx.(*ScriptContext).stmts, localctx.(*ScriptContext)._statement)
	p.SetState(279)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 24, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			p.SetState(272)
			p.GetErrorHandler().Sync(p)
			_la = p.GetTokenStream().LA(1)

			for ok := true; ok; ok = _la == FaRlParserNEWLINE {
				{
					p.SetState(271)
					p.Match(FaRlParserNEWLINE)
				}

				p.SetState(274)
				p.GetErrorHandler().Sync(p)
				_la = p.GetTokenStream().LA(1)
			}
			{
				p.SetState(276)

				var _x = p.Statement()

				localctx.(*ScriptContext)._statement = _x
			}
			localctx.(*ScriptContext).stmts = append(localctx.(*ScriptContext).stmts, localctx.(*ScriptContext)._statement)

		}
		p.SetState(281)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 24, p.GetParserRuleContext())
	}
	p.SetState(283)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == FaRlParserNEWLINE {
		{
			p.SetState(282)
			p.Match(FaRlParserNEWLINE)
		}

	}
	{
		p.SetState(285)
		p.Match(FaRlParserEOF)
	}

	return localctx
}

func (p *FaRlParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 3:
		var t *ExpressionContext = nil
		if localctx != nil {
			t = localctx.(*ExpressionContext)
		}
		return p.Expression_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *FaRlParser) Expression_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 3)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
