package eval

// Env is a chained variable environment. The global Env persists across REPL
// submissions so variables survive.
type Env struct {
	vars   map[string]Value
	parent *Env
}

// NewEnv returns a new scope nested in parent (nil for the global scope).
func NewEnv(parent *Env) *Env {
	return &Env{vars: map[string]Value{}, parent: parent}
}

// Get looks a name up through the scope chain.
func (e *Env) Get(name string) (Value, bool) {
	for s := e; s != nil; s = s.parent {
		if v, ok := s.vars[name]; ok {
			return v, true
		}
	}
	return nil, false
}

// Set binds name in this scope.
func (e *Env) Set(name string, v Value) {
	e.vars[name] = v
}
