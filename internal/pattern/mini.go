package pattern

// Mini-notation compiler (milestone M3): turns the parser's pattern AST into a
// tree of combinators. Splices ($name, (expr)) are evaluated eagerly before
// compilation, so querying a pattern never runs user code.
