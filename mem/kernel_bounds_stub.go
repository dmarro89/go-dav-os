//go:build !gccgo

package mem

func bootstrapEnd() uint64 { return 0 }
func kernelEnd() uint64    { return 0 }
