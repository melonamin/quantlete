// Package analysis provides power analysis functions.
// This package re-exports functions from the shared algorithms package
// to maintain backwards compatibility with existing code.
package analysis

import algorithms "github.com/melonamin/quantlete/algorithms/go"

// RollingMaxAverage finds the maximum rolling average over a given window.
func RollingMaxAverage(values []float64, windowSeconds int) float64 {
	return algorithms.RollingMaxAverage(values, windowSeconds)
}

// NormalizedPower calculates normalized power from power data.
func NormalizedPower(watts []float64) float64 {
	return algorithms.NormalizedPower(watts)
}

// IntensityFactor calculates the intensity factor: IF = NP / FTP
func IntensityFactor(normalizedPower, ftp float64) float64 {
	return algorithms.IntensityFactor(normalizedPower, ftp)
}

// TrainingStressScore calculates TSS.
func TrainingStressScore(durationSeconds int, normalizedPower, ftp float64) float64 {
	return algorithms.TrainingStressScore(durationSeconds, normalizedPower, ftp)
}
