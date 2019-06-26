const {Token, Lexer, CstParser, createToken} = require("chevrotain")

//FullStop statement terminator
const Terminator = createToken({
    name: "Terminator",
    pattern: /\./
})

//Comments
const Comment = createToken({
    name: "Comment",
    pattern: Lexer.NA
})

const LineStartCommentChar = createToken({
    name: "LineStartCommentChar",
    pattern: /\*/,
    categories: Comment
})


const AnyCommentChar = createToken({
    name: "AnyCommentChar",
    pattern: /\"/,
    categories: Comment
})

const CommentLine = createToken({
    name: "CommentLine",
    pattern: /^[*].+$/,
    categories: Comment
})

const InlineComment = createToken({
    name: "InlineComment",
    pattern: /(?<=.)\".*$/,
    categories: Comment
})

// User defined message
var msg = new RegExp(/(?<=\s)MESSAGE.+\./, 'i')
const Message = createToken({
    name: "Message",
    pattern: msg
})

// Literals
const StringLiteral = createToken({
    name: "StringLiteral",
    pattern: /\'.+\'/
})

const NumberLiteral = createToken({
    name: "NumberLiteral",
    pattern: /-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d+)?/
})

const WhiteSpace = createToken({
    name: "WhiteSpace",
    pattern: /\s+/,
    group: Lexer.SKIPPED
})

// Conditional
const IfStatement = createToken({
    name: "IfStatement",
    pattern: Lexer.NA
})

var ifstart = new RegExp(/(?<=\s)IF(?=\s)/, 'i')
const IfStart = createToken({
    name: "IfStart",
    pattern: ifstart,
    categories: IfStatement
})

var ifend = new RegExp(/(?<=\s)ENDIF\./, 'i')
const IfEnd = createToken({
    name: "IfEnd",
    pattern: ifend,
    categories: IfStatement
})

//Loop
const LoopStatement = createToken({
    name: "LoopStatement",
    pattern: Lexer.NA
})

var loopstart = new RegExp(/(?<=\s)LOOP(?=\s)/, 'i')
const LoopStart = createToken({
    name: "LoopStart",
    pattern: loopstart,
    categories: LoopStatement
})

var loopend = new RegExp(/(?<=\s)ENDLOOP\./, 'i')
const LoopEnd = createToken({
    name: "LoopEnd",
    pattern: loopend,
    categories: LoopStatement
})

// Arithmetic Operators
const ArithmeticOps = createToken({
    name: "LogicalOps",
    pattern: Lexer.NA
})

const ArithmeticPlus = createToken({
    name: "ArithmeticPlus",
    pattern: /\+/,
    categories: ArithmeticOps
})

const ArithmeticMinus = createToken({
    name: "ArithmeticMinus",
    pattern: /\-/,
    categories: ArithmeticOps
})

const ArithmeticMult = createToken({
    name: "ArithmeticMult",
    pattern: /\*/,
    categories: ArithmeticOps
})

const ArithmeticPow = createToken({
    name: "ArithmeticPow",
    pattern: /\*{2}/,
    categories: ArithmeticOps
})

var mod = new RegExp(/(?<=\s)MOD\s+/, 'i')
const ArithmeticMod = createToken({
    name: "ArithmeticMod",
    pattern: mod,
    categories: ArithmeticOps
})

const ArithmeticDiv = createToken({
    name: "ArithmeticDiv",
    pattern: /\//,
    categories: ArithmeticOps
})


// Comparative Operators
const ComparitiveOps = createToken({
    name: "ComparitiveOps",
    pattern: Lexer.NA
})

const LessThan = createToken({
    name: "LessThan",
    pattern: /(?<=\s)(<|LT)(?=\s)/,
    categories: ComparitiveOps
})

const LessThanEqualTo = createToken({
    name: "LessThanEqualTo",
    pattern: /(?<=\s)(<=|LE)(?=\s)/,
    categories: ComparitiveOps
})

const GreaterThan = createToken({
    name: "GreaterThan",
    pattern: /(?<=\s)(>|GT)(?=\s)/,
    categories: ComparitiveOps
})

const GreaterThanEqualTo = createToken({
    name: "GreaterThanEqualTo",
    pattern: /(?<=\s)(>=|GE)(?=\s)/,
    categories: ComparitiveOps
})

const EqualTo = createToken({
    name: "EqualTo",
    pattern: /(?<=\s)(=|EQ)(?=\s)/,
    categories: ComparitiveOps
})

const NotEqual = createToken({
    name: "NotEqual",
    pattern: /(?<=\s)(<>|NE)(?=\s)/,
    categories: ComparitiveOps
})

const LogicalNot = createToken({
    name: "LogicalNot",
    pattern: /\s+NOT\s+/,
    categories: ComparitiveOps
})

// Logical Operators
const LogicalOps = createToken({
    name: "LogicalOps",
    pattern: Lexer.NA
})

const LogicalAnd = createToken({
    name: "LogicalAnd",
    pattern: /\s+AND\s+/,
    categories: LogicalOps
})

const LogicalOr = createToken({
    name: "LogicalOr",
    pattern: /\s+OR\s+/,
    categories: LogicalOps
})

const LogicalNot = createToken({
    name: "LogicalNot",
    pattern: /\s+NOT\s+/,
    categories: LogicalOps
})

//Header declaration
var header = new RegExp(/\s+(PROGRAM|REPORT)\s+/, 'i')
const Header = createToken({
    name: "Header",
    pattern: header
})



const allTokens = [

]