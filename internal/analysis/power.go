package analysis

import "math"

func RollingMaxAverage(values []float64, windowSeconds int) float64 {
	if windowSeconds <= 0 || len(values) < windowSeconds {
		return 0
	}

	prefix := make([]float64, len(values)+1)
	for i, v := range values {
		if v < 0 {
			v = 0
		}
		prefix[i+1] = prefix[i] + v
	}

	best := 0.0
	w := float64(windowSeconds)
	for end := windowSeconds; end <= len(values); end++ {
		sum := prefix[end] - prefix[end-windowSeconds]
		avg := sum / w
		if avg > best {
			best = avg
		}
	}
	return best
}

func NormalizedPower(watts []float64) float64 {
	if len(watts) == 0 {
		return 0
	}

	// Standard cycling NP uses a 30-second rolling average.
	const window = 30
	if len(watts) < window {
		sum := 0.0
		for _, w := range watts {
			if w < 0 {
				w = 0
			}
			sum += w
		}
		return sum / float64(len(watts))
	}

	prefix := make([]float64, len(watts)+1)
	for i, w := range watts {
		if w < 0 {
			w = 0
		}
		prefix[i+1] = prefix[i] + w
	}

	sumFourth := 0.0
	count := 0
	for end := window; end <= len(watts); end++ {
		avg := (prefix[end] - prefix[end-window]) / float64(window)
		sumFourth += math.Pow(avg, 4)
		count++
	}
	if count == 0 {
		return 0
	}
	return math.Pow(sumFourth/float64(count), 0.25)
}

func IntensityFactor(normalizedPower float64, ftp float64) float64 {
	if ftp <= 0 {
		return 0
	}
	return normalizedPower / ftp
}

func TrainingStressScore(durationSeconds int, normalizedPower float64, ftp float64) float64 {
	if durationSeconds <= 0 || ftp <= 0 {
		return 0
	}
	ifactor := IntensityFactor(normalizedPower, ftp)
	// TSS = (sec * NP * IF) / (FTP * 3600) * 100
	return (float64(durationSeconds) * normalizedPower * ifactor) / (ftp * 3600.0) * 100.0
}
