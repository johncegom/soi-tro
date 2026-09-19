package main

import "os"

// Throwaway: deliberate unchecked error to confirm the Lint job fails (task 017, 2.5).
var _ = func() { os.Remove("canary") }
