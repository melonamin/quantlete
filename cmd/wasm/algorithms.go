//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	algorithms "github.com/melonamin/quantlete/algorithms/go"
)

// ============================================================================
// Algorithm Functions - Power
// ============================================================================

//wasm:category Algorithms - Power

// normalizedPowerFn calculates normalized power from power data
// Called from JS: goStorage.normalizedPower(watts)
//
//wasm:export normalizedPower
var normalizedPowerFn = wrapWasmRaw("normalizedPower", func(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing watts data"))
	}

	watts, err := jsArrayToFloat64(args[0])
	if err != nil {
		return errorJSON(fmt.Errorf("parsing watts: %w", err))
	}

	result := algorithms.NormalizedPower(watts)
	return toJSON(map[string]interface{}{
		"ok":    true,
		"value": result,
	})
})

// rollingMaxAverageFn finds the maximum rolling average over a given window
// Called from JS: goStorage.rollingMaxAverage(values, windowSeconds)
//
//wasm:export rollingMaxAverage
var rollingMaxAverageFn = wrapWasmRaw("rollingMaxAverage", func(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return errorJSON(fmt.Errorf("missing values or window"))
	}

	values, err := jsArrayToFloat64(args[0])
	if err != nil {
		return errorJSON(fmt.Errorf("parsing values: %w", err))
	}

	windowSeconds := args[1].Int()
	result := algorithms.RollingMaxAverage(values, windowSeconds)
	return toJSON(map[string]interface{}{
		"ok":    true,
		"value": result,
	})
})

// intensityFactorFn calculates the intensity factor: IF = NP / FTP
// Called from JS: goStorage.intensityFactor(np, ftp)
//
//wasm:export intensityFactor
var intensityFactorFn = wrapWasmRaw("intensityFactor", func(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return errorJSON(fmt.Errorf("missing np or ftp"))
	}

	np := args[0].Float()
	ftp := args[1].Float()
	result := algorithms.IntensityFactor(np, ftp)
	return toJSON(map[string]interface{}{
		"ok":    true,
		"value": result,
	})
})

// trainingStressScoreFn calculates TSS
// Called from JS: goStorage.trainingStressScore(durationSeconds, np, ftp)
//
//wasm:export trainingStressScore
var trainingStressScoreFn = wrapWasmRaw("trainingStressScore", func(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return errorJSON(fmt.Errorf("missing duration, np, or ftp"))
	}

	durationSeconds := args[0].Int()
	np := args[1].Float()
	ftp := args[2].Float()
	result := algorithms.TrainingStressScore(durationSeconds, np, ftp)
	return toJSON(map[string]interface{}{
		"ok":    true,
		"value": result,
	})
})

// ============================================================================
// Algorithm Functions - Eddington
// ============================================================================

//wasm:category Algorithms - Eddington

// eddingtonNumberFn calculates the Eddington number from daily distances
// Called from JS: goStorage.eddingtonNumber(distances)
//
//wasm:export eddingtonNumber
var eddingtonNumberFn = wrapWasmRaw("eddingtonNumber", func(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing distances data"))
	}

	distances, err := jsArrayToFloat64(args[0])
	if err != nil {
		return errorJSON(fmt.Errorf("parsing distances: %w", err))
	}

	result := algorithms.EddingtonNumber(distances)
	return toJSON(map[string]interface{}{
		"ok":    true,
		"value": result,
	})
})

// eddingtonNextStepsFn calculates days needed for next Eddington numbers
// Called from JS: goStorage.eddingtonNextSteps(distances, currentE, stepsToCalculate)
//
//wasm:export eddingtonNextSteps
var eddingtonNextStepsFn = wrapWasmRaw("eddingtonNextSteps", func(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return errorJSON(fmt.Errorf("missing distances, currentE, or stepsToCalculate"))
	}

	distances, err := jsArrayToFloat64(args[0])
	if err != nil {
		return errorJSON(fmt.Errorf("parsing distances: %w", err))
	}

	currentE := args[1].Int()
	stepsToCalculate := args[2].Int()
	result := algorithms.EddingtonNextSteps(distances, currentE, stepsToCalculate)

	// Convert to response format
	steps := make([]map[string]interface{}, len(result))
	for i, s := range result {
		steps[i] = map[string]interface{}{
			"target":      s.Target,
			"days_needed": s.DaysNeeded,
		}
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": steps,
	})
})

// eddingtonHistoryFn calculates progressive Eddington numbers over time
// Called from JS: goStorage.eddingtonHistory(distances)
//
//wasm:export eddingtonHistory
var eddingtonHistoryFn = wrapWasmRaw("eddingtonHistory", func(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing distances data"))
	}

	distances, err := jsArrayToFloat64(args[0])
	if err != nil {
		return errorJSON(fmt.Errorf("parsing distances: %w", err))
	}

	result := algorithms.EddingtonHistory(distances)
	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": result,
	})
})

// ============================================================================
// Algorithm Functions - Training Load
// ============================================================================

//wasm:category Algorithms - Training Load

// calculateTrainingLoadFn calculates training load metrics from daily TSS
// Called from JS: goStorage.calculateTrainingLoad(dailyTss, ctlTau, atlTau)
//
//wasm:export calculateTrainingLoad
var calculateTrainingLoadFn = wrapWasmRaw("calculateTrainingLoad", func(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return errorJSON(fmt.Errorf("missing dailyTss, ctlTau, or atlTau"))
	}

	dailyTss, err := jsArrayToFloat64(args[0])
	if err != nil {
		return errorJSON(fmt.Errorf("parsing dailyTss: %w", err))
	}

	ctlTau := args[1].Float()
	atlTau := args[2].Float()
	result := algorithms.CalculateTrainingLoad(dailyTss, ctlTau, atlTau)

	// Convert to response format
	points := make([]map[string]interface{}, len(result))
	for i, p := range result {
		points[i] = map[string]interface{}{
			"ctl": p.CTL,
			"atl": p.ATL,
			"tsb": p.TSB,
		}
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": points,
	})
})

// calculateTrainingLoadWithInitialFn calculates training load from existing CTL/ATL
// Called from JS: goStorage.calculateTrainingLoadWithInitial(dailyTss, initialCtl, initialAtl, ctlTau, atlTau)
//
//wasm:export calculateTrainingLoadWithInitial
var calculateTrainingLoadWithInitialFn = wrapWasmRaw("calculateTrainingLoadWithInitial", func(this js.Value, args []js.Value) interface{} {
	if len(args) < 5 {
		return errorJSON(fmt.Errorf("missing dailyTss, initialCtl, initialAtl, ctlTau, or atlTau"))
	}

	dailyTss, err := jsArrayToFloat64(args[0])
	if err != nil {
		return errorJSON(fmt.Errorf("parsing dailyTss: %w", err))
	}

	initialCtl := args[1].Float()
	initialAtl := args[2].Float()
	ctlTau := args[3].Float()
	atlTau := args[4].Float()
	result := algorithms.CalculateTrainingLoadWithInitial(dailyTss, initialCtl, initialAtl, ctlTau, atlTau)

	// Convert to response format
	points := make([]map[string]interface{}, len(result))
	for i, p := range result {
		points[i] = map[string]interface{}{
			"ctl": p.CTL,
			"atl": p.ATL,
			"tsb": p.TSB,
		}
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": points,
	})
})

// predictAfterWorkoutFn calculates predicted TSB after a planned workout
// Called from JS: goStorage.predictAfterWorkout(currentCtl, currentAtl, plannedTss, ctlTau, atlTau)
//
//wasm:export predictAfterWorkout
var predictAfterWorkoutFn = wrapWasmRaw("predictAfterWorkout", func(this js.Value, args []js.Value) interface{} {
	if len(args) < 5 {
		return errorJSON(fmt.Errorf("missing currentCtl, currentAtl, plannedTss, ctlTau, or atlTau"))
	}

	currentCtl := args[0].Float()
	currentAtl := args[1].Float()
	plannedTss := args[2].Float()
	ctlTau := args[3].Float()
	atlTau := args[4].Float()
	result := algorithms.PredictAfterWorkout(currentCtl, currentAtl, plannedTss, ctlTau, atlTau)

	return toJSON(map[string]interface{}{
		"ok":  true,
		"ctl": result.CTL,
		"atl": result.ATL,
		"tsb": result.TSB,
	})
})

// tssForTargetTsbFn calculates TSS needed to reach a target TSB
// Called from JS: goStorage.tssForTargetTsb(currentCtl, currentAtl, targetTsb, ctlTau, atlTau)
//
//wasm:export tssForTargetTsb
var tssForTargetTsbFn = wrapWasmRaw("tssForTargetTsb", func(this js.Value, args []js.Value) interface{} {
	if len(args) < 5 {
		return errorJSON(fmt.Errorf("missing currentCtl, currentAtl, targetTsb, ctlTau, or atlTau"))
	}

	currentCtl := args[0].Float()
	currentAtl := args[1].Float()
	targetTsb := args[2].Float()
	ctlTau := args[3].Float()
	atlTau := args[4].Float()
	result := algorithms.TssForTargetTsb(currentCtl, currentAtl, targetTsb, ctlTau, atlTau)

	return toJSON(map[string]interface{}{
		"ok":    true,
		"value": result,
	})
})
