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
    "testing"
    "fmt"
)

func TestForecast(t *testing.T) {
    y := []float64{362, 385, 432, 341, 382, 409, 498, 387, 473, 513,
        582, 474, 544, 582, 681, 557, 628, 707, 773, 592, 627, 725,
        854, 661}
    l := 4
    m := 4

    alpha := 0.5
    beta := 0.4
    gamma := 0.6

    this := NewTripleExponentialSmoothing()
    this.Train(y, alpha, beta, gamma, l)
    prediction := this.Forecast(m)
    fmt.Printf("%v\n",prediction)
    expected := [...]float64{599.1184450128665, 733.227872348479, 949.0708357438998,
        748.6618488792186}
    if Compare(expected[:], prediction) != 0 {
        t.Fatal("failed")
    }
}

func Compare(a, b []float64) int {
    for i := 0; i < len(a) && i < len(b); i++ {
        switch {
        case a[i] > b[i]:
            return 1
        case a[i] < b[i]:
            return -1
        }
    }
    switch {
    case len(a) < len(b):
        return -1
    case len(a) > len(b):
        return 1
    }
    return 0
}