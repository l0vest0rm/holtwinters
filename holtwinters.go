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


// Package holtwinters import http://www.itl.nist.gov/div898/handbook/pmc/section4/pmc435.htm
// st[i] = alpha * y[i] / it[i - period] + (1.0 - alpha) * (st[i - 1] + bt[i - 1])
// bt[i] = gamma * (st[i] - st[i - 1]) + (1 - gamma) * bt[i - 1]
// it[i] = beta * y[i] / st[i] + (1.0 - beta) * it[i - period]
// ft[i + m] = (st[i] + (m * bt[i])) * it[i - period + m]

type TripleExponentialSmoothing struct {
    alpha float64
    beta float64
    gamma float64
    y []float64
    ylen int
    bufSize int
    st []float64
    bt []float64
    it []float64
    l int
}

func NewTripleExponentialSmoothing() *TripleExponentialSmoothing {
    return &TripleExponentialSmoothing{}
}

func (t *TripleExponentialSmoothing) Train(y []float64, alpha, beta, gamma float64, l int){
    t.y = y
    t.ylen = len(y)
    t.bufSize = 2*l
    t.alpha = alpha
    t.beta = beta
    t.gamma = gamma
    t.st = make([]float64, t.bufSize)
    t.bt = make([]float64, t.bufSize)
    t.it = make([]float64, t.bufSize)
    t.l = l

    t.st[1] = t.y[0]
    t.bt[1] = initialTrend(t.y, l)
    t.initialSeasonalIndicies(t.y, l)

    for i := 2; i < t.ylen; i++ {
        // overall smoothing
        if (i - l) >= 0 {
            t.st[i%t.bufSize] = alpha*y[i]/t.it[(i-l)%t.bufSize] + (1.0-alpha)*(t.st[(i-1)%t.bufSize]+t.bt[(i-1)%t.bufSize])
        } else {
            t.st[i] = alpha*y[i] + (1.0-alpha)*(t.st[(i-1)%t.bufSize]+t.bt[(i-1)%t.bufSize])
        }

        // trend smoothing
        t.bt[i%t.bufSize] = gamma*(t.st[i%t.bufSize]-t.st[(i-1)%t.bufSize]) + (1-gamma)*t.bt[(i-1)%t.bufSize]

        // seasonal smoothing
        if (i - l) >= 0 {
            t.it[i%t.bufSize] = beta*y[i]/t.st[i%t.bufSize] + (1.0-beta)*t.it[(i-l)%t.bufSize]
        }
    }
}

func (t *TripleExponentialSmoothing) Forecast(m int) []float64 {
    ft := make([]float64, m)
    for k := 0; k < m; k++ {
        i := t.ylen + k - m
        ft[k] = (t.st[i%t.bufSize] + float64(m) * t.bt[i%t.bufSize]) * t.it[(i-t.l+m)%t.bufSize]
    }

    return ft
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
func (t *TripleExponentialSmoothing)initialSeasonalIndicies(y []float64, l int) {
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
            t.it[i] += averagedObservations[(j*l)+i]
        }
        t.it[i] /= float64(seasons)
    }
}
