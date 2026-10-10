package eval

// Built-ins (milestones M2-M6):
//
//   pattern functions  fast slow rev every stack cat shift degrade euclid
//   constructors       channel bus
//   controls           bpm hush solo mute unmute undo show
//   units              generated from the engine/units registry (sine, saw, lpf, ...)
//
// Every registered unit automatically becomes a curried function here.
