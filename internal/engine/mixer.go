package engine

// Mixer (milestone M6): per block, process channels (voices, insert FX, gain,
// pan), then sends into buses, then buses with their FX, then master.
// Gain, pan and param changes are smoothed over ~10 ms.
