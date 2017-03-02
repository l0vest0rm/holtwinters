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
    "os"
    "bufio"
    "strings"
    "strconv"
    "io/ioutil"
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

    this := NewTripleExponentialSmoothing(l)
    this.Train(y, alpha, beta, gamma)
    prediction := this.Forecast(m)
    fmt.Printf("%v\n",prediction)
    expected := [...]float64{599.1184450128665, 733.227872348479, 949.0708357438998,
        748.6618488792186}
    if Compare(expected[:], prediction) != 0 {
        t.Fatal("failed")
    }
}

func TestFit(t *testing.T) {
    y := []float64{362, 385, 432, 341, 382, 409, 498, 387, 473, 513,
        582, 474, 544, 582, 681, 557, 628, 707, 773, 592, 627, 725,
        854, 661}
    l := 4

    tolerance := float64(0.001)

    this := NewTripleExponentialSmoothing(l)
    this.Fit(y, tolerance)
}

func TestRealFit(t *testing.T) {
    y := ReadY("time_series1.csv")
    fmt.Printf("ylen:%d\n", len(y))
    l := 1440*7
    m := l
    tolerance := float64(0.00001)

    this := NewTripleExponentialSmoothing(l)
    this.Fit(y[:l*2], tolerance)

    this.Train(y, this.alpha, this.beta, this.gamma)
    prediction := this.Forecast(m)

    render("index.tpl", "index.html", y[l*2:], prediction)
}

func render(tplFile string, htmlFile string, real []float64, estimate []float64) {
    b, err := ioutil.ReadFile(tplFile)
    if err != nil {
        panic(fmt.Sprintf("ioutil.ReadFile,err:%s\n", err.Error()))
    }

    realStr := strings.Replace(fmt.Sprintf("%v", real), " ", ",", -1)
    estimateStr := strings.Replace(fmt.Sprintf("%v", estimate), " ", ",", -1)

    html := string(b)
    html = strings.Replace(html, "[real]", realStr, -1)
    html = strings.Replace(html, "[estimate]", estimateStr, -1)
    ioutil.WriteFile(htmlFile, []byte(html), os.ModePerm)
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

func ReadY(fileName string) []float64 {
    f, err := os.Open(fileName)
    if err != nil {
        panic(fmt.Sprintf("fail to open %s\n", fileName))
    }

    y := make([]float64, 0)
    buf := bufio.NewReader(f)
    //skip header
    _,_ = buf.ReadString('\n')
    for {
        line, err := buf.ReadString('\n')
        if err == nil {
            line = strings.TrimSpace(line)
            att := strings.Split(line, ",")
            data, err := strconv.ParseFloat(att[1], 64)
            if err != nil {
                panic(fmt.Sprintf("err line:%s\n", line))
            }
            y = append(y, data)
        }else {
            fmt.Printf("err:%s\n",err.Error())
            break
        }
    }
    return y
}