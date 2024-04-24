package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {

	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	// inputData := []int{0,1}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
	}

	start := time.Now()

	ExecutePipeline(hashSignJobs...)

	fmt.Println("Время работы пайплайна:", time.Since(start))

}

// сюда писать код
func SingleHash(in, out chan interface{}) {
	wgG := &sync.WaitGroup{}
	for data := range in {
		dataRaw := fmt.Sprint(data)
		md5Str := DataSignerMd5(dataRaw)
		wgG.Add(1)
		go func(md5Str string, dataRaw string) {
			defer wgG.Done()
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				md5Str = DataSignerCrc32(md5Str)
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				dataRaw = DataSignerCrc32(dataRaw)
			}()
			wg.Wait()
			result := dataRaw + "~" + md5Str
			fmt.Printf("%s SingleHash result: %s\n", dataRaw, result)
			out <- result
		}(md5Str, dataRaw)
	}
	wgG.Wait()
}

func MultiHash(in, out chan interface{}) {
	wgG := &sync.WaitGroup{}
	for data := range in {
		dataRaw := fmt.Sprint(data)
		wgG.Add(1)
		go func(out chan interface{}, dataRaw string) {
			defer wgG.Done()
			wg := &sync.WaitGroup{}
			arr := make([]string, 7)
			for i := 0; i < 6; i++ {
				wg.Add(1)
				go func(val int, dataRaw string) {
					defer wg.Done()
					arr[val] = DataSignerCrc32(strconv.Itoa(val) + dataRaw)
				}(i, dataRaw)
			}
			wg.Wait()
			result := strings.Join(arr, "")
			fmt.Printf("%s MultiHash result: %s\n", dataRaw, result)
			out <- result
		}(out, dataRaw)
	}
	wgG.Wait()
}

func CombineResults(in, out chan interface{}) {
	tmp := len(in)
	arr := make([]string, tmp)
	for i := range in {
		arr = append(arr, fmt.Sprint(i))
	}
	fmt.Println(arr)
	sort.Strings(arr)
	fmt.Println("Sort arr:", arr)
	result := strings.Join(arr[tmp:], "_")
	fmt.Printf("CombineResults result: %s\n", result)
	out <- result
}

func ExecutePipeline(works ...job) {
	wg := &sync.WaitGroup{}
	in, out := make(chan interface{}, MaxInputDataLen), make(chan interface{}, MaxInputDataLen)

	for _, work := range works {
		wg.Add(1)
		go func(job job, in, out chan interface{}) {
			defer close(out)
			defer wg.Done()
			job(in, out)
		}(work, in, out)
		in = out
		out = make(chan interface{}, MaxInputDataLen)
	}
	wg.Wait()
}
