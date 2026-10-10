package parser

// Binding powers, lowest first. They follow docs/grammar.ebnf: pipe is the
// loosest, then compose, sum, product, prefix minus, then application. Field
// access and atoms bind tighter still and are handled by parseAccess/parseAtom.
const (
	bpPipe    = 1 // |>
	bpCompose = 2 // >>
	bpAdd     = 3 // + -
	bpMul     = 4 // * /
	bpUnary   = 5 // prefix -
	bpApply   = 6 // juxtaposition
)
