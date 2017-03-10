/**
 * Copyright 2017  authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License"): you may
 * not use this file except in compliance with the License. You may obtain
 * a copy of the License at
 *
 *     http: *www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 */

// Created by xuning on 2017/3/1

package holtwinters

import (
	"log"
	"math"
	"sync"
)

// Package holtwinters import http://www.itl.nist.gov/div898/handbook/pmc/section4/pmc435.htm
// st[i] = alpha * y[i] / it[i - period] + (1.0 - alpha) * (st[i - 1] + bt[i - 1])
// bt[i] = gamma * (st[i] - st[i - 1]) + (1 - gamma) * bt[i - 1]
// it[i] = beta * y[i] / st[i] + (1.0 - beta) * it[i - period]
// ft[i + m] = (st[i] + (m * bt[i])) * it[i - period + m]

type TripleExponentialSmoothing struct {
	mu      sync.RWMutex
	mse     float64
	alpha   float64
	beta    float64
	gamma   float64
	ylen    int
	bufSize int
	st      []float64
	bt      []float64
	it      []float64
	ft      []float64
	l       int
}

func NewTripleExponentialSmoothing(l int) *TripleExponentialSmoothing {
	t := &TripleExponentialSmoothing{}
	t.l = l
	t.bufSize = 2 * l
	t.st = make([]float64, t.bufSize)
	t.bt = make([]float64, t.bufSize)
	t.it = make([]float64, t.bufSize)

	return t
}

func (t *TripleExponentialSmoothing) Train(y []float64, alpha, beta, gamma float64) {
	t.ylen = len(y)
	t.ft = make([]float64, t.ylen+t.l)

	t.st[1] = y[0]
	t.bt[1] = initialTrend(y, t.l)
	initialSeasonalIndicies(y, t.l, t.it)

	for i := 2; i < t.ylen; i++ {
		// overall smoothing
		if (i - t.l) >= 0 {
			t.st[i%t.bufSize] = alpha*y[i]/t.it[(i-t.l)%t.bufSize] + (1.0-alpha)*(t.st[(i-1)%t.bufSize]+t.bt[(i-1)%t.bufSize])
		} else {
			t.st[i] = alpha*y[i] + (1.0-alpha)*(t.st[(i-1)%t.bufSize]+t.bt[(i-1)%t.bufSize])
		}

		// trend smoothing
		t.bt[i%t.bufSize] = gamma*(t.st[i%t.bufSize]-t.st[(i-1)%t.bufSize]) + (1-gamma)*t.bt[(i-1)%t.bufSize]

		// seasonal smoothing
		if (i - t.l) >= 0 {
			t.it[i%t.bufSize] = beta*y[i]/t.st[i%t.bufSize] + (1.0-beta)*t.it[(i-t.l)%t.bufSize]
		}

		// forecast
		t.ft[i+t.l] = (t.st[i%t.bufSize] + (float64(t.l) * t.bt[i%t.bufSize])) * t.it[i%t.bufSize]
	}
}

func (t *TripleExponentialSmoothing) Forecast(m int) []float64 {
	ft := make([]float64, m)
	for k := 0; k < m; k++ {
		i := t.ylen + k - m
		ft[k] = (t.st[i%t.bufSize] + (float64(m) * t.bt[i%t.bufSize])) * t.it[(i-t.l+m)%t.bufSize]
	}

	return ft
}

func try(y []float64, alpha, beta, gamma float64, l int) float64 {
	ylen := len(y)
	bufSize := 2 * l
	st := make([]float64, bufSize)
	bt := make([]float64, bufSize)
	it := make([]float64, bufSize)
	ft := make([]float64, ylen+l)

	st[1] = y[0]
	bt[1] = initialTrend(y, l)
	initialSeasonalIndicies(y, l, it)

	for i := 2; i < ylen; i++ {
		// overall smoothing
		if (i - l) >= 0 {
			st[i%bufSize] = alpha*y[i]/it[(i-l)%bufSize] + (1.0-alpha)*(st[(i-1)%bufSize]+bt[(i-1)%bufSize])
		} else {
			st[i] = alpha*y[i] + (1.0-alpha)*(st[(i-1)%bufSize]+bt[(i-1)%bufSize])
		}

		// trend smoothing
		bt[i%bufSize] = gamma*(st[i%bufSize]-st[(i-1)%bufSize]) + (1-gamma)*bt[(i-1)%bufSize]

		// seasonal smoothing
		if (i - l) >= 0 {
			it[i%bufSize] = beta*y[i]/st[i%bufSize] + (1.0-beta)*it[(i-l)%bufSize]
		}

		// forecast
		if (i - l) >= 0 {
			ft[i+l] = (st[i%bufSize] + float64(l)*bt[i%bufSize]) * it[i%bufSize]
		}
	}

	return mse(y, ft, l)
}

func (t *TripleExponentialSmoothing) goTryBest(wg *sync.WaitGroup, y []float64, alpha, beta, gamma float64, l int) {
	defer wg.Done()

	mse := try(y, alpha, beta, gamma, l)
	//log.Printf("goTryBest,alpha:%f,beta:%f,gamma:%f,mse:%f\n", alpha, beta, gamma, mse)
	t.mu.Lock()
	if mse < t.mse {
		t.mse = mse
		t.alpha = alpha
		t.beta = beta
		t.gamma = gamma
	}
	t.mu.Unlock()
	//log.Printf("goTryBest,alpha:%f,beta:%f,gamma:%f,mse:%f\n", alpha, beta, gamma, mse)
}

//auto choose alpha, beta, gamma with tolerance
func (t *TripleExponentialSmoothing) Fit(y []float64, tolerance float64) {
	alphaMin := float64(0)
	alphaMax := float64(1)
	betaMin := float64(0)
	betaMax := float64(1)
	gammaMin := float64(0)
	gammaMax := float64(1)
	step := 0.1

	for {
		t.FindBest(y, alphaMin, alphaMax, betaMin, betaMax, gammaMin, gammaMax, step)
		if step <= tolerance {
			break
		}

		alphaMin = correctParam(t.alpha - step)
		alphaMax = correctParam(t.alpha + step)
		betaMin = correctParam(t.beta - step)
		betaMax = correctParam(t.beta + step)
		gammaMin = correctParam(t.gamma - step)
		gammaMax = correctParam(t.gamma + step)
		step /= float64(10)
	}

	log.Printf("end fit with alpha:%f,beta:%f,gamma:%f,mse:%f\n", t.alpha, t.beta, t.gamma, t.mse)
}

func (t *TripleExponentialSmoothing) FindBest(y []float64, alphaMin, alphaMax, betaMin, betaMax, gammaMin, gammaMax, step float64) {
	var wg sync.WaitGroup

	t.mse = math.MaxFloat64
	for alpha := alphaMin; alpha <= alphaMax; alpha += step {
		for beta := betaMin; beta <= betaMax; beta += step {
			for gamma := gammaMin; gamma <= gammaMax; gamma += step {
				wg.Add(1)
				go t.goTryBest(&wg, y, alpha, beta, gamma, t.l)
			}
		}
	}

	wg.Wait()
	//log.Printf("FindBest,alpha:%f,beta:%f,gamma:%f,step:%f, mse:%f\n", t.alpha, t.beta, t.gamma, step, t.mse)
}

// See: http://www.itl.nist.gov/div898/handbook/pmc/section4/pmc435.htm
func initialTrend(y []float64, l int) float64 {
	sum := float64(0)

	for i := 0; i < l; i++ {
		sum += (y[l+i] - y[i])
	}

	return sum / float64(l*l)
}

// See: http://www.itl.nist.gov/div898/handbook/pmc/section4/pmc435.htm
func initialSeasonalIndicies(y []float64, l int, it []float64) {
	seasons := len(y) / l
	seasonalAverage := make([]float64, seasons)

	averagedObservations := make([]float64, len(y))

	for i := 0; i < seasons; i++ {
		for j := 0; j < l; j++ {
			seasonalAverage[i] += y[(i*l)+j]
		}
		seasonalAverage[i] /= float64(l)
	}

	for i := 0; i < seasons; i++ {
		for j := 0; j < l; j++ {
			averagedObservations[(i*l)+j] = y[(i*l)+j] / seasonalAverage[i]
		}
	}

	for i := 0; i < l; i++ {
		for j := 0; j < seasons; j++ {
			it[i] += averagedObservations[(j*l)+i]
		}
		it[i] /= float64(seasons)
	}
}

func mse(y []float64, ft []float64, l int) float64 {
	sse := float64(0)
	for i := l + 2; i < len(y); i++ {
		sse += (ft[i] - y[i]) * (ft[i] - y[i])
	}

	mse := sse / float64(len(y)-l-2)
	//log.Printf("mse:%f\n", mse)
	return mse
}

func correctParam(param float64) float64 {
	if param < 0 {
		return 0
	} else if param > 1 {
		return 1
	} else {
		return param
	}
}
