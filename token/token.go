// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package token defines constants representing the lexical tokens of the Go
// programming language and basic operations on tokens (printing, predicates).
//
package token

import "strconv"

// Token is the set of lexical tokens of the Go programming language.
type Token int

// The list of tokens.
const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	COMMENT

	literal_beg
	// Identifiers and basic type literals
	// (these tokens stand for classes of literals)
	IDENT  // main
	INT    // 12345
	FLOAT  // 123.45
	IMAG   // 123.45i
	CHAR   // 'a'
	STRING // "abc"
	literal_end

	operator_beg
	// Operators and delimiters
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	AND     // &
	OR      // |
	XOR     // ^
	SHL     // <<
	SHR     // >>
	AND_NOT // &^

	ADD_ASSIGN // +=
	SUB_ASSIGN // -=
	MUL_ASSIGN // *=
	QUO_ASSIGN // /=
	REM_ASSIGN // %=

	AND_ASSIGN     // &=
	OR_ASSIGN      // |=
	XOR_ASSIGN     // ^=
	SHL_ASSIGN     // <<=
	SHR_ASSIGN     // >>=
	AND_NOT_ASSIGN // &^=

	LAND  // &&
	LOR   // ||
	ARROW // <-
	INC   // ++
	DEC   // --

	EQL    // ==
	LSS    // <
	GTR    // >
	ASSIGN // =
	NOT    // !

	NEQ      // !=
	LEQ      // <=
	GEQ      // >=
	DEFINE   // :=
	ELLIPSIS // ...

	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	PERIOD // .

	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	COLON     // :
	operator_end

	keyword_beg
	// Keywords
	BREAK
	CASE
	CHAN
	CONST
	CONTINUE

	DEFAULT
	DEFER
	ELSE
	FALLTHROUGH
	FOR

	FUNC
	GO
	GOTO
	IF
	IMPORT

	INTERFACE
	MAP
	PACKAGE
	RANGE
	RETURN

	SELECT
	STRUCT
	SWITCH
	TYPE
	VAR
	keyword_end

	unihan_beg

	u_keyword_beg
	// Unihan Keywords
	U包 //package
	U入 //import; MAYBE: change to 输shu
	U久 //const; MAYBE: change to 永yong or 常chang
	U变 //var
	U种 //type; MAYBE: change to 式shi
	U功 //func
	U构 //struct
	U图 //map
	U面 //interface
	U通 //channel

	U如 //if; TODO: change to 若ruo
	U否 //else
	U考 //switch
	U事 //case
	U别 //default
	U掉 //fallthrough
	U选 //select
	U为 //for
	U围 //range
	U终 //defer
	U去 //go
	U回 //return
	U破 //break
	U继 //continue; TODO: change to 续xu
	U跳 //goto
	u_keyword_end

	u_groword_beg
	// Gro-specific Unihan Keywords
	U用 //use
	U源 //source
 	U英 //ascii
	U做 //do
	U对 //assert

	U这 //this
	U尖 //pointer
	U任 //any
	U引 //quote

	U出 //exit
	U前 //prev
	U后 //next
	U学 //learn
	u_groword_end

	unihan_end

	/*
	u_identifier_beg
	// Unihan Identifiers
	U正 //main
	U开 //init

	U真 //true
	U假 //false
	U空 //nil; MAYBE: change to 无wu
	U毫 //iota

	U能 //cap
	U度 //len
	U实 //real
	U虚 //imag
	U造 //make
	U新 //new
	U关 //close
	U加 //append
	U副 //copy
	U删 //delete
	U丢 //panic
	U抓 //recover
	U写 //print
	U线 //println; MAYBE: change to 印yin

	U节 //byte
	U字 //rune
	U串 //string
	U双 //bool; MAYBE: change to 布bu
	U错 //error
	U镇 //uintptr
	u_identifier_end

	u_suffixable_beg
	// Unihan Suffixable Identifiers
	U整 //int, int8, int16, int32, int64
	U绝 //uint, uint8, uint16, uint32, uint64
	U漂 //float32, float64; MAYBE: change to 浮fu
	U复 //complex, complex64, complex128
	u_suffixable_end

	u_macro_beg
	// Unihan Macros
	U鲜 //fresh - gro/macro/fresh
	U准 //prepare - gro/macro/command
	U执 //execute - gro/macro/command
	U跑 //run - gro/macro/command
	U叫 //call - gro/macro/expr
	U动 //dyn - groo/cmd/groo
	u_macro_end

	// Unihan Tentatives
	U让 //let
	U显 //vars
	U预 //prepack
	U先 //first
	U黑 //blacklist
	U白 //whitelist
	U特 //special
	U试 //try
	U具 //util
	U指 //spec
	U羔 //lamb
	U程 //proc
	U冲 //flush
	U建 //build
	U洗 //clean
	U解 //parse
	U类 //class
	U是 //is //[reflect.TypeOf]
	U侯 //while
	U自 //self
	U滤 //filter
	U减 //reduce
	U组 //groupby
	U颠 //reverse
	U长 //long //长(long)=int64 短/矮(short)=int8
	U除 //exception
	U摸 //pattern
	U田 //array
	U片 //slice //切
	U甲 //first
	U乙 //second
	U丙 //third
	U丁 //fourth
	U戊 //fifth
	U己 //sixth
	U庚 //seventh
	U辛 //eighth
	U壬 //ninth
	U癸 //tenth
	U他 //him
	U她 //her
	U它 //it
	*/

	DUD_LAST
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:     "EOF",
	COMMENT: "COMMENT",

	IDENT:  "IDENT",
	INT:    "INT",
	FLOAT:  "FLOAT",
	IMAG:   "IMAG",
	CHAR:   "CHAR",
	STRING: "STRING",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "%",

	AND:     "&",
	OR:      "|",
	XOR:     "^",
	SHL:     "<<",
	SHR:     ">>",
	AND_NOT: "&^",

	ADD_ASSIGN: "+=",
	SUB_ASSIGN: "-=",
	MUL_ASSIGN: "*=",
	QUO_ASSIGN: "/=",
	REM_ASSIGN: "%=",

	AND_ASSIGN:     "&=",
	OR_ASSIGN:      "|=",
	XOR_ASSIGN:     "^=",
	SHL_ASSIGN:     "<<=",
	SHR_ASSIGN:     ">>=",
	AND_NOT_ASSIGN: "&^=",

	LAND:  "&&",
	LOR:   "||",
	ARROW: "<-",
	INC:   "++",
	DEC:   "--",

	EQL:    "==",
	LSS:    "<",
	GTR:    ">",
	ASSIGN: "=",
	NOT:    "!",

	NEQ:      "!=",
	LEQ:      "<=",
	GEQ:      ">=",
	DEFINE:   ":=",
	ELLIPSIS: "...",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",
	COMMA:  ",",
	PERIOD: ".",

	RPAREN:    ")",
	RBRACK:    "]",
	RBRACE:    "}",
	SEMICOLON: ";",
	COLON:     ":",

	BREAK:    "break",
	CASE:     "case",
	CHAN:     "chan",
	CONST:    "const",
	CONTINUE: "continue",

	DEFAULT:     "default",
	DEFER:       "defer",
	ELSE:        "else",
	FALLTHROUGH: "fallthrough",
	FOR:         "for",

	FUNC:   "func",
	GO:     "go",
	GOTO:   "goto",
	IF:     "if",
	IMPORT: "import",

	INTERFACE: "interface",
	MAP:       "map",
	PACKAGE:   "package",
	RANGE:     "range",
	RETURN:    "return",

	SELECT: "select",
	STRUCT: "struct",
	SWITCH: "switch",
	TYPE:   "type",
	VAR:    "var",

	U包: "包", //package
	U入: "入", //import
	U久: "久", //const
	U变: "变", //var
	U种: "种", //type
	U功: "功", //func
	U构: "构", //struct
	U图: "图", //map
	U面: "面", //interface
	U通: "通", //channel
	U如: "如", //if
	U否: "否", //else
	U考: "考", //switch
	U事: "事", //case
	U别: "别", //default
	U掉: "掉", //fallthrough
	U选: "选", //select
	U为: "为", //for
	U围: "围", //range
	U终: "终", //defer
	U去: "去", //go
	U回: "回", //return
	U破: "破", //break
	U继: "继", //continue
	U跳: "跳", //goto

	U用: "用", //use
	U源: "源", //source
	U英: "英", //ascii
	U做: "做", //do
	U对: "对", //assert

	U这: "这", //this
	U尖: "尖", //pointer
	U任: "任", //any
	U引: "引", //quote

	U出: "出", //exit
	U前: "前", //prev
	U后: "后", //next
	U学: "学", //learn

	/*
	U正: "正", //main
	U开: "开", //init

	U真: "真", //true
	U假: "假", //false
	U空: "空", //nil
	U毫: "毫", //iota

	U能: "能", //cap
	U度: "度", //len
	U实: "实", //real
	U虚: "虚", //imag
	U造: "造", //make
	U新: "新", //new
	U关: "关", //close
	U加: "加", //append
	U副: "副", //copy
	U删: "删", //delete
	U丢: "丢", //panic
	U抓: "抓", //recover
	U写: "写", //print
	U线: "线", //println

	U节: "节", //byte
	U字: "字", //rune
	U串: "串", //string
	U双: "双", //bool
	U错: "错", //error
	U镇: "镇", //uintptr

	U整: "整", //int, int8, int16, int32, int64
	U绝: "绝", //uint, uint8, uint16, uint32, uint64
	U漂: "漂", //float32, float64
	U复: "复", //complex, complex64, complex128

	U鲜: "鲜", //fresh
	U准: "准", //prepare
	U执: "执", //execute
	U跑: "跑", //run
	U叫: "叫", //call
	U动: "动", //dyn
	*/

	DUD_LAST: "",
}

var Identifiers =  map[rune]string{}

func init() {
	for r, d:= range unihanMap {
		Identifiers[r] = d
	}
}

var unihanMap = map[rune]string {
	'正': "main",
	'开': "init",

	'真': "true",
	'假': "false",
	'空': "nil",
	'毫': "iota",

	'能': "cap",
	'度': "len",
	'实': "real",
	'虚': "imag",
	'造': "make",
	'新': "new",
	'关': "close",
	'加': "append",
	'副': "copy",
	'删': "delete",
	'丢': "panic",
	'抓': "recover",
	'写': "print",
	'线': "println",

	'节': "byte",
	'字': "rune",
	'串': "string",
	'双': "bool",
	'错': "error",
	'镇': "uintptr",
}

func IsSuffixable(r rune) bool {
	return r == '整' || r == '绝' || r == '漂' || r == '复'
}

var SuffixedIdents = map[string]string{
	"整": "int",
	"整8": "int8",
	"整16": "int16",
	"整32": "int32",
	"整64": "int64",

	"绝": "uint",
	"绝8": "uint8",
	"绝16": "uint16",
	"绝32": "uint32",
	"绝64": "uint64",

	"漂32": "float32",
	"漂64": "float64",

	"复": "complex",
	"复64": "complex64",
	"复128": "complex128",
}

var PackageNames = map[string]string{
	"形": "fmt",
	"网": "net",
	"序": "sort",
	"数": "math",
	"大": "math/big",
	"时": "time",
}

// String returns the string corresponding to the token tok.
// For operators, delimiters, and keywords the string is the actual
// token character sequence (e.g., for the token ADD, the string is
// "+"). For all other tokens the string corresponds to the token
// constant name (e.g. for the token IDENT, the string is "IDENT").
//
func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

// A set of constants for precedence-based expression parsing.
// Non-operators have lowest precedence, followed by operators
// starting with precedence 1 up to unary operators. The highest
// precedence serves as "catch-all" precedence for selector,
// indexing, and other operator and delimiter tokens.
//
const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 6
	HighestPrec = 7
)

// Precedence returns the operator precedence of the binary
// operator op. If op is not a binary operator, the result
// is LowestPrecedence.
//
func (op Token) Precedence() int {
	switch op {
	case LOR:
		return 1
	case LAND:
		return 2
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return 3
	case ADD, SUB, OR, XOR:
		return 4
	case MUL, QUO, REM, SHL, SHR, AND, AND_NOT:
		return 5
	}
	return LowestPrec
}

var keywords map[string]Token
var unihans  map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}

	unihans = make(map[string]Token)
	for i := unihan_beg + 1; i < unihan_end; i++ {
		unihans[tokens[i]] = i
	}
}

// Lookup maps an identifier to its keyword or Unihan token,
// or IDENT (if not a keyword or Unihan).
//
func Lookup(ident string) Token {
	if tok, is_keyword := keywords[ident]; is_keyword {
		return tok
	}
	if tok, is_unihan := unihans[ident]; is_unihan {
		return tok
	}
	return IDENT
}

// Predicates

// IsLiteral returns true for tokens corresponding to identifiers
// and basic type literals; it returns false otherwise.
//
func (tok Token) IsLiteral() bool { return literal_beg < tok && tok < literal_end }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
//
func (tok Token) IsOperator() bool { return operator_beg < tok && tok < operator_end }

// IsKeyword returns true for tokens corresponding to keywords;
// it returns false otherwise.
//
func (tok Token) IsKeyword() bool { return keyword_beg < tok && tok < keyword_end }

// IsUnihan returns true for tokens corresponding to Unihan tokens;
// it returns false otherwise.
//
func (tok Token) IsUnihan() bool { return unihan_beg < tok && tok < unihan_end }

// KouRadicalChars are all Unihan characters with a kou-radical on the left-hand side.
//
var KouRadicalChars =
	`㕤㕥㕧㕨㕩㕪㕫㕬㕭㕮㕰㕱㕲㕳㕴㕵㕶㕷㕸㕹㕼㕽㖀㖁㖂㖃㖄㖅㖆㖇㖉㖊㖏㖑㖒㖓㖔㖕㖗㖘㖞㖟㖠㖡㖢㖣㖤㖥㖦㖧㖨㖩㖪㖫㖬㖭㖮㖴㖵㖶`+
	`㖷㖸㖹㖺㖻㖼㖽㖿㗀㗁㗂㗃㗄㗅㗆㗇㗈㗋㗌㗍㗎㗏㗐㗑㗒㗓㗔㗕㗖㗘㗙㗚㗛㗜㗝㗞㗢㗣㗥㗦㗧㗩㗪㗫㗭㗰㗱㗲㗳㗴㗵㗶㗷㗸㗹㗺㗻㗼㗾㗿`+
	`㘀㘁㘂㘃㘄㘅㘆㘇㘈㘉㘊㘋㘌㘍㘎㘐㘑㘓㘔㘕㘖㘗㘙㘚㘛卟叨叩叫叭叮叱叶叹叺叻叼叽叿吀吁吃吅吆吇吋吐吒吓吔吖吗吘吙吚吜吟吠吡吣`+
	`吤吥吧吨吩吪听吭吮吰吱吲吵吶吷吸吹吺吻吼吽呀呁呃呅呋呌呍呎呏呐呒呓呔呕呖呗呚呛呜呝呞呟呠呡呢呣呤呥呦呧呩呪呫呬呭呮呯呱呲`+
	`味呴呵呶呷呸呹呺呻呼呾呿咀咁咂咃咄咆咇咈咉咊咋咍咏咐咑咓咔咕咖咗咘咙咚咛咜咝咞咟咡咣咤咥咦咧咩咪咬咭咮咯咰咱咲咳咴咵咶咷`+
	`咹咺咻咽咾咿哂哃哄哅哆哇哈哊哋哌响哎哏哐哑哒哓哔哕哖哗哘哙哚哜哝哞哟哠哢哣哤哦哧哨哩哪哫哬哮哯哰哱哳哴哵哶哷哸哹哺哻哼哽`+
	`哾唀唁唂唃唄唅唆唈唉唊唋唌唍唎唏唑唒唓唔唕唖唗唙唚唛唝唞唠唡唢唣唤唥唦唧唨唩唪唫唬唭唯唰唱唲唳唴唵唶唷唸唹唺唻唼唽唾唿啀`+
	`啁啂啃啄啅啈啉啊啋啌啍啐啑啒啕啖啗啘啛啜啝啞啡啢啣啤啥啦啧啨啩啪啫啭啮啯啰啱啲啳啴啵啶啷啸啹啺啼啽啾啿喀喁喂喃喅喇喈喉喊`+
	`喋喍喎喏喐喑喒喓喔喕喖喗喘喙喚喛喝喞喟喠喡喢喣喤喥喧喨喩喫喭喯喰喱喲喳喴喵喷喹喺喻喼喽嗁嗂嗃嗄嗅嗆嗈嗉嗊嗋嗌嗍嗎嗏嗐嗑嗒`+
	`嗓嗔嗕嗖嗗嗘嗙嗚嗛嗜嗝嗞嗟嗡嗢嗤嗥嗦嗨嗩嗪嗫嗬嗮嗯嗰嗱嗲嗳嗴嗵嗶嗷嗹嗺嗻嗼嗽嗾嗿嘀嘁嘃嘄嘅嘆嘇嘈嘊嘋嘌嘍嘎嘐嘑嘒嘓嘔嘕嘖`+
	`嘘嘙嘚嘛嘜嘝嘞嘟嘠嘡嘢嘣嘤嘥嘧嘨嘩嘪嘫嘬嘭嘮嘯嘰嘱嘲嘳嘴嘵嘶嘷嘸嘹嘺嘻嘽嘾嘿噀噁噂噃噄噅噆噇噈噉噊噋噌噍噎噏噑噒噓噔噖噗`+
	`噘噙噚噛噜噝噞噠噡噢噣噤噥噦噧噪噫噬噭噮噯噰噱噲噳噴噵噶噷噸噹噺噻噼噾噿嚀嚁嚂嚃嚄嚅嚆嚇嚈嚉嚊嚋嚌嚍嚎嚏嚐嚑嚒嚓嚔嚕嚖嚗`+
	`嚘嚙嚛嚜嚝嚟嚠嚡嚤嚥嚦嚧嚨嚩嚪嚫嚬嚯嚰嚱嚵嚶嚷嚸嚹嚺嚼嚽嚾嚿囀囁囃囄囆囇囈囉囋囌囎囐囑囒囓囔囕囖鳴鸣𠮙𠮜𠮝𠮟𠮤𠮧𠮨𠮩𠮪𠮬`+
	`𠮭𠮱𠮵𠮶𠮹𠮺𠮻𠮼𠮾𠮿𠯀𠯄𠯅𠯆𠯇𠯈𠯋𠯍𠯎𠯏𠯐𠯔𠯖𠯗𠯘𠯙𠯜𠯝𠯞𠯟𠯠𠯡𠯢𠯤𠯥𠯦𠯩𠯪𠯫𠯬𠯯𠯰𠯱𠯲𠯴𠯷𠯸𠯹𠯻𠯼𠯽𠯾𠯿𠰀𠰁𠰂𠰃𠰄𠰆𠰈`+
	`𠰉𠰊𠰋𠰌𠰍𠰏𠰐𠰑𠰒𠰖𠰗𠰘𠰙𠰚𠰜𠰠𠰢𠰧𠰩𠰪𠰭𠰮𠰯𠰱𠰲𠰳𠰴𠰵𠰷𠰸𠰹𠰺𠰻𠰼𠰽𠰾𠰿𠱀𠱁𠱂𠱃𠱅𠱆𠱇𠱈𠱉𠱊𠱋𠱌𠱍𠱎𠱏𠱐𠱓𠱔𠱕𠱖𠱘𠱙𠱚`+
	`𠱜𠱝𠱞𠱟𠱠𠱡𠱢𠱣𠱤𠱥𠱨𠱪𠱱𠱲𠱳𠱴𠱶𠱷𠱸𠱹𠱺𠱻𠱼𠱽𠱾𠱿𠲂𠲃𠲄𠲅𠲇𠲈𠲊𠲋𠲌𠲍𠲎𠲏𠲐𠲓𠲔𠲕𠲖𠲗𠲙𠲚𠲛𠲜𠲝𠲞𠲟𠲠𠲡𠲢𠲣𠲤𠲥𠲦𠲧𠲨`+
	`𠲪𠲫𠲬𠲭𠲮𠲰𠲲𠲳𠲴𠲵𠲶𠲷𠲸𠲺𠲼𠲽𠲾𠲿𠳀𠳁𠳂𠳃𠳈𠳉𠳍𠳎𠳏𠳐𠳑𠳒𠳓𠳔𠳕𠳖𠳗𠳘𠳚𠳜𠳝𠳞𠳟𠳠𠳡𠳣𠳤𠳥𠳦𠳧𠳨𠳩𠳪𠳭𠳰𠳱𠳲𠳳𠳴𠳶𠳷𠳸`+
	`𠳹𠳺𠳻𠳼𠳽𠳾𠳿𠴀𠴁𠴂𠴃𠴄𠴆𠴇𠴈𠴉𠴊𠴋𠴌𠴍𠴎𠴏𠴐𠴑𠴒𠴓𠴔𠴕𠴖𠴗𠴘𠴙𠴚𠴛𠴜𠴝𠴞𠴟𠴠𠴡𠴢𠴣𠴤𠴥𠴧𠴨𠴪𠴫𠴬𠴭𠴮𠴯𠴰𠴱𠴲𠴳𠴴𠴵𠴶𠴷`+
	`𠴹𠴺𠴻𠴼𠴽𠴾𠵃𠵄𠵅𠵆𠵇𠵈𠵉𠵋𠵌𠵎𠵏𠵐𠵑𠵒𠵔𠵕𠵖𠵘𠵙𠵚𠵜𠵝𠵟𠵠𠵡𠵢𠵣𠵨𠵩𠵫𠵭𠵮𠵯𠵰𠵱𠵴𠵷𠵸𠵹𠵺𠵻𠵼𠵽𠵾𠵿𠶀𠶁𠶂𠶃𠶄𠶅𠶆𠶈𠶉`+
	`𠶊𠶋𠶌𠶍𠶎𠶏𠶐𠶑𠶒𠶓𠶔𠶕𠶖𠶗𠶙𠶚𠶛𠶜𠶝𠶞𠶟𠶠𠶡𠶢𠶣𠶤𠶥𠶦𠶧𠶨𠶩𠶪𠶫𠶭𠶯𠶲𠶴𠶸𠶹𠶺𠶻𠶼𠶽𠶾𠶿𠷀𠷁𠷂𠷃𠷄𠷅𠷆𠷇𠷈𠷉𠷊𠷋𠷌𠷍𠷐`+
	`𠷑𠷕𠷖𠷘𠷙𠷚𠷝𠷟𠷢𠷣𠷤𠷥𠷦𠷧𠷨𠷩𠷪𠷬𠷭𠷮𠷯𠷲𠷴𠷵𠷶𠷸𠷹𠷺𠷻𠷼𠷾𠷿𠸀𠸁𠸂𠸃𠸄𠸇𠸉𠸊𠸋𠸌𠸍𠸎𠸏𠸐𠸑𠸒𠸓𠸔𠸕𠸖𠸘𠸚𠸝𠸞𠸟𠸠𠸡𠸢`+
	`𠸣𠸤𠸥𠸦𠸧𠸨𠸩𠸪𠸫𠸬𠸯𠸰𠸳𠸴𠸵𠸷𠸸𠸹𠸺𠸻𠸼𠸽𠸾𠹀𠹁𠹂𠹃𠹄𠹅𠹆𠹇𠹊𠹋𠹌𠹍𠹎𠹏𠹐𠹑𠹒𠹓𠹔𠹕𠹖𠹗𠹘𠹙𠹚𠹛𠹞𠹠𠹡𠹤𠹥𠹦𠹭𠹮𠹯𠹰𠹱`+
	`𠹲𠹳𠹴𠹵𠹶𠹷𠹸𠹹𠹺𠹻𠹼𠹽𠹿𠺀𠺁𠺂𠺄𠺅𠺆𠺈𠺉𠺊𠺋𠺌𠺍𠺏𠺑𠺒𠺓𠺔𠺕𠺖𠺗𠺘𠺙𠺚𠺜𠺝𠺟𠺠𠺡𠺢𠺣𠺦𠺧𠺨𠺩𠺪𠺫𠺬𠺭𠺮𠺰𠺱𠺲𠺳𠺴𠺵𠺶𠺷`+
	`𠺸𠺹𠺺𠺻𠺼𠺽𠺾𠺿𠻀𠻂𠻃𠻄𠻅𠻆𠻈𠻉𠻊𠻋𠻍𠻎𠻏𠻐𠻑𠻒𠻓𠻔𠻕𠻗𠻘𠻙𠻛𠻜𠻞𠻟𠻠𠻢𠻣𠻤𠻥𠻦𠻧𠻨𠻩𠻪𠻫𠻬𠻯𠻱𠻲𠻳𠻴𠻵𠻶𠻷𠻹𠻺𠻻𠻼𠻽𠻾`+
	`𠻿𠼀𠼁𠼂𠼄𠼇𠼈𠼉𠼊𠼋𠼌𠼍𠼎𠼏𠼐𠼒𠼓𠼔𠼕𠼖𠼗𠼘𠼙𠼚𠼜𠼝𠼟𠼠𠼢𠼣𠼤𠼥𠼦𠼩𠼪𠼫𠼬𠼭𠼮𠼯𠼰𠼱𠼲𠼳𠼴𠼵𠼶𠼸𠼹𠼺𠼻𠼼𠼽𠼾𠽀𠽁𠽂𠽃𠽄𠽅`+
	`𠽆𠽇𠽈𠽉𠽊𠽋𠽌𠽍𠽎𠽏𠽐𠽑𠽒𠽓𠽔𠽕𠽖𠽗𠽙𠽛𠽜𠽞𠽟𠽡𠽢𠽣𠽤𠽥𠽦𠽧𠽨𠽩𠽪𠽫𠽬𠽭𠽮𠽯𠽰𠽱𠽲𠽳𠽴𠽵𠽶𠽹𠽻𠽼𠽾𠽿𠾀𠾁𠾆𠾇𠾈𠾊𠾋𠾌𠾍𠾎`+
	`𠾏𠾐𠾑𠾒𠾓𠾔𠾕𠾗𠾘𠾙𠾚𠾛𠾜𠾝𠾞𠾠𠾡𠾢𠾣𠾦𠾨𠾩𠾪𠾫𠾬𠾭𠾮𠾯𠾰𠾱𠾲𠾴𠾵𠾶𠾷𠾸𠾺𠾻𠾼𠾽𠾾𠾿𠿀𠿁𠿂𠿃𠿄𠿅𠿆𠿇𠿈𠿊𠿋𠿌𠿍𠿎𠿏𠿐𠿑𠿒`+
	`𠿓𠿔𠿖𠿗𠿘𠿙𠿚𠿛𠿜𠿝𠿞𠿠𠿢𠿣𠿤𠿥𠿨𠿩𠿪𠿫𠿬𠿭𠿮𠿯𠿰𠿱𠿳𠿴𠿵𠿶𠿷𠿸𠿹𠿺𠿼𠿾𠿿𡀀𡀁𡀂𡀃𡀄𡀅𡀇𡀊𡀌𡀍𡀎𡀏𡀐𡀑𡀔𡀕𡀖𡀗𡀘𡀙𡀚𡀛𡀜`+
	`𡀝𡀞𡀟𡀠𡀡𡀢𡀣𡀥𡀦𡀧𡀨𡀩𡀫𡀬𡀭𡀮𡀯𡀰𡀱𡀲𡀳𡀴𡀵𡀶𡀷𡀹𡀺𡀼𡀽𡀾𡀿𡁀𡁁𡁂𡁃𡁄𡁅𡁆𡁇𡁈𡁊𡁋𡁌𡁍𡁎𡁏𡁐𡁑𡁒𡁓𡁔𡁕𡁖𡁙𡁚𡁛𡁜𡁝𡁞𡁟`+
	`𡁠𡁡𡁣𡁤𡁦𡁧𡁪𡁫𡁬𡁭𡁮𡁯𡁰𡁱𡁲𡁴𡁵𡁶𡁷𡁸𡁹𡁺𡁻𡁼𡁽𡁾𡁿𡂀𡂁𡂂𡂃𡂄𡂅𡂆𡂈𡂉𡂊𡂋𡂌𡂍𡂎𡂏𡂐𡂑𡂒𡂓𡂔𡂕𡂖𡂗𡂘𡂙𡂚𡂛𡂜𡂝𡂞𡂠𡂡𡂢`+
	`𡂣𡂥𡂩𡂪𡂫𡂭𡂮𡂰𡂱𡂳𡂴𡂵𡂷𡂸𡂹𡂺𡂻𡂼𡂿𡃀𡃁𡃂𡃃𡃄𡃅𡃆𡃇𡃈𡃉𡃊𡃌𡃍𡃎𡃏𡃐𡃑𡃒𡃓𡃔𡃕𡃖𡃗𡃘𡃙𡃚𡃛𡃜𡃝𡃞𡃢𡃤𡃥𡃦𡃧𡃨𡃩𡃪𡃮𡃰𡃱`+
	`𡃲𡃳𡃴𡃵𡃶𡃹𡃺𡃻𡃼𡃽𡃾𡃿𡄁𡄃𡄄𡄆𡄇𡄊𡄋𡄍𡄎𡄏𡄐𡄑𡄓𡄔𡄕𡄖𡄗𡄘𡄙𡄟𡄠𡄡𡄢𡄣𡄤𡄥𡄦𡄧𡄨𡄩𡄪𡄫𡄭𡄮𡄯𡄱𡄳𡄴𡄵𡄶𡄷𡄸𡄺𡄼𡄽𡄾𡅁𡅂`+
	`𡅃𡅅𡅆𡅇𡅈𡅉𡅊𡅋𡅌𡅍𡅎𡅏𡅑𡅒𡅓𡅗𡅘𡅙𡅛𡅜𡅞𡅠𡅢𡅣𡅥𡅧𡅨𡅩𡅪𡅫𡅬𡅭𡅯𡅰𡅲𡅳𡅵𡅶𡅷𡅹𡅺𡅼𡅿𡆀𡆁𡆂𡆄𡆅𡆆𡆇𡆈𡆋𡆌𡆍𡆏𡆑𡆓𡆕𡆖𡆗`+
	`𡆘𡆙𡆚𡆜𡆝𡆞𡆟𢒯𧛧𨙫𩐉𩒻𪄨𪄼𪚩𪠳𪠴𪠵𪠶𪠸𪠺𪠻𪠼𪠽𪠾𪠿𪡀𪡁𪡂𪡃𪡄𪡆𪡇𪡈𪡊𪡋𪡏𪡓𪡔𪡕𪡗𪡙𪡚𪡛𪡝𪡞𪡟𪡠𪡡𪡣𪡥𪡦𪡧𪡨𪡩𪡫𪡭𪡮𪡱𪡴`+
	`𪡵𪡶𪡷𪡸𪡺𪡽𪡾𪡿𪢀𪢁𪢂𪢃𪢄𪢅𪢆𪢇𪢉𪢊𪢋𪢌𪢍𪢎𪢐𪢑𪢒𪢔𪢕𪢖𪢗𪢘𪢙𪢚𪢜𪢝𪢟𪢠𪢤𪢥𪢧𫛗𫝘𫝚𫝜𫝞`
	//the last few are CJK ext-D chars that often display incorrectly in the Windows fonts

