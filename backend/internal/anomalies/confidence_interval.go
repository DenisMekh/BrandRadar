package anomalies

import "math"

const lambda = 1.96

type Interval struct {
	Lower float64
	Upper float64
}

type PointResult struct {
	Index    int
	Value    float64
	Pred     float64
	Interval Interval
	IsAnom   bool
	Z        float64
}

func stddev(array []float64, from, to int) float64 {
	if from < 0 || to > len(array) || from >= to {
		return math.NaN()
	}
	n := to - from
	if n == 0 {
		return math.NaN()
	}

	var sum float64
	for i := from; i < to; i++ {
		sum += array[i]
	}
	mean := sum / float64(n)

	var ss float64
	for i := from; i < to; i++ {
		d := array[i] - mean
		ss += d * d
	}
	return math.Sqrt(ss / float64(n))
}

func mean(array []float64, from, to int) float64 {
	if from < 0 || to > len(array) || from >= to {
		return math.NaN()
	}
	var sum float64
	for i := from; i < to; i++ {
		sum += array[i]
	}
	return sum / float64(to-from)
}

func Analyze(array []float64, w int) []PointResult {
	res := make([]PointResult, 0, len(array))

	for i := 0; i < len(array); i++ {
		r := PointResult{
			Index: i,
			Value: array[i],
			Pred:  math.NaN(),
			Interval: Interval{
				Lower: math.NaN(),
				Upper: math.NaN(),
			},
			IsAnom: false,
			Z:      math.NaN(),
		}

		if i < w {
			res = append(res, r)
			continue
		}

		mu := mean(array, i-w, i)
		sigma := stddev(array, i-w, i)

		if sigma == 0 || math.IsNaN(sigma) {

			r.Pred = mu
			r.Interval = Interval{Lower: mu, Upper: mu}
			r.IsAnom = array[i] != mu
			if array[i] == mu {
				r.Z = 0
			} else {
				r.Z = math.Inf(1)
			}
			res = append(res, r)
			continue
		}

		lower := mu - lambda*sigma
		upper := mu + lambda*sigma

		r.Pred = mu
		r.Interval = Interval{Lower: lower, Upper: upper}
		r.Z = (array[i] - mu) / sigma
		r.IsAnom = array[i] < lower || array[i] > upper

		res = append(res, r)
	}

	return res
}
