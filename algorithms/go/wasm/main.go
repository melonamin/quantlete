//go:build tinygo

// Package main provides WASM exports for the algorithms package.
// This file is compiled with TinyGo to produce a WASM module.
package main

import (
	"unsafe"

	algorithms "github.com/melonamin/quantlete/algorithms/go"
)

// Buffer size for input/output arrays (64K elements should be plenty)
const maxArraySize = 65536

// Fixed buffers for WASM memory interface
var (
	inputBuffer     [maxArraySize]float64
	outputBuffer    [maxArraySize]float64
	intOutputBuffer [maxArraySize]int32
)

// Buffer pointer exports - called by JavaScript to get buffer addresses

//go:export getInputBufferPtr
func getInputBufferPtr() uintptr {
	return uintptr(unsafe.Pointer(&inputBuffer[0]))
}

//go:export getOutputBufferPtr
func getOutputBufferPtr() uintptr {
	return uintptr(unsafe.Pointer(&outputBuffer[0]))
}

//go:export getIntOutputBufferPtr
func getIntOutputBufferPtr() uintptr {
	return uintptr(unsafe.Pointer(&intOutputBuffer[0]))
}

// Power Analysis Exports

//go:export normalizedPower
func normalizedPower(length int32) float64 {
	if length <= 0 || length > maxArraySize {
		return 0
	}
	data := inputBuffer[:length]
	return algorithms.NormalizedPower(data)
}

//go:export rollingMaxAverage
func rollingMaxAverage(length int32, windowSeconds int32) float64 {
	if length <= 0 || length > maxArraySize {
		return 0
	}
	data := inputBuffer[:length]
	return algorithms.RollingMaxAverage(data, int(windowSeconds))
}

//go:export intensityFactor
func intensityFactor(np, ftp float64) float64 {
	return algorithms.IntensityFactor(np, ftp)
}

//go:export trainingStressScore
func trainingStressScore(durationSeconds int32, np, ftp float64) float64 {
	return algorithms.TrainingStressScore(int(durationSeconds), np, ftp)
}

// Eddington Number Exports

//go:export eddingtonNumber
func eddingtonNumber(length int32) int32 {
	if length <= 0 || length > maxArraySize {
		return 0
	}
	data := inputBuffer[:length]
	return int32(algorithms.EddingtonNumber(data))
}

//go:export eddingtonNextSteps
func eddingtonNextSteps(length int32, currentE int32, stepsToCalculate int32) int32 {
	if length <= 0 || length > maxArraySize || stepsToCalculate <= 0 {
		return 0
	}
	data := inputBuffer[:length]
	result := algorithms.EddingtonNextSteps(data, int(currentE), int(stepsToCalculate))

	// Write to intOutputBuffer: [target, daysNeeded, target, daysNeeded, ...]
	for i, step := range result {
		if i*2+1 >= maxArraySize {
			break
		}
		intOutputBuffer[i*2] = int32(step.Target)
		intOutputBuffer[i*2+1] = int32(step.DaysNeeded)
	}

	return int32(len(result))
}

//go:export eddingtonHistory
func eddingtonHistory(length int32) int32 {
	if length <= 0 || length > maxArraySize {
		return 0
	}
	data := inputBuffer[:length]
	result := algorithms.EddingtonHistory(data)

	// Write to intOutputBuffer
	for i, v := range result {
		if i >= maxArraySize {
			break
		}
		intOutputBuffer[i] = int32(v)
	}

	return int32(len(result))
}

// Training Load Exports

//go:export calculateTrainingLoad
func calculateTrainingLoad(length int32, ctlTau, atlTau float64) int32 {
	if length <= 0 || length > maxArraySize {
		return 0
	}
	data := inputBuffer[:length]
	result := algorithms.CalculateTrainingLoad(data, ctlTau, atlTau)

	// Write to outputBuffer: [CTL, ATL, TSB, CTL, ATL, TSB, ...]
	for i, p := range result {
		if i*3+2 >= maxArraySize {
			break
		}
		outputBuffer[i*3] = p.CTL
		outputBuffer[i*3+1] = p.ATL
		outputBuffer[i*3+2] = p.TSB
	}

	return int32(len(result))
}

//go:export calculateTrainingLoadWithInitial
func calculateTrainingLoadWithInitial(length int32, initialCtl, initialAtl, ctlTau, atlTau float64) int32 {
	if length <= 0 || length > maxArraySize {
		return 0
	}
	data := inputBuffer[:length]
	result := algorithms.CalculateTrainingLoadWithInitial(data, initialCtl, initialAtl, ctlTau, atlTau)

	// Write to outputBuffer
	for i, p := range result {
		if i*3+2 >= maxArraySize {
			break
		}
		outputBuffer[i*3] = p.CTL
		outputBuffer[i*3+1] = p.ATL
		outputBuffer[i*3+2] = p.TSB
	}

	return int32(len(result))
}

//go:export predictAfterWorkout
func predictAfterWorkout(currentCtl, currentAtl, plannedTss, ctlTau, atlTau float64) {
	result := algorithms.PredictAfterWorkout(currentCtl, currentAtl, plannedTss, ctlTau, atlTau)
	outputBuffer[0] = result.CTL
	outputBuffer[1] = result.ATL
	outputBuffer[2] = result.TSB
}

//go:export tssForTargetTsb
func tssForTargetTsb(currentCtl, currentAtl, targetTsb, ctlTau, atlTau float64) float64 {
	return algorithms.TssForTargetTsb(currentCtl, currentAtl, targetTsb, ctlTau, atlTau)
}

// main is required for TinyGo WASM but does nothing
func main() {}
